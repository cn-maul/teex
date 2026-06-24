package apperr

import (
	"errors"
	"log/slog"
)

// AppError is a lightweight application error that carries an HTTP status code.
// Handlers use HTTPStatus() to extract the code and pass Message to the client.
type AppError struct {
	Code    int    // HTTP status code
	Message string // safe to show to the client
}

func (e *AppError) Error() string {
	return e.Message
}

// NotFound returns a 404 error.
func NotFound(msg string) *AppError {
	return &AppError{Code: 404, Message: msg}
}

// Conflict returns a 409 error.
func Conflict(msg string) *AppError {
	return &AppError{Code: 409, Message: msg}
}

// BadRequest returns a 400 error.
func BadRequest(msg string) *AppError {
	return &AppError{Code: 400, Message: msg}
}

// Unauthorized returns a 401 error.
func Unauthorized(msg string) *AppError {
	return &AppError{Code: 401, Message: msg}
}

// Forbidden returns a 403 error.
func Forbidden(msg string) *AppError {
	return &AppError{Code: 403, Message: msg}
}

// Internal returns a 500 error.
func Internal(msg string) *AppError {
	return &AppError{Code: 500, Message: msg}
}

// Wrapf wraps an existing error with a safe message and status code.
// The internal err is logged server-side but never exposed to the client.
func Wrapf(code int, msg string, err error) *AppError {
	slog.Error(msg, "error", err)
	return &AppError{Code: code, Message: msg}
}

// HTTPStatus extracts the HTTP status code from err.
// Returns 500 if err is not an *AppError.
func HTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return 500
}

// SafeMessage returns the error message if err is an *AppError,
// otherwise returns a generic fallback message.
func SafeMessage(err error, fallback string) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return fallback
}
