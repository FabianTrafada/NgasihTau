package errors

import (
	"fmt"
)

// AppError represents a structured application error with code, message, and details.
// It implements the error interface and provides additional context for error handling.
type AppError struct {
	Code    Code            `json:"code"`
	Message string          `json:"message"`
	Details []ErrorDetail   `json:"details,omitempty"`
	Err     error           `json:"-"` // Original error, not serialized
}

// ErrorDetail provides additional context about a specific error,
// typically used for validation errors with field-level information.
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for errors.Is and errors.As support.
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus returns the HTTP status code for this error.
func (e *AppError) HTTPStatus() int {
	return e.Code.HTTPStatus()
}

// WithDetails adds error details to the error.
func (e *AppError) WithDetails(details ...ErrorDetail) *AppError {
	e.Details = append(e.Details, details...)
	return e
}

// WithError wraps an underlying error.
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// New creates a new AppError with the given code and message.
func New(code Code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with an AppError.
func Wrap(code Code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
