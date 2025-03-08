package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SeriesRepository interface {
	// Basic CRUD
	Store(ctx context.Context, show *entity.Series) error
	FindByID(ctx context.Context, id uint) (*entity.Series, error)
	FindAll(ctx context.Context) ([]*entity.Series, error)
	Update(ctx context.Context, show *entity.Series) error

	// Soft Delete Management
	Delete(ctx context.Context, id uint) error
	HardDelete(ctx context.Context, id uint) error
	FindAllDeleted(ctx context.Context) ([]*entity.Series, error)
	Restore(ctx context.Context, id uint) error
	CleanupDeletedShows(ctx context.Context, olderThan time.Duration) error

	// Scanning Management
	UpdateScannedStatus(ctx context.Context, id uint) error
	FindStaleShows(ctx context.Context, threshold time.Duration) ([]*entity.Series, error)
}

type seriesRepository struct {
	db        *gorm.DB
	appLogger logger.Logger
}

func NewSeriesRepository(db *gorm.DB, appLogger logger.Logger) SeriesRepository {
	return &seriesRepository{
		db:        db,
		appLogger: appLogger,
	}
}

// TV Series Management
// Basic CRUD
func (r *seriesRepository) Store(ctx context.Context, show *entity.Series) error {
	result := r.db.WithContext(ctx).Clauses(clause.Returning{}).Create(show)
	if result.Error != nil {
		return fmt.Errorf("failed to create show: %w", result.Error)
	}
	return nil
}

func (r *seriesRepository) FindByID(ctx context.Context, id uint) (*entity.Series, error) {
	var show entity.Series
	result := r.db.WithContext(ctx).Preload("Seasons.Episodes").First(&show, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find show by id: %w", result.Error)
	}
	return &show, nil
}

func (r *seriesRepository) FindAll(ctx context.Context) ([]*entity.Series, error) {
	var shows []*entity.Series
	result := r.db.WithContext(ctx).Preload("Seasons.Episodes").Find(&shows)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list shows: %w", result.Error)
	}
	return shows, nil
}

func (r *seriesRepository) Update(ctx context.Context, show *entity.Series) error {
	result := r.db.WithContext(ctx).Save(show)
	if result.Error != nil {
		return fmt.Errorf("failed to update show: %w", result.Error)
	}
	return nil
}

// Soft Delete Management
// Delete will delete a series and cascade down to seasons and episodes.
func (r *seriesRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entity.Series{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete series: %w", result.Error)
	}
	return nil
}

func (r *seriesRepository) HardDelete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&entity.Series{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to hard delete series: %w", result.Error)
	}
	return nil
}
func (r *seriesRepository) FindAllDeleted(ctx context.Context) ([]*entity.Series, error) {
	var shows []*entity.Series
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Find(&shows)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find all deleted shows: %w", result.Error)
	}
	return shows, nil
}

func (r *seriesRepository) Restore(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().Model(&entity.Series{}).Where("id = ?", id).Update("deleted_at", nil)
	if result.Error != nil {
		return fmt.Errorf("failed to restore series: %w", result.Error)
	}
	return nil
}

func (r *seriesRepository) CleanupDeletedShows(ctx context.Context, olderThan time.Duration) error {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at < ?", time.Now().Add(-olderThan)).Delete(&entity.Series{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup deleted shows: %w", result.Error)
	}
	return nil
}

// Scanning Management
func (r *seriesRepository) UpdateScannedStatus(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Model(&entity.Series{}).Where("id=?", id).Update("last_scanned", time.Now())
	if result.Error != nil {
		return fmt.Errorf("failed to update scanned status")
	}
	return nil
}

func (r *seriesRepository) FindStaleShows(ctx context.Context, threshold time.Duration) ([]*entity.Series, error) {
	var shows []*entity.Series
	result := r.db.WithContext(ctx).Where("last_scanned < ?", time.Now().Add(-threshold)).Find(&shows)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find stale shows: %w", result.Error)
	}
	return shows, nil
}
