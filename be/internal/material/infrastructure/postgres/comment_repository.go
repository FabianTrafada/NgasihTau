package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"ngasihtau/internal/material/domain"
)

// CommentRepository implements domain.CommentRepository using PostgreSQL.
type CommentRepository struct {
	db DBTX
}

// NewCommentRepository creates a new CommentRepository.
func NewCommentRepository(db DBTX) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create creates a new comment.
func (r *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	query := `
		INSERT INTO comments (id, material_id, user_id, parent_id, content, edited, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		comment.ID,
		comment.MaterialID,
		comment.UserID,
		comment.ParentID,
		comment.Content,
		comment.Edited,
		comment.CreatedAt,
		comment.UpdatedAt,
	)

	return err
}

// FindByID finds a comment by ID.
func (r *CommentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	query := `
		SELECT id, material_id, user_id, parent_id, content, edited, created_at, updated_at, deleted_at
		FROM comments
		WHERE id = $1 AND deleted_at IS NULL
	`

	var c domain.Comment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.MaterialID,
		&c.UserID,
		&c.ParentID,
		&c.Content,
		&c.Edited,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("comment not found")
	}

	return &c, err
}

// FindByMaterialID finds all comments for a material.
func (r *CommentRepository) FindByMaterialID(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*domain.Comment, int, error) {
	// Get total count (top-level comments only)
	countQuery := `SELECT COUNT(*) FROM comments WHERE material_id = $1 AND parent_id IS NULL AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, materialID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get comments (top-level only)
	query := `
		SELECT id, material_id, user_id, parent_id, content, edited, created_at, updated_at, deleted_at
		FROM comments
		WHERE material_id = $1 AND parent_id IS NULL AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, materialID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(
			&c.ID,
			&c.MaterialID,
			&c.UserID,
			&c.ParentID,
			&c.Content,
			&c.Edited,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		comments = append(comments, &c)
	}

	return comments, total, rows.Err()
}

// FindByMaterialIDWithUsers finds all comments for a material with user details.
func (r *CommentRepository) FindByMaterialIDWithUsers(ctx context.Context, materialID uuid.UUID, limit, offset int) ([]*domain.CommentWithUser, int, error) {
	// Get total count (top-level comments only)
	countQuery := `SELECT COUNT(*) FROM comments WHERE material_id = $1 AND parent_id IS NULL AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, materialID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get comments with user info (top-level only)
	// Note: In a real implementation, this would join with the users table via gRPC or a shared view
	query := `
		SELECT c.id, c.material_id, c.user_id, c.parent_id, c.content, c.edited, c.created_at, c.updated_at, c.deleted_at
		FROM comments c
		WHERE c.material_id = $1 AND c.parent_id IS NULL AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, materialID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var comments []*domain.CommentWithUser
	for rows.Next() {
		var c domain.CommentWithUser
		if err := rows.Scan(
			&c.ID,
			&c.MaterialID,
			&c.UserID,
			&c.ParentID,
			&c.Content,
			&c.Edited,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		// User info would be populated via gRPC call to User Service
		c.UserName = "Unknown"
		comments = append(comments, &c)
	}

	return comments, total, rows.Err()
}

// FindReplies finds all replies to a comment.
func (r *CommentRepository) FindReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*domain.CommentWithUser, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM comments WHERE parent_id = $1 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, parentID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get replies
	query := `
		SELECT id, material_id, user_id, parent_id, content, edited, created_at, updated_at, deleted_at
		FROM comments
		WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, parentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var comments []*domain.CommentWithUser
	for rows.Next() {
		var c domain.CommentWithUser
		if err := rows.Scan(
			&c.ID,
			&c.MaterialID,
			&c.UserID,
			&c.ParentID,
			&c.Content,
			&c.Edited,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		c.UserName = "Unknown"
		comments = append(comments, &c)
	}

	return comments, total, rows.Err()
}

// Update updates a comment.
func (r *CommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	query := `
		UPDATE comments
		SET content = $2, edited = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		comment.ID,
		comment.Content,
		comment.Edited,
		time.Now(),
	)

	return err
}

// Delete soft-deletes a comment.
func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE comments SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id, time.Now())
	return err
}

// CountByMaterialID returns the comment count for a material.
func (r *CommentRepository) CountByMaterialID(ctx context.Context, materialID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM comments WHERE material_id = $1 AND deleted_at IS NULL`
	var count int
	err := r.db.QueryRow(ctx, query, materialID).Scan(&count)
	return count, err
}
