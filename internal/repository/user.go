package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository interface {
	Store(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uint) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]*entity.User, error)
	UpdateLastLogin(ctx context.Context, id uint) error
}

type userRepository struct {
	db        *gorm.DB
	appLogger logger.Logger
}

func NewUserRepository(db *gorm.DB, appLogger logger.Logger) UserRepository {
	return &userRepository{
		db:        db,
		appLogger: appLogger,
	}
}

func (r *userRepository) Store(ctx context.Context, user *entity.User) error {
	// Log the intention *before* the operation.
	r.appLogger.Debug().
		Str("username", user.Username).
		Msg("Storing user")

	start := time.Now()
	result := r.db.WithContext(ctx).Clauses(clause.Returning{}).Create(user)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Msgf("Failed to store user: %v", result.Error)
		return fmt.Errorf("failed to store user: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", user.ID).
		Dur("duration", duration).
		Msg("User stored successfully")
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*entity.User, error) {
	r.appLogger.Debug().
		Uint("userID", id).
		Msg("Finding user by ID")

	var user entity.User
	start := time.Now()
	result := r.db.WithContext(ctx).First(&user, id)
	duration := time.Since(start)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Log at Debug level for "not found" - it's not an application error.
			r.appLogger.Debug().
				Uint("userID", id).
				Dur("duration", duration).
				Msg("User not found")
			return nil, nil
		}
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msgf("Failed to find user by ID: %v", result.Error)
		return nil, fmt.Errorf("failed to find user by ID: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", id).
		Dur("duration", duration).
		Msg("User found by ID")
	return &user, nil
}

// Add similar logging to ALL repository methods: FindByUsername, FindByEmail, Update, Delete, etc.
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	r.appLogger.Debug().
		Str("username", username).
		Msg("Finding user by username")

	var user entity.User
	start := time.Now()
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&user)
	duration := time.Since(start)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.appLogger.Debug().
				Str("username", username).
				Dur("duration", duration).
				Msg("User not found")
			return nil, nil // User not found
		}
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to find user by username")
		return nil, fmt.Errorf("failed to find user by username: %w", result.Error)
	}

	r.appLogger.Info().
		Str("username", username).
		Dur("duration", duration).
		Msg("User found by username")
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	r.appLogger.Debug().
		Str("email", email).
		Msg("Finding user by email")

	var user entity.User
	start := time.Now()
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	duration := time.Since(start)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.appLogger.Debug().
				Str("email", email).
				Dur("duration", duration).
				Msg("User not found")
			return nil, nil
		}
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to find user by email")
		return nil, fmt.Errorf("failed to find user by email: %w", result.Error)
	}

	r.appLogger.Info().
		Str("email", email).
		Dur("duration", duration).
		Msg("User found by email")
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	r.appLogger.Debug().
		Uint("userID", user.ID).
		Msg("Updating user")

	start := time.Now()
	result := r.db.WithContext(ctx).Save(user)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to update user")
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", user.ID).
		Dur("duration", duration).
		Msg("User updated successfully")
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	r.appLogger.Debug().
		Uint("userID", id).
		Msg("Deleting user")

	start := time.Now()
	result := r.db.WithContext(ctx).Delete(&entity.User{}, id)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", id).
		Dur("duration", duration).
		Msg("User deleted successfully")
	return nil
}

func (r *userRepository) List(ctx context.Context) ([]*entity.User, error) {
	r.appLogger.Debug().
		Msg("Getting all users (admin)")

	var users []*entity.User
	start := time.Now()
	result := r.db.Find(&users)
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to get all users")
		return nil, fmt.Errorf("failed to get all users: %w", result.Error)
	}

	r.appLogger.Info().
		Int("userCount", len(users)).
		Dur("duration", duration).
		Msg("All users retrieved successfully")
	return users, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id uint) error {
	r.appLogger.Debug().
		Uint("userID", id).
		Msg("Updating last login for user")

	start := time.Now()
	result := r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("last_login", gorm.Expr("CURRENT_TIMESTAMP"))
	duration := time.Since(start)

	if result.Error != nil {
		r.appLogger.Error().
			Err(result.Error).
			Str("sql", result.Statement.SQL.String()).
			Any("args", result.Statement.Vars).
			Dur("duration", duration).
			Msg("Failed to update last login")
		return fmt.Errorf("failed to update last login: %w", result.Error)
	}

	r.appLogger.Info().
		Uint("userID", id).
		Dur("duration", duration).
		Msg("Last login updated successfully")
	return nil
}
