package services

import (
	"context"
	"fmt"

	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/persistence"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Authenticate(ctx context.Context, username, password string) (*entity.User, error)
	CreateUser(ctx context.Context, username, email, password string) (*entity.User, error)
	ListUsers(ctx context.Context) ([]*entity.User, error)
}

type userService struct {
	userRepo persistence.UserRepository
}

func NewUserService(userRepo persistence.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// Authenticate
func (s *userService) Authenticate(ctx context.Context, username string, password string) (*entity.User, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("username or password is incorrect")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("username or password is incorrect")
	}

	return user, nil
}

// CreateUser
func (s *userService) CreateUser(ctx context.Context, username string, email string, password string) (*entity.User, error) {
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &entity.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Role:     entity.RoleUser,
	}

	if err := s.userRepo.Store(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return newUser, nil
}

// ListUsers
func (s *userService) ListUsers(ctx context.Context) ([]*entity.User, error) {
	return s.userRepo.AdminGetUsers()
}
