package metadata

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type SeriesSearchResult struct {
	Page         int
	Results      []Series
	TotalPages   int
	TotalResults int
}

type Series struct {
	Adult            bool
	BackdropPath     *string
	GenreIDs         []int
	ID               uint
	OriginCountry    []string
	OriginalLanguage string
	OriginalName     string
	Overview         string
	Popularity       float64
	PosterPath       *string
	FirstAirDate     string
	Name             string
	VoteAverage      float64
	VoteCount        int

	Container string `db:"container"`
	Codec     string `db:"codec"`
}

func (s *TMDbService) SearchTV(ctx context.Context, query string, opts ...SearchOption) (*SeriesSearchResult, error) {
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

	fullURL := fmt.Sprintf("%s/search/tv?%s", s.baseURL, params.Encode())

	var result SeriesSearchResult
	if err := s.fetch(ctx, fullURL, &result); err != nil {
		return nil, fmt.Errorf("search TV show error: %w", err)
	}

	return &result, nil
}
