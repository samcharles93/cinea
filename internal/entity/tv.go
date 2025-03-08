package entity

import (
	"time"
)

type Series struct {
	LibraryItem
	Title         string `gorm:"not null"`
	OriginalTitle string
	TMDbID        uint
	Overview      string
	FirstAirDate  time.Time
	BackdropPath  string
	PosterPath    string
	VoteAverage   float64
	VoteCount     int
	LastScanned   time.Time

	AirsDayOfWeek *time.Weekday
	AirsTime      *time.Time

	Seasons []Season `gorm:"foreignKey:SeriesID"`
}

func (s Series) SeasonCount() int {
	return len(s.Seasons)
}

type Season struct {
	LibraryItem
	SeriesID     uint   `gorm:"not null"`
	Series       Series `gorm:"foreignKey:SeriesID"`
	SeasonNumber int    `gorm:"not null"`
	Overview     string
	AirDate      time.Time
	PosterPath   string
	LastScanned  time.Time

	Episodes []Episode `gorm:"foreignKey:SeasonID"`
}

func (s Season) EpisodeCount() int {
	return len(s.Episodes)
}

type Episode struct {
	LibraryItem
	SeriesID      uint   `gorm:"not null"`
	Series        Series `gorm:"foreignKey:SeriesID"`
	SeasonID      uint   `gorm:"not null"`
	Season        Season `gorm:"foreignKey:SeasonID"`
	EpisodeNumber int    `gorm:"not null"`
	Title         string `gorm:"not null"`
	Overview      string
	AirDate       time.Time
	StillPath     string
	LastScanned   time.Time
}
