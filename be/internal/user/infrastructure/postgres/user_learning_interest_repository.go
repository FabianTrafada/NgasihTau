// Package postgres provides PostgreSQL implementations of the user learning interest repository.
package postgres

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// UserLearningInterestRepository implements domain.UserLearningInterestRepository using PostgreSQL.
type UserLearningInterestRepository struct {
	db DBTX
}

// NewUserLearningInterestRepository creates a new UserLearningInterestRepository.
func NewUserLearningInterestRepository(db DBTX) *UserLearningInterestRepository {
	return &UserLearningInterestRepository{db: db}
}

// Create creates a new user learning interest.
func (r *UserLearningInterestRepository) Create(ctx context.Context, interest *domain.UserLearningInterest) error {
	query := `
		INSERT INTO user_learning_interests (id, user_id, predefined_interest_id, custom_interest, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		interest.ID,
		interest.UserID,
		interest.PredefinedInterestID,
		interest.CustomInterest,
		interest.CreatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "user_predefined_interest_unique") {
			return errors.Conflict("user learning interest", "predefined interest already selected")
		}
		if strings.Contains(err.Error(), "user_custom_interest_unique") {
			return errors.Conflict("user learning interest", "custom interest already exists")
		}
		return errors.Internal("failed to create user learning interest", err)
	}

	return nil
}

// CreateBatch creates multiple user learning interests.
func (r *UserLearningInterestRepository) CreateBatch(ctx context.Context, interests []*domain.UserLearningInterest) error {
	if len(interests) == 0 {
		return nil
	}

	query := `
		INSERT INTO user_learning_interests (id, user_id, predefined_interest_id, custom_interest, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING
	`

	for _, interest := range interests {
		_, err := r.db.Exec(ctx, query,
			interest.ID,
			interest.UserID,
			interest.PredefinedInterestID,
			interest.CustomInterest,
			interest.CreatedAt,
		)
		if err != nil {
			return errors.Internal("failed to create user learning interest batch", err)
		}
	}

	return nil
}

// FindByUserID returns all learning interests for a user.
func (r *UserLearningInterestRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.UserLearningInterest, error) {
	query := `
		SELECT id, user_id, predefined_interest_id, custom_interest, created_at
		FROM user_learning_interests
		WHERE user_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Internal("failed to query user learning interests", err)
	}
	defer rows.Close()

	var interests []*domain.UserLearningInterest
	for rows.Next() {
		interest := &domain.UserLearningInterest{}
		if err := rows.Scan(
			&interest.ID,
			&interest.UserID,
			&interest.PredefinedInterestID,
			&interest.CustomInterest,
			&interest.CreatedAt,
		); err != nil {
			return nil, errors.Internal("failed to scan user learning interest", err)
		}
		interests = append(interests, interest)
	}

	return interests, nil
}

