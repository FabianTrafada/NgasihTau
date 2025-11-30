// Package errors provides custom error types and error handling utilities
// for consistent error responses across all microservices.
package errors

// Code represents a standardized error code used across all services.
type Code string

// Standard error codes as defined in requirement 10.4.1.
const (
	// CodeValidationError indicates invalid input data.
	CodeValidationError Code = "VALIDATION_ERROR"

	// CodeUnauthorized indicates missing or invalid authentication.
	CodeUnauthorized Code = "UNAUTHORIZED"

	// CodeForbidden indicates insufficient permissions.
	CodeForbidden Code = "FORBIDDEN"

	// CodeNotFound indicates the requested resource was not found.
	CodeNotFound Code = "NOT_FOUND"

	// CodeConflict indicates the resource already exists or state conflict.
	CodeConflict Code = "CONFLICT"

	// CodeRateLimited indicates too many requests.
	CodeRateLimited Code = "RATE_LIMITED"

	// CodeInternalError indicates an internal server error.
	CodeInternalError Code = "INTERNAL_ERROR"

	// CodeServiceUnavailable indicates the service is temporarily unavailable.
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"

	// CodeBadRequest indicates a malformed request.
	CodeBadRequest Code = "BAD_REQUEST"

	// CodeTimeout indicates the operation timed out.
	CodeTimeout Code = "TIMEOUT"

	// CodeUnprocessable indicates the request was well-formed but semantically invalid.
	CodeUnprocessable Code = "UNPROCESSABLE_ENTITY"
)

// HTTPStatus returns the corresponding HTTP status code for an error code.
func (c Code) HTTPStatus() int {
	switch c {
	case CodeValidationError:
		return 400
	case CodeBadRequest:
		return 400
	case CodeUnauthorized:
		return 401
	case CodeForbidden:
		return 403
	case CodeNotFound:
		return 404
	case CodeConflict:
		return 409
	case CodeUnprocessable:
		return 422
	case CodeRateLimited:
		return 429
	case CodeInternalError:
		return 500
	case CodeServiceUnavailable:
		return 503
	case CodeTimeout:
		return 504
	default:
		return 500
	}
}

// String returns the string representation of the error code.
func (c Code) String() string {
	return string(c)
}
