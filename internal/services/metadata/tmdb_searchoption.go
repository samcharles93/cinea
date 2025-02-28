package metadata

import (
	"net/url"
	"strconv"
)

// SearchOption is a function type that modifies URL parameters
type SearchOption func(*url.Values)

// Search options for both movies and TV shows
func WithPage(page int) SearchOption {
	return func(v *url.Values) {
		v.Set("page", strconv.Itoa(page))
	}
}

func WithRegion(region string) SearchOption {
	return func(v *url.Values) {
		v.Set("region", region)
	}
}

// Movie-specific search options
func WithPrimaryReleaseYear(year string) SearchOption {
	return func(v *url.Values) {
		v.Set("primary_release_year", year)
	}
}

func WithMovieYear(year string) SearchOption {
	return func(v *url.Values) {
		v.Set("year", year)
	}
}

// TV-specific search options
func WithFirstAirDateYear(year int) SearchOption {
	return func(v *url.Values) {
		if year >= 1000 && year <= 9999 {
			v.Set("first_air_date_year", strconv.Itoa(year))
		}
	}
}

func WithTVYear(year int) SearchOption {
	return func(v *url.Values) {
		if year >= 1000 && year <= 9999 {
			v.Set("year", strconv.Itoa(year))
		}
	}
}
