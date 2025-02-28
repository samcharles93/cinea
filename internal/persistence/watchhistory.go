package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WatchHistoryRepository interface {
	GetWatchHistory(ctx context.Context, userID uint) ([]entity.WatchHistory, error)
	AddToWatchHistory(ctx context.Context, history *entity.WatchHistory) error
	UpdateWatchProgress(ctx context.Context, historyID uint, progress float64) error
	ClearHistory(ctx context.Context, userId uint) ([]entity.WatchHistory, error)
}

type watchHistoryRepository struct {
	db        *gorm.DB
	appLogger logger.Logger
}

func NewWatchHistoryRepository(db *gorm.DB, appLogger logger.Logger) WatchHistoryRepository {
	return &watchHistoryRepository{db: db, appLogger: appLogger}
}

func (r *watchHistoryRepository) GetWatchHistory(ctx context.Context, userID uint) ([]entity.WatchHistory, error) {
	r.appLogger.Debug().
		Uint("userID", userID).
		Msg("Getting watch history for user")

	var history []entity.WatchHistory
	start := time.Now()
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&history)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to get watch history")
		return nil, fmt.Errorf("failed to get watch history: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", userID).
		Int("historyCount", len(history)).
		Dur("duration", duration).
		Msg("Watch history retrieved successfully")
	return history, nil
}

func (r *watchHistoryRepository) AddToWatchHistory(ctx context.Context, history *entity.WatchHistory) error {
	r.appLogger.Debug().
		Uint("userID", history.UserID).
		Str("mediaType", history.MediaType).
		Uint("mediaID", history.MediaID).
		Msg("Adding to watch history")

	start := time.Now()
	result := r.db.WithContext(ctx).Clauses(clause.Returning{}).Create(history)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to add to watch history")
		return fmt.Errorf("failed to add to watch history: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", history.UserID).
		Str("mediaType", history.MediaType).
		Uint("mediaID", history.MediaID).
		Uint("historyID", history.ID).
		Dur("duration", duration).
		Msg("Added to watch history successfully")
	return nil
}

func (r *watchHistoryRepository) UpdateWatchProgress(ctx context.Context, historyID uint, progress float64) error {
	r.appLogger.Debug().
		Uint("historyID", historyID).
		Float64("progress", progress).
		Msg("Updating watch progress")

	start := time.Now()
	result := r.db.WithContext(ctx).Model(&entity.WatchHistory{}).Where("id = ?", historyID).Update("progress", progress)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to update watch progress")
		return fmt.Errorf("failed to update watch progress: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("historyID", historyID).
		Float64("progress", progress).
		Dur("duration", duration).
		Msg("Watch progress updated successfully")
	return nil
}

func (r *watchHistoryRepository) ClearHistory(ctx context.Context, userId uint) ([]entity.WatchHistory, error) {
	r.appLogger.Debug().
		Uint("userID", userId).
		Msg("Clearing watch history for user")

	start := time.Now()
	// First, get the history to return it
	var history []entity.WatchHistory
	r.db.Unscoped().Where("user_id = ?", userId).Find(&history)

	result := r.db.WithContext(ctx).Where("user_id = ?", userId).Delete(&entity.WatchHistory{})
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to clear watch history")
		return nil, fmt.Errorf("failed to clear history: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", userId).
		Dur("duration", duration).
		Msg("Watch history cleared successfully")
	return history, nil
}
