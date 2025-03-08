package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/samcharles93/cinea/internal/auth"
	"github.com/samcharles93/cinea/internal/service"
	"github.com/samcharles93/cinea/internal/services/metadata"
)

type MovieHandler interface {
	RegisterRoutes(r chi.Router)
	List(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Stream(w http.ResponseWriter, r *http.Request)
}

type movieHandler struct {
	movieService service.MediaService
	tmdb         *metadata.TMDbService
	jwtVerifier  *auth.JWTVerifier
}

func NewMovieHandler(movieService service.MediaService, tmdb *metadata.TMDbService, jwtVerifier *auth.JWTVerifier) MovieHandler {
	return &movieHandler{
		movieService: movieService,
		tmdb:         tmdb,
		jwtVerifier:  jwtVerifier,
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
	movies, err := h.movieService.GetAllMovies(r.Context())
	if err != nil {
		h.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, movies)
}

func (h *movieHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.writeJSONError(w, http.StatusBadRequest, errors.New("invalid ID format"))
		return
	}

	movie, err := h.movieService.GetMovieByID(r.Context(), uint(id))
	if err != nil {
		h.writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	if movie == nil {
		h.writeJSONError(w, http.StatusNotFound, errors.New("movie not found"))
		return
	}

	h.writeJSON(w, http.StatusOK, movie)
}

func (h *movieHandler) Stream(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement streaming logic
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *movieHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *movieHandler) writeJSONError(w http.ResponseWriter, status int, err error) {
	h.writeJSON(w, status, map[string]string{"error": err.Error()})
}
