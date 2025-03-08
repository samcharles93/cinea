package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/samcharles93/cinea/internal/auth"
	"github.com/samcharles93/cinea/internal/repository"
	"github.com/samcharles93/cinea/internal/services/metadata"
)

type SeriesHandler interface {
	RegisterRoutes(r chi.Router)
	List(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	ListSeasons(w http.ResponseWriter, r *http.Request)
	GetSeason(w http.ResponseWriter, r *http.Request)
	ListEpisodes(w http.ResponseWriter, r *http.Request)
	GetEpisode(w http.ResponseWriter, r *http.Request)
	StreamEpisode(w http.ResponseWriter, r *http.Request)
}

type seriesHandler struct {
	mediaService service.MediaService
	tmdb         *metadata.TMDbService
	jwtVerifier  *auth.JWTVerifier
}

func NewSeriesHandler(mediaService service.MediaService, tmdb *metadata.TMDbService, jwtVerifier *auth.JWTVerifier) SeriesHandler {
	return &seriesHandler{
		mediaService: mediaService,
		tmdb:         tmdb,
		jwtVerifier:  jwtVerifier,
	}
}

func (h *seriesHandler) RegisterRoutes(r chi.Router) {
	r.Route("/tv", func(r chi.Router) {
		// Protected Routes
		r.Group(func(r chi.Router) {
			r.Use(h.jwtVerifier.Verify)
			r.Get("/", h.List)
			r.Get("/{id}", h.Get)
			r.Get("/{id}/seasons", h.ListSeasons)
			r.Get("/{id}/seasons/{seasonNumber}", h.GetSeason)
			r.Get("/{id}/seasons/{seasonNumber}/episodes", h.ListEpisodes)
			r.Get("/{id}/seasons/{seasonNumber}/episodes/{episodeNumber}", h.GetEpisode)
			r.Get("/{id}/episodes/{episodeId}/stream", h.StreamEpisode)
		})
	})
}

func (h *seriesHandler) List(w http.ResponseWriter, r *http.Request) {
	shows, err := h.mediaService.GetAllSeries(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shows)
}

func (h *seriesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid TV show ID", http.StatusBadRequest)
		return
	}

	show, err := h.mediaService.GetSeriesByID(r.Context(), id)
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

func (h *seriesHandler) ListSeasons(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid TV show ID", http.StatusBadRequest)
		return
	}

	seasons, err := h.mediaService.GetAllSeasons(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(seasons)
}

func (h *seriesHandler) GetSeason(w http.ResponseWriter, r *http.Request) {
	seriesID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid TV show ID", http.StatusBadRequest)
		return
	}

	seasonNumber, err := strconv.Atoi(chi.URLParam(r, "seasonNumber"))
	if err != nil {
		http.Error(w, "Invalid season number", http.StatusBadRequest)
		return
	}

	season, err := h.mediaService.GetSeasonByNumber(r.Context(), seriesID, seasonNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if season == nil {
		http.Error(w, "Season not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(season)
}
--
func (h *seriesHandler) ListEpisodes(w http.ResponseWriter, r *http.Request) {
	seriesID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid TV show ID", http.StatusBadRequest)
		return
	}

	seasonNumber, err := strconv.Atoi(chi.URLParam(r, "seasonNumber"))
	if err != nil {
		http.Error(w, "Invalid season number", http.StatusBadRequest)
		return
	}

	// First get the season to get its ID
	season, err := h.mediaService.GetSeasonByNumber(r.Context(), seriesID, seasonNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if season == nil {
		http.Error(w, "Season not found", http.StatusNotFound)
		return
	}

	episodes, err := h.mediaService.GetAllEpisodes(r.Context(), season.ID, seriesID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episodes)
}

func (h *seriesHandler) GetEpisode(w http.ResponseWriter, r *http.Request) {
	seriesID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid TV show ID", http.StatusBadRequest)
		return
	}

	seasonNumber, err := strconv.Atoi(chi.URLParam(r, "seasonNumber"))
	if err != nil {
		http.Error(w, "Invalid season number", http.StatusBadRequest)
		return
	}

	episodeNumber, err := strconv.Atoi(chi.URLParam(r, "episodeNumber"))
	if err != nil {
		http.Error(w, "Invalid episode number", http.StatusBadRequest)
		return
	}

	episode, err := h.mediaService.GetEpisodeByNumber(r.Context(), seriesID, seasonNumber, episodeNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if episode == nil {
		http.Error(w, "Episode not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episode)
}

func (h *seriesHandler) StreamEpisode(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement streaming logic
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// Helper function to parse ID parameters
func parseID(idStr string) (uint, error) {
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id64), nil
}
