package httputil

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrValidation    = errors.New("validation error")
	ErrRateLimited   = errors.New("rate limited")
	ErrInternal      = errors.New("internal error")
)

type AppError struct {
	Err     error
	Message string
	Code    string
	Details map[string]any
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NotFound(resource string) *AppError {
	return &AppError{
		Err:     ErrNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Code:    "NOT_FOUND",
	}
}

func AlreadyExists(resource string) *AppError {
	return &AppError{
		Err:     ErrAlreadyExists,
		Message: fmt.Sprintf("%s already exists", resource),
		Code:    "ALREADY_EXISTS",
	}
}

func Validation(field, msg string) *AppError {
	return &AppError{
		Err:     ErrValidation,
		Message: msg,
		Code:    "VALIDATION_ERROR",
		Details: map[string]any{"field": field},
	}
}

func Unauthorized(msg string) *AppError {
	return &AppError{
		Err:     ErrUnauthorized,
		Message: msg,
		Code:    "UNAUTHORIZED",
	}
}

func Forbidden(msg string) *AppError {
	return &AppError{
		Err:     ErrForbidden,
		Message: msg,
		Code:    "FORBIDDEN",
	}
}

func RateLimited() *AppError {
	return &AppError{
		Err:     ErrRateLimited,
		Message: "rate limit exceeded",
		Code:    "RATE_LIMITED",
	}
}

func Wrap(err error, msg string) *AppError {
	return &AppError{
		Err:     err,
		Message: msg,
		Code:    "INTERNAL_ERROR",
	}
}

func MapToHTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		err = appErr.Err
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, ErrRateLimited):
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
