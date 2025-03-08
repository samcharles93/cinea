package cleanup

import (
	"context"

	"github.com/samcharles93/cinea/config"
	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/logger"
	"github.com/samcharles93/cinea/internal/repository"
)

type Service interface {
	Run(ctx context.Context) error

	// Task scheduler methods
	Execute(ctx context.Context, config string) error
	Description() string
}

type service struct {
	config      *config.Config
	appLogger   logger.Logger
	libraryRepo repository.LibraryRepository
}

func NewCleanupService(config *config.Config, appLogger logger.Logger, libraryRepo repository.LibraryRepository) Service {
	return &service{
		config:      config,
		appLogger:   appLogger,
		libraryRepo: libraryRepo,
	}
}

// Cleanup movies that have been soft-deleted for more than cfg.Cleanup.MaxAge days
func (s *service) Run(ctx context.Context) error {
	// Get all libraries
	libraries, err := s.libraryRepo.ListLibraries(ctx)
	if err != nil {
		return err
	}

	for _, lib := range libraries {
		if err := s.cleanupLibrary(ctx, lib); err != nil {
			s.appLogger.Error().
				Err(err).
				Str("library", lib.Name).
				Msg("Failed to cleanup library")
		}
	}

	return nil
}

func (s *service) cleanupLibrary(ctx context.Context, lib *entity.Library) error {
	// Find items with missing files
	if s.config.Jobs.Cleanup.DeleteMissing {
		if err := s.cleanupMissingFiles(ctx, lib); err != nil {
			return err
		}
	}

	// Find orphaned files
	if s.config.Jobs.Cleanup.DeleteOrphaned {
		if err := s.cleanupOrphanedFiles(ctx, lib); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) cleanupOrphanedFiles(ctx context.Context, lib *entity.Library) error {
	// Find and cleanup media files which don't have database entries
	return nil
}

func (s *service) cleanupMissingFiles(ctx context.Context, lib *entity.Library) error {
	// Find and cleanup database entries where media files don't exist
	return nil
}

func (s *service) Execute(ctx context.Context, config string) error {
	// Process the config file to determine what settings are being used in the runner.

	return s.Run(ctx)
}

func (s *service) Description() string {
	return "Cleans up media files or database entries"
}
