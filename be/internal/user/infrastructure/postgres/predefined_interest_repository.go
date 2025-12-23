// Package postgres provides PostgreSQL implementations of the predefined interest repository.
package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// PredefinedInterestRepository implements domain.PredefinedInterestRepository using PostgreSQL.
type PredefinedInterestRepository struct {
	db DBTX
}

// NewPredefinedInterestRepository creates a new PredefinedInterestRepository.
func NewPredefinedInterestRepository(db DBTX) *PredefinedInterestRepository {
	return &PredefinedInterestRepository{db: db}
}

// FindAll returns all active predefined interests ordered by display_order.
func (r *PredefinedInterestRepository) FindAll(ctx context.Context) ([]*domain.PredefinedInterest, error) {
	query := `
		SELECT id, name, slug, description, icon, category, display_order, is_active, created_at, updated_at
		FROM predefined_interests
		WHERE is_active = TRUE
		ORDER BY display_order ASC, name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, errors.Internal("failed to query predefined interests", err)
	}
	defer rows.Close()

	var interests []*domain.PredefinedInterest
	for rows.Next() {
		interest := &domain.PredefinedInterest{}
		if err := rows.Scan(
			&interest.ID,
			&interest.Name,
			&interest.Slug,
			&interest.Description,
			&interest.Icon,
			&interest.Category,
			&interest.DisplayOrder,
			&interest.IsActive,
			&interest.CreatedAt,
			&interest.UpdatedAt,
		); err != nil {
			return nil, errors.Internal("failed to scan predefined interest", err)
		}
		interests = append(interests, interest)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Internal("error iterating predefined interests", err)
	}

	return interests, nil
}

// FindByID finds a predefined interest by ID.
func (r *PredefinedInterestRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.PredefinedInterest, error) {
	query := `
		SELECT id, name, slug, description, icon, category, display_order, is_active, created_at, updated_at
		FROM predefined_interests
		WHERE id = $1
	`

	interest := &domain.PredefinedInterest{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&interest.ID,
		&interest.Name,
		&interest.Slug,
		&interest.Description,
		&interest.Icon,
		&interest.Category,
		&interest.DisplayOrder,
		&interest.IsActive,
		&interest.CreatedAt,
		&interest.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound("predefined interest", id.String())
		}
		return nil, errors.Internal("failed to query predefined interest", err)
	}

	return interest, nil
}

// FindBySlug finds a predefined interest by slug.
func (r *PredefinedInterestRepository) FindBySlug(ctx context.Context, slug string) (*domain.PredefinedInterest, error) {
	query := `
		SELECT id, name, slug, description, icon, category, display_order, is_active, created_at, updated_at
		FROM predefined_interests
		WHERE slug = $1
	`

	interest := &domain.PredefinedInterest{}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&interest.ID,
		&interest.Name,
		&interest.Slug,
		&interest.Description,
		&interest.Icon,
		&interest.Category,
		&interest.DisplayOrder,
		&interest.IsActive,
		&interest.CreatedAt,
		&interest.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound("predefined interest", slug)
		}
		return nil, errors.Internal("failed to query predefined interest by slug", err)
	}

	return interest, nil
}

// FindByIDs finds predefined interests by multiple IDs.
func (r *PredefinedInterestRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*domain.PredefinedInterest, error) {
	if len(ids) == 0 {
		return []*domain.PredefinedInterest{}, nil
	}

	query := `
		SELECT id, name, slug, description, icon, category, display_order, is_active, created_at, updated_at
		FROM predefined_interests
		WHERE id = ANY($1)
		ORDER BY display_order ASC
	`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, errors.Internal("failed to query predefined interests by IDs", err)
	}
	defer rows.Close()

	var interests []*domain.PredefinedInterest
	for rows.Next() {
		interest := &domain.PredefinedInterest{}
		if err := rows.Scan(
			&interest.ID,
			&interest.Name,
			&interest.Slug,
			&interest.Description,
			&interest.Icon,
			&interest.Category,
			&interest.DisplayOrder,
			&interest.IsActive,
			&interest.CreatedAt,
			&interest.UpdatedAt,
		); err != nil {
			return nil, errors.Internal("failed to scan predefined interest", err)
		}
		interests = append(interests, interest)
	}

	return interests, nil
}

// FindByCategory returns all active predefined interests in a category.
func (r *PredefinedInterestRepository) FindByCategory(ctx context.Context, category string) ([]*domain.PredefinedInterest, error) {
	query := `
		SELECT id, name, slug, description, icon, category, display_order, is_active, created_at, updated_at
		FROM predefined_interests
		WHERE category = $1 AND is_active = TRUE
		ORDER BY display_order ASC, name ASC
	`

	rows, err := r.db.Query(ctx, query, category)
	if err != nil {
		return nil, errors.Internal("failed to query predefined interests by category", err)
	}
	defer rows.Close()

	var interests []*domain.PredefinedInterest
	for rows.Next() {
		interest := &domain.PredefinedInterest{}
		if err := rows.Scan(
			&interest.ID,
			&interest.Name,
			&interest.Slug,
			&interest.Description,
			&interest.Icon,
			&interest.Category,
			&interest.DisplayOrder,
			&interest.IsActive,
			&interest.CreatedAt,
			&interest.UpdatedAt,
		); err != nil {
			return nil, errors.Internal("failed to scan predefined interest", err)
		}
		interests = append(interests, interest)
	}

	return interests, nil
}

// GetCategories returns all unique categories.
func (r *PredefinedInterestRepository) GetCategories(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT category
		FROM predefined_interests
		WHERE is_active = TRUE AND category IS NOT NULL
		ORDER BY category ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, errors.Internal("failed to query categories", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, errors.Internal("failed to scan category", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// Ensure interface compliance
var _ domain.PredefinedInterestRepository = (*PredefinedInterestRepository)(nil)
