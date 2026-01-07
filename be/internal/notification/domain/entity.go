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

	// Student-Teacher roles notification types
	NotificationTypeUploadRequest         NotificationType = "upload_request"          // Teacher requests upload permission
	NotificationTypeUploadRequestApproved NotificationType = "upload_request_approved" // Upload request approved
	NotificationTypeUploadRequestRejected NotificationType = "upload_request_rejected" // Upload request rejected
	NotificationTypePodShared             NotificationType = "pod_shared"              // Teacher shares pod with student
	NotificationTypeTeacherVerified       NotificationType = "teacher_verified"        // Teacher verification approved
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

	// Student-Teacher roles notification fields
	UploadRequestID *uuid.UUID `json:"upload_request_id,omitempty"` // For upload request notifications
	RejectionReason *string    `json:"rejection_reason,omitempty"`  // For rejection notifications
	SharedByID      *uuid.UUID `json:"shared_by_id,omitempty"`      // For shared pod notifications
	SharedByName    string     `json:"shared_by_name,omitempty"`    // For shared pod notifications
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

// NewUploadRequestNotification creates a notification for when a teacher requests upload permission.
func NewUploadRequestNotification(podOwnerID uuid.UUID, requesterName string, podName string, podID uuid.UUID, uploadRequestID uuid.UUID) *Notification {
	return NewNotification(
		podOwnerID,
		NotificationTypeUploadRequest,
		"New Upload Request",
		fmt.Sprintf("%s has requested permission to upload materials to your pod '%s'", requesterName, podName),
		&NotificationData{
			PodID:           &podID,
			PodName:         podName,
			UserName:        requesterName,
			UploadRequestID: &uploadRequestID,
			ActionURL:       fmt.Sprintf("/pods/%s/upload-requests", podID.String()),
		},
	)
}

// NewUploadRequestApprovedNotification creates a notification for when an upload request is approved.
func NewUploadRequestApprovedNotification(requesterID uuid.UUID, podOwnerName string, podName string, podID uuid.UUID, uploadRequestID uuid.UUID) *Notification {
	return NewNotification(
		requesterID,
		NotificationTypeUploadRequestApproved,
		"Upload Request Approved",
		fmt.Sprintf("Your request to upload materials to '%s' has been approved by %s", podName, podOwnerName),
		&NotificationData{
			PodID:           &podID,
			PodName:         podName,
			UserName:        podOwnerName,
			UploadRequestID: &uploadRequestID,
			ActionURL:       fmt.Sprintf("/pods/%s", podID.String()),
		},
	)
}

// NewUploadRequestRejectedNotification creates a notification for when an upload request is rejected.
func NewUploadRequestRejectedNotification(requesterID uuid.UUID, podOwnerName string, podName string, podID uuid.UUID, uploadRequestID uuid.UUID, reason *string) *Notification {
	message := fmt.Sprintf("Your request to upload materials to '%s' has been rejected by %s", podName, podOwnerName)
	if reason != nil && *reason != "" {
		message = fmt.Sprintf("%s. Reason: %s", message, *reason)
	}

	return NewNotification(
		requesterID,
		NotificationTypeUploadRequestRejected,
		"Upload Request Rejected",
		message,
		&NotificationData{
			PodID:           &podID,
			PodName:         podName,
			UserName:        podOwnerName,
			UploadRequestID: &uploadRequestID,
			RejectionReason: reason,
			ActionURL:       fmt.Sprintf("/pods/%s", podID.String()),
		},
	)
}

// NewPodSharedNotification creates a notification for when a teacher shares a pod with a student.
func NewPodSharedNotification(studentID uuid.UUID, teacherID uuid.UUID, teacherName string, podName string, podID uuid.UUID, message *string) *Notification {
	notifMessage := fmt.Sprintf("%s has shared the pod '%s' with you", teacherName, podName)
	if message != nil && *message != "" {
		notifMessage = fmt.Sprintf("%s: \"%s\"", notifMessage, *message)
	}

	return NewNotification(
		studentID,
		NotificationTypePodShared,
		"Pod Shared With You",
		notifMessage,
		&NotificationData{
			PodID:        &podID,
			PodName:      podName,
			SharedByID:   &teacherID,
			SharedByName: teacherName,
			ActionURL:    fmt.Sprintf("/pods/%s", podID.String()),
		},
	)
}

// NewTeacherVerifiedNotification creates a notification for when a teacher verification is approved.
func NewTeacherVerifiedNotification(userID uuid.UUID) *Notification {
	return NewNotification(
		userID,
		NotificationTypeTeacherVerified,
		"Teacher Verification Approved",
		"Congratulations! Your teacher verification has been approved. You can now create verified knowledge pods.",
		&NotificationData{
			ActionURL: "/dashboard",
		},
	)
}
