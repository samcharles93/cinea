package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/errors"
	"github.com/samcharles93/cinea/internal/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MovieRepository interface {
	Store(ctx context.Context, movie *entity.Movie) error
	FindByID(ctx context.Context, id uint) (*entity.Movie, error)
	FindByPath(ctx context.Context, path string) (*entity.Movie, error)
	FindAll(ctx context.Context) ([]*entity.Movie, error)
	Update(ctx context.Context, movie *entity.Movie) error
	Delete(ctx context.Context, id uint) error

	HardDelete(ctx context.Context, id uint) error
	FindAllDeleted(ctx context.Context) ([]*entity.Movie, error)
	Restore(ctx context.Context, id uint) error
	UpdateScannedStatus(ctx context.Context, id uint) error
}

type movieRepository struct {
	db        *gorm.DB
	appLogger logger.Logger
}

func NewMovieRepository(db *gorm.DB, appLogger logger.Logger) MovieRepository {
	return &movieRepository{
		db:        db,
		appLogger: appLogger,
	}
}

func (r *movieRepository) Store(ctx context.Context, movie *entity.Movie) error {
	result := r.db.WithContext(ctx).Clauses(clause.Returning{}).Create(movie)
	if result.Error != nil {
		return fmt.Errorf("failed to store movie: %w", result.Error)
	}
	return nil
}

func (r *movieRepository) FindByID(ctx context.Context, id uint) (*entity.Movie, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid movie ID: %w", errors.ErrBadRequest)
	}

	var movie entity.Movie
	result := r.db.WithContext(ctx).First(&movie, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("movie with ID %d not found: %w", id, errors.ErrNotFound)
		}
		return nil, fmt.Errorf("database error finding movie %d: %w", id, result.Error)
	}

	return &movie, nil
}

func (r *movieRepository) FindByPath(ctx context.Context, path string) (*entity.Movie, error) {
	var movie entity.Movie
	result := r.db.WithContext(ctx).Where("file_path = ?", path).First(&movie)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find movie by path: %w", result.Error)
	}
	return &movie, nil
}
func (r *movieRepository) FindAll(ctx context.Context) ([]*entity.Movie, error) {
	var movies []*entity.Movie
	result := r.db.WithContext(ctx).Find(&movies)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find all movies: %w", result.Error)
	}
	return movies, nil
}

func (r *movieRepository) Update(ctx context.Context, movie *entity.Movie) error {
	result := r.db.WithContext(ctx).Save(movie)
	if result.Error != nil {
		return fmt.Errorf("failed to update movie: %w", result.Error)
	}
	return nil
}

func (r *movieRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entity.Movie{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete movie: %w", result.Error)
	}
	return nil
}

func (r *movieRepository) HardDelete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&entity.Movie{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to hard delete movie: %w", result.Error)
	}
	return nil
}

func (r *movieRepository) FindAllDeleted(ctx context.Context) ([]*entity.Movie, error) {
	var movies []*entity.Movie
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Find(&movies)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find all deleted movies: %w", result.Error)
	}
	return movies, nil
}

func (r *movieRepository) Restore(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().Model(&entity.Movie{}).Where("id = ?", id).Update("deleted_at", nil)
	if result.Error != nil {
		return fmt.Errorf("failed to restore movie: %w", result.Error)
	}
	return nil
}

func (r *movieRepository) UpdateScannedStatus(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Model(&entity.Movie{}).Where("id = ?", id).Update("last_scanned", time.Now())
	if result.Error != nil {
		return fmt.Errorf("failed to update scanned status: %w", result.Error)
	}
	return nil
}
