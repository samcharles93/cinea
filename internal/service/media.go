package service

import (
	"context"
	"fmt"

	"github.com/samcharles93/cinea/internal/dto"
	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/errors"
	"github.com/samcharles93/cinea/internal/repository"
)

type MediaService interface {
	// Movie
	GetAllMovies(ctx context.Context) ([]*dto.MovieDTO, error)
	GetMovieByID(ctx context.Context, id uint) (*dto.MovieDTO, error)
	CreateMovie(ctx context.Context, movie *dto.CreateMovieDTO) (*dto.MovieDTO, error)
	UpdateMovie(ctx context.Context, id uint, movie *dto.UpdateMovieDTO) (*dto.MovieDTO, error)
	DeleteMovie(ctx context.Context, id uint) error

	// Series
	GetAllSeries(ctx context.Context) ([]*dto.SeriesDTO, error)
	GetSeriesByID(ctx context.Context, id uint) (*dto.SeriesDTO, error)
	GetSeriesWithDetails(ctx context.Context, id uint) (*dto.SeriesDTO, error)
	CreateSeries(ctx context.Context, series *dto.CreateSeriesDTO) (*dto.SeriesDTO, error)
	UpdateSeries(ctx context.Context, id uint, series *dto.UpdateSeriesDTO) (*dto.SeriesDTO, error)
	DeleteSeries(ctx context.Context, id uint) error

	// Season
	GetAllSeasons(ctx context.Context, seriesID uint) ([]*dto.SeasonDTO, error)
	GetSeasonByID(ctx context.Context, id uint) (*dto.SeasonDTO, error)
	GetSeasonByNumber(ctx context.Context, seriesID uint, seasonNumber int) (*dto.SeasonDTO, error)

	// Episode
	GetAllEpisodes(ctx context.Context, seasonID uint, seriesID uint) ([]*dto.EpisodeDTO, error)
	GetEpisodeByID(ctx context.Context, id uint) (*dto.EpisodeDTO, error)
	GetEpisodeByNumber(ctx context.Context, seriesID uint, seasonNumber int, episodeNumber int) (*dto.EpisodeDTO, error)

	// Stream
	GetStreamURL(ctx context.Context, mediaType string, mediaID uint) (string, error)
}

type mediaService struct {
	movieRepo   repository.MovieRepository
	seriesRepo  repository.SeriesRepository
	seasonRepo  repository.SeasonRepository
	episodeRepo repository.EpisodeRepository
}

func NewMediaService(
	movieRepo repository.MovieRepository,
	seriesRepo repository.SeriesRepository,
	seasonRepo repository.SeasonRepository,
	episodeRepo repository.EpisodeRepository) MediaService {
	return &mediaService{
		movieRepo:   movieRepo,
		seriesRepo:  seriesRepo,
		seasonRepo:  seasonRepo,
		episodeRepo: episodeRepo,
	}
}

// Movie functions
func (s *mediaService) GetMovieByID(ctx context.Context, id uint) (*dto.MovieDTO, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid movie ID: %w", errors.ErrBadRequest)
	}

	movie, err := s.movieRepo.FindByID(ctx, id)
	if err != nil {
		// Don't wrap errors that are already wrapped
		if errors.Is(err, errors.ErrNotFound) || errors.Is(err, errors.ErrBadRequest) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get movie by ID: %w", err)
	}

	if movie == nil {
		return nil, fmt.Errorf("movie with ID %d not found: %w", id, errors.ErrNotFound)
	}

	return dto.MovieToDTO(movie), nil
}

func (s *mediaService) GetAllMovies(ctx context.Context) ([]*dto.MovieDTO, error) {
	movies, err := s.movieRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all movies: %w", err)
	}
	return dto.MoviesToDTO(movies), nil
}

// Series functions
func (s *mediaService) GetSeriesByID(ctx context.Context, id uint) (*dto.SeriesDTO, error) {
	series, err := s.seriesRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get series by id: %w", err)
	}
	if series == nil {
		return nil, nil
	}
	// Return basic series information without detailed episode data
	return dto.SeriesToDTO(series), nil
}

