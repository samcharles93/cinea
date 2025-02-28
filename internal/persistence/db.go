package persistence

import (
	"fmt"

	"github.com/samcharles93/cinea/config"
	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/logger"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDB(cfg *config.Config, appLogger logger.Logger) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch cfg.DB.Driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DB.SQLite.Path), &gorm.Config{})
		if err != nil {
			appLogger.Error().
				Err(err).
				Str("database_driver", "sqlite").
				Msg("Failed to connect to SQLite Database")
			return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
		}

		if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
			appLogger.Error().
				Err(err).
				Str("database_driver", "sqlite").
				Msg("Failed to enable foreign keys for SQLite")
			return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
		}
	case "postgres":
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?SSLMode=%s",
			cfg.DB.Postgres.User,
			cfg.DB.Postgres.Password,
			cfg.DB.Postgres.Host,
			cfg.DB.Postgres.Port,
			cfg.DB.Postgres.DBName,
			cfg.DB.Postgres.SSLMode,
		)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			appLogger.Error().
				Err(err).
				Str("step", fmt.Sprintf("Initialise DB Driver: %s", cfg.DB.Driver)).
				Msg("Failed to connect to PostgreSQL Database")
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
	case "mariadb", "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DB.MariaDB.User,
			cfg.DB.MariaDB.Password,
			cfg.DB.MariaDB.Host,
			cfg.DB.MariaDB.Port,
			cfg.DB.MariaDB.DBName,
		)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			appLogger.Error().
				Err(err).
				Str("step", fmt.Sprintf("Initialise DB Driver: %s", cfg.DB.Driver)).
				Msg("Failed to connect to MariaDB/MySQL Database")
			return nil, fmt.Errorf("failed to connect to MariaDB/MySQL: %w", err)
		}

	default:
		appLogger.Error().
			Err(err).
			Str("step", fmt.Sprintf("Initialise DB Driver: %s", cfg.DB.Driver)).
			Msg("Unsupported database driver")
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.DB.Driver)
	}

	// Perform database auto-migration
	if err := db.AutoMigrate(
		&entity.User{},
		&entity.Library{},
		&entity.LibraryItem{},
		&entity.LibraryPath{},
		&entity.LibraryAccess{},
		&entity.Movie{},
		&entity.Series{},
		&entity.Season{},
		&entity.Episode{},
		&entity.ScheduledTask{},
		&entity.WatchHistory{},
		&entity.Watchlist{},
		&entity.Favorite{},
		&entity.Rating{},
	); err != nil {
		appLogger.Error().
			Err(err).
			Str("step", "auto-migrate").
			Msg("Failed to migrate database")
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	appLogger.Info().
		Msgf("Successfully connected to and migrated %s database", cfg.DB.Driver)
	return db, nil
}
