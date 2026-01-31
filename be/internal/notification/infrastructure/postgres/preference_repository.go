package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/notification/domain"
)

type NotificationPreferenceRepository struct {
	db *pgxpool.Pool
}

func NewNotificationPreferenceRepository(db *pgxpool.Pool) *NotificationPreferenceRepository {
	return &NotificationPreferenceRepository{db: db}
}

func (r *NotificationPreferenceRepository) Create(ctx context.Context, pref *domain.NotificationPreference) error {
	query := `
		INSERT INTO notification_preferences (
			user_id, email_pod_invite, email_new_material, email_comment_reply,
			inapp_pod_invite, inapp_new_material, inapp_comment_reply, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		pref.UserID,
		pref.EmailPodInvite,
		pref.EmailNewMaterial,
		pref.EmailCommentReply,
		pref.InAppPodInvite,
		pref.InAppNewMaterial,
		pref.InAppCommentReply,
		pref.UpdatedAt,
	)

	return err
}

func (r *NotificationPreferenceRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreference, error) {
	query := `
		SELECT user_id, email_pod_invite, email_new_material, email_comment_reply,
			   inapp_pod_invite, inapp_new_material, inapp_comment_reply, updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`

	var pref domain.NotificationPreference
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&pref.UserID,
		&pref.EmailPodInvite,
		&pref.EmailNewMaterial,
		&pref.EmailCommentReply,
		&pref.InAppPodInvite,
		&pref.InAppNewMaterial,
		&pref.InAppCommentReply,
		&pref.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &pref, nil
}

func (r *NotificationPreferenceRepository) Update(ctx context.Context, pref *domain.NotificationPreference) error {
	pref.UpdatedAt = time.Now()

	query := `
		UPDATE notification_preferences
		SET email_pod_invite = $2, email_new_material = $3, email_comment_reply = $4,
			inapp_pod_invite = $5, inapp_new_material = $6, inapp_comment_reply = $7,
			updated_at = $8
		WHERE user_id = $1
	`

	_, err := r.db.Exec(ctx, query,
		pref.UserID,
		pref.EmailPodInvite,
		pref.EmailNewMaterial,
		pref.EmailCommentReply,
		pref.InAppPodInvite,
		pref.InAppNewMaterial,
		pref.InAppCommentReply,
		pref.UpdatedAt,
	)

	return err
}

func (r *NotificationPreferenceRepository) Upsert(ctx context.Context, pref *domain.NotificationPreference) error {
	pref.UpdatedAt = time.Now()

	query := `
		INSERT INTO notification_preferences (
			user_id, email_pod_invite, email_new_material, email_comment_reply,
			inapp_pod_invite, inapp_new_material, inapp_comment_reply, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) DO UPDATE SET
			email_pod_invite = EXCLUDED.email_pod_invite,
			email_new_material = EXCLUDED.email_new_material,
			email_comment_reply = EXCLUDED.email_comment_reply,
			inapp_pod_invite = EXCLUDED.inapp_pod_invite,
			inapp_new_material = EXCLUDED.inapp_new_material,
			inapp_comment_reply = EXCLUDED.inapp_comment_reply,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		pref.UserID,
		pref.EmailPodInvite,
		pref.EmailNewMaterial,
		pref.EmailCommentReply,
		pref.InAppPodInvite,
		pref.InAppNewMaterial,
		pref.InAppCommentReply,
		pref.UpdatedAt,
	)

	return err
}
