package handler

import (
	"encoding/json"
	"net/http"

	"github.com/samcharles93/cinea/internal/entity"
)

type WatchlistHandler interface {
	AddToWatchlist(w http.ResponseWriter, r *http.Request)
}

type watchlistHandler struct {
	authSvc      services.AuthService
	watchlistSvc services.WatchlistService
}

func NewWatchlistHandler(authSvc services.AuthService, watchlistSvc services.WatchlistService) WatchlistHandler {
	return &watchlistHandler{
		authSvc:      authSvc,
		watchlistSvc: watchlistSvc,
	}
}

func (h *watchlistHandler) AddToWatchlist(w http.ResponseWriter, r *http.Request) {
	userFromCtx, err := h.authSvc.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var item entity.Watchlist
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	item.UserID = userFromCtx.ID
	if err := h.watchlistSvc.AddToWatchlist(r.Context(), &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
