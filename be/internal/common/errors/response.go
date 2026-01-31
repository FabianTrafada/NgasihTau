package errors

import (
	"time"
)

// ErrorResponse represents the standard error response envelope
// as defined in requirement 10.4.1.
type ErrorResponse struct {
	Success bool           `json:"success"`
	Error   ErrorBody      `json:"error"`
	Meta    ResponseMeta   `json:"meta"`
}

// ErrorBody contains the error details in the response.
type ErrorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ResponseMeta contains metadata for the response.
type ResponseMeta struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id"`
}

// ResponseBuilder builds error responses with consistent formatting.
type ResponseBuilder struct {
	requestID string
}

// NewResponseBuilder creates a new ResponseBuilder with the given request ID.
func NewResponseBuilder(requestID string) *ResponseBuilder {
	return &ResponseBuilder{
		requestID: requestID,
	}
}

// Build creates an ErrorResponse from an error.
// If the error is an AppError, it uses its code and details.
// Otherwise, it creates a generic internal error response.
func (b *ResponseBuilder) Build(err error) ErrorResponse {
	appErr := GetAppError(err)
	if appErr == nil {
		// Wrap non-AppError as internal error
		appErr = Internal("an unexpected error occurred", err)
	}

	return ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    appErr.Code.String(),
			Message: appErr.Message,
			Details: appErr.Details,
		},
		Meta: ResponseMeta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: b.requestID,
		},
	}
}

// BuildFromCode creates an ErrorResponse from an error code and message.
func (b *ResponseBuilder) BuildFromCode(code Code, message string, details ...ErrorDetail) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code.String(),
			Message: message,
			Details: details,
		},
		Meta: ResponseMeta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: b.requestID,
		},
	}
}

// BuildResponse is a convenience function to build an error response without a builder.
func BuildResponse(requestID string, err error) ErrorResponse {
	return NewResponseBuilder(requestID).Build(err)
}

// HTTPStatusFromError returns the HTTP status code for an error.
func HTTPStatusFromError(err error) int {
	return GetHTTPStatus(err)
}
