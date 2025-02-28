package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/samcharles93/cinea/internal/entity"
)

type RatingHandler interface {
	AddRating(w http.ResponseWriter, r *http.Request)
	RemoveRating(w http.ResponseWriter, r *http.Request)
}

type ratingHandler struct {
	authSvc   services.AuthService
	ratingSvc services.RatingService
}

func NewRatingHandler(r chi.Router, authSvc services.AuthService, ratingSvc services.RatingService) RatingHandler {
	hdl := &ratingHandler{
		authSvc:   authSvc,
		ratingSvc: ratingSvc,
	}

	r.Route("/user", func(r chi.Router) {
		r.Post("/ratings", hdl.AddRating)
		r.Delete("/likes/{id}", hdl.RemoveRating)
	})

	return hdl
}

func (h *ratingHandler) AddRating(w http.ResponseWriter, r *http.Request) {
	userFromCtx, err := h.authSvc.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var item entity.Rating
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	item.UserID = userFromCtx.ID
	if err := h.ratingSvc.AddRating(r.Context(), &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *ratingHandler) RemoveRating(w http.ResponseWriter, r *http.Request) {
	userFromCtx, err := h.authSvc.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ratingIdParam := chi.URLParam(r, "id")
	ratingId, err := strconv.Atoi(ratingIdParam)
	if err != nil {
		http.Error(w, "Invalid rating ID", http.StatusBadRequest)
		return
	}

	if err := h.ratingSvc.RemoveRating(r.Context(), userFromCtx.ID, uint(ratingId), ""); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
