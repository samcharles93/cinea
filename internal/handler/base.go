package handler

import (
	"encoding/json"
	"net/http"

	"github.com/samcharles93/cinea/internal/errors"
	"github.com/samcharles93/cinea/internal/logger"
)

type BaseHandler struct {
	logger logger.Logger
}

func NewBaseHandler(logger logger.Logger) *BaseHandler {
	return &BaseHandler{
		logger: logger,
	}
}

func (h *BaseHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Error().Err(err).Msg("Failed to encode JSON response")
		}
	}
}

func (h *BaseHandler) writeJSONError(w http.ResponseWriter, status int, err error) {
	resp := errors.ErrorResponse{
		Error: http.StatusText(status),
		Code:  status,
	}

	if err != nil {
		resp.Message = err.Error()

		switch status {
		case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
			h.logger.Debug().Err(err).Int("status", status).Msg("Client error")
		default:
			h.logger.Error().Err(err).Int("status", status).Msg("Server error")
		}
	}

	h.writeJSON(w, status, resp)
}

// HandleError determines the appropriate HTTP status code based on the error type
func (h *BaseHandler) HandleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errors.ErrNotFound):
		h.writeJSONError(w, http.StatusNotFound, err)
	case errors.Is(err, errors.ErrUnauthorized):
		h.writeJSONError(w, http.StatusUnauthorized, err)
	case errors.Is(err, errors.ErrForbidden):
		h.writeJSONError(w, http.StatusForbidden, err)
	case errors.Is(err, errors.ErrBadRequest):
		h.writeJSONError(w, http.StatusBadRequest, err)
	case errors.Is(err, errors.ErrAlreadyExists):
		h.writeJSONError(w, http.StatusConflict, err)
	default:
		h.writeJSONError(w, http.StatusInternalServerError, err)
	}
}
