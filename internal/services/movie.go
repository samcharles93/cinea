package services

import (
	"context"
	"fmt"

	"github.com/samcharles93/cinea/internal/dto"
	"github.com/samcharles93/cinea/internal/persistence"
)

type MovieService interface {
	GetMovieByID(ctx context.Context, id uint) (*dto.MovieDTO, error)
	GetAllMovies(ctx context.Context) ([]*dto.MovieDTO, error)
}

type movieService struct {
	movieRepo persistence.MovieRepository
}

func NewMovieService(movieRepo persistence.MovieRepository) MovieService {
	return &movieService{
		movieRepo: movieRepo,
	}
}

func (s *movieService) GetMovieByID(ctx context.Context, id uint) (*dto.MovieDTO, error) {
	movie, err := s.movieRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie by ID: %w", err)
	}
	if movie == nil {
		return nil, nil // Or a specific "not found" error
	}

	// Convert entity to DTO
	movieDTO := &dto.MovieDTO{
		ID:    movie.ID,
		Title: movie.Title,
		// ... map other fields ...
	}
	return movieDTO, nil
}

func (s *movieService) GetAllMovies(ctx context.Context) ([]*dto.MovieDTO, error) {
	movies, err := s.movieRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all movies: %w", err)
	}

	movieDTOs := make([]*dto.MovieDTO, len(movies))
	for i, m := range movies {
		movieDTOs[i] = &dto.MovieDTO{
			ID:    m.ID,
			Title: m.Title,
		}
	}
	return movieDTOs, nil
}
