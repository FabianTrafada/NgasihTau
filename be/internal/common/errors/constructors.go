package errors

import "fmt"

// Validation creates a validation error with optional field details.
func Validation(message string, details ...ErrorDetail) *AppError {
	return &AppError{
		Code:    CodeValidationError,
		Message: message,
		Details: details,
	}
}

// ValidationField creates a validation error for a specific field.
func ValidationField(field, message string) *AppError {
	return &AppError{
		Code:    CodeValidationError,
		Message: fmt.Sprintf("validation failed for field '%s'", field),
		Details: []ErrorDetail{{Field: field, Message: message}},
	}
}

// Unauthorized creates an unauthorized error.
func Unauthorized(message string) *AppError {
	if message == "" {
		message = "authentication required"
	}
	return &AppError{
		Code:    CodeUnauthorized,
		Message: message,
	}
}

// Forbidden creates a forbidden error.
func Forbidden(message string) *AppError {
	if message == "" {
		message = "access denied"
	}
	return &AppError{
		Code:    CodeForbidden,
		Message: message,
	}
}

// NotFound creates a not found error for a resource.
func NotFound(resource, identifier string) *AppError {
	return &AppError{
		Code:    CodeNotFound,
		Message: fmt.Sprintf("%s not found: %s", resource, identifier),
	}
}

// NotFoundMsg creates a not found error with a custom message.
func NotFoundMsg(message string) *AppError {
	return &AppError{
		Code:    CodeNotFound,
		Message: message,
	}
}

// Conflict creates a conflict error for duplicate resources.
func Conflict(resource, identifier string) *AppError {
	return &AppError{
		Code:    CodeConflict,
		Message: fmt.Sprintf("%s already exists: %s", resource, identifier),
	}
}

// ConflictMsg creates a conflict error with a custom message.
func ConflictMsg(message string) *AppError {
	return &AppError{
		Code:    CodeConflict,
		Message: message,
	}
}

// RateLimited creates a rate limited error.
func RateLimited(retryAfter int) *AppError {
	return &AppError{
		Code:    CodeRateLimited,
		Message: fmt.Sprintf("rate limit exceeded, retry after %d seconds", retryAfter),
		Details: []ErrorDetail{{Field: "retry_after", Value: retryAfter}},
	}
}

// Internal creates an internal server error.
// The original error is wrapped but not exposed in the response.
func Internal(message string, err error) *AppError {
	if message == "" {
		message = "an internal error occurred"
	}
	return &AppError{
		Code:    CodeInternalError,
		Message: message,
		Err:     err,
	}
}

// ServiceUnavailable creates a service unavailable error.
func ServiceUnavailable(service string, err error) *AppError {
	return &AppError{
		Code:    CodeServiceUnavailable,
		Message: fmt.Sprintf("service unavailable: %s", service),
		Err:     err,
	}
}

// BadRequest creates a bad request error.
func BadRequest(message string) *AppError {
	if message == "" {
		message = "invalid request"
	}
	return &AppError{
		Code:    CodeBadRequest,
		Message: message,
	}
}

// Timeout creates a timeout error.
func Timeout(operation string) *AppError {
	return &AppError{
		Code:    CodeTimeout,
		Message: fmt.Sprintf("operation timed out: %s", operation),
	}
}

// Unprocessable creates an unprocessable entity error.
func Unprocessable(message string) *AppError {
	return &AppError{
		Code:    CodeUnprocessable,
		Message: message,
	}
}
