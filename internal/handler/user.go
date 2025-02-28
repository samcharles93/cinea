package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/samcharles93/cinea/internal/auth"
	"github.com/samcharles93/cinea/internal/dto"
	"github.com/samcharles93/cinea/internal/entity"
)

type UserHandler interface {
	RegisterRoutes(r chi.Router)
	AdminGetUsers(w http.ResponseWriter, r *http.Request)
	AdminDeleteUser(w http.ResponseWriter, r *http.Request)

	UpdateLastSeen(w http.ResponseWriter, r *http.Request)
	UpdateUserProfile(w http.ResponseWriter, r *http.Request)

	AddToWatchHistory(w http.ResponseWriter, r *http.Request)
	ClearHistory(w http.ResponseWriter, r *http.Request)
}

type userHandler struct {
	authSvc services.AuthService
	userSvc services.UserService
}

func NewUserHandler(authSvc services.AuthService, userSvc services.UserService) UserHandler {
	return &userHandler{
		authSvc: authSvc,
		userSvc: userSvc,
	}
}

func (h *userHandler) RegisterRoutes(r chi.Router) {
	r.Route("/user", func(r chi.Router) {
		r.Use(h.jwtVerifier.Verify)

		r.Get("/", h.AdminGetUsers)
		r.Delete("/{userId}", h.AdminDeleteUser)
		// r.Post("/", h.AdminCreateUser)
		// r.Patch("/{userId}", h.AdminUpdateUser)
		// r.Post("/{userId}/roles", h.AdminUpdateUserRole)

		// r.Post("/verify/{verificationToken}", h.VerifyEmail)
		r.Patch("/{userId}", h.UpdateUserProfile)
		r.Post("/last-seen", h.UpdateLastSeen)

		r.Post("/watchlist", h.AddToWatchlist)
		// r.Delete("/watchlist/{watchlistId}", h.RemoveFromWatchlist)

		r.Post("/history", h.AddToWatchHistory)
		r.Delete("/history", h.ClearHistory)

		r.Get("/favorites", h.GetFavorites)
		r.Post("/favorites", h.AddToFavorites)
		r.Delete("/favorites/{favoriteId}", h.RemoveFromFavorites)

	})
}

func (h *userHandler) AdminGetUsers(w http.ResponseWriter, r *http.Request) {
	userFromCtx, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check user is an admin
	if string(userFromCtx.Role) != string(entity.RoleAdmin) {
		http.Error(w, "Insufficient access", http.StatusForbidden)
		return
	}

	users, err := h.userRepo.AdminGetUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userDTOs := make([]*dto.UserDTO, 0, len(users))
	for _, user := range users {
		userDTOs = append(userDTOs, dto.UserToDTO(user))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userDTOs)
}

func (h *userHandler) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	userFromCtx, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if string(userFromCtx.Role) != string(entity.RoleAdmin) {
		http.Error(w, "Insufficient access", http.StatusForbidden)
		return
	}

	userIdParam := chi.URLParam(r, "userId")
	userId, err := strconv.Atoi(userIdParam)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	if err := h.userRepo.Delete(r.Context(), uint(userId)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *userHandler) UpdateLastSeen(w http.ResponseWriter, r *http.Request) {
	userFromCtx, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.userRepo.UpdateLastLogin(r.Context(), userFromCtx.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *userHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement UpdateUserProfile
	w.WriteHeader(http.StatusNotImplemented)
}
