package entity

import (
	"time"

	"gorm.io/gorm"
)

type TaskStatus string

const (
	StatusIdle    TaskStatus = "idle"
	StatusRunning TaskStatus = "running"
	StatusFailed  TaskStatus = "failed"
)

type ScheduledTask struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Type        string `gorm:"not null"`
	Description string
	Enabled     bool   `gorm:"default:true"`
	Interval    string `gorm:"not null"`
	LastRun     time.Time
	NextRun     time.Time
	Status      TaskStatus
	Config      string
}
