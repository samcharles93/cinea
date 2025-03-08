package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RatingRepository interface {
	GetRatings(ctx context.Context, userID uint) ([]entity.Rating, error)
	AddRating(ctx context.Context, rating *entity.Rating) error
	UpdateRating(ctx context.Context, rating *entity.Rating) error
	RemoveRating(ctx context.Context, userID uint, mediaID uint, mediaType string) error
}

type ratingRepository struct {
	db        *gorm.DB
	appLogger logger.Logger
}

func NewRatingRepository(db *gorm.DB, appLogger logger.Logger) RatingRepository {
	return &ratingRepository{db: db, appLogger: appLogger}
}

func (r *ratingRepository) GetRatings(ctx context.Context, userID uint) ([]entity.Rating, error) {
	r.appLogger.Debug().
		Uint("userID", userID).
		Msg("Getting ratings for user")

	var ratings []entity.Rating
	start := time.Now()
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&ratings)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to get ratings")
		return nil, fmt.Errorf("failed to get ratings: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", userID).
		Int("ratingCount", len(ratings)).
		Dur("duration", duration).
		Msg("Ratings retrieved successfully")
	return ratings, nil
}

func (r *ratingRepository) AddRating(ctx context.Context, rating *entity.Rating) error {
	r.appLogger.Debug().
		Uint("userID", rating.UserID).
		Str("mediaType", rating.MediaType).
		Uint("mediaID", rating.MediaID).
		Msg("Adding rating")

	start := time.Now()
	result := r.db.WithContext(ctx).Clauses(clause.Returning{}).Create(rating)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to add rating")
		return fmt.Errorf("failed to add rating: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", rating.UserID).
		Str("mediaType", rating.MediaType).
		Uint("mediaID", rating.MediaID).
		Dur("duration", duration).
		Msg("Rating added successfully")
	return nil
}

func (r *ratingRepository) UpdateRating(ctx context.Context, rating *entity.Rating) error {
	r.appLogger.Debug().
		Uint("userID", rating.UserID).
		Str("mediaType", rating.MediaType).
		Uint("mediaID", rating.MediaID).
		Msg("Updating rating")

	start := time.Now()
	// Updates only non-zero fields
	result := r.db.WithContext(ctx).Model(rating).Updates(rating)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to update rating")
		return fmt.Errorf("failed to update rating: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", rating.UserID).
		Str("mediaType", rating.MediaType).
		Uint("mediaID", rating.MediaID).
		Dur("duration", duration).
		Msg("Rating updated successfully")
	return nil
}

func (r *ratingRepository) RemoveRating(ctx context.Context, userID uint, mediaID uint, mediaType string) error {
	r.appLogger.Debug().
		Uint("userID", userID).
		Str("mediaType", mediaType).
		Uint("mediaID", mediaID).
		Msg("Removing rating")

	start := time.Now()
	result := r.db.WithContext(ctx).Where("user_id = ? AND media_id = ? AND media_type = ?", userID, mediaID, mediaType).Delete(&entity.Rating{})
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to remove rating")
		return fmt.Errorf("failed to remove rating: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", userID).
		Str("mediaType", mediaType).
		Uint("mediaID", mediaID).
		Dur("duration", duration).
		Msg("Rating removed successfully")
	return nil
}
