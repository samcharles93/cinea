package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/service/metadata"
)

func (s *service) processMovieFile(ctx context.Context, lib *entity.Library, filePath string) error {
	// 1. Check if the movie already exists (by path)
	existingMovie, err := s.movieRepo.FindByPath(ctx, filePath)
	if err != nil {
		return fmt.Errorf("error checking for existing movie: %w", err)
	}

	if existingMovie != nil {
		existingMovie.LastScanned = time.Now()
		return s.movieRepo.Update(ctx, existingMovie)
	}

	// 2. Extract metadata
	fileMeta, err := s.mediaExtractor.Extract(ctx, filePath)
	if err != nil {
		s.appLogger.Warn().Err(err).Str("filepath", filePath).Msg("Failed to extract movie metadata")
	}

	// 3. Extract movie info (title, year) from the filename.
	movieInfo := extractMovieInfo(filePath)

	// 4. Search TMDb
	searchResult, err := s.tmdb.SearchMovie(ctx, movieInfo.Title, metadata.WithMovieYear(movieInfo.Year))
	if err != nil {
		s.appLogger.Error().Err(err).Str("title", movieInfo.Title).Str("year", movieInfo.Year).Msg("TMDb search failed")
	}
	var tmdbMovie *metadata.Movie

	if searchResult != nil && len(searchResult.Results) > 0 {
		tmdbMovie = &searchResult.Results[0]
		s.appLogger.Info().Str("title", tmdbMovie.Title).Int("tmdb_id", tmdbMovie.ID).Msg("Found movie on TMDb")
	} else {
		s.appLogger.Warn().Str("title", movieInfo.Title).Str("year", movieInfo.Year).Msg("No results found on TMDb")
	}

	// 5. Create and store the movie entity
	movie := &entity.Movie{
		LibraryItem: entity.LibraryItem{
			LibraryID:        lib.ID,
			DateAdded:        time.Now(),
			FilePath:         filePath,
			Container:        fileMeta.Container,
			Codec:            fileMeta.Codec,
			ResolutionWidth:  fileMeta.ResolutionWidth,
			ResolutionHeight: fileMeta.ResolutionHeight,
		},
		LastScanned: time.Now(),
	}
	if len(fileMeta.AudioTracks) > 0 {
		movie.LibraryItem.AudioChannels = fileMeta.AudioTracks[0].Channels
	}

	// If we found a match on TMDb, populate more fields.
	if tmdbMovie != nil {
		movie.Title = tmdbMovie.Title
		movie.OriginalTitle = tmdbMovie.OriginalTitle
		movie.TMDbID = tmdbMovie.ID
		movie.Overview = tmdbMovie.Overview
		if tmdbMovie.ReleaseDate != "" {
			releaseDate, err := time.Parse("2006-01-02", tmdbMovie.ReleaseDate)
			if err == nil {
				movie.ReleaseDate = releaseDate
			} else {
				s.appLogger.Warn().Err(err).Str("date_str", tmdbMovie.ReleaseDate).Msg("Failed to parse release date")
			}
		}
		if tmdbMovie.BackdropPath != nil {
			movie.BackdropPath = *tmdbMovie.BackdropPath
		}
		if tmdbMovie.PosterPath != nil {
			movie.PosterPath = *tmdbMovie.PosterPath
		}
		movie.VoteAverage = tmdbMovie.VoteAverage
		movie.VoteCount = tmdbMovie.VoteCount
	} else {
		movie.Title = movieInfo.Title
	}

	if err := s.movieRepo.Store(ctx, movie); err != nil {
		return fmt.Errorf("failed to store movie: %w", err)
	}

	return nil
}
