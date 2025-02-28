package persistence

import (
	"github.com/samcharles93/cinea/internal/logger"
	"gorm.io/gorm"
)

type LibraryAccessRepo interface {
}

type libraryAccessRepo struct {
	db        *gorm.DB
	appLogger logger.Logger
}

func NewLibraryAccessRepo(db *gorm.DB, appLogger logger.Logger) LibraryAccessRepo {
	return &libraryAccessRepo{
		db:        db,
		appLogger: appLogger,
	}
}
