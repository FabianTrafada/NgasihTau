package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/ai/domain"
)

type ChatMessageRepository struct {
	db *pgxpool.Pool
}

func NewChatMessageRepository(db *pgxpool.Pool) *ChatMessageRepository {
	return &ChatMessageRepository{db: db}
}

func (r *ChatMessageRepository) Create(ctx context.Context, message *domain.ChatMessage) error {
	query := `
		INSERT INTO chat_messages (id, session_id, role, content, sources, feedback, feedback_text, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		message.ID,
		message.SessionID,
		message.Role,
		message.Content,
		domain.Sources(message.Sources),
		message.Feedback,
		message.FeedbackText,
		message.CreatedAt,
	)

	return err
}

func (r *ChatMessageRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ChatMessage, error) {
	query := `
		SELECT id, session_id, role, content, sources, feedback, feedback_text, created_at
		FROM chat_messages
		WHERE id = $1
	`

	var message domain.ChatMessage
	var sources domain.Sources
	err := r.db.QueryRow(ctx, query, id).Scan(
		&message.ID,
		&message.SessionID,
		&message.Role,
		&message.Content,
		&sources,
		&message.Feedback,
		&message.FeedbackText,
		&message.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	message.Sources = sources
	return &message, nil
}

func (r *ChatMessageRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]domain.ChatMessage, error) {
	query := `
		SELECT id, session_id, role, content, sources, feedback, feedback_text, created_at
		FROM chat_messages
		WHERE session_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, sessionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.ChatMessage
	for rows.Next() {
		var message domain.ChatMessage
		var sources domain.Sources
		err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&message.Role,
			&message.Content,
			&sources,
			&message.Feedback,
			&message.FeedbackText,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		message.Sources = sources
		messages = append(messages, message)
	}

	return messages, rows.Err()
}

func (r *ChatMessageRepository) UpdateFeedback(ctx context.Context, id uuid.UUID, feedback domain.FeedbackType, feedbackText *string) error {
	query := `
		UPDATE chat_messages
		SET feedback = $2, feedback_text = $3
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, feedback, feedbackText)
	return err
}

func (r *ChatMessageRepository) CountBySessionID(ctx context.Context, sessionID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM chat_messages WHERE session_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, sessionID).Scan(&count)
	return count, err
}
