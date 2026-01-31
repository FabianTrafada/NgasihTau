package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// OAuthAccountRepository implements domain.OAuthAccountRepository using PostgreSQL.
type OAuthAccountRepository struct {
	db DBTX
}

// NewOAuthRepository creates a new OAuthAccountRepository.
func NewOAuthRepository(db DBTX) *OAuthAccountRepository {
	return &OAuthAccountRepository{db: db}
}

// Create creates a new OAuth account link.
func (r *OAuthAccountRepository) Create(ctx context.Context, account *domain.OAuthAccount) error {
	query := `
		INSERT INTO oauth_accounts (id, user_id, provider, provider_user_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}
	account.CreatedAt = time.Now().UTC()

	_, err := r.db.Exec(ctx, query,
		account.ID,
		account.UserID,
		account.Provider,
		account.ProviderUserID,
		account.CreatedAt,
	)
	if err != nil {
		return errors.Internal("failed to create oauth account", err)
	}

	return nil
}

// FindByProvider finds an OAuth account by provider and provider user ID.
func (r *OAuthAccountRepository) FindByProvider(ctx context.Context, provider domain.OAuthProvider, providerUserID string) (*domain.OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, created_at
		FROM oauth_accounts
		WHERE provider = $1 AND provider_user_id = $2
	`

	account := &domain.OAuthAccount{}
	err := r.db.QueryRow(ctx, query, provider, providerUserID).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderUserID,
		&account.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.NotFoundMsg("oauth account not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to find oauth account", err)
	}

	return account, nil
}

// FindByUserID finds all OAuth accounts for a user.
func (r *OAuthAccountRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, created_at
		FROM oauth_accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Internal("failed to find oauth accounts", err)
	}
	defer rows.Close()

	var accounts []*domain.OAuthAccount
	for rows.Next() {
		account := &domain.OAuthAccount{}
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Provider,
			&account.ProviderUserID,
			&account.CreatedAt,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan oauth account", err)
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Internal("error iterating oauth accounts", err)
	}

	return accounts, nil
}

// Delete removes an OAuth account link.
func (r *OAuthAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM oauth_accounts WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Internal("failed to delete oauth account", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFound("oauth account", id.String())
	}

	return nil
}