func (s *mediaService) GetSeriesWithDetails(ctx context.Context, id uint) (*dto.SeriesDTO, error) {
	series, err := s.seriesRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get series by id: %w", err)
	}
	if series == nil {
		return nil, nil
	}
	// Return complete series information with detailed episode data
	return dto.GetSeriesWithDetails(series), nil
}

func (s *mediaService) GetAllSeries(ctx context.Context) ([]*dto.SeriesDTO, error) {
	series, err := s.seriesRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get series: %w", err)
	}
	return dto.SeriesToDTOs(series), nil
}

// Season functions

func (s *mediaService) GetAllSeasons(ctx context.Context, seriesID uint) ([]*dto.SeasonDTO, error) {
	series, err := s.seriesRepo.FindByID(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to get series: %w", err)
	}
	if series == nil {
		return nil, fmt.Errorf("series not found")
	}

	// Convert entity seasons to DTO
	seriesDTO := dto.SeriesToDTO(series)
	seasons := make([]*dto.SeasonDTO, len(seriesDTO.Seasons))

	for i, season := range seriesDTO.Seasons {
		seasonCopy := season // Create a copy to avoid referencing the loop variable
		seasons[i] = &seasonCopy
	}

	return seasons, nil
}

func (s *mediaService) GetSeasonByID(ctx context.Context, id uint) (*dto.SeasonDTO, error) {
	// Since we don't have direct method to get a season by ID in the repository,
	// we need to find the series first and then find the season
	// This is a bit inefficient, but works with the current repository structure

	// We'll have to query the DB to find a season by ID
	// Assuming we can get the series ID from the season ID
	season, err := s.seasonRepo.FindBySeriesID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get season: %w", err)
	}
	if season == nil {
		return nil, nil
	}

	// Convert to DTO
	return dto.GetSeasonWithDetails(season), nil
}

func (s *mediaService) GetSeasonByNumber(ctx context.Context, seriesID uint, seasonNumber int) (*dto.SeasonDTO, error) {
	// First get the entire series
	series, err := s.seriesRepo.FindByID(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to get series: %w", err)
	}
	if series == nil {
		return nil, fmt.Errorf("series not found")
	}

	// Find the requested season
	var targetSeason *entity.Season
	for _, season := range series.Seasons {
		if season.SeasonNumber == seasonNumber {
			targetSeason = &season
			break
		}
	}

	if targetSeason == nil {
		return nil, nil // Season not found
	}

	// Convert to DTO with details
	return dto.GetSeasonWithDetails(targetSeason), nil
}

// Episode functions

func (s *mediaService) GetAllEpisodes(ctx context.Context, seasonID uint, seriesID uint) ([]*dto.EpisodeDTO, error) {
	// First get the season
	season, err := s.seasonRepo.FindBySeriesID(ctx, seasonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get season: %w", err)
	}
	if season == nil {
		return nil, fmt.Errorf("season not found")
	}

	// Convert episodes to DTOs
	episodes := make([]*dto.EpisodeDTO, len(season.Episodes))
	for i, episode := range season.Episodes {
		episodeCopy := episode // Create a copy to avoid referencing the loop variable
		episodes[i] = dto.GetEpisodeDetails(&episodeCopy)
	}

	return episodes, nil
}

func (s *mediaService) GetEpisodeByID(ctx context.Context, id uint) (*dto.EpisodeDTO, error) {
	episode, err := s.episodeRepo.FindEpisodeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode: %w", err)
	}
	if episode == nil {
		return nil, nil
	}

	return dto.GetEpisodeDetails(episode), nil
}

func (s *mediaService) GetEpisodeByNumber(ctx context.Context, seriesID uint, seasonNumber int, episodeNumber int) (*dto.EpisodeDTO, error) {
	episode, err := s.episodeRepo.FindEpisodeByNumber(ctx, seriesID, seasonNumber, episodeNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode: %w", err)
	}
	if episode == nil {
		return nil, nil
	}

	return dto.GetEpisodeDetails(episode), nil
}
