package services

import (
	"context"
	"fmt"

	"github.com/samcharles93/cinea/internal/dto"
	"github.com/samcharles93/cinea/internal/persistence"
)

type SeriesService interface {
	GetSeriesByID(ctx context.Context, id uint) (*dto.SeriesDTO, error)
	GetAllSeries(ctx context.Context) ([]*dto.SeriesDTO, error)
}

type seriesService struct {
	seriesRepo persistence.SeriesRepository
}

func NewSeriesService(seriesRepo persistence.SeriesRepository) SeriesService {
	return &seriesService{
		seriesRepo: seriesRepo,
	}
}

func (s *seriesService) GetSeriesByID(ctx context.Context, id uint) (*dto.SeriesDTO, error) {
	series, err := s.seriesRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get series by id: %w", err)
	}
	if series == nil {
		return nil, nil
	}

	seriesDTO := &dto.SeriesDTO{
		ID:    series.ID,
		Title: series.Title,
	}

	return seriesDTO, nil
}

func (s *seriesService) GetAllSeries(ctx context.Context) ([]*dto.SeriesDTO, error) {
	series, err := s.seriesRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get series: %w", err)
	}
	seriesDTOs := make([]*dto.SeriesDTO, len(series))
	for i, tvShow := range series {
		seriesDTOs[i] = &dto.SeriesDTO{
			ID:    tvShow.ID,
			Title: tvShow.Title,
		}
	}
	return seriesDTOs, nil
}
