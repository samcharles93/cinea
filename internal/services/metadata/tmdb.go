package metadata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/samcharles93/cinea/config"
)

type TMDbService struct {
	config    *config.Config
	client    *http.Client
	baseURL   string
	sessionID string
}

type SessionRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	RequestToken string `json:"request_token"`
}

func NewTMDbService(cfg *config.Config) *TMDbService {
	return &TMDbService{
		config:  cfg,
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: "https://api.themoviedb.org/3",
	}
}

func (s *TMDbService) fetch(ctx context.Context, url string, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.config.Meta.TMDb.BearerToken))
	req.Header.Add("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var tmdbError struct {
			StatusMessage string `json:"status_message"`
			StatusCode    int    `json:"status_code"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tmdbError); err != nil {
			return fmt.Errorf("TMDb API error: %s", resp.Status)
		}
		return fmt.Errorf("TMDb API error: %s (code: %d)", tmdbError.StatusMessage, tmdbError.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

func (s *TMDbService) createUserSession(ctx context.Context, username, password string) error {
	var tokenResp struct {
		Success      bool   `json:"success"`
		RequestToken string `json:"request_token"`
	}

	if err := s.fetch(ctx, fmt.Sprintf("%s/authentication/token/new", s.baseURL), &tokenResp); err != nil {
		return fmt.Errorf("failed to get request token: %w", err)
	}

	loginURL := fmt.Sprintf("%s/authentication/token/validate_with_login", s.baseURL)
	loginReq := SessionRequest{
		Username:     username,
		Password:     password,
		RequestToken: tokenResp.RequestToken,
	}

	var loginResp struct {
		Success bool `json:"success"`
	}

	if err := s.postJSON(ctx, loginURL, loginReq, &loginResp); err != nil {
		return fmt.Errorf("failed to validate login: %w", err)
	}

	var sessionResp struct {
		Success   bool   `json:"success"`
		SessionID string `json:"session_id"`
	}

	if err := s.postJSON(ctx,
		fmt.Sprintf("%s/authentication/session/new", s.baseURL),
		map[string]string{"request_token": tokenResp.RequestToken},
		&sessionResp,
	); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	s.sessionID = sessionResp.SessionID
	return nil
}

// Helper method for POST requests
func (s *TMDbService) postJSON(ctx context.Context, url string, body, response interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.config.Meta.TMDb.BearerToken))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var tmdbError struct {
			StatusMessage string `json:"status_message"`
			StatusCode    int    `json:"status_code"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tmdbError); err != nil {
			return fmt.Errorf("TMDb API error: %s", resp.Status)
		}
		return fmt.Errorf("TMDb API error: %s (code: %d)", tmdbError.StatusMessage, tmdbError.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(response)
}
