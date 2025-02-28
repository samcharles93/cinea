package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/services/metadata"
)

func (s *service) processSeriesFile(ctx context.Context, lib *entity.Library, filePath string) error {
	// 1. Extract show name, season, episode from filename
	tvInfo := extractTVShowInfo(filePath)
	if tvInfo.Season == 0 || tvInfo.Episode == 0 {
		s.appLogger.Warn().Str("filepath", filePath).Msg("Could not extract TV show info from filename")
		return nil
	}

	// 2. Check if the *episode* already exists (by path).
	existingEpisode, err := s.episodeRepo.FindByPath(ctx, filePath)
	if err != nil {
		return fmt.Errorf("error checking for existing episode: %w", err)
	}
	if existingEpisode != nil {
		existingEpisode.LastScanned = time.Now()
		return s.episodeRepo.UpdateEpisode(ctx, existingEpisode)
	}

	// 3. Extract technical metadata
	fileMeta, err := s.mediaExtractor.Extract(ctx, filePath)
	if err != nil {
		s.appLogger.Warn().Err(err).Str("filepath", filePath).Msg("Failed to extract TV show metadata")
	}

	// 4. Search TMDb for the *show*.
	searchResult, err := s.tmdb.SearchTV(ctx, tvInfo.Title, metadata.WithPage(1))
	if err != nil {
		s.appLogger.Error().Err(err).Str("title", tvInfo.Title).Msg("TMDb search failed for TV show")
	}

	var tmdbShow *metadata.Series

	if searchResult != nil && len(searchResult.Results) > 0 {
		tmdbShow = &searchResult.Results[0]
		s.appLogger.Info().Str("title", tmdbShow.Name).Uint("tmdb_id", tmdbShow.ID).Msg("Found TV show on TMDb")
	} else {
		s.appLogger.Warn().Str("title", tvInfo.Title).Msg("No results found on TMDb for TV show")
	}

	// 5. Create/Update Series, Season, and Episode entities.

	// 5.1 Find or Create Series
	var series *entity.Series
	if tmdbShow != nil {
		series, err = s.seriesRepo.FindByID(ctx, tmdbShow.ID)
		if err != nil {
			return fmt.Errorf("error checking for existing series: %w", err)
		}
	}

	if series == nil {
		series = &entity.Series{
			LibraryItem: entity.LibraryItem{
				LibraryID: lib.ID,
				DateAdded: time.Now(),
			},
			Title:       tvInfo.Title,
			LastScanned: time.Now(),
		}
		if tmdbShow != nil {
			series.Title = tmdbShow.Name
			series.OriginalTitle = tmdbShow.OriginalName
			series.TMDbID = tmdbShow.ID
			series.Overview = tmdbShow.Overview
			if tmdbShow.FirstAirDate != "" {
				firstAirDate, _ := time.Parse("2006-01-02", tmdbShow.FirstAirDate)
				series.FirstAirDate = firstAirDate
			}
			if tmdbShow.BackdropPath != nil {
				series.BackdropPath = *tmdbShow.BackdropPath
			}
			if tmdbShow.PosterPath != nil {
				series.PosterPath = *tmdbShow.PosterPath
			}

			series.VoteAverage = tmdbShow.VoteAverage
			series.VoteCount = tmdbShow.VoteCount
		}
		if err := s.seriesRepo.Store(ctx, series); err != nil {
			return fmt.Errorf("failed to store series: %w", err)
		}
	} else {
		series.LastScanned = time.Now()
		s.seriesRepo.Update(ctx, series)
	}

	// 5.2 Find or Create Season
	season, err := s.seasonRepo.FindSeasonByNumber(ctx, series.ID, tvInfo.Season)
	if err != nil {
		return fmt.Errorf("error checking for existing season: %w", err)
	}

	if season == nil {
		season = &entity.Season{
			SeriesID:     series.ID,
			SeasonNumber: tvInfo.Season,
			LibraryItem: entity.LibraryItem{
				LibraryID: lib.ID,
				DateAdded: time.Now(),
			},
		}
		if err := s.seasonRepo.AddSeason(ctx, season); err != nil {
			return fmt.Errorf("failed to store season: %w", err)
		}
	} else {
		season.LastScanned = time.Now()
		s.seasonRepo.UpdateSeason(ctx, season)
	}

	// 5.3 Create Episode
	episode := &entity.Episode{
		LibraryItem: entity.LibraryItem{
			LibraryID:        lib.ID,
			DateAdded:        time.Now(),
			FilePath:         filePath,
			Container:        fileMeta.Container,
			Codec:            fileMeta.Codec,
			ResolutionWidth:  fileMeta.ResolutionWidth,
			ResolutionHeight: fileMeta.ResolutionHeight,
		},
		SeriesID:      series.ID,
		SeasonID:      season.ID,
		EpisodeNumber: tvInfo.Episode,
		Title:         fmt.Sprintf("Episode %d", tvInfo.Episode),
		LastScanned:   time.Now(),
	}
	if len(fileMeta.AudioTracks) > 0 {
		episode.LibraryItem.AudioChannels = fileMeta.AudioTracks[0].Channels
	}

	// TODO: Look into getting episode title/overview from TMDb.

	if err := s.episodeRepo.AddEpisode(ctx, episode); err != nil {
		return fmt.Errorf("failed to store episode: %w", err)
	}

	return nil
}
