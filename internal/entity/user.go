package entity

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null" `
	Email    string `gorm:"uniqueIndex;not null" `
	Password string `gorm:"not null" json:"-"`
	Name     string
	Role     UserRole `gorm:"type:string;default:'user'"`

	// Library Access
	LibraryAccess []LibraryAccess `gorm:"foreignKey:UserID"`

	IsActive        bool       `gorm:"default:true" `
	EmailVerified   bool       `gorm:"default:false"`
	LastLogin       *time.Time `json:"omitempty"`
	LastAccessToken string     `gorm:"-" json:"-"`

	// OAuth related fields
	OAuthProvider string `gorm:"default:''"`
	OAuthID       string `gorm:"default:''" json:"-"`

	// User preferences
	PreferredLanguage string `gorm:"default:'en-US'"`
	Theme             string `gorm:"default:'light'"`

	// Relationships
	WatchHistory []WatchHistory `gorm:"foreignKey:UserID" json:"-"`
	Watchlist    []Watchlist    `gorm:"foreignKey:UserID" json:"-"`
	Favorites    []Favorite     `gorm:"foreignKey:UserID" json:"-"`
	Ratings      []Rating       `gorm:"foreignKey:UserID" json:"-"`
}

// UserRole defines the type of user
type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
	RoleGuest UserRole = "guest"
)

type LibraryAccess struct {
	gorm.Model
	UserID    uint `gorm:"not null"`
	LibraryID uint `gorm:"not null"`
	CanManage bool `gorm:"default:false"`
}

// WatchHistory tracks what users have watched
type WatchHistory struct {
	gorm.Model
	UserID    uint      `gorm:"not null"`
	MediaType string    `gorm:"not null"`
	MediaID   uint      `gorm:"not null"`
	Progress  float64   `gorm:"default:0"`
	WatchedAt time.Time `gorm:"not null"`
}

// Watchlist tracks what users want to watch
type Watchlist struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	MediaType string `gorm:"not null"`
	MediaID   uint   `gorm:"not null"`
}

// Favorite tracks user's favorite content
type Favorite struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	MediaType string `gorm:"not null"`
	MediaID   uint   `gorm:"not null"`
}

// Rating stores user ratings for content
type Rating struct {
	gorm.Model
	UserID    uint    `gorm:"not null"`
	MediaType string  `gorm:"not null"`
	MediaID   uint    `gorm:"not null"`
	Score     float32 `gorm:"not null"`
	Review    string
}

// BeforeCreate hook for User model
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = RoleUser
	}
	return nil
}

// AfterFind hook for User model
func (u *User) AfterFind(tx *gorm.DB) error {
	return nil
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// UpdateLastLogin updates the user's last login timestamp
func (u *User) UpdateLastLogin(tx *gorm.DB) error {
	now := time.Now()
	u.LastLogin = &now
	return tx.Model(u).Update("last_seen", now).Error
}

// MarkEmailAsVerified marks the user's email as verified
func (u *User) MarkEmailAsVerified(tx *gorm.DB) error {
	u.EmailVerified = true
	return tx.Model(u).Update("email_verified", true).Error
}
