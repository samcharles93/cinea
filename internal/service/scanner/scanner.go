package scanner

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/samcharles93/cinea/config"
	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/logger"
	"github.com/samcharles93/cinea/internal/repository"
	"github.com/samcharles93/cinea/internal/service/extractor"
	"github.com/samcharles93/cinea/internal/service/metadata"
)

type Service interface {
	ScanLibrary(ctx context.Context, lib *entity.Library) error
	ScanLibraries(ctx context.Context) error
	scanPath(ctx context.Context, lib *entity.Library, path string) error

	// Task scheduler methods
	Execute(ctx context.Context, config string) error
	Description() string
}

type service struct {
	config         *config.Config
	appLogger      logger.Logger
	libraryRepo    repository.LibraryRepository
	movieRepo      repository.MovieRepository
	seriesRepo     repository.SeriesRepository
	seasonRepo     repository.SeasonRepository
	episodeRepo    repository.EpisodeRepository
	tmdb           *metadata.TMDbService
	mediaExtractor extractor.Service
}

type tvShowInfo struct {
	Title   string
	Season  int
	Episode int
}

type mediaInfo struct {
	Title string
	Year  string
}

func NewScannerService(
	cfg *config.Config,
	appLogger logger.Logger,
	libraryRepo repository.LibraryRepository,
	movieRepo repository.MovieRepository,
	seriesRepo repository.SeriesRepository,
	seasonRepo repository.SeasonRepository,
	episodeRepo repository.EpisodeRepository,
	tmdb *metadata.TMDbService,
	mediaExtractor extractor.Service,
) Service {
	return &service{
		config:         cfg,
		appLogger:      appLogger,
		libraryRepo:    libraryRepo,
		movieRepo:      movieRepo,
		seriesRepo:     seriesRepo,
		seasonRepo:     seasonRepo,
		episodeRepo:    episodeRepo,
		tmdb:           tmdb,
		mediaExtractor: mediaExtractor,
	}
}

// Execute implements the scheduler.TaskExecutor interface
func (s *service) Execute(ctx context.Context, config string) error {
	// TODO - Parse the configStr to get scanner configuration (e.g. which libraries to scan)

	s.appLogger.Info().Str("package", "scanner").Msg("Starting scan from the scheduler")
	return s.ScanLibraries(ctx)
}

func (s *service) Description() string {
	return "Scans media libraries for new files."
}

func (s *service) ScanLibraries(ctx context.Context) error {
	libraries, err := s.libraryRepo.ListLibraries(ctx)
	if err != nil {
		return err
	}

	for _, lib := range libraries {
		if !lib.AutoScan {
			continue
		}

		if err := s.ScanLibrary(ctx, lib); err != nil {
			s.appLogger.Error().
				Err(err).
				Str("library", lib.Name).
				Msg("Failed to scan library")
		}
	}
	return nil
}

func (s *service) ScanLibrary(ctx context.Context, lib *entity.Library) error {
	s.appLogger.Info().
		Str("library", lib.Name).
		Str("type", string(lib.Type)).
		Msg("Starting library scan")

	for _, path := range lib.Paths {
		if !path.Enabled {
			continue
		}

		if err := s.scanPath(ctx, lib, path.Path); err != nil {
			s.appLogger.Error().
				Err(err).
				Str("library", lib.Name).
				Str("path", path.Path).
				Msg("Failed to scan path")
		}
	}

	lib.LastScanned = time.Now()
	return s.libraryRepo.UpdateLibrary(ctx, lib)
}

func (s *service) scanPath(ctx context.Context, lib *entity.Library, path string) error {
	return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !isVideoFile(filePath) {
			return nil
		}

		return s.processFile(ctx, lib, filePath)
	})
}

func (s *service) processFile(ctx context.Context, lib *entity.Library, filePath string) error {
	// Determine if file is likely tv show episode or a movie
	if isLikelyTVFile(filePath) {
		return s.processSeriesFile(ctx, lib, filePath)
	} else {
		return s.processMovieFile(ctx, lib, filePath)
	}
}
