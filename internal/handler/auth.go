package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/samcharles93/cinea/internal/auth"
	"github.com/samcharles93/cinea/internal/dto"
	"github.com/samcharles93/cinea/internal/service"
)

type AuthHandler interface {
	RegisterRoutes(r chi.Router)
	Login(w http.ResponseWriter, r *http.Request)
	GetCurrentUser(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

type authHandler struct {
	authSvc     service.AuthService
	jwtVerifier *auth.JWTVerifier
}

func NewAuthHandler(authSvc service.AuthService) AuthHandler {
	return &authHandler{
		authSvc:     authSvc,
		jwtVerifier: auth.NewJWTVerifier(),
	}
}

func (h *authHandler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/logout", h.Logout)
		r.With(h.jwtVerifier.Verify).Get("/me", h.GetCurrentUser)
	})
}

func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Missing credentials", http.StatusBadRequest)
		return
	}

	userDTO, err := h.authSvc.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userDTO)
}

func (h *authHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.authSvc.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userDTO := dto.UserToDTO(user)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userDTO)
}

func (h *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" || req.Email == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	user, err := h.authSvc.CreateUser(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode()
}

func (h *authHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Since we're using JWT, we just need to return success
	// The frontend should handle removing the token
	w.WriteHeader(http.StatusOK)
}
