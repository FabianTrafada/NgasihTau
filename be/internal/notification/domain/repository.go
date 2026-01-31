// Package domain contains the core business entities and repository interfaces
// for the Notification Service. This layer is independent of external frameworks and databases.
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// NotificationRepository defines the interface for notification data access.
// Implements the Repository pattern for data access abstraction.
// Supports Requirement 11: Notification Service.
type NotificationRepository interface {
	// Create creates a new notification.
	Create(ctx context.Context, notification *Notification) error

	// FindByID finds a notification by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Notification, error)

	// FindByUserID returns paginated notifications for a user.
	// Implements Requirement 11.5: paginated notifications with read/unread status.
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Notification, int, error)

	// FindUnreadByUserID returns paginated unread notifications for a user.
	FindUnreadByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Notification, int, error)

	// MarkAsRead marks a notification as read.
	// Implements Requirement 11.6: update read status.
	MarkAsRead(ctx context.Context, id uuid.UUID) error

	// MarkAllAsRead marks all notifications for a user as read.
	// Implements Requirement 11.7: batch marking all notifications as read.
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error

	// CountUnread returns the count of unread notifications for a user.
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)

	// Delete removes a notification.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteOlderThan removes notifications older than the specified time.
	// Used for cleanup of old notifications.
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}

// NotificationPreferenceRepository defines the interface for notification preference data access.
// Implements Requirement 11.4: notification preferences for enabling/disabling notification types.
type NotificationPreferenceRepository interface {
	// Create creates a new notification preference record.
	Create(ctx context.Context, pref *NotificationPreference) error

	// FindByUserID finds notification preferences for a user.
	FindByUserID(ctx context.Context, userID uuid.UUID) (*NotificationPreference, error)

	// Update updates notification preferences.
	Update(ctx context.Context, pref *NotificationPreference) error

	// Upsert creates or updates notification preferences.
	Upsert(ctx context.Context, pref *NotificationPreference) error
}

// EmailTemplateRepository defines the interface for email template data access.
// Used for managing email templates for notifications.
type EmailTemplateRepository interface {
	// Create creates a new email template.
	Create(ctx context.Context, template *EmailTemplate) error

	// FindByID finds an email template by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*EmailTemplate, error)

	// FindByName finds an email template by name.
	FindByName(ctx context.Context, name string) (*EmailTemplate, error)

	// FindAll returns all email templates.
	FindAll(ctx context.Context) ([]*EmailTemplate, error)

	// Update updates an email template.
	Update(ctx context.Context, template *EmailTemplate) error

	// Delete removes an email template.
	Delete(ctx context.Context, id uuid.UUID) error
}
