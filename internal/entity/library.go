package entity

import (
	"time"

	"gorm.io/gorm"
)

type LibraryType string

const (
	LibraryTypeMovie LibraryType = "movie"
	LibraryTypeTV    LibraryType = "tv"

	// Can be built upon:
	// LibraryTypeMusic LibraryType = "music"
	// LibraryTypeBook LibraryType = "book"
)

// Library is the media collection
type Library struct {
	gorm.Model
	Name        string      `gorm:"size:128;not null"`
	Type        LibraryType `gorm:"type:string;not null"`
	Description string

	Paths []LibraryPath `gorm:"foreignKey:LibraryID"`

	AutoScan     bool          `gorm:"default:true"`
	ScanInterval time.Duration `gorm:"default:12h"`
	LastScanned  time.Time

	Items []LibraryItem `gorm:"foreignKey:LibraryID"`
}

type LibraryPath struct {
	gorm.Model
	LibraryID uint   `gorm:"not null"`
	Path      string `gorm:"not null"`
	Enabled   bool   `gorm:"default:true"`
}

type LibraryItem struct {
	gorm.Model
	LibraryID uint      `gorm:"not null"`
	Library   Library   `gorm:"foreignKey:LibraryID"`
	DateAdded time.Time `gorm:"not null"`
	FilePath  string    `gorm:"not null"`

	Container        string
	Codec            string
	ResolutionWidth  int
	ResolutionHeight int
	AudioChannels    int
}
