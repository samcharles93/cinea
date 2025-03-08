package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/samcharles93/cinea/config"
	"github.com/samcharles93/cinea/internal/auth"
	"github.com/samcharles93/cinea/internal/logger"
	"github.com/samcharles93/cinea/internal/service"
)

type WebService interface {
	JWTMiddleware(next http.Handler) http.Handler
	GetStaticFS() fs.FS

	DashboardHandler(w http.ResponseWriter, r *http.Request)
	LoginHandler(w http.ResponseWriter, r *http.Request)
	LogoutHandler(w http.ResponseWriter, r *http.Request)
	RegisterHandler(w http.ResponseWriter, r *http.Request)
	GetCurrentUser(w http.ResponseWriter, r *http.Request)
	UserManagerHandler(w http.ResponseWriter, r *http.Request)
	MediaBrowserHandler(w http.ResponseWriter, r *http.Request)
	MediaDetailsHandler(w http.ResponseWriter, r *http.Request)
}

type webService struct {
	webFS        embed.FS
	config       *config.Config
	appLogger    logger.Logger
	tokenAuth    *jwtauth.JWTAuth
	templates    *template.Template
	userService  service.UserService
	mediaService service.MediaService
	jwtVerifier  *auth.JWTVerifier
}

// NewWebService creates a new web service
func NewWebService(
	cfg *config.Config,
	appLogger logger.Logger,
	userService service.UserService,
	mediaService service.MediaService,
	tokenAuth *jwtauth.JWTAuth,
	webFS embed.FS,
) WebService {
	jwtVerifier := auth.NewJWTVerifier(tokenAuth)

	// Try to parse all templates
	tmpl, err := template.ParseFS(webFS, "web/templates/**/*.html")
	if err != nil {
		appLogger.Error().Err(err).Str("package", "web").Str("method", "NewWebService").Msg("Failed to parse templates")
		// Don't panic, but log the error
	}

	return &webService{
		config:       cfg,
		appLogger:    appLogger,
		tokenAuth:    tokenAuth,
		webFS:        webFS,
		templates:    tmpl,
		userService:  userService,
		mediaService: mediaService,
		jwtVerifier:  jwtVerifier,
	}
}

// JWTMiddleware applies the JWT verification middleware
func (s *webService) JWTMiddleware(next http.Handler) http.Handler {
	return s.jwtVerifier.Verify(next)
}

// GetStaticFS returns a filesystem with static files
func (s *webService) GetStaticFS() fs.FS {
	staticFS, err := fs.Sub(s.webFS, "web/static")
	if err != nil {
		// Log the error but return an empty FS
		s.appLogger.Error().Err(err).Msg("Failed to get static file system")
		return nil
	}
	return staticFS
}
