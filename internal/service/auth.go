package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/samcharles93/cinea/config"
	"github.com/samcharles93/cinea/internal/dto"
	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/logger"
	"github.com/samcharles93/cinea/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	GenerateToken(user *entity.User) (string, error)
	GetUserFromContext(ctx context.Context) (*entity.User, error)
	IsAdmin(ctx context.Context) bool
	IsAuthenticated(ctx context.Context) bool
	Authenticate(ctx context.Context, username, password string) (*dto.AuthResponse, error)
	CreateUser(ctx context.Context, username, email, password string) (*dto.AuthResponse, error)
	ListUsers(ctx context.Context) ([]*entity.User, error)
}

type authService struct {
	config    *config.Config
	appLogger logger.Logger
	tokenAuth *jwtauth.JWTAuth
	userRepo  repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config, appLogger logger.Logger, tokenAuth *jwtauth.JWTAuth) AuthService {
	return &authService{
		tokenAuth: tokenAuth,
		userRepo:  userRepo,
		appLogger: appLogger,
		config:    cfg,
	}
}

func (s *authService) Authenticate(ctx context.Context, username, password string) (*dto.AuthResponse, error) {
	// Find user
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("username or password is incorrect")
	}

	// Compare hash and password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("username or password is incorrect")
	}

	// Generate user token
	tokenString, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		s.appLogger.Warn().Err(err).Msg("failed to update last login")
	}

	// Prepare response
	resp := dto.AuthResponse{
		Token: tokenString,
		User: struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
			Role     string `json:"role"`
		}{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     string(user.Role),
		},
	}

	return &resp, nil
}

func (s *authService) CreateUser(ctx context.Context, username, email, password string) (*dto.AuthResponse, error) {
	// Check user exists
	var existingUser *entity.User
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	newUser := &entity.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Role:     entity.RoleUser,
	}

	// Create the new user
	if err := s.userRepo.Store(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate Token
	tokenString, err := s.GenerateToken(newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Prepare response
	resp := dto.AuthResponse{
		Token: tokenString,
		User: struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
			Role     string `json:"role"`
		}{
			ID:       newUser.ID,
			Username: newUser.Username,
			Email:    newUser.Email,
			Role:     string(newUser.Role),
		},
	}

	return &resp, nil
}

func (s *authService) GenerateToken(user *entity.User) (string, error) {
	tokenAuth := jwtauth.New("HS256", []byte(s.config.Auth.JWTSecret), nil)
	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

func (s *authService) GetUserFromContext(ctx context.Context) (*dto.UserDTO, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from context: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, uint(claims["user"].(map[string]interface{})["id"].(float64)))
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return dto.UserToDTO(user), nil
}

func (s *authService) IsAdmin(ctx context.Context) bool {
	user, err := s.GetUserFromContext(ctx)
	if err != nil {
		return false
	}

	return user.Role == string(entity.RoleAdmin)
}

func (s *authService) IsAuthenticated(ctx context.Context) bool {
	_, _, err := jwtauth.FromContext(ctx)
	return err == nil
}

func (s *authService) ListUsers(ctx context.Context) ([]*entity.User, error) {
	// Only allow admins to list users
	user, err := s.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from context: %w", err)
	}
	if user.Role != string(entity.RoleAdmin) {
		return nil, fmt.Errorf("only admins can list users")
	}

	// List users
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}
