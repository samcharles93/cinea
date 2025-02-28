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

type FavoriteRepository interface {
	GetFavorites(ctx context.Context, userID uint) ([]entity.Favorite, error)
	AddToFavorites(ctx context.Context, favorite *entity.Favorite) error
	RemoveFromFavorites(ctx context.Context, userID uint, mediaID uint, mediaType string) error
}

type favoriteRepository struct {
	db        *gorm.DB
	appLogger logger.Logger
}

func NewFavoriteRepository(db *gorm.DB, appLogger logger.Logger) FavoriteRepository {
	return &favoriteRepository{
		db:        db,
		appLogger: appLogger,
	}
}

func (r *favoriteRepository) GetFavorites(ctx context.Context, userID uint) ([]entity.Favorite, error) {
	r.appLogger.Debug().
		Uint("userID", userID).
		Msg("Getting favorites for user")

	var favorites []entity.Favorite
	start := time.Now()
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&favorites)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to get favorites")
		return nil, fmt.Errorf("failed to get favorites: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", userID).
		Int("favoriteCount", len(favorites)).
		Dur("duration", duration).
		Msg("Favorites retrieved successfully")
	return favorites, nil
}

func (r *favoriteRepository) AddToFavorites(ctx context.Context, favorite *entity.Favorite) error {
	r.appLogger.Debug().
		Uint("userID", favorite.UserID).
		Str("mediaType", favorite.MediaType).
		Uint("mediaID", favorite.MediaID).
		Msg("Adding to favorites")

	start := time.Now()
	result := r.db.WithContext(ctx).Clauses(clause.Returning{}).Create(favorite)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to add to favorites")
		return fmt.Errorf("failed to add to favorites: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", favorite.UserID).
		Str("mediaType", favorite.MediaType).
		Uint("mediaID", favorite.MediaID).
		Dur("duration", duration).
		Msg("Added to favorites successfully")
	return nil
}

func (r *favoriteRepository) RemoveFromFavorites(ctx context.Context, userID uint, mediaID uint, mediaType string) error {
	r.appLogger.Debug().
		Uint("userID", userID).
		Str("mediaType", mediaType).
		Uint("mediaID", mediaID).
		Msg("Removing from favorites")

	start := time.Now()
	result := r.db.WithContext(ctx).Where("user_id = ? AND media_id = ? AND media_type = ?", userID, mediaID, mediaType).Delete(&entity.Favorite{})
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to remove from favorites")
		return fmt.Errorf("failed to remove from favorites: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", userID).
		Str("mediaType", mediaType).
		Uint("mediaID", mediaID).
		Dur("duration", duration).
		Msg("Removed from favorites successfully")
	return nil
}
