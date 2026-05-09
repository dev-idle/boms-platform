package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// AppError is an application-level error with HTTP status and stable code.
type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Err        error
	Details    map[string]string
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetail returns a copy with an extra detail entry.
func (e *AppError) WithDetail(key, value string) *AppError {
	cp := *e
	if cp.Details == nil {
		cp.Details = map[string]string{}
	} else {
		cp.Details = mapsClone(e.Details)
	}
	cp.Details[key] = value
	return &cp
}

func mapsClone(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// New creates an AppError.
func New(status int, code, message string) *AppError {
	return &AppError{StatusCode: status, Code: code, Message: message}
}

// Wrap adds an underlying error.
func Wrap(status int, code, message string, err error) *AppError {
	return &AppError{StatusCode: status, Code: code, Message: message, Err: err}
}

// Common constructors for future features (no auth logic here).
var (
	ErrNotFound           = New(http.StatusNotFound, "not_found", "Resource not found")
	ErrConflict           = New(http.StatusConflict, "conflict", "Resource conflict")
	ErrValidation         = New(http.StatusBadRequest, "validation_error", "Validation failed")
	ErrUnauthorized       = New(http.StatusUnauthorized, "unauthorized", "Unauthorized")
	ErrForbidden          = New(http.StatusForbidden, "forbidden", "Forbidden")
	ErrInternal           = New(http.StatusInternalServerError, "internal_error", "Internal server error")
	ErrServiceUnavailable = New(http.StatusServiceUnavailable, "service_unavailable", "Service unavailable")
	ErrTooManyRequests    = New(http.StatusTooManyRequests, "rate_limited", "Too many requests")
)

// FromFiberError maps *fiber.Error to AppError.
func FromFiberError(err *fiber.Error) *AppError {
	if err == nil {
		return nil
	}
	return New(err.Code, httpStatusText(err.Code), err.Message)
}

func httpStatusText(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusTooManyRequests:
		return "rate_limited"
	default:
		return "http_error"
	}
}

// AsAppError unwraps err into *AppError when possible.
func AsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if stderrors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

// Errorf wraps fmt.Errorf as an internal AppError.
func Errorf(format string, args ...any) *AppError {
	return Wrap(http.StatusInternalServerError, "internal_error", "Internal server error", fmt.Errorf(format, args...))
}
