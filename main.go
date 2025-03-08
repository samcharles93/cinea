package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/samcharles93/cinea/config"
	"github.com/samcharles93/cinea/internal/auth"
	"github.com/samcharles93/cinea/internal/ffmpeg"
	"github.com/samcharles93/cinea/internal/handler"
	"github.com/samcharles93/cinea/internal/logger"
	"github.com/samcharles93/cinea/internal/repository"
	"github.com/samcharles93/cinea/internal/router"
	"github.com/samcharles93/cinea/internal/service"
	"github.com/samcharles93/cinea/internal/service/cleanup"
	"github.com/samcharles93/cinea/internal/service/extractor"
	"github.com/samcharles93/cinea/internal/service/metadata"
	"github.com/samcharles93/cinea/internal/service/scanner"
	"github.com/samcharles93/cinea/internal/service/scheduler"
	"github.com/samcharles93/cinea/web"
	"gorm.io/gorm"
)

//go:embed web/templates/* web/static/*
var webFS embed.FS

type app struct {
	// Configuration
	config *config.Config

	// Core infrastructure
	db        *gorm.DB
	appLogger logger.Logger
	tokenAuth *jwtauth.JWTAuth

	// Repositories
	repositories *repositories

	// Services
	services *services

	// Handlers
	handlers *handlers

	// HTTP Server
	router     *chi.Mux
	server     *http.Server
	webService *web.WebService

	// Background Services
	schedulerService *scheduler.Scheduler
	ffmpegService    ffmpeg.Service
}

type repositories struct {
	libraryRepo      repository.LibraryRepository
	userRepo         repository.UserRepository
	movieRepo        repository.MovieRepository
	seriesRepo       repository.SeriesRepository
	seasonRepo       repository.SeasonRepository
	episodeRepo      repository.EpisodeRepository
	schedulerRepo    repository.SchedulerRepository
	watchHistoryRepo repository.WatchHistoryRepository
	watchlistRepo    repository.WatchlistRepository
	favoriteRepo     repository.FavoriteRepository
	ratingRepo       repository.RatingRepository
}

type services struct {
	authService      service.AuthService
	userService      service.UserService
	mediaService     service.MediaService
	scannerService   scanner.Service
	tmdbService      *metadata.TMDbService
	cleanupService   cleanup.Service
	extractorService extractor.Service
}

type handlers struct {
	authHandler   handler.AuthHandler
	movieHandler  handler.MovieHandler
	seriesHandler handler.SeriesHandler
	userHandler   handler.UserHandler
	webHandler    handler.WebHandler
}

func (a *app) initServices() *services {
	// Create the JWT auth once
	tokenAuth := jwtauth.New("HS256", []byte(a.config.Auth.JWTSecret), nil)
	a.tokenAuth = tokenAuth

	// Initialise services
	return &services{
		authService: service.NewAuthService(a.repositories.userRepo, a.config, tokenAuth),
		userService: service.NewUserService(a.repositories.userRepo),
		mediaService: service.NewMediaService(
			a.repositories.movieRepo,
			a.repositories.seriesRepo,
			a.repositories.seasonRepo,
			a.repositories.episodeRepo,
		),
		tmdbService:      metadata.NewTMDbService(a.config),
		extractorService: extractor.NewExtractor(a.appLogger, a.ffmpegService),
		scannerService: scanner.NewScannerService(
			a.config,
			a.appLogger,
			a.repositories.libraryRepo,
			a.repositories.movieRepo,
			a.repositories.seriesRepo,
			a.repositories.seasonRepo,
			a.repositories.episodeRepo,
			a.services.tmdbService,
			a.services.extractorService,
		),
		cleanupService: cleanup.NewCleanupService(a.config, a.appLogger, a.repositories.libraryRepo),
	}
}

func (a *app) initHandlers() *handlers {
	// Initialise the JWT verifier
	jwtVerifier := auth.NewJWTVerifier(a.tokenAuth)

	return &handlers{
		authHandler:   handler.NewAuthHandler(a.services.authService, a.config),
		movieHandler:  handler.NewMovieHandler(a.services.mediaService, jwtVerifier),
		seriesHandler: handler.NewSeriesHandler(a.services.mediaService, jwtVerifier),
		userHandler:   handler.NewUserHandler(a.services.userService, a.services.authService, jwtVerifier),
	}
}

func (a *app) initWebService(webFS embed.FS) {
	a.webService = web.NewWebService(
		a.config,
		a.appLogger,
		a.services.userService,
		a.services.mediaService,
		a.tokenAuth,
		webFS,
	)
}

func (a *app) initRouter() {
	handlers := a.initHandlers()
	a.router = router.NewRouter(
		a.config,
		handler.movieHandler,
		handler.seriesHandler,
		handler.userHandler,
		handler.authHandler,
		handler.webHandler,
	)
}

func initConfig() (*config.Config, error) {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	return cfg, nil
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("Cinea failed to start: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, webFS embed.FS) error {
	// Create app instance
	app := &app{}

	// Initialise Configuration
	cfg, err := initConfig()
	if err != nil {
		return fmt.Errorf("failed to initialise config: %w", err)
	}
	app.config = cfg

	// Initialise Logger
	appLogger, err := logger.NewLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialise logger: %w", err)
	}
	app.appLogger = appLogger

	// Initialize FFmpeg service
	ffmpegService, err := ffmpeg.NewFFMpegService(appLogger)
	if err != nil {
		return fmt.Errorf("failed to initialise FFmpeg service")
	}
	app.ffmpegService = ffmpegService

	// Ensure FFmpeg binaries are installed
	if err := ffmpegService.EnsureInstalled(); err != nil {
		return fmt.Errorf("failed to verify FFmpeg is installed")
	}

	// Database and Repositories
	db, err := repository.NewDB(cfg, appLogger)
	if err != nil {
		return fmt.Errorf("failed to initialise the database: %w", err)
	}
	app.db = db

	app.repositories = app.initRepositories(db)
	app.services = app.initServices()
	app.initWebService(webFS)
	app.initRouter()

	// Initialise Scheduler
	schedulerService, err := scheduler.NewScheduler(app.appLogger, app.repositories.schedulerRepo)
	if err != nil {
		return fmt.Errorf("failed to initialise scheduler: %w", err)
	}
	app.schedulerService = &schedulerService

	schedulerService.RegisterTask("scanner", app.services.scannerService)
	schedulerService.RegisterTask("cleanup", app.services.cleanupService)

	if err := schedulerService.LoadTasks(ctx); err != nil {
		return fmt.Errorf("failed to load scheduler tasks: %w", err)
	}

	schedulerService.Start(ctx)
	defer schedulerService.Shutdown(ctx)

	// Initialise HTTP Server
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      app.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	app.server = server

	go func() {
		app.appLogger.Info().Msgf("Starting server on port %d", cfg.Server.Port)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.appLogger.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	app.appLogger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := app.server.Shutdown(ctx); err != nil {
		app.appLogger.Fatal().Err(err).Msg("Server forced to shutdown")
		return err
	}

	app.appLogger.Info().Msg("Server exiting")
	return nil
}
