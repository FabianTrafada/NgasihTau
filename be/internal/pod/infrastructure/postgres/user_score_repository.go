// Package postgres provides PostgreSQL implementation of user score repositories.
package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// UserCategoryScoreRepository implements domain.UserCategoryScoreRepository.
type UserCategoryScoreRepository struct {
	db *pgxpool.Pool
}

// NewUserCategoryScoreRepository creates a new UserCategoryScoreRepository.
func NewUserCategoryScoreRepository(db *pgxpool.Pool) *UserCategoryScoreRepository {
	return &UserCategoryScoreRepository{db: db}
}

// Ensure UserCategoryScoreRepository implements the interface.
var _ domain.UserCategoryScoreRepository = (*UserCategoryScoreRepository)(nil)

// Upsert creates or updates a category score.
func (r *UserCategoryScoreRepository) Upsert(ctx context.Context, score *domain.UserCategoryScore) error {
	query := `
		INSERT INTO user_category_scores (
			id, user_id, category, score, view_count, star_count, follow_count, fork_count,
			total_time_spent_seconds, last_interaction_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (user_id, category) DO UPDATE SET
			score = EXCLUDED.score,
			view_count = EXCLUDED.view_count,
			star_count = EXCLUDED.star_count,
			follow_count = EXCLUDED.follow_count,
			fork_count = EXCLUDED.fork_count,
			total_time_spent_seconds = EXCLUDED.total_time_spent_seconds,
			last_interaction_at = EXCLUDED.last_interaction_at,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query,
		score.ID,
		score.UserID,
		score.Category,
		score.Score,
		score.ViewCount,
		score.StarCount,
		score.FollowCount,
		score.ForkCount,
		score.TotalTimeSpentSeconds,
		score.LastInteractionAt,
		score.CreatedAt,
		score.UpdatedAt,
	)
	if err != nil {
		return errors.Internal("failed to upsert category score", err)
	}

	return nil
}

// FindByUserID returns all category scores for a user, ordered by score.
func (r *UserCategoryScoreRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.UserCategoryScore, error) {
	query := `
		SELECT id, user_id, category, score, view_count, star_count, follow_count, fork_count,
			   total_time_spent_seconds, last_interaction_at, created_at, updated_at
		FROM user_category_scores
		WHERE user_id = $1
		ORDER BY score DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, errors.Internal("failed to find category scores", err)
	}
	defer rows.Close()

	return r.scanCategoryScores(rows)
}

// FindByUserAndCategory returns a specific category score.
func (r *UserCategoryScoreRepository) FindByUserAndCategory(ctx context.Context, userID uuid.UUID, category string) (*domain.UserCategoryScore, error) {
	query := `
		SELECT id, user_id, category, score, view_count, star_count, follow_count, fork_count,
			   total_time_spent_seconds, last_interaction_at, created_at, updated_at
		FROM user_category_scores
		WHERE user_id = $1 AND category = $2
	`

	var score domain.UserCategoryScore
	err := r.db.QueryRow(ctx, query, userID, category).Scan(
		&score.ID,
		&score.UserID,
		&score.Category,
		&score.Score,
		&score.ViewCount,
		&score.StarCount,
		&score.FollowCount,
		&score.ForkCount,
		&score.TotalTimeSpentSeconds,
		&score.LastInteractionAt,
		&score.CreatedAt,
		&score.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Internal("failed to find category score", err)
	}

	return &score, nil
}

