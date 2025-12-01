package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// RefreshTokenRepository implements domain.RefreshTokenRepository using PostgreSQL.
type RefreshTokenRepository struct {
	db DBTX
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository.
func NewRefreshTokenRepository(db DBTX) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create creates a new refresh token.
func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, device_info, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	token.CreatedAt = time.Now().UTC()

	_, err := r.db.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.DeviceInfo,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return errors.Internal("failed to create refresh token", err)
	}

	return nil
}

// FindByTokenHash finds a refresh token by its hash.
func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	token := &domain.RefreshToken{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.DeviceInfo,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.NotFoundMsg("refresh token not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to find refresh token", err)
	}

	return token, nil
}


// FindByUserID finds all refresh tokens for a user.
func (r *RefreshTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, expires_at, created_at
		FROM refresh_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Internal("failed to find refresh tokens", err)
	}
	defer rows.Close()

	var tokens []*domain.RefreshToken
	for rows.Next() {
		token := &domain.RefreshToken{}
		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.TokenHash,
			&token.DeviceInfo,
			&token.ExpiresAt,
			&token.CreatedAt,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan refresh token", err)
		}
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Internal("error iterating refresh tokens", err)
	}

	return tokens, nil
}

// Delete removes a refresh token.
func (r *RefreshTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Internal("failed to delete refresh token", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFound("refresh token", id.String())
	}

	return nil
}

// DeleteByTokenHash removes a refresh token by its hash.
func (r *RefreshTokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`

	tag, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return errors.Internal("failed to delete refresh token", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFoundMsg("refresh token not found")
	}

	return nil
}

// DeleteAllByUserID removes all refresh tokens for a user.
func (r *RefreshTokenRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return errors.Internal("failed to delete refresh tokens", err)
	}

	return nil
}


// GetActiveSessions returns all active (non-expired) sessions for a user.
// Used for the "view active sessions" feature (requirement 1.3).
func (r *RefreshTokenRepository) GetActiveSessions(ctx context.Context, userID uuid.UUID, currentTokenHash string) ([]*domain.Session, error) {
	query := `
		SELECT id, device_info, created_at, created_at as last_used_at, token_hash
		FROM refresh_tokens
		WHERE user_id = $1 AND expires_at > $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID, time.Now().UTC())
	if err != nil {
		return nil, errors.Internal("failed to get active sessions", err)
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		session := &domain.Session{}
		var tokenHash string
		err := rows.Scan(
			&session.ID,
			&session.DeviceInfo,
			&session.CreatedAt,
			&session.LastUsedAt,
			&tokenHash,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan session", err)
		}
		session.IsCurrent = tokenHash == currentTokenHash
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Internal("error iterating sessions", err)
	}

	return sessions, nil
}

// DeleteExpired removes all expired refresh tokens.
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1`

	tag, err := r.db.Exec(ctx, query, time.Now().UTC())
	if err != nil {
		return 0, errors.Internal("failed to delete expired refresh tokens", err)
	}

	return tag.RowsAffected(), nil
}