// FindByUserIDWithDetails returns all learning interests for a user with predefined interest details.
func (r *UserLearningInterestRepository) FindByUserIDWithDetails(ctx context.Context, userID uuid.UUID) ([]*domain.UserLearningInterest, error) {
	query := `
		SELECT 
			uli.id, uli.user_id, uli.predefined_interest_id, uli.custom_interest, uli.created_at,
			pi.id, pi.name, pi.slug, pi.description, pi.icon, pi.category, pi.display_order, pi.is_active, pi.created_at, pi.updated_at
		FROM user_learning_interests uli
		LEFT JOIN predefined_interests pi ON uli.predefined_interest_id = pi.id
		WHERE uli.user_id = $1
		ORDER BY COALESCE(pi.display_order, 9999) ASC, uli.created_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Internal("failed to query user learning interests with details", err)
	}
	defer rows.Close()

	var interests []*domain.UserLearningInterest
	for rows.Next() {
		interest := &domain.UserLearningInterest{}
		var pi struct {
			ID           *uuid.UUID
			Name         *string
			Slug         *string
			Description  *string
			Icon         *string
			Category     *string
			DisplayOrder *int
			IsActive     *bool
			CreatedAt    *interface{}
			UpdatedAt    *interface{}
		}

		if err := rows.Scan(
			&interest.ID,
			&interest.UserID,
			&interest.PredefinedInterestID,
			&interest.CustomInterest,
			&interest.CreatedAt,
			&pi.ID,
			&pi.Name,
			&pi.Slug,
			&pi.Description,
			&pi.Icon,
			&pi.Category,
			&pi.DisplayOrder,
			&pi.IsActive,
			&pi.CreatedAt,
			&pi.UpdatedAt,
		); err != nil {
			return nil, errors.Internal("failed to scan user learning interest with details", err)
		}

		// Populate predefined interest if it exists
		if pi.ID != nil {
			interest.PredefinedInterest = &domain.PredefinedInterest{
				ID:           *pi.ID,
				Name:         *pi.Name,
				Slug:         *pi.Slug,
				Description:  pi.Description,
				Icon:         pi.Icon,
				Category:     pi.Category,
				DisplayOrder: *pi.DisplayOrder,
				IsActive:     *pi.IsActive,
			}
		}

		interests = append(interests, interest)
	}

	return interests, nil
}

// Delete removes a user learning interest.
func (r *UserLearningInterestRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM user_learning_interests WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Internal("failed to delete user learning interest", err)
	}

	if result.RowsAffected() == 0 {
		return errors.NotFound("user learning interest", id.String())
	}

	return nil
}

// DeleteByUserID removes all learning interests for a user.
func (r *UserLearningInterestRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM user_learning_interests WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return errors.Internal("failed to delete user learning interests", err)
	}

	return nil
}

// DeleteByUserIDAndPredefinedID removes a specific predefined interest for a user.
func (r *UserLearningInterestRepository) DeleteByUserIDAndPredefinedID(ctx context.Context, userID, predefinedInterestID uuid.UUID) error {
	query := `DELETE FROM user_learning_interests WHERE user_id = $1 AND predefined_interest_id = $2`

	result, err := r.db.Exec(ctx, query, userID, predefinedInterestID)
	if err != nil {
		return errors.Internal("failed to delete user predefined interest", err)
	}

	if result.RowsAffected() == 0 {
		return errors.NotFound("user learning interest", predefinedInterestID.String())
	}

	return nil
}

// ExistsByUserIDAndPredefinedID checks if a user already has a specific predefined interest.
func (r *UserLearningInterestRepository) ExistsByUserIDAndPredefinedID(ctx context.Context, userID, predefinedInterestID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_learning_interests WHERE user_id = $1 AND predefined_interest_id = $2)`

	var exists bool
	err := r.db.QueryRow(ctx, query, userID, predefinedInterestID).Scan(&exists)
	if err != nil {
		return false, errors.Internal("failed to check user predefined interest existence", err)
	}

	return exists, nil
}

// ExistsByUserIDAndCustom checks if a user already has a specific custom interest.
func (r *UserLearningInterestRepository) ExistsByUserIDAndCustom(ctx context.Context, userID uuid.UUID, customInterest string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_learning_interests WHERE user_id = $1 AND LOWER(custom_interest) = LOWER($2))`

	var exists bool
	err := r.db.QueryRow(ctx, query, userID, customInterest).Scan(&exists)
	if err != nil {
		return false, errors.Internal("failed to check user custom interest existence", err)
	}

	return exists, nil
}

// CountByUserID returns the number of interests for a user.
func (r *UserLearningInterestRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM user_learning_interests WHERE user_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, errors.Internal("failed to count user learning interests", err)
	}

	return count, nil
}

// GetInterestSummaries returns simplified interest summaries for a user.
func (r *UserLearningInterestRepository) GetInterestSummaries(ctx context.Context, userID uuid.UUID) ([]*domain.InterestSummary, error) {
	query := `
		SELECT 
			uli.id,
			COALESCE(pi.name, uli.custom_interest) as name,
			pi.slug,
			pi.icon,
			pi.category,
			CASE WHEN uli.custom_interest IS NOT NULL THEN TRUE ELSE FALSE END as is_custom
		FROM user_learning_interests uli
		LEFT JOIN predefined_interests pi ON uli.predefined_interest_id = pi.id
		WHERE uli.user_id = $1
		ORDER BY COALESCE(pi.display_order, 9999) ASC, uli.created_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Internal("failed to query interest summaries", err)
	}
	defer rows.Close()

	var summaries []*domain.InterestSummary
	for rows.Next() {
		summary := &domain.InterestSummary{}
		if err := rows.Scan(
			&summary.ID,
			&summary.Name,
			&summary.Slug,
			&summary.Icon,
			&summary.Category,
			&summary.IsCustom,
		); err != nil {
			return nil, errors.Internal("failed to scan interest summary", err)
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// Ensure interface compliance
var _ domain.UserLearningInterestRepository = (*UserLearningInterestRepository)(nil)
