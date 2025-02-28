package ffmpeg

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/samcharles93/cinea/internal/logger"
)

type Service interface {
	Install() error
	SetPaths() error
	CheckInstallation() (bool, error)
	EnsureInstalled() error
	ExtractMetadata(ctx context.Context, filePath string) (*MediaMetadata, error)
	GetFFmpegPath() string
	GetFFprobePath() string
	RunFFmpeg(ctx context.Context, args []string) ([]byte, error)
	RunFFprobe(ctx context.Context, args []string) ([]byte, error)
}

type service struct {
	ffmpegPath  string
	ffprobePath string
	appLogger   logger.Logger
}

func NewFFMpegService(appLogger logger.Logger) (Service, error) {
	svc := &service{
		appLogger: appLogger,
	}

	if err := svc.Install(); err != nil {
		return nil, fmt.Errorf("failed to install ffmpeg: %w", err)
	}

	if err := svc.SetPaths(); err != nil {
		return nil, fmt.Errorf("failed to set ffmpeg paths: %w", err)
	}

	return svc, nil
}

// SetPaths initializes the paths to the FFmpeg and FFprobe binaries
func (s *service) SetPaths() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config directory: %w", err)
	}

	targetDir := filepath.Join(configDir, "ffmpeg")
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	s.ffmpegPath = filepath.Join(targetDir, "ffmpeg"+ext)
	s.ffprobePath = filepath.Join(targetDir, "ffprobe"+ext)

	return nil
}

// Install copies the FFmpeg and FFprobe binaries from the app's bin directory
// to the user's config directory
func (s *service) Install() error {
	s.appLogger.Info().Msg("Installing FFmpeg binaries")

	// Get the user's config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config dir: %w", err)
	}

	// Create the target directory
	targetDir := filepath.Join(configDir, "ffmpeg")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target ffmpeg directory: %w", err)
	}

	// Set the extension if Windows
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	// Determine the source directory based on architecture
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	exeDir := filepath.Dir(exePath)
	sourceDir := filepath.Join(exeDir, "bin", runtime.GOARCH, runtime.GOOS)

	// Copy binaries
	binaries := []string{"ffmpeg", "ffprobe"}
	for _, binaryName := range binaries {
		srcPath := filepath.Join(sourceDir, binaryName+ext)
		destPath := filepath.Join(targetDir, binaryName+ext)

		s.appLogger.Debug().
			Str("source", srcPath).
			Str("destination", destPath).
			Msg("Copying binary")

		if err := s.copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", binaryName, err)
		}
	}

	s.appLogger.Info().Msg("FFmpeg binaries installed successfully")
	return nil
}

// GetFFmpegPath returns the path to the FFmpeg binary
func (s *service) GetFFmpegPath() string {
	return s.ffmpegPath
}

// GetFFprobePath returns the path to the FFprobe binary
func (s *service) GetFFprobePath() string {
	return s.ffprobePath
}

// copyFile copies a file from src to dest, preserving file permissions
func (s *service) copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	info, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	return os.Chmod(dest, info.Mode())
}

// CheckInstallation verifies if FFmpeg is properly installed
func (s *service) CheckInstallation() (bool, error) {
	if _, err := os.Stat(s.ffmpegPath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error checking FFmpeg installation: %w", err)
	}

	if _, err := os.Stat(s.ffprobePath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error checking FFprobe installation: %w", err)
	}

	return true, nil
}

// EnsureInstalled checks if FFmpeg is installed and installs it if not
func (s *service) EnsureInstalled() error {
	installed, err := s.CheckInstallation()
	if err != nil {
		return err
	}

	if !installed {
		s.appLogger.Info().Msg("FFmpeg not found, installing...")
		return s.Install()
	}

	log.Println("FFmpeg is already installed")
	return nil
}
