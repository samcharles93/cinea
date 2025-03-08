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

	// Web routes
	webHandler.RegisterRoutes(r)
	//r.Get("/", webHandler.DashboardHandler)
	//r.Get("/login", webHandler.LoginHandler)
	//r.Post("/login", webHandler.LoginHandler)
	//r.Get("/register", webHandler.RegisterHandler)
	//r.Post("/register", webHandler.RegisterHandler)
	//r.Post("/logout", webHandler.LogoutHandler)
	//r.Get("/users", webHandler.UserManagerHandler)
	//r.Get("/media", webHandler.MediaBrowserHandler)
	//r.Get("/media/{id}", webHandler.MediaDetailsHandler)

	r.Group(func(r chi.Router) {
		r.Use(webHandler.JWTMiddleware)
		r.Get("/me", webHandler.GetCurrentUser)
	})

	return r
}
