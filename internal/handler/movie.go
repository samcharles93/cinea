package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/samcharles93/cinea/internal/auth"
	"github.com/samcharles93/cinea/internal/persistence"
	"github.com/samcharles93/cinea/internal/services/metadata"
)

type MovieHandler interface {
	RegisterRoutes(r chi.Router)
	List(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Stream(w http.ResponseWriter, r *http.Request)
}

type movieHandler struct {
	movieRepo   persistence.MovieRepository
	tmdb        *metadata.TMDbService
	jwtVerifier *auth.JWTVerifier
}

func NewMovieHandler(movieRepo persistence.MovieRepository, tmdb *metadata.TMDbService, jwtVerifier *auth.JWTVerifier) MovieHandler {
	return &movieHandler{
		movieRepo:   movieRepo,
		tmdb:        tmdb,
		jwtVerifier: jwtVerifier,
	}
}

func (h *movieHandler) RegisterRoutes(r chi.Router) {
	r.Route("/movies", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(h.jwtVerifier.Verify)
			r.Get("/", h.List)
			r.Get("/{id}", h.Get)
			r.Get("/{id}/stream", h.Stream)
		})
	})
}

func (h *movieHandler) List(w http.ResponseWriter, r *http.Request) {
	movies, err := h.movieRepo.FindAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

func (h *movieHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	id := uint(id64)
	movie, err := h.movieRepo.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if movie == nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

func (h *movieHandler) Stream(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement streaming logic
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
