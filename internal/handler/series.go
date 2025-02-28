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

type SeriesHandler interface {
	RegisterRoutes(r chi.Router)
	List(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	ListEpisodes(w http.ResponseWriter, r *http.Request)
	StreamEpisode(w http.ResponseWriter, r *http.Request)
}

type seriesHandler struct {
	seriesRepo  persistence.SeriesRepository
	tmdb        *metadata.TMDbService
	jwtVerifier *auth.JWTVerifier
}

func NewSeriesHandler(seriesRepo persistence.SeriesRepository, tmdb *metadata.TMDbService, jwtVerifier *auth.JWTVerifier) SeriesHandler {
	return &seriesHandler{
		seriesRepo:  seriesRepo,
		tmdb:        tmdb,
		jwtVerifier: jwtVerifier,
	}
}

func (h *seriesHandler) RegisterRoutes(r chi.Router) {
	r.Route("/tv", func(r chi.Router) {
		// Protected Routes
		r.Group(func(r chi.Router) {
			r.Use(h.jwtVerifier.Verify)
			r.Get("/", h.List)
			r.Get("/{id}", h.Get)
			r.Get("/{id}/episodes", h.ListEpisodes)
			r.Get("/{id}/episodes/{episodeId}/stream", h.StreamEpisode)
		})
	})
}

func (h *seriesHandler) List(w http.ResponseWriter, r *http.Request) {
	shows, err := h.seriesRepo.FindAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shows)
}

func (h *seriesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id64, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid TV show ID", http.StatusBadRequest)
		return
	}

	id := uint(id64)
	show, err := h.seriesRepo.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if show == nil {
		http.Error(w, "TV show not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(show)
}

func (h *seriesHandler) ListEpisodes(w http.ResponseWriter, r *http.Request) {
	id64, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid TV show ID", http.StatusBadRequest)
		return
	}

	id := uint(id64)
	show, err := h.seriesRepo.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if show == nil {
		http.Error(w, "TV show not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(show.Seasons)
}

func (h *seriesHandler) StreamEpisode(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement streaming logic
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
