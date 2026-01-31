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

// StorageQuotaDetails contains details about storage quota exceeded error.
// Implements Requirement 3.4: Error response with usage details.
type StorageQuotaDetails struct {
	CurrentUsage int64  `json:"current_usage"`
	QuotaLimit   int64  `json:"quota_limit"`
	RequiredSize int64  `json:"required_size"`
	Tier         string `json:"tier"`
}

// StorageQuotaExceeded creates a storage quota exceeded error with usage details.
// Implements Requirement 3.4: Upload rejection due to quota.
func StorageQuotaExceeded(currentUsage, quotaLimit, requiredSize int64, tier string) *AppError {
	return &AppError{
		Code:    CodeStorageQuotaExceeded,
		Message: "storage quota exceeded",
		Details: []ErrorDetail{
			{Field: "current_usage", Value: currentUsage},
			{Field: "quota_limit", Value: quotaLimit},
			{Field: "required_size", Value: requiredSize},
			{Field: "tier", Value: tier},
		},
	}
}

// AILimitExceeded creates an AI daily limit exceeded error.
// Implements Requirement 9.5: AI limit rejection.
func AILimitExceeded(usedToday, dailyLimit int, tier string) *AppError {
	return &AppError{
		Code:    CodeAILimitExceeded,
		Message: "daily AI message limit exceeded",
		Details: []ErrorDetail{
			{Field: "used_today", Value: usedToday},
			{Field: "daily_limit", Value: dailyLimit},
			{Field: "tier", Value: tier},
		},
	}
}

// PremiumFeatureRequired creates an error for premium-only features.
// Implements Requirement 11.3: Premium feature access rejection.
func PremiumFeatureRequired(feature string) *AppError {
	return &AppError{
		Code:    CodePremiumFeatureRequired,
		Message: fmt.Sprintf("premium subscription required for feature: %s", feature),
		Details: []ErrorDetail{
			{Field: "feature", Value: feature},
			{Field: "required_tier", Value: "premium"},
		},
	}
}

// ProFeatureRequired creates an error for pro-only features.
// Implements Requirement 12.6: Pro feature access rejection.
func ProFeatureRequired(feature string) *AppError {
	return &AppError{
		Code:    CodeProFeatureRequired,
		Message: fmt.Sprintf("pro subscription required for feature: %s", feature),
		Details: []ErrorDetail{
			{Field: "feature", Value: feature},
			{Field: "required_tier", Value: "pro"},
		},
	}
}