// IncrementScore increments the score for a category.
func (r *UserCategoryScoreRepository) IncrementScore(ctx context.Context, userID uuid.UUID, category string, delta float64, interactionType domain.InteractionType) error {
	// Build increment fields based on interaction type
	var countField string
	switch interactionType {
	case domain.InteractionView:
		countField = "view_count = view_count + 1"
	case domain.InteractionStar:
		countField = "star_count = star_count + 1"
	case domain.InteractionFollow:
		countField = "follow_count = follow_count + 1"
	case domain.InteractionFork:
		countField = "fork_count = fork_count + 1"
	case domain.InteractionTimeSpent:
		// Extract seconds from delta (delta = weight = 0.1 * seconds)
		seconds := int(delta / 0.1)
		countField = "total_time_spent_seconds = total_time_spent_seconds + " + string(rune(seconds))
	default:
		countField = "view_count = view_count" // No-op for other types
	}

	query := `
		INSERT INTO user_category_scores (id, user_id, category, score, last_interaction_at, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW(), NOW())
		ON CONFLICT (user_id, category) DO UPDATE SET
			score = user_category_scores.score + $3,
			` + countField + `,
			last_interaction_at = NOW(),
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, userID, category, delta)
	if err != nil {
		return errors.Internal("failed to increment category score", err)
	}

	return nil
}

// ApplyDecay applies time decay to scores older than the threshold.
func (r *UserCategoryScoreRepository) ApplyDecay(ctx context.Context, decayFactor float64, olderThan time.Time) error {
	query := `
		UPDATE user_category_scores
		SET score = score * $1, updated_at = NOW()
		WHERE last_interaction_at < $2 AND score > 0.01
	`

	_, err := r.db.Exec(ctx, query, decayFactor, olderThan)
	if err != nil {
		return errors.Internal("failed to apply decay", err)
	}

	return nil
}

func (r *UserCategoryScoreRepository) scanCategoryScores(rows pgx.Rows) ([]*domain.UserCategoryScore, error) {
	var scores []*domain.UserCategoryScore

	for rows.Next() {
		var score domain.UserCategoryScore
		err := rows.Scan(
			&score.ID,
			&score.UserID,
			&score.Category,
			&score.Score,
			&score.ViewCount,
			&score.StarCount,
			&score.FollowCount,
			&score.ForkCount,
			&score.TotalTimeSpentSeconds,
			&score.LastInteractionAt,
			&score.CreatedAt,
			&score.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan category score", err)
		}
		scores = append(scores, &score)
	}

	return scores, nil
}

// ===========================================
// User Tag Score Repository
// ===========================================

// UserTagScoreRepository implements domain.UserTagScoreRepository.
type UserTagScoreRepository struct {
	db *pgxpool.Pool
}

// NewUserTagScoreRepository creates a new UserTagScoreRepository.
func NewUserTagScoreRepository(db *pgxpool.Pool) *UserTagScoreRepository {
	return &UserTagScoreRepository{db: db}
}

// Ensure UserTagScoreRepository implements the interface.
var _ domain.UserTagScoreRepository = (*UserTagScoreRepository)(nil)

// Upsert creates or updates a tag score.
func (r *UserTagScoreRepository) Upsert(ctx context.Context, score *domain.UserTagScore) error {
	query := `
		INSERT INTO user_tag_scores (id, user_id, tag, score, interaction_count, last_interaction_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, tag) DO UPDATE SET
			score = EXCLUDED.score,
			interaction_count = EXCLUDED.interaction_count,
			last_interaction_at = EXCLUDED.last_interaction_at,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query,
		score.ID,
		score.UserID,
		score.Tag,
		score.Score,
		score.InteractionCount,
		score.LastInteractionAt,
		score.CreatedAt,
		score.UpdatedAt,
	)
	if err != nil {
		return errors.Internal("failed to upsert tag score", err)
	}

	return nil
}

// FindByUserID returns all tag scores for a user, ordered by score.
func (r *UserTagScoreRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.UserTagScore, error) {
	query := `
		SELECT id, user_id, tag, score, interaction_count, last_interaction_at, created_at, updated_at
		FROM user_tag_scores
		WHERE user_id = $1
		ORDER BY score DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, errors.Internal("failed to find tag scores", err)
	}
	defer rows.Close()

	var scores []*domain.UserTagScore
	for rows.Next() {
		var score domain.UserTagScore
		err := rows.Scan(
			&score.ID,
			&score.UserID,
			&score.Tag,
			&score.Score,
			&score.InteractionCount,
			&score.LastInteractionAt,
			&score.CreatedAt,
			&score.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan tag score", err)
		}
		scores = append(scores, &score)
	}

	return scores, nil
}

// IncrementScore increments the score for a tag.
func (r *UserTagScoreRepository) IncrementScore(ctx context.Context, userID uuid.UUID, tag string, delta float64) error {
	query := `
		INSERT INTO user_tag_scores (id, user_id, tag, score, interaction_count, last_interaction_at, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 1, NOW(), NOW(), NOW())
		ON CONFLICT (user_id, tag) DO UPDATE SET
			score = user_tag_scores.score + $3,
			interaction_count = user_tag_scores.interaction_count + 1,
			last_interaction_at = NOW(),
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, userID, tag, delta)
	if err != nil {
		return errors.Internal("failed to increment tag score", err)
	}

	return nil
}

// IncrementScoreBatch increments scores for multiple tags.
func (r *UserTagScoreRepository) IncrementScoreBatch(ctx context.Context, userID uuid.UUID, tags []string, delta float64) error {
	if len(tags) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO user_tag_scores (id, user_id, tag, score, interaction_count, last_interaction_at, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 1, NOW(), NOW(), NOW())
		ON CONFLICT (user_id, tag) DO UPDATE SET
			score = user_tag_scores.score + $3,
			interaction_count = user_tag_scores.interaction_count + 1,
			last_interaction_at = NOW(),
			updated_at = NOW()
	`

	for _, tag := range tags {
		batch.Queue(query, userID, tag, delta)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for range tags {
		if _, err := br.Exec(); err != nil {
			return errors.Internal("failed to increment tag scores batch", err)
		}
	}

	return nil
}
