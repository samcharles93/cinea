package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/logger"
	"gorm.io/gorm"
)

type WatchlistRepository interface {
	GetWatchlist(ctx context.Context, userID uint) ([]entity.Watchlist, error)
	AddToWatchlist(ctx context.Context, item *entity.Watchlist) error
	RemoveFromWatchlist(ctx context.Context, userID uint, mediaID uint, mediaType string) error
}

type watchlistRepository struct {
	db        *gorm.DB
	appLogger logger.Logger
}

func NewWatchlistRepository(db *gorm.DB, appLogger logger.Logger) WatchlistRepository {
	return &watchlistRepository{db: db, appLogger: appLogger}
}

func (r *watchlistRepository) GetWatchlist(ctx context.Context, userID uint) ([]entity.Watchlist, error) {
	r.appLogger.Debug().
		Uint("userID", userID).
		Msg("Getting watchlist for user")

	var watchlist []entity.Watchlist
	start := time.Now()
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&watchlist)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to get watchlist")
		return nil, fmt.Errorf("failed to get watchlist: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", userID).
		Int("watchlistCount", len(watchlist)).
		Dur("duration", duration).
		Msg("Watchlist retrieved successfully")
	return watchlist, nil
}

func (r *watchlistRepository) AddToWatchlist(ctx context.Context, item *entity.Watchlist) error {
	r.appLogger.Debug().
		Uint("userID", item.UserID).
		Str("mediaType", item.MediaType).
		Uint("mediaID", item.MediaID).
		Msg("Adding to watchlist")

	start := time.Now()
	result := r.db.WithContext(ctx).Create(item)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to add to watchlist")
		return fmt.Errorf("failed to add to watchlist: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", item.UserID).
		Str("mediaType", item.MediaType).
		Uint("mediaID", item.MediaID).
		Dur("duration", duration).
		Msg("Added to watchlist successfully")
	return nil
}

func (r *watchlistRepository) RemoveFromWatchlist(ctx context.Context, userID uint, mediaID uint, mediaType string) error {
	r.appLogger.Debug().
		Uint("userID", userID).
		Str("mediaType", mediaType).
		Uint("mediaID", mediaID).
		Msg("Removing from watchlist")

	start := time.Now()
	result := r.db.WithContext(ctx).Where("user_id = ? AND media_id = ? AND media_type = ?", userID, mediaID, mediaType).Delete(&entity.Watchlist{})
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to remove from watchlist")
		return fmt.Errorf("failed to remove from watchlist: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", userID).
		Str("mediaType", mediaType).
		Uint("mediaID", mediaID).
		Dur("duration", duration).
		Msg("Removed from watchlist successfully")
	return nil
}
