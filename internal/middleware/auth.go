// internal/api/middleware/auth.go
package middleware

import (
	"context"
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

// Verify and Authenticate
func (j *JWTVerifier) Verify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, claims, err := jwtauth.FromContext(r.Context())

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

		user := &entity.User{
			Model:    gorm.Model{ID: userID},
			Username: username,
			Role:     role,
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) *entity.User {
	user, _ := ctx.Value(userContextKey).(*entity.User)
	return user
}
