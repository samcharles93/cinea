package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/samcharles93/cinea/config"
)

type Service struct {
	config *config.Config
}

func New(cfg *config.Config) *Service {
	return &Service{config: cfg}
}

func (s *Service) CreateBackup() (string, error) {
	timestamp := time.Now().Format(time.RFC3339)
	backupFile := filepath.Join(s.config.Backup.BackupDir, fmt.Sprintf("cinea-backup-%s.tar.gz", timestamp))

	// Create the backup file
	file, err := os.Create(backupFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create the gzip writer
	gw := gzip.NewWriter(file)
	defer gw.Close()

	// Create the tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Backup the database
	if err := s.backupDatabase(tw); err != nil {
		return "", err
	}

	// Backup the config
	if err := s.backupConfig(tw); err != nil {
		return "", err
	}

	// Backup the metadata
	if err := s.backupMetadata(tw); err != nil {
		return "", err
	}

	return backupFile, nil
}

func (s *Service) Restore(backupPath string) error {
	/*
		Implement Restore Fucntionality

		1. Verify the backup integrity
		2. Stop Services
		3. Restore the database
		4. Restore the config
		5. Restore the metadata
		6. Start the services
	*/
	return nil
}

func (s *Service) backupDatabase(tw *tar.Writer) error {
	return nil
}

func (s *Service) backupConfig(tw *tar.Writer) error {
	return nil
}

func (s *Service) backupMetadata(tw *tar.Writer) error {
	return nil
}
