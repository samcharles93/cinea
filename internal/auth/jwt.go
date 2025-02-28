package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/samcharles93/cinea/internal/entity"
	"gorm.io/gorm"
)

// contextKey is a custom type for context keys
type contextKey string

// userContextKey is the context key for storing the user
const userContextKey contextKey = "user"

// JWTVerifier is a middleware to verify JWTs and add user info to the context
type JWTVerifier struct {
	TokenAuth *jwtauth.JWTAuth
}

func NewJWTVerifier(tokenAuth *jwtauth.JWTAuth) *JWTVerifier {
	return &JWTVerifier{TokenAuth: tokenAuth}
}

// Verify is the JWT verification middleware.
func (j *JWTVerifier) Verify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Initialize context with the token and claims
		ctx := r.Context()
		token, claims, err := jwtauth.FromContext(ctx)

		if err != nil || token == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get user data from claims
		userData, ok := claims["user"].(map[string]interface{})
		if !ok {
			http.Error(w, "Invalid user data in token", http.StatusUnauthorized)
			return
		}

		userIDFloat, ok := userData["id"].(float64)
		if !ok {
			http.Error(w, "Invalid user ID in token", http.StatusInternalServerError)
			return
		}

		userID := uint(userIDFloat)
		username, _ := userData["username"].(string)
		roleStr, _ := userData["role"].(string)
		role := entity.UserRole(roleStr)
		email, _ := userData["email"].(string)

		user := &entity.User{
			Model:    gorm.Model{ID: userID},
			Username: username,
			Email:    email,
			Role:     role,
		}

		// Add the user to the context
		ctx = context.WithValue(ctx, userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext retrieves the user from the request context.
func GetUserFromContext(ctx context.Context) (*entity.User, error) {
	user, ok := ctx.Value(userContextKey).(*entity.User)
	if !ok || user == nil {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}
