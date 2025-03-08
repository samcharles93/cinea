package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/logger"
	"gorm.io/gorm"
)

type LibraryRepository interface {
	// Library Management
	CreateLibrary(ctx context.Context, lib *entity.Library) error
	GetLibrary(ctx context.Context, id uint) (*entity.Library, error)
	ListLibraries(ctx context.Context) ([]*entity.Library, error)
	UpdateLibrary(ctx context.Context, lib *entity.Library) error
	DeleteLibrary(ctx context.Context, id uint) error

	// Library Items
	AddItem(ctx context.Context, item *entity.LibraryItem) error
	GetItem(ctx context.Context, id uint) (*entity.LibraryItem, error)
	FindItemByPath(ctx context.Context, path string) (*entity.LibraryItem, error)
	ListItems(ctx context.Context, libraryID uint) ([]*entity.LibraryItem, error)
	UpdateItem(ctx context.Context, item *entity.LibraryItem) error
	DeleteItem(ctx context.Context, id uint) error

	GetStaleItems(ctx context.Context, threshold time.Duration) ([]*entity.LibraryItem, error)
	FindMissingItems(ctx context.Context, lib *entity.Library) ([]*entity.LibraryItem, error)
}

type libraryRepository struct {
	db        *gorm.DB
	appLogger logger.Logger
}

func NewLibraryRepository(db *gorm.DB, appLogger logger.Logger) LibraryRepository {
	return &libraryRepository{
		db:        db,
		appLogger: appLogger,
	}
}

// CreateLibrary implements repository.LibraryRepository.
func (r *libraryRepository) CreateLibrary(ctx context.Context, lib *entity.Library) error {
	result := r.db.WithContext(ctx).Create(lib)
	if result.Error != nil {
		return fmt.Errorf("failed to create library: %w", result.Error)
	}
	return nil
}

// DeleteLibrary implements repository.LibraryRepository.
func (r *libraryRepository) DeleteLibrary(ctx context.Context, id uint) error {
	// This will also delete associated LibraryPaths due to cascading deletes
	result := r.db.WithContext(ctx).Delete(&entity.Library{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete library: %w", result.Error)
	}
	return nil
}

// GetLibrary implements repository.LibraryRepository.
func (r *libraryRepository) GetLibrary(ctx context.Context, id uint) (*entity.Library, error) {
	var lib entity.Library
	result := r.db.WithContext(ctx).Preload("Paths").First(&lib, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get library: %w", result.Error)
	}
	return &lib, nil
}

// ListLibraries implements repository.LibraryRepository.
func (r *libraryRepository) ListLibraries(ctx context.Context) ([]*entity.Library, error) {
	var libraries []*entity.Library
	result := r.db.WithContext(ctx).Preload("Paths").Find(&libraries)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list libraries: %w", result.Error)
	}
	return libraries, nil
}

// UpdateLibrary implements repository.LibraryRepository.
func (r *libraryRepository) UpdateLibrary(ctx context.Context, lib *entity.Library) error {
	// Use Save to handle both new and existing records
	result := r.db.WithContext(ctx).Save(lib)
	if result.Error != nil {
		return fmt.Errorf("failed to update library: %w", result.Error)
	}
	return nil
}

// Library Item Management

// FindItemByPath implements repository.LibraryRepository.
func (r *libraryRepository) FindItemByPath(ctx context.Context, path string) (*entity.LibraryItem, error) {
	var item entity.LibraryItem
	result := r.db.WithContext(ctx).Where("file_path = ?", path).First(&item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Not found is not an error
		}
		return nil, fmt.Errorf("failed to find item by path: %w", result.Error)
	}
	return &item, nil
}

func (r *libraryRepository) ListItems(ctx context.Context, libraryID uint) ([]*entity.LibraryItem, error) {
	var items []*entity.LibraryItem
	result := r.db.WithContext(ctx).Where("library_id = ?", libraryID).Find(&items)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list library items: %w", result.Error)
	}
	return items, nil
}

func (r *libraryRepository) UpdateItem(ctx context.Context, item *entity.LibraryItem) error {
	result := r.db.WithContext(ctx).Save(item)
	if result.Error != nil {
		return fmt.Errorf("failed to update library item: %w", result.Error)
	}
	return nil
}

func (r *libraryRepository) GetItem(ctx context.Context, id uint) (*entity.LibraryItem, error) {
	var item entity.LibraryItem
	result := r.db.WithContext(ctx).First(&item, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get library item: %w", result.Error)
	}
	return &item, nil
}

func (r *libraryRepository) AddItem(ctx context.Context, item *entity.LibraryItem) error {
	result := r.db.WithContext(ctx).Create(item)
	if result.Error != nil {
		return fmt.Errorf("failed to add library item: %w", result.Error)
	}
	return nil
}

func (r *libraryRepository) DeleteItem(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entity.LibraryItem{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete library item: %w", result.Error)
	}
	return nil
}

func (r *libraryRepository) GetStaleItems(ctx context.Context, threshold time.Duration) ([]*entity.LibraryItem, error) {
	var items []*entity.LibraryItem
	result := r.db.WithContext(ctx).Where("last_scanned &lt; ?", time.Now().Add(-threshold)).Find(&items)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get stale items: %w", result.Error)
	}
	return items, nil
}

func (r *libraryRepository) FindMissingItems(ctx context.Context, lib *entity.Library) ([]*entity.LibraryItem, error) {
	var items []*entity.LibraryItem
	err := r.db.WithContext(ctx).
		Where("library_id = ? AND file_path NOT IN (?)", lib.ID, r.db.WithContext(ctx).Table("library_paths").Select("path").Where("library_id = ?", lib.ID)).
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get missing items: %w", err)
	}

	return items, nil
}
