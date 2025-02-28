package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EpisodeRepository interface {
	AddEpisode(ctx context.Context, episode *entity.Episode) error
	FindEpisodeByNumber(ctx context.Context, showID uint, seasonNumber, episodeNumber int) (*entity.Episode, error)
	FindEpisodeByID(ctx context.Context, episodeID uint) (*entity.Episode, error)
	UpdateEpisode(ctx context.Context, episode *entity.Episode) error
	DeleteEpisode(ctx context.Context, id uint) error
	FindByPath(ctx context.Context, filePath string) (*entity.Episode, error)
}

type episodeRepository struct {
	db        *gorm.DB
	appLogger logger.Logger
}

func NewEpisodeRepository(db *gorm.DB, appLogger logger.Logger) EpisodeRepository {
	return &episodeRepository{db: db, appLogger: appLogger}
}

func (r *episodeRepository) AddEpisode(ctx context.Context, episode *entity.Episode) error {
	result := r.db.WithContext(ctx).Clauses(clause.Returning{}).Create(episode)
	if result.Error != nil {
		return fmt.Errorf("failed to add episode: %w", result.Error)
	}
	return nil
}

func (r *episodeRepository) FindEpisodeByNumber(ctx context.Context, showID uint, seasonNumber, episodeNumber int) (*entity.Episode, error) {
	var episode entity.Episode
	var season entity.Season

	seasonResult := r.db.WithContext(ctx).Select("id").Where("series_id = ? AND season_number = ?", showID, seasonNumber).First(&season)
	if seasonResult.Error != nil {
		if errors.Is(seasonResult.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find season id: %w", seasonResult.Error)
	}

	result := r.db.WithContext(ctx).
		Where("season_id = ? AND episode_number = ?", season.ID, episodeNumber).
		First(&episode)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get episode: %w", result.Error)
	}
	return &episode, nil
}

func (r *episodeRepository) FindEpisodeByID(ctx context.Context, episodeID uint) (*entity.Episode, error) {
	var episode entity.Episode
	result := r.db.WithContext(ctx).First(&episode, episodeID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find episode by ID: %w", result.Error)
	}
	return &episode, nil
}

func (r *episodeRepository) UpdateEpisode(ctx context.Context, episode *entity.Episode) error {
	result := r.db.WithContext(ctx).Save(episode)
	if result.Error != nil {
		return fmt.Errorf("failed to update episode: %w", result.Error)
	}
	return nil
}

func (r *episodeRepository) DeleteEpisode(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entity.Episode{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete episode: %w", result.Error)
	}
	return nil
}

func (r *episodeRepository) FindByPath(ctx context.Context, filePath string) (*entity.Episode, error) {
	var episode entity.Episode
	result := r.db.WithContext(ctx).Where("file_path = ?", filePath).First(&episode)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find episode with path '%s': %w", filePath, result.Error)

	}
	return &episode, nil
}
