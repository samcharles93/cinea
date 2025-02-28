package handler

import (
	"encoding/json"
	"net/http"

	"github.com/samcharles93/cinea/internal/entity"
)

type WatchHistoryHandler interface {
	AddToWatchHistory(w http.ResponseWriter, r *http.Request)
	ClearHistory(w http.ResponseWriter, r *http.Request)
}

type watchHistoryHandler struct {
	authSvc         services.AuthService
	watchHistorySvc services.WatchHistoryService
}

func NewWatchHistoryHandler(authSvc services.AuthService, watchHistoryService services.WatchHistoryService) WatchHistoryHandler {
	return &watchHistoryHandler{
		authSvc:         authSvc,
		watchHistorySvc: watchHistoryService,
	}
}

func (h *watchHistoryHandler) AddToWatchHistory(w http.ResponseWriter, r *http.Request) {
	userFromCtx, err := h.authSvc.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var item entity.WatchHistory
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	item.UserID = userFromCtx.ID
	if err := h.watchHistorySvc.AddToWatchHistory(r.Context(), &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *watchHistoryHandler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	userFromCtx, err := h.authSvc.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	history, err := h.watchHistorySvc.ClearHistory(r.Context(), userFromCtx.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
