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

	// CodeStorageQuotaExceeded indicates the user's storage quota has been exceeded.
	// Implements Requirement 3.4: Upload rejection due to quota.
	CodeStorageQuotaExceeded Code = "STORAGE_QUOTA_EXCEEDED"

	// CodeAILimitExceeded indicates the user's daily AI message limit has been exceeded.
	// Implements Requirement 9.5: AI limit rejection.
	CodeAILimitExceeded Code = "AI_LIMIT_EXCEEDED"

	// CodePremiumFeatureRequired indicates the feature requires premium tier or higher.
	// Implements Requirement 11.3: Premium feature access rejection.
	CodePremiumFeatureRequired Code = "PREMIUM_FEATURE_REQUIRED"

	// CodeProFeatureRequired indicates the feature requires pro tier.
	// Implements Requirement 12.6: Pro feature access rejection.
	CodeProFeatureRequired Code = "PRO_FEATURE_REQUIRED"
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
	case CodeStorageQuotaExceeded:
		return 403
	case CodeAILimitExceeded:
		return 429
	case CodePremiumFeatureRequired:
		return 403
	case CodeProFeatureRequired:
		return 403
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
