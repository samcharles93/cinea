package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SeasonRepository interface {
	AddSeason(ctx context.Context, season *entity.Season) error
	FindBySeriesID(ctx context.Context, seriesID uint) (*entity.Season, error)
	UpdateSeason(ctx context.Context, season *entity.Season) error
	DeleteSeason(ctx context.Context, id uint) error
}

type seasonRepository struct {
	db        *gorm.DB
	appLogger logger.Logger
}

func NewSeasonRepository(db *gorm.DB, appLogger logger.Logger) SeasonRepository {
	return &seasonRepository{
		db:        db,
		appLogger: appLogger,
	}
}

func (r *seasonRepository) AddSeason(ctx context.Context, season *entity.Season) error {
	result := r.db.WithContext(ctx).Clauses(clause.Returning{}).Create(season)
	if result.Error != nil {
		return fmt.Errorf("failed to add season: %w", result.Error)
	}
	return nil
}

func (r *seasonRepository) FindBySeriesID(ctx context.Context, seriesID uint) (*entity.Season, error) {
	var season entity.Season
	result := r.db.WithContext(ctx).Preload("Episodes").First(&season, seriesID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find season by id: %w", result.Error)
	}
	return &season, nil
}

func (r *seasonRepository) UpdateSeason(ctx context.Context, season *entity.Season) error {
	result := r.db.WithContext(ctx).Save(season)
	if result.Error != nil {
		return fmt.Errorf("failed to update season: %w", result.Error)
	}
	return nil
}

func (r *seasonRepository) DeleteSeason(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entity.Season{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete season: %w", result.Error)
	}
	return nil
}
