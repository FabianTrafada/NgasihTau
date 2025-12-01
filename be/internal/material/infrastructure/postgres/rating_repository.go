package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"ngasihtau/internal/material/domain"
)

// RatingRepository implements domain.RatingRepository using PostgreSQL.
type RatingRepository struct {
	db DBTX
}

// NewRatingRepository creates a new RatingRepository.
func NewRatingRepository(db DBTX) *RatingRepository {
	return &RatingRepository{db: db}
}

// Create creates a new rating.
func (r *RatingRepository) Create(ctx context.Context, rating *domain.Rating) error {
	query := `
		INSERT INTO ratings (id, material_id, user_id, score, review, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query,
		rating.ID,
		rating.MaterialID,
		rating.UserID,
		rating.Score,
		rating.Review,
		rating.CreatedAt,
		rating.UpdatedAt,
	)

	return err
}

// FindByID finds a rating by ID.
func (r *RatingRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Rating, error) {
	query := `
		SELECT id, material_id, user_id, score, review, created_at, updated_at
		FROM ratings
		WHERE id = $1
	`

	var rating domain.Rating
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rating.ID,
		&rating.MaterialID,
		&rating.UserID,
		&rating.Score,
		&rating.Review,
		&rating.CreatedAt,
		&rating.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("rating not found")
	}

	return &rating, err
}

// FindByMaterialAndUser finds a rating by material and user.
func (r *RatingRepository) FindByMaterialAndUser(ctx context.Context, materialID, userID uuid.UUID) (*domain.Rating, error) {
	query := `
		SELECT id, material_id, user_id, score, review, created_at, updated_at
		FROM ratings
		WHERE material_id = $1 AND user_id = $2
	`

	var rating domain.Rating
	err := r.db.QueryRow(ctx, query, materialID, userID).Scan(
		&rating.ID,
		&rating.MaterialID,
		&rating.UserID,
		&rating.Score,
		&rating.Review,
		&rating.CreatedAt,
		&rating.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // Not found is not an error for this use case
	}

	return &rating, err
}

// FindByMaterialID finds all ratings for a material.
func (r *RatingRepository) FindByMaterialID(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*domain.Rating, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM ratings WHERE material_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, materialID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get ratings
	query := `
		SELECT id, material_id, user_id, score, review, created_at, updated_at
		FROM ratings
		WHERE material_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, materialID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ratings []*domain.Rating
	for rows.Next() {
		var rating domain.Rating
		if err := rows.Scan(
			&rating.ID,
			&rating.MaterialID,
			&rating.UserID,
			&rating.Score,
			&rating.Review,
			&rating.CreatedAt,
			&rating.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		ratings = append(ratings, &rating)
	}

	return ratings, total, rows.Err()
}

// FindByMaterialIDWithUsers finds all ratings for a material with user details.
func (r *RatingRepository) FindByMaterialIDWithUsers(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*domain.RatingWithUser, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM ratings WHERE material_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, materialID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get ratings
	query := `
		SELECT id, material_id, user_id, score, review, created_at, updated_at
		FROM ratings
		WHERE material_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, materialID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ratings []*domain.RatingWithUser
	for rows.Next() {
		var rating domain.RatingWithUser
		if err := rows.Scan(
			&rating.ID,
			&rating.MaterialID,
			&rating.UserID,
			&rating.Score,
			&rating.Review,
			&rating.CreatedAt,
			&rating.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		// User info would be populated via gRPC call to User Service
		rating.UserName = "Unknown"
		ratings = append(ratings, &rating)
	}

	return ratings, total, rows.Err()
}

// Update updates a rating.
func (r *RatingRepository) Update(ctx context.Context, rating *domain.Rating) error {
	query := `
		UPDATE ratings
		SET score = $2, review = $3, updated_at = $4
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		rating.ID,
		rating.Score,
		rating.Review,
		time.Now(),
	)

	return err
}

// Delete removes a rating.
func (r *RatingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ratings WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// GetSummary returns the rating summary for a material.
func (r *RatingRepository) GetSummary(ctx context.Context, materialID uuid.UUID) (*domain.RatingSummary, error) {
	query := `
		SELECT 
			COALESCE(AVG(score), 0) as avg_rating,
			COUNT(*) as rating_count,
			COUNT(*) FILTER (WHERE score = 1) as one_star,
			COUNT(*) FILTER (WHERE score = 2) as two_star,
			COUNT(*) FILTER (WHERE score = 3) as three_star,
			COUNT(*) FILTER (WHERE score = 4) as four_star,
			COUNT(*) FILTER (WHERE score = 5) as five_star
		FROM ratings
		WHERE material_id = $1
	`

	var summary domain.RatingSummary
	err := r.db.QueryRow(ctx, query, materialID).Scan(
		&summary.AverageRating,
		&summary.RatingCount,
		&summary.Distribution.OneStar,
		&summary.Distribution.TwoStar,
		&summary.Distribution.ThreeStar,
		&summary.Distribution.FourStar,
		&summary.Distribution.FiveStar,
	)

	if err != nil {
		return nil, err
	}

	return &summary, nil
}

// CalculateAverage calculates the average rating for a material.
func (r *RatingRepository) CalculateAverage(ctx context.Context, materialID uuid.UUID) (float64, int, error) {
	query := `SELECT COALESCE(AVG(score), 0), COUNT(*) FROM ratings WHERE material_id = $1`

	var avgRating float64
	var count int
	err := r.db.QueryRow(ctx, query, materialID).Scan(&avgRating, &count)
	return avgRating, count, err
}
