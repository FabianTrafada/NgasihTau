package errors

import (
	"errors"
)

// Is reports whether any error in err's chain matches target.
// This is a convenience wrapper around errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target.
// This is a convenience wrapper around errors.As.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// GetAppError extracts an AppError from an error chain.
// Returns nil if no AppError is found.
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// GetCode extracts the error code from an error.
// Returns CodeInternalError if the error is not an AppError.
func GetCode(err error) Code {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.Code
	}
	return CodeInternalError
}

// GetHTTPStatus extracts the HTTP status code from an error.
// Returns 500 if the error is not an AppError.
func GetHTTPStatus(err error) int {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.HTTPStatus()
	}
	return 500
}

// IsValidationError checks if the error is a validation error.
func IsValidationError(err error) bool {
	return GetCode(err) == CodeValidationError
}

// IsNotFound checks if the error is a not found error.
func IsNotFound(err error) bool {
	return GetCode(err) == CodeNotFound
}

// IsUnauthorized checks if the error is an unauthorized error.
func IsUnauthorized(err error) bool {
	return GetCode(err) == CodeUnauthorized
}

// IsForbidden checks if the error is a forbidden error.
func IsForbidden(err error) bool {
	return GetCode(err) == CodeForbidden
}

// IsConflict checks if the error is a conflict error.
func IsConflict(err error) bool {
	return GetCode(err) == CodeConflict
}

// IsRateLimited checks if the error is a rate limited error.
func IsRateLimited(err error) bool {
	return GetCode(err) == CodeRateLimited
}

// IsInternal checks if the error is an internal error.
func IsInternal(err error) bool {
	return GetCode(err) == CodeInternalError
}

// IsServiceUnavailable checks if the error is a service unavailable error.
func IsServiceUnavailable(err error) bool {
	return GetCode(err) == CodeServiceUnavailable
}
