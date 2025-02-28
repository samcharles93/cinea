package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/samcharles93/cinea/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger interface {
	Debug() *zerolog.Event
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
	Fatal() *zerolog.Event
	WithError(err error) *zerolog.Event
	With() zerolog.Context
}

type logger struct {
	zlog zerolog.Logger
}

func NewLogger(cfg *config.Config) (Logger, error) {
	logDir, err := getLogDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to get log directory: %w", err)
	}

	fileLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "cinea.log"),
		MaxSize:    cfg.Logging.Rotation.MaxSize,
		MaxAge:     cfg.Logging.Rotation.MaxAge,
		MaxBackups: cfg.Logging.Rotation.MaxBackups,
		Compress:   true,
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	multi := zerolog.MultiLevelWriter(consoleWriter, fileLogger)

	level, err := zerolog.ParseLevel(strings.ToLower(cfg.Logging.Level))
	if err != nil {
		level = zerolog.ErrorLevel
	}
	zerolog.SetGlobalLevel(level)

	zlog := zerolog.New(multi).With().Timestamp().Caller().Logger()

	return &logger{zlog: zlog}, nil
}

func getLogDirectory() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	logDir := filepath.Join(configDir, "cinea", "logs")
	if err := os.MkdirAll(logDir, 0744); err != nil {
		return "", fmt.Errorf("failed to create log directory: %w", err)
	}

	return logDir, nil
}

func (l *logger) Debug() *zerolog.Event {
	return l.zlog.Debug()
}

func (l *logger) Info() *zerolog.Event {
	return l.zlog.Info()
}

func (l *logger) Warn() *zerolog.Event {
	return l.zlog.Warn()
}

func (l *logger) Error() *zerolog.Event {
	return l.zlog.Error()
}

func (l *logger) Fatal() *zerolog.Event {
	return l.zlog.Fatal()
}

func (l *logger) WithError(err error) *zerolog.Event {
	return l.zlog.Error().Err(err)
}

func (l *logger) With() zerolog.Context {
	return l.zlog.With()
}
