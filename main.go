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

	"github.com/samcharles93/cinea/config"
	"github.com/samcharles93/cinea/internal/ffmpeg"
	"github.com/samcharles93/cinea/internal/logger"
	"github.com/samcharles93/cinea/internal/persistence"
	"github.com/samcharles93/cinea/internal/router"
	"github.com/samcharles93/cinea/internal/services/cleanup"
	"github.com/samcharles93/cinea/internal/services/extractor"
	"github.com/samcharles93/cinea/internal/services/metadata"
	"github.com/samcharles93/cinea/internal/services/scanner"
	"github.com/samcharles93/cinea/internal/services/scheduler"
	"github.com/samcharles93/cinea/web"
)

//go:embed web/templates/* web/static/*
var webFS embed.FS

type app struct {
	config           *config.Config
	appLogger        logger.Logger
	schedulerService *scheduler.Scheduler
	server           *http.Server
	webService       *web.WebService
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("Cinea failed to start: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Configuration
	cfg, err := initConfig()
	if err != nil {
		return fmt.Errorf("failed to initialise config: %w", err)
	}

	// Logger
	appLogger, err := logger.NewLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialise logger: %w", err)
	}

	// Initialize FFmpeg service
	ffmpegService, err := ffmpeg.NewFFMpegService(appLogger)
	if err != nil {
		return fmt.Errorf("failed to initialise FFmpeg service")
	}

	// Ensure FFmpeg binaries are installed
	if err := ffmpegService.EnsureInstalled(); err != nil {
		return fmt.Errorf("failed to verify FFmpeg is installed")
	}

	// Database and Repositories
	db, err := persistence.NewDB(cfg, appLogger)
	if err != nil {
		return fmt.Errorf("failed to initialise the database: %w", err)
	}

	libraryRepo := persistence.NewLibraryRepository(db, appLogger)
	userRepo := persistence.NewUserRepository(db, appLogger)
	movieRepo := persistence.NewMovieRepository(db, appLogger)
	seriesRepo := persistence.NewSeriesRepository(db, appLogger)
	seasonRepo := persistence.NewSeasonRepository(db, appLogger)
	episodeRepo := persistence.NewEpisodeRepository(db, appLogger)
	schedulerRepo := persistence.NewSchedulerRepository(db, appLogger)
	watchHistoryRepo := persistence.NewWatchHistoryRepository(db, appLogger)
	watchlistRepo := persistence.NewWatchlistRepository(db, appLogger)
	favoriteRepo := persistence.NewFavoriteRepository(db, appLogger)
	ratingRepo := persistence.NewRatingRepository(db, appLogger)

	// db.LibraryRepo()
	// Initialize repositories
	// libraryRepo := db.LibraryRepository()
	// userRepo := db.UserRepository()
	// movieRepo := db.MovieRepository()
	// seriesRepo := db.SeriesRepository()
	// seasonRepo := db.SeasonRepository()
	// episodeRepo := db.EpisodeRepository()
	// schedulerRepo := db.SchedulerRepository()
	// watchHistoryRepo := db.WatchHistoryRepository()
	// watchlistRepo := db.WatchlistRepository()
	// favoriteRepo := db.FavoriteRepository()
	// ratingRepo := db.RatingRepository()

	// Services
	tmdbService := metadata.NewTMDbService(cfg)
	extractorService, err := extractor.NewExtractor()
	if err != nil {
		log.Fatalf("Failed to initialise extractor service: %v", err)
	}
	scannerService := scanner.NewScannerService(cfg, appLogger, libraryRepo, movieRepo, seriesRepo, seasonRepo, episodeRepo, tmdbService, extractorService)
	cleanupService := cleanup.NewCleanupService(cfg, appLogger, libraryRepo)

	schedulerService, err := scheduler.NewScheduler(appLogger, schedulerRepo)
	if err != nil {
		return fmt.Errorf("failed to initialise scheduler: %w", err)
	}

	schedulerService.RegisterTask("scanner", scannerService)
	schedulerService.RegisterTask("cleanup", cleanupService)

	if err := schedulerService.LoadTasks(ctx); err != nil {
		return fmt.Errorf("failed to load scheduler tasks: %w", err)
	}

	schedulerService.Start(ctx)
	defer schedulerService.Shutdown(ctx)

	// Web Service
	webService := web.NewWebService(cfg, userRepo, libraryRepo, movieRepo, seriesRepo, seasonRepo, episodeRepo, tokenAuth, webFS)

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	movieHandler := handlers.NewMovieHandler(movieRepo, tmdbService)
	seriesHandler := handlers.NewSeriesHandler(seriesRepo, tmdbService)
	userHandler := handlers.NewUserHandler(userRepo, cfg)

	// Router (inject handlers)
	r := router.NewRouter(cfg, movieHandler, seriesHandler, userHandler, authHandler, webService)

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	app := &app{
		config:           cfg,
		logger:           logService,
		schedulerService: schedulerService,
		server:           server,
		webService:       webService,
	}

	// Start the server
	go func() {
		logger.Info().Msgf("Starting server on port %d", cfg.Server.Port)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := app.server.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Server forced to shutdown")
		return err
	}

	logger.Info().Msg("Server exiting")
	return nil
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
