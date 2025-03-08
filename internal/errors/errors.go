package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("resource not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrBadRequest    = errors.New("bad request")
	ErrInternal      = errors.New("internal server error")
	ErrAlreadyExists = errors.New("resource already exists")
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// NotFoundError returns a not found error with additional context
func NotFoundError(resource string) error {
	return fmt.Errorf("%s: %w", resource, ErrNotFound)
}

// UnauthorisedError returns an unauthorized error with additional context
func UnauthorizedError(reason string) error {
	return fmt.Errorf("%s: %w", reason, ErrUnauthorized)
}

// ForbiddenError
func ForbiddenError(resource string) error {
	return fmt.Errorf("%s: %w", resource, ErrForbidden)
}

// Is checks if the target error is contained in the error chain
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in the error chain that matches target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
