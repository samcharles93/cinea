package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/samcharles93/cinea/config"
	"github.com/samcharles93/cinea/internal/handler"
)

func NewRouter(
	cfg *config.Config,
	movieHandler *handler.MovieHandler,
	seriesHandler *handler.SeriesHandler,
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
	webHandler *handler.WebHandler,
) *chi.Mux {
	r := chi.NewRouter()

	// Base middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Configure Cors
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8937"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "HX-Request", "HX-Target"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API routes
	r.Route("/api", func(r chi.Router) {
		authHandler.RegisterRoutes(r)
		movieHandler.RegisterRoutes(r)
		seriesHandler.RegisterRoutes(r)
		userHandler.RegisterRoutes(r)
	})

	webHandler.
		r.Get("/", webService.DashboardHandler)
	r.Get("/login", webService.LoginHandler)
	r.Post("/login", webService.LoginHandler)
	r.Get("/register", webService.RegisterHandler)
	r.Post("/register", webService.RegisterHandler)
	r.Post("/logout", webService.LogoutHandler)
	r.Get("/users", webService.UserManagerHandler)
	r.Get("/media", webService.MediaBrowserHandler)
	r.Get("/media/{id}", webService.MediaDetailsHandler)

	r.Group(func(r chi.Router) {
		r.Use(webService.JWTMiddleware)
		r.Get("/me", webService.GetCurrentUser)
	})

	return r
}
