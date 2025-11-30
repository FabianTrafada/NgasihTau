package response

import (
	"time"
)

// OK creates a success response with the given data using the provided request ID.
// This is a convenience function for simple use cases.
func OK[T any](requestID string, data T) Response[T] {
	return Response[T]{
		Success: true,
		Data:    data,
		Meta: Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: requestID,
		},
	}
}

// List creates a paginated response with the given data and pagination info.
// This is a convenience function for simple use cases.
func List[T any](requestID string, data []T, page, perPage, total int) PaginatedResponse[T] {
	totalPages := 0
	if perPage > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	// Ensure data is never nil in JSON output
	if data == nil {
		data = []T{}
	}

	return PaginatedResponse[T]{
		Success: true,
		Data:    data,
		Pagination: Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
		Meta: Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: requestID,
		},
	}
}

// Empty creates a success response with no data (null).
// Useful for DELETE operations or operations that don't return data.
func Empty(requestID string) Response[any] {
	return Response[any]{
		Success: true,
		Data:    nil,
		Meta: Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: requestID,
		},
	}
}

// Created creates a success response for resource creation.
// Typically used with HTTP 201 status code.
func Created[T any](requestID string, data T) Response[T] {
	return OK(requestID, data)
}

// PaginationParams holds pagination request parameters.
type PaginationParams struct {
	Page    int
	PerPage int
}

// DefaultPaginationParams returns default pagination parameters.
// Default: page 1, 20 items per page.
func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Page:    1,
		PerPage: 20,
	}
}

// Normalize ensures pagination parameters are within valid bounds.
// Page must be >= 1, PerPage must be between 1 and maxPerPage.
func (p *PaginationParams) Normalize(maxPerPage int) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 20
	}
	if maxPerPage > 0 && p.PerPage > maxPerPage {
		p.PerPage = maxPerPage
	}
}

// Offset calculates the database offset for the current page.
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Limit returns the number of items per page.
func (p PaginationParams) Limit() int {
	return p.PerPage
}
