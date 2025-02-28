package web

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/samcharles93/cinea/internal/auth"
	"github.com/samcharles93/cinea/internal/entity"
)

type PageData struct {
	User       *entity.User
	Flashes    []string
	ActivePage string
	Data       interface{}
	Title      string
}

// Simplified notification component
func (s *webService) notificationComponent(flashes []string) template.HTML {
	if len(flashes) == 0 {
		return ""
	}

	tmpl, err := template.ParseFS(s.webFS, "web/templates/components/notification.html")
	if err != nil {
		s.appLogger.Error().Err(err).Msg("Error parsing notification component template")
		return template.HTML(`<div class="notification error">Error loading notifications</div>`)
	}

	data := map[string]interface{}{
		"Flashes": flashes,
	}

	var outputBuffer bytes.Buffer
	if err := tmpl.ExecuteTemplate(&outputBuffer, "notification", data); err != nil {
		s.appLogger.Error().Err(err).Msg("Error executing notification component template")
		return template.HTML(`<div class="notification error">Error displaying notifications</div>`)
	}

	return template.HTML(outputBuffer.String())
}

// Serve a page with a consistent layout
func (s *webService) servePage(w http.ResponseWriter, r *http.Request, pageName string, data interface{}, flashes ...string) {
	// Get user from context if available
	user, _ := auth.GetUserFromContext(r.Context())

	pageData := PageData{
		User:       user,
		Flashes:    flashes,
		ActivePage: pageName,
		Data:       data,
		Title:      pageName,
	}

	// Set more descriptive title based on page
	switch pageName {
	case "dashboard":
		pageData.Title = "Dashboard - Cinea"
	case "login":
		pageData.Title = "Login - Cinea"
	case "register":
		pageData.Title = "Register - Cinea"
	case "media":
		pageData.Title = "Media Browser - Cinea"
	case "media_details":
		pageData.Title = "Media Details - Cinea"
	case "users":
		pageData.Title = "User Manager - Cinea"
	case "server":
		pageData.Title = "Server Manager - Cinea"
	default:
		pageData.Title = "Cinea Media Server"
	}

	// Determine which template files to parse
	baseTemplateFiles := []string{
		"web/templates/layout.html",
		"web/templates/components/notification.html",
	}

	contentTemplateFile := ""

	switch pageName {
	case "dashboard":
		contentTemplateFile = "web/templates/dashboard.html"
	case "login":
		contentTemplateFile = "web/templates/auth/login.html"
	case "register":
		contentTemplateFile = "web/templates/auth/register.html"
	case "media":
		contentTemplateFile = "web/templates/media_browser.html"
	case "media_details":
		contentTemplateFile = "web/templates/media_details.html"
	case "users":
		contentTemplateFile = "web/templates/user_manager.html"
	case "server":
		contentTemplateFile = "web/templates/server_manager.html"
	default:
		// For anything else, return 404
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Page not found: %s", pageName)
		return
	}

	templateFiles := append(baseTemplateFiles, contentTemplateFile)

	// Parse templates
	tmpl, err := template.ParseFS(s.webFS, templateFiles...)
	if err != nil {
		s.appLogger.Error().Err(err).Strs("templates", templateFiles).Msg("Error parsing templates")
		http.Error(w, "Internal Server Error: Failed to load template", http.StatusInternalServerError)
		return
	}

	// Execute template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", pageData); err != nil {
		s.appLogger.Error().Err(err).Str("template", "layout").Msg("Error executing template")
		http.Error(w, "Internal Server Error: Failed to render page", http.StatusInternalServerError)
	}
}

// Generate JWT token
func (s *webService) generateToken(user *entity.User) (string, error) {
	tokenLifetime, err := time.ParseDuration(s.config.Auth.TokenLifetime)
	if err != nil {
		tokenLifetime = 24 * time.Hour
	}

	now := time.Now()

	claims := jwt.MapClaims{
		"jti": uuid.New().String(),
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"exp": now.Add(tokenLifetime).Unix(),
		"sub": strconv.FormatUint(uint64(user.ID), 10),
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	}

	_, tokenString, err := s.tokenAuth.Encode(claims)
	return tokenString, err
}
