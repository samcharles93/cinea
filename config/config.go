package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`

	Logging struct {
		Level    string `yaml:"level"`
		Rotation struct {
			MaxSize    int `yaml:"max_size"`
			MaxAge     int `yaml:"max_age"`
			MaxBackups int `yaml:"max_backups"`
		} `yaml:"rotation"`
	} `yaml:"logging"`

	Auth struct {
		JWTSecret     string `yaml:"jwt_secret"`
		TokenLifetime string `yaml:"token_lifetime"`
	} `yaml:"auth"`

	DB struct {
		Driver string `yaml:"driver"`
		SQLite struct {
			Path string `yaml:"path"`
		} `yaml:"sqlite"`
		Postgres struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			DBName   string `yaml:"dbname"`
			SSLMode  string `yaml:"sslMode"`
		} `yaml:"postgres"`
		MariaDB struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			DBName   string `yaml:"dbname"`
		} `yaml:"mariadb"`
	}

	Meta struct {
		TMDb struct {
			BearerToken  string
			Language     string
			IncludeAdult bool
		} `yaml:"tmdb"`
	} `yaml:"meta"`

	Backup struct {
		Enabled    bool
		AutoBackup bool
		BackupDir  string
		Interval   string
	}

	Jobs struct {
		Cleanup struct {
			Enabled        bool   // Enable the scheduled cleanup job
			DeleteOrphaned bool   // Cleanup files which don't have corresponding database entries
			DeleteMissing  bool   // Cleanup database entries where files have been deleted and the database entries remain
			CleanInterval  string // String representation of cleanup duration (e.g. 2d = 2 days, 36h = 36 hours, 30d = 30 days)
		} `yaml:"cleanup"`
		Scanner struct {
			MovieDirs    []string // List of directories to search for movies
			SeriesDirs   []string // List of directories to search for tv shows
			AutoScan     bool     // Enable autoscan to periodically scan directories at specified intervals
			WatchDirs    bool     // Can be used with or without autoscan, will watch the media directories for changes and import any new media
			ScanInterval string   // Specify the intervals the autoscan runs (e.g. 2d = 2 days, 36h = 36 hours, 30d = 30 days)
		} `yaml:"scanner"`
	} `yaml:"jobs"`
}

func Load(path string) (*Config, error) {
	config := &Config{}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)

	const maxRetries = 3
	const retryDelay = 100 * time.Millisecond

	var decodeErr error
	for i := 0; i < maxRetries; i++ {
		decodeErr = decoder.Decode(config)
		if decodeErr == nil {
			break
		}
		log.Printf("Failed to decode config file: %v, retrying in %v", decodeErr, retryDelay)
		time.Sleep(retryDelay)
	}

	if decodeErr != nil {
		return nil, decodeErr
	}

	// Expand environment variables in paths
	config.DB.SQLite.Path = os.ExpandEnv(config.DB.SQLite.Path)
	config.Backup.BackupDir = os.ExpandEnv(config.Backup.BackupDir)

	// Expand movie directories
	for i, dir := range config.Jobs.Scanner.MovieDirs {
		config.Jobs.Scanner.MovieDirs[i] = os.ExpandEnv(dir)
	}

	// Expand TV show directories
	for i, dir := range config.Jobs.Scanner.SeriesDirs {
		config.Jobs.Scanner.SeriesDirs[i] = os.ExpandEnv(dir)
	}

	return config, nil
}
