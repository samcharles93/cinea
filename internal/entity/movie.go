package entity

import (
	"time"
)

type Movie struct {
	LibraryItem
	Title         string `gorm:"not null"`
	OriginalTitle string
	TMDbID        int
	Overview      string
	ReleaseDate   time.Time
	Runtime       int
	BackdropPath  string
	PosterPath    string
	VoteAverage   float64
	VoteCount     int
	LastScanned   time.Time
}
