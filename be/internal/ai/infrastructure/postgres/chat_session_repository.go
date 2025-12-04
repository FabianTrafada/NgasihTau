package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/ai/domain"
)

type ChatSessionRepository struct {
	db *pgxpool.Pool
}

func NewChatSessionRepository(db *pgxpool.Pool) *ChatSessionRepository {
	return &ChatSessionRepository{db: db}
}

func (r *ChatSessionRepository) Create(ctx context.Context, session *domain.ChatSession) error {
	query := `
		INSERT INTO chat_sessions (id, user_id, material_id, pod_id, mode, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.MaterialID,
		session.PodID,
		session.Mode,
		session.CreatedAt,
		session.UpdatedAt,
	)

	return err
}

func (r *ChatSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ChatSession, error) {
	query := `
		SELECT id, user_id, material_id, pod_id, mode, created_at, updated_at
		FROM chat_sessions
		WHERE id = $1
	`

	var session domain.ChatSession
	err := r.db.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.MaterialID,
		&session.PodID,
		&session.Mode,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *ChatSessionRepository) FindByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) (*domain.ChatSession, error) {
	query := `
		SELECT id, user_id, material_id, pod_id, mode, created_at, updated_at
		FROM chat_sessions
		WHERE user_id = $1 AND material_id = $2
		ORDER BY updated_at DESC
		LIMIT 1
	`

	var session domain.ChatSession
	err := r.db.QueryRow(ctx, query, userID, materialID).Scan(
		&session.ID,
		&session.UserID,
		&session.MaterialID,
		&session.PodID,
		&session.Mode,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *ChatSessionRepository) FindByUserAndPod(ctx context.Context, userID, podID uuid.UUID) (*domain.ChatSession, error) {
	query := `
		SELECT id, user_id, material_id, pod_id, mode, created_at, updated_at
		FROM chat_sessions
		WHERE user_id = $1 AND pod_id = $2 AND mode = 'pod'
		ORDER BY updated_at DESC
		LIMIT 1
	`

	var session domain.ChatSession
	err := r.db.QueryRow(ctx, query, userID, podID).Scan(
		&session.ID,
		&session.UserID,
		&session.MaterialID,
		&session.PodID,
		&session.Mode,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *ChatSessionRepository) Update(ctx context.Context, session *domain.ChatSession) error {
	query := `
		UPDATE chat_sessions
		SET updated_at = $2
		WHERE id = $1
	`

	session.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, query, session.ID, session.UpdatedAt)
	return err
}

func (r *ChatSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM chat_sessions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
