package dto

import "github.com/samcharles93/cinea/internal/entity"

// SeriesDTO represents the basic information of a series
type SeriesDTO struct {
	ID           uint   `json:"id"`
	TMDbID       uint   `json:"tmdb_id"`
	Title        string `json:"title"`
	Overview     string `json:"overview"`
	BackdropPath string `json:"backdrop_path"`
	PosterPath   string `json:"poster_path"`
	SeasonCount  int    `json:"season_count"`

	// Seasons will be populated when converting from entity.Series
	Seasons []SeasonDTO `json:"seasons,omitempty"`
}

// SeasonDTO represents the basic information of a season
type SeasonDTO struct {
	ID           uint   `json:"id"`
	SeasonNumber int    `json:"season_number"`
	EpisodeCount int    `json:"episode_count"`
	AirDate      string `json:"air_date"`
	PosterPath   string `json:"poster_path"`

	// Episodes will be populated when needed for detail view
	Episodes []EpisodeDTO `json:"episodes,omitempty"`
}

// EpisodeDTO represents the basic information of an episode
type EpisodeDTO struct {
	ID            uint   `json:"id"`
	EpisodeNumber int    `json:"episode_number"`
	Title         string `json:"title"`
	Overview      string `json:"overview"`
	AirDate       string `json:"air_date"`
	StillPath     string `json:"still_path,omitempty"`
}

type CreateSeriesDTO struct {
	TMDbID uint   `json:"tmdb_id"`
	Title  string `json:"title"`
}

// SeriesToDTO converts an entity.Series to a SeriesDTO with basic season information
func SeriesToDTO(series *entity.Series) *SeriesDTO {
	if series == nil {
		return nil
	}

	seriesDTO := &SeriesDTO{
		ID:           series.ID,
		TMDbID:       series.TMDbID,
		Title:        series.Title,
		Overview:     series.Overview,
		BackdropPath: series.BackdropPath,
		PosterPath:   series.PosterPath,
		SeasonCount:  series.SeasonCount(),
		Seasons:      make([]SeasonDTO, 0, len(series.Seasons)),
	}

	// Add basic season information, without episodes
	for _, season := range series.Seasons {
		airDateStr := ""
		if !season.AirDate.IsZero() {
			airDateStr = season.AirDate.Format("2006-01-02")
		}

		seriesDTO.Seasons = append(seriesDTO.Seasons, SeasonDTO{
			ID:           season.ID,
			SeasonNumber: season.SeasonNumber,
			EpisodeCount: season.EpisodeCount(),
			AirDate:      airDateStr,
			PosterPath:   season.PosterPath,
			// Episodes will be empty here
		})
	}

	return seriesDTO
}

// SeriesToDTOs converts a slice of entity.Series to a slice of SeriesDTO
func SeriesToDTOs(series []*entity.Series) []*SeriesDTO {
	seriesDTOs := make([]*SeriesDTO, len(series))
	for i, s := range series {
		seriesDTOs[i] = SeriesToDTO(s)
	}
	return seriesDTOs
}

// GetSeriesWithDetails gets full details for a series, including episodes
func GetSeriesWithDetails(series *entity.Series) *SeriesDTO {
	if series == nil {
		return nil
	}

	// First get the basic series info
	seriesDTO := SeriesToDTO(series)

	// Now add episode details to each season
	for i, season := range series.Seasons {
		// Skip if already processed or season index out of range
		if i >= len(seriesDTO.Seasons) {
			continue
		}

		// Create episodes for this season
		episodes := make([]EpisodeDTO, 0, len(season.Episodes))
		for _, episode := range season.Episodes {
			airDateStr := ""
			if !episode.AirDate.IsZero() {
				airDateStr = episode.AirDate.Format("2006-01-02")
			}

			episodes = append(episodes, EpisodeDTO{
				ID:            episode.ID,
				EpisodeNumber: episode.EpisodeNumber,
				Title:         episode.Title,
				Overview:      episode.Overview,
				AirDate:       airDateStr,
				StillPath:     episode.StillPath,
			})
		}

		// Add episodes to the season
		seriesDTO.Seasons[i].Episodes = episodes
	}

	return seriesDTO
}

// GetSeasonWithDetails gets detailed information for a specific season
func GetSeasonWithDetails(season *entity.Season) *SeasonDTO {
	if season == nil {
		return nil
	}

	airDateStr := ""
	if !season.AirDate.IsZero() {
		airDateStr = season.AirDate.Format("2006-01-02")
	}

	seasonDTO := &SeasonDTO{
		ID:           season.ID,
		SeasonNumber: season.SeasonNumber,
		EpisodeCount: season.EpisodeCount(),
		AirDate:      airDateStr,
		PosterPath:   season.PosterPath,
		Episodes:     make([]EpisodeDTO, 0, len(season.Episodes)),
	}

	// Add episodes to the season
	for _, episode := range season.Episodes {
		airDateStr := ""
		if !episode.AirDate.IsZero() {
			airDateStr = episode.AirDate.Format("2006-01-02")
		}

		seasonDTO.Episodes = append(seasonDTO.Episodes, EpisodeDTO{
			ID:            episode.ID,
			EpisodeNumber: episode.EpisodeNumber,
			Title:         episode.Title,
			Overview:      episode.Overview,
			AirDate:       airDateStr,
			StillPath:     episode.StillPath,
		})
	}

	return seasonDTO
}

// GetEpisodeDetails gets detailed information for a specific episode
func GetEpisodeDetails(episode *entity.Episode) *EpisodeDTO {
	if episode == nil {
		return nil
	}

	airDateStr := ""
	if !episode.AirDate.IsZero() {
		airDateStr = episode.AirDate.Format("2006-01-02")
	}

	return &EpisodeDTO{
		ID:            episode.ID,
		EpisodeNumber: episode.EpisodeNumber,
		Title:         episode.Title,
		Overview:      episode.Overview,
		AirDate:       airDateStr,
		StillPath:     episode.StillPath,
	}
}
