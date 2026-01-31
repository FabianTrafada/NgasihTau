// Package response provides standard response envelope formats
// for consistent API responses across all microservices.
// Implements requirement 10.4.1 for consistent API response format.
package response

import (
	"time"
)

// Response represents the standard success response envelope
// as defined in requirement 10.4.1.
type Response[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data"`
	Meta    Meta   `json:"meta"`
}

// PaginatedResponse represents a paginated response envelope
// as defined in requirement 10.4.1.
type PaginatedResponse[T any] struct {
	Success    bool       `json:"success"`
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
	Meta       Meta       `json:"meta"`
}

// Meta contains metadata for the response including timestamp and request ID.
type Meta struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id"`
}

// Pagination contains pagination information for list responses.
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Builder builds responses with consistent formatting.
type Builder struct {
	requestID string
}

// NewBuilder creates a new Builder with the given request ID.
func NewBuilder(requestID string) *Builder {
	return &Builder{
		requestID: requestID,
	}
}

// newMeta creates a new Meta with current timestamp and request ID.
func (b *Builder) newMeta() Meta {
	return Meta{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: b.requestID,
	}
}

// Success creates a success response with the given data.
func (b *Builder) Success(data any) Response[any] {
	return Response[any]{
		Success: true,
		Data:    data,
		Meta:    b.newMeta(),
	}
}

// Paginated creates a paginated response with the given data and pagination info.
func (b *Builder) Paginated(data any, page, perPage, total int) PaginatedResponse[any] {
	totalPages := 0
	if perPage > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	return PaginatedResponse[any]{
		Success: true,
		Data:    toSlice(data),
		Pagination: Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
		Meta: b.newMeta(),
	}
}

// toSlice converts data to []any for consistent JSON serialization.
// If data is already a slice, it converts each element.
// If data is nil, returns an empty slice.
func toSlice(data any) []any {
	if data == nil {
		return []any{}
	}

	// Use type assertion for common slice types
	switch v := data.(type) {
	case []any:
		return v
	default:
		// For other slice types, we return as single-element slice
		// In practice, callers should pass []any or use generic version
		return []any{data}
	}
}
