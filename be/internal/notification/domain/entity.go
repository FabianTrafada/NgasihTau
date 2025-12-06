package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypePodInvite         NotificationType = "pod_invite"
	NotificationTypeNewMaterial       NotificationType = "new_material"
	NotificationTypeCommentReply      NotificationType = "comment_reply"
	NotificationTypeNewFollower       NotificationType = "new_follower"
	NotificationTypeMaterialProcessed NotificationType = "material_processed"
)

type NotificationData struct {
	PodID         *uuid.UUID `json:"pod_id,omitempty"`
	PodName       string     `json:"pod_name,omitempty"`
	MaterialID    *uuid.UUID `json:"material_id,omitempty"`
	MaterialTitle string     `json:"material_title,omitempty"`
	CommentID     *uuid.UUID `json:"comment_id,omitempty"`
	UserID        *uuid.UUID `json:"user_id,omitempty"`
	UserName      string     `json:"user_name,omitempty"`
	ActionURL     string     `json:"action_url,omitempty"`
}

func (d *NotificationData) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	return json.Marshal(d)
}

func (d *NotificationData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into NotificationData", value)
	}

	return json.Unmarshal(data, d)
}

type Notification struct {
	ID        uuid.UUID         `json:"id"`
	UserID    uuid.UUID         `json:"user_id"`
	Type      NotificationType  `json:"type"`
	Title     string            `json:"title"`
	Message   string            `json:"message,omitempty"`
	Data      *NotificationData `json:"data,omitempty"`
	Read      bool              `json:"read"`
	CreatedAt time.Time         `json:"created_at"`
}

type NotificationPreference struct {
	UserID            uuid.UUID `json:"user_id"`
	EmailPodInvite    bool      `json:"email_pod_invite"`
	EmailNewMaterial  bool      `json:"email_new_material"`
	EmailCommentReply bool      `json:"email_comment_reply"`
	InAppPodInvite    bool      `json:"inapp_pod_invite"`
	InAppNewMaterial  bool      `json:"inapp_new_material"`
	InAppCommentReply bool      `json:"inapp_comment_reply"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type EmailTemplate struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	HTMLBody  string    `json:"html_body"`
	TextBody  string    `json:"text_body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewNotification(userID uuid.UUID, notifType NotificationType, title, message string, data *NotificationData) *Notification {
	return &Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Message:   message,
		Data:      data,
		Read:      false,
		CreatedAt: time.Now(),
	}
}

func NewDefaultNotificationPreference(userID uuid.UUID) *NotificationPreference {
	return &NotificationPreference{
		UserID:            userID,
		EmailPodInvite:    true,
		EmailNewMaterial:  true,
		EmailCommentReply: true,
		InAppPodInvite:    true,
		InAppNewMaterial:  true,
		InAppCommentReply: true,
		UpdatedAt:         time.Now(),
	}
}

// NewEmailTemplate creates a new email template.
func NewEmailTemplate(name, subject, htmlBody, textBody string) *EmailTemplate {
	now := time.Now()
	return &EmailTemplate{
		ID:        uuid.New(),
		Name:      name,
		Subject:   subject,
		HTMLBody:  htmlBody,
		TextBody:  textBody,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
