package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/samcharles93/cinea/internal/dto"
	"github.com/samcharles93/cinea/internal/services"
)

type FavoriteHandler interface {
	GetFavorites(w http.ResponseWriter, r *http.Request)
	AddToFavorites(w http.ResponseWriter, r *http.Request)
	RemoveFromFavorites(w http.ResponseWriter, r *http.Request)
}

type favoriteHandler struct {
	authSvc     services.AuthService
	favoriteSvc services.FavoriteService
}

func NewFavoriteHandler(authSvc services.AuthService, favoriteSvc services.FavoriteService) FavoriteHandler {
	return &favoriteHandler{
		authSvc:     authSvc,
		favoriteSvc: favoriteSvc,
	}
}

func (h *favoriteHandler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	userFromCtx, err := h.authSvc.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	favorites, err := h.favoriteSvc.GetFavorites(r.Context(), userFromCtx.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}

func (h *favoriteHandler) AddToFavorites(w http.ResponseWriter, r *http.Request) {
	userFromCtx, err := h.authSvc.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var item dto.FavoriteDTO
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	item.UserID = userFromCtx.ID
	if err := h.favoriteSvc.AddToFavorites(r.Context(), &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *favoriteHandler) RemoveFromFavorites(w http.ResponseWriter, r *http.Request) {
	userFromCtx, err := h.authSvc.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	favoriteIdParam := chi.URLParam(r, "favoriteId")
	favoriteId, err := strconv.Atoi(favoriteIdParam)
	if err != nil {
		http.Error(w, "Invalid favorite ID", http.StatusBadRequest)
		return
	}

	if err := h.favoriteSvc.RemoveFromFavorites(r.Context(), userFromCtx.ID, uint(favoriteId), ""); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
