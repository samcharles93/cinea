package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/samcharles93/cinea/internal/auth"
	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/web"
	"golang.org/x/crypto/bcrypt"
)

type WebHandler interface {
}

type webHandler struct {
	webSvc web.WebService
}

func NewWebHandler(webSvc web.WebService) WebHandler {
	return &webHandler{webSvc: webSvc}
}

// GetCurrentUser returns the current user information
func (s *webHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userData := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userData)
		return
	}

	s.servePage(w, r, "dashboard", userData)
}

func (h *webHandler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	h.servePage(w, r, "dashboard", nil)
}

func (h *webHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.servePage(w, r, "login", nil)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		h.servePage(w, r, "login", nil, "Missing credentials")
		return
	}

	user, err := s.userRepo.FindByUsername(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		h.servePage(w, r, "login", nil, "Invalid credentials")
		return
	}

	tokenString, err := s.generateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := AuthResponse{
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

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	s.servePage(w, r, "dashboard", nil, "Logged In")
}

// RegisterHandler handles the registration page and registration requests
func (h *webHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.servePage(w, r, "register", nil)
		return
	}

	// Process registration (POST)
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" || req.Email == "" {
		h.servePage(w, r, "register", nil, "Missing required fields")
		return
	}

	// Check if user already exists
	existingUser, err := h.userRepo.FindByUsername(r.Context(), req.Username)
	if err != nil {
		h.appLogger.Error().Err(err).Str("username", req.Username).Msg("Error checking for existing user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existingUser != nil {
		h.servePage(w, r, "register", nil, "Username already taken")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.appLogger.Error().Err(err).Msg("Error hashing password")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	user := &entity.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     entity.RoleUser,
	}

	// Store the user in the database
	if err := h.userRepo.Store(r.Context(), user); err != nil {
		h.appLogger.Error().Err(err).Msg("Failed to create user")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate token
	tokenString, err := s.generateToken(user)
	if err != nil {
		h.appLogger.Error().Err(err).Str("username", user.Username).Msg("Failed to generate token")
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		resp := AuthResponse{
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// For regular requests, set cookie and redirect to dashboard
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		MaxAge:   86400, // 1 day
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// LogoutHandler handles logout requests
func (h *webHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// For regular requests, redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// UserManagerHandler displays the user management page
func (s *webHandler) UserManagerHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserFromContext(r.Context())
	if err != nil || user.Role != entity.RoleAdmin {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	users, err := s.userRepo.AdminGetUsers()
	if err != nil {
		s.appLogger.Error().Err(err).Msg("Failed to get users")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.servePage(w, r, "users", users)
}

// MediaBrowserHandler displays the media browser page
func (s *webHandler) MediaBrowserHandler(w http.ResponseWriter, r *http.Request) {
	// Get movies
	movies, err := s.movieRepo.FindAll(r.Context())
	if err != nil {
		s.appLogger.Error().Err(err).Msg("Failed to get movies")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get TV shows
	tvShows, err := s.seriesRepo.FindAll(r.Context())
	if err != nil {
		s.appLogger.Error().Err(err).Msg("Failed to get TV shows")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Convert to media items
	mediaItems := []MediaItem{}
	for _, m := range movies {
		mediaItems = append(mediaItems, web.MediaItem{
			ID:        m.ID,
			Title:     m.Title,
			Type:      "movie",
			Overview:  m.Overview,
			PosterURL: m.PosterPath,
		})
	}

	for _, s := range tvShows {
		mediaItems = append(mediaItems, web.MediaItem{
			ID:        s.ID,
			Title:     s.Title,
			Type:      "tvshow",
			Overview:  s.Overview,
			PosterURL: s.PosterPath,
		})
	}

	h.servePage(w, r, "media", mediaItems)
}

// MediaDetailsHandler displays the details of a specific media item
func (h *webHandler) MediaDetailsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the media ID from the URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	// Try to find a movie first
	movie, err := h.movieRepo.FindByID(r.Context(), uint(id))
	if err != nil {
		h.appLogger.Error().Err(err).Uint64("id", id).Msg("Error finding movie")
	}

	if movie != nil {
		// It's a movie
		mediaItem := web.MediaItem{
			ID:        movie.ID,
			Title:     movie.Title,
			Type:      "movie",
			Overview:  movie.Overview,
			PosterURL: movie.PosterPath,
		}
		h.servePage(w, r, "media_details", mediaItem)
		return
	}

	// If it's not a movie, try to find a TV show
	tvShow, err := h.seriesRepo.FindByID(r.Context(), uint(id))
	if err != nil {
		s.appLogger.Error().Err(err).Uint64("id", id).Msg("Error finding TV show")
	}

	if tvShow != nil {
		// It's a TV show
		mediaItem := web.MediaItem{
			ID:        tvShow.ID,
			Title:     tvShow.Title,
			Type:      "tvshow",
			Overview:  tvShow.Overview,
			PosterURL: tvShow.PosterPath,
		}
		h.servePage(w, r, "media_details", mediaItem)
		return
	}

	// If we didn't find either, return a 404
	http.Error(w, "Media not found", http.StatusNotFound)
}
