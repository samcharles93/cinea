package metadata

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type MovieSearchResult struct {
	Page         int     `json:"page"`
	Results      []Movie `json:"results"`
	TotalPages   int     `json:"total_pages"`
	TotalResults int     `json:"total_results"`
}

type Movie struct {
	Adult            bool    `json:"adult"`
	BackdropPath     *string `json:"backdrop_path"`
	GenreIDs         []int   `json:"genre_ids"`
	ID               int     `json:"id"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       *string `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`

	Container string `db:"container"`
	Codec     string `db:"codec"`
}

func (s *TMDbService) SearchMovie(ctx context.Context, query string, opts ...SearchOption) (*MovieSearchResult, error) {
	params := url.Values{}
	params.Add("api_key", s.config.Meta.TMDb.BearerToken)
	params.Add("query", query)
	params.Add("language", s.config.Meta.TMDb.Language)
	params.Add("include_adult", strconv.FormatBool(s.config.Meta.TMDb.IncludeAdult))
	params.Add("page", "1")

	// Apply any additional search options
	for _, opt := range opts {
		opt(&params)
	}

	fullURL := fmt.Sprintf("%s/search/movie?%s", s.baseURL, params.Encode())

	var result MovieSearchResult
	if err := s.fetch(ctx, fullURL, &result); err != nil {
		return nil, fmt.Errorf("search movie error: %w", err)
	}

	return &result, nil
}
