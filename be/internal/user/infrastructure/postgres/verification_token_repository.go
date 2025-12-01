// Package postgres provides PostgreSQL implementations of the User Service repositories.
package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// VerificationTokenRepository implements domain.VerificationTokenRepository using PostgreSQL.
type VerificationTokenRepository struct {
	db DBTX
}

// NewVerificationTokenRepository creates a new VerificationTokenRepository.
func NewVerificationTokenRepository(db DBTX) *VerificationTokenRepository {
	return &VerificationTokenRepository{db: db}
}

// Create creates a new verification token in the database.
func (r *VerificationTokenRepository) Create(ctx context.Context, token *domain.VerificationToken) error {
	query := `
		INSERT INTO verification_tokens (id, user_id, token_hash, token_type, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now().UTC()
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	token.CreatedAt = now

	_, err := r.db.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.TokenType,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return errors.Internal("failed to create verification token", err)
	}

	return nil
}

// FindByTokenHash finds a verification token by its hash.
func (r *VerificationTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.VerificationToken, error) {
	query := `
		SELECT id, user_id, token_hash, token_type, expires_at, used_at, created_at
		FROM verification_tokens
		WHERE token_hash = $1
	`

	token := &domain.VerificationToken{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.TokenType,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.NotFound("verification_token", "token")
	}
	if err != nil {
		return nil, errors.Internal("failed to find verification token", err)
	}

	return token, nil
}

// FindActiveByUserIDAndType finds an active (unused, non-expired) token for a user and type.
func (r *VerificationTokenRepository) FindActiveByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) (*domain.VerificationToken, error) {
	query := `
		SELECT id, user_id, token_hash, token_type, expires_at, used_at, created_at
		FROM verification_tokens
		WHERE user_id = $1 AND token_type = $2 AND used_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	token := &domain.VerificationToken{}
	err := r.db.QueryRow(ctx, query, userID, tokenType).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.TokenType,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.NotFound("verification_token", userID.String())
	}
	if err != nil {
		return nil, errors.Internal("failed to find verification token", err)
	}

	return token, nil
}

// MarkAsUsed marks a verification token as used.
func (r *VerificationTokenRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE verification_tokens
		SET used_at = $2
		WHERE id = $1 AND used_at IS NULL
	`

	now := time.Now().UTC()
	tag, err := r.db.Exec(ctx, query, id, now)
	if err != nil {
		return errors.Internal("failed to mark token as used", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFound("verification_token", id.String())
	}

	return nil
}

// DeleteByUserIDAndType deletes all tokens for a user and type.
func (r *VerificationTokenRepository) DeleteByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) error {
	query := `DELETE FROM verification_tokens WHERE user_id = $1 AND token_type = $2`

	_, err := r.db.Exec(ctx, query, userID, tokenType)
	if err != nil {
		return errors.Internal("failed to delete verification tokens", err)
	}

	return nil
}

// DeleteExpired removes all expired verification tokens.
func (r *VerificationTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM verification_tokens WHERE expires_at < NOW()`

	tag, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, errors.Internal("failed to delete expired tokens", err)
	}

	return tag.RowsAffected(), nil
}
