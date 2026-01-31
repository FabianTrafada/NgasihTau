package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/notification/domain"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, type, title, message, data, read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.Data,
		notification.Read,
		notification.CreatedAt,
	)

	return err
}

func (r *NotificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data, read, created_at
		FROM notifications
		WHERE id = $1
	`

	var notification domain.Notification
	err := r.db.QueryRow(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.Data,
		&notification.Read,
		&notification.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &notification, nil
}

func (r *NotificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Notification, int, error) {
	countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, type, title, message, data, read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&n.Title,
			&n.Message,
			&n.Data,
			&n.Read,
			&n.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		notifications = append(notifications, &n)
	}

	return notifications, total, nil
}

func (r *NotificationRepository) FindUnreadByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Notification, int, error) {
	countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, type, title, message, data, read, created_at
		FROM notifications
		WHERE user_id = $1 AND read = false
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&n.Title,
			&n.Message,
			&n.Data,
			&n.Read,
			&n.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		notifications = append(notifications, &n)
	}

	return notifications, total, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE notifications SET read = true WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE notifications SET read = true WHERE user_id = $1 AND read = false`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *NotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *NotificationRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM notifications WHERE created_at < $1`
	result, err := r.db.Exec(ctx, query, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
