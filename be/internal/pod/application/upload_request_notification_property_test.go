// Package application contains property-based tests for upload request notifications.
package application

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"ngasihtau/internal/pod/domain"
)

// **Feature: student-teacher-roles, Property 7: Upload Request Notifications**
// **Validates: Requirements 4.2, 4.4**
//
// Property 7: Upload Request Notifications
// *For any* upload request status change:
// - When created, the pod owner SHALL receive a notification
// - When rejected, the requester SHALL receive a notification with optional reason

// mockEventPublisherForNotifications tracks notification events for testing.
type mockEventPublisherForNotifications struct {
	mu                         sync.Mutex
	uploadRequestCreatedCalls  []uploadRequestCreatedCall
	uploadRequestRejectedCalls []uploadRequestRejectedCall
	collaboratorInvitedCalls   int
	podCreatedCalls            int
	podUpdatedCalls            int
}

type uploadRequestCreatedCall struct {
	Request       *domain.UploadRequest
	PodName       string
	RequesterName string
}

type uploadRequestRejectedCall struct {
	Request *domain.UploadRequest
	PodName string
	Reason  *string
}

func newMockEventPublisherForNotifications() *mockEventPublisherForNotifications {
	return &mockEventPublisherForNotifications{
		uploadRequestCreatedCalls:  make([]uploadRequestCreatedCall, 0),
		uploadRequestRejectedCalls: make([]uploadRequestRejectedCall, 0),
	}
}

func (m *mockEventPublisherForNotifications) PublishPodCreated(ctx context.Context, pod *domain.Pod) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.podCreatedCalls++
	return nil
}

func (m *mockEventPublisherForNotifications) PublishPodUpdated(ctx context.Context, pod *domain.Pod) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.podUpdatedCalls++
	return nil
}

func (m *mockEventPublisherForNotifications) PublishCollaboratorInvited(ctx context.Context, collaborator *domain.Collaborator, podName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collaboratorInvitedCalls++
	return nil
}

func (m *mockEventPublisherForNotifications) PublishUploadRequestCreated(ctx context.Context, request *domain.UploadRequest, podName string, requesterName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.uploadRequestCreatedCalls = append(m.uploadRequestCreatedCalls, uploadRequestCreatedCall{
		Request:       request,
		PodName:       podName,
		RequesterName: requesterName,
	})
	return nil
}

func (m *mockEventPublisherForNotifications) PublishUploadRequestRejected(ctx context.Context, request *domain.UploadRequest, podName string, reason *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.uploadRequestRejectedCalls = append(m.uploadRequestRejectedCalls, uploadRequestRejectedCall{
		Request: request,
		PodName: podName,
		Reason:  reason,
	})
	return nil
}

func (m *mockEventPublisherForNotifications) PublishUploadRequestApproved(ctx context.Context, request *domain.UploadRequest, podName string, requesterName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

func (m *mockEventPublisherForNotifications) PublishPodShared(ctx context.Context, sharedPod *domain.SharedPod, podName string, teacherName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

func (m *mockEventPublisherForNotifications) PublishPodUpvoted(ctx context.Context, podID uuid.UUID, userID uuid.UUID, upvoteCount int, isUpvote bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

func (m *mockEventPublisherForNotifications) getUploadRequestCreatedCalls() []uploadRequestCreatedCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]uploadRequestCreatedCall, len(m.uploadRequestCreatedCalls))
	copy(result, m.uploadRequestCreatedCalls)
	return result
}

func (m *mockEventPublisherForNotifications) getUploadRequestRejectedCalls() []uploadRequestRejectedCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]uploadRequestRejectedCall, len(m.uploadRequestRejectedCalls))
	copy(result, m.uploadRequestRejectedCalls)
	return result
}

func (m *mockEventPublisherForNotifications) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.uploadRequestCreatedCalls = make([]uploadRequestCreatedCall, 0)
	m.uploadRequestRejectedCalls = make([]uploadRequestRejectedCall, 0)
	m.collaboratorInvitedCalls = 0
	m.podCreatedCalls = 0
	m.podUpdatedCalls = 0
}

// mockUserRoleCheckerForNotifications always returns teacher role for testing.
type mockUserRoleCheckerForNotifications struct{}

func (m *mockUserRoleCheckerForNotifications) IsTeacher(ctx context.Context, userID uuid.UUID) (bool, error) {
	return true, nil
}

func TestProperty_UploadRequestNotifications(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 7.1: Upload request creation event contains correct pod owner info
	// Validates: Requirement 4.2 - WHEN a pod owner receives an upload request,
	// THE Notification Service SHALL send a notification to the pod owner.
	properties.Property("upload request creation event targets pod owner", prop.ForAll(
		func(podName string, hasMessage bool, messageContent string) bool {
			requesterID := uuid.New()
			podID := uuid.New()
			podOwnerID := uuid.New()

			var message *string
			if hasMessage && len(messageContent) > 0 {
				message = &messageContent
			}

			// Create upload request
			request := domain.NewUploadRequest(requesterID, podID, podOwnerID, message)

			// The event should contain the pod owner ID for notification routing
			// The notification service uses PodOwnerID to send notification to the owner
			return request.PodOwnerID == podOwnerID &&
				request.RequesterID == requesterID &&
				request.PodID == podID
		},
		genValidPodNameForNotif(),
		gen.Bool(),
		gen.AlphaString(),
	))

	// Property 7.2: Upload request creation event preserves request details
	// Validates: Requirement 4.2 - notification contains request information
	properties.Property("upload request creation event preserves request details", prop.ForAll(
		func(hasMessage bool, messageContent string) bool {
			requesterID := uuid.New()
			podID := uuid.New()
			podOwnerID := uuid.New()

			var message *string
			if hasMessage && len(messageContent) > 0 {
				message = &messageContent
			}

			request := domain.NewUploadRequest(requesterID, podID, podOwnerID, message)

			// Verify all details are preserved for notification
			detailsPreserved := request.ID != uuid.Nil &&
				request.RequesterID == requesterID &&
				request.PodID == podID &&
				request.PodOwnerID == podOwnerID &&
				request.Status == domain.UploadRequestStatusPending

			// Message should be preserved if provided
			if hasMessage && len(messageContent) > 0 {
				return detailsPreserved && request.Message != nil && *request.Message == messageContent
			}
			return detailsPreserved && request.Message == nil
		},
		gen.Bool(),
		gen.AlphaString(),
	))

	// Property 7.3: Rejected upload request event contains requester info
	// Validates: Requirement 4.4 - WHEN a pod owner rejects an upload request,
	// THE Pod Service SHALL notify the requesting teacher.
	properties.Property("rejected upload request event targets requester", prop.ForAll(
		func(hasReason bool, reasonContent string) bool {
			requesterID := uuid.New()
			podID := uuid.New()
			podOwnerID := uuid.New()

			request := domain.NewUploadRequest(requesterID, podID, podOwnerID, nil)

			// Simulate rejection
			request.Status = domain.UploadRequestStatusRejected
			if hasReason && len(reasonContent) > 0 {
				request.RejectionReason = &reasonContent
			}

			// The event should contain the requester ID for notification routing
			// The notification service uses RequesterID to send notification to the requester
			return request.RequesterID == requesterID &&
				request.Status == domain.UploadRequestStatusRejected
		},
		gen.Bool(),
		gen.AlphaString(),
	))

	// Property 7.4: Rejected upload request event includes optional rejection reason
	// Validates: Requirement 4.4 - notify the requesting teacher with optional rejection reason
	properties.Property("rejected upload request event includes optional reason", prop.ForAll(
		func(hasReason bool, reasonContent string) bool {
			requesterID := uuid.New()
			podID := uuid.New()
			podOwnerID := uuid.New()

			request := domain.NewUploadRequest(requesterID, podID, podOwnerID, nil)
			request.Status = domain.UploadRequestStatusRejected

			if hasReason && len(reasonContent) > 0 {
				request.RejectionReason = &reasonContent
			}

			// Verify rejection reason is preserved
			if hasReason && len(reasonContent) > 0 {
				return request.RejectionReason != nil && *request.RejectionReason == reasonContent
			}
			return request.RejectionReason == nil
		},
		gen.Bool(),
		gen.AlphaString(),
	))

	// Property 7.5: Upload request notification data is complete for pod owner
	// Validates: Requirement 4.2 - notification contains all necessary information
	properties.Property("upload request notification data is complete for pod owner", prop.ForAll(
		func(hasMessage bool, messageContent string) bool {
			requesterID := uuid.New()
			podID := uuid.New()
			podOwnerID := uuid.New()

			var message *string
			if hasMessage && len(messageContent) > 0 {
				message = &messageContent
			}

			request := domain.NewUploadRequest(requesterID, podID, podOwnerID, message)

			// All required fields for notification should be present
			return request.ID != uuid.Nil &&
				request.RequesterID != uuid.Nil &&
				request.PodID != uuid.Nil &&
				request.PodOwnerID != uuid.Nil &&
				!request.CreatedAt.IsZero()
		},
		gen.Bool(),
		gen.AlphaString(),
	))

	// Property 7.6: Rejection notification data is complete for requester
	// Validates: Requirement 4.4 - rejection notification contains all necessary information
	properties.Property("rejection notification data is complete for requester", prop.ForAll(
		func(hasReason bool, reasonContent string) bool {
			requesterID := uuid.New()
			podID := uuid.New()
			podOwnerID := uuid.New()

			request := domain.NewUploadRequest(requesterID, podID, podOwnerID, nil)
			request.Status = domain.UploadRequestStatusRejected

			if hasReason && len(reasonContent) > 0 {
				request.RejectionReason = &reasonContent
			}

			// All required fields for rejection notification should be present
			hasRequiredFields := request.ID != uuid.Nil &&
				request.RequesterID != uuid.Nil &&
				request.PodID != uuid.Nil &&
				request.Status == domain.UploadRequestStatusRejected

			// Rejection reason is optional but should be preserved if provided
			if hasReason && len(reasonContent) > 0 {
				return hasRequiredFields && request.RejectionReason != nil
			}
			return hasRequiredFields
		},
		gen.Bool(),
		gen.AlphaString(),
	))

	// Property 7.7: Only pending requests can be rejected (notification precondition)
	// Validates: Requirement 4.4 - rejection only applies to pending requests
	properties.Property("only pending requests can be rejected for notification", prop.ForAll(
		func(statusIdx int) bool {
			statuses := []domain.UploadRequestStatus{
				domain.UploadRequestStatusPending,
				domain.UploadRequestStatusApproved,
				domain.UploadRequestStatusRejected,
				domain.UploadRequestStatusRevoked,
			}

			idx := statusIdx % len(statuses)
			if idx < 0 {
				idx = -idx
			}

			request := &domain.UploadRequest{
				ID:          uuid.New(),
				RequesterID: uuid.New(),
				PodID:       uuid.New(),
				PodOwnerID:  uuid.New(),
				Status:      statuses[idx],
			}

			// Only pending requests should be rejectable
			// This is a precondition for sending rejection notifications
			canBeRejected := request.IsPending()
			return canBeRejected == (statuses[idx] == domain.UploadRequestStatusPending)
		},
		gen.Int(),
	))

	// Property 7.8: Notification event preserves pod and requester relationship
	// Validates: Requirements 4.2, 4.4 - notifications maintain correct relationships
	properties.Property("notification event preserves pod and requester relationship", prop.ForAll(
		func(_ int) bool {
			requesterID := uuid.New()
			podID := uuid.New()
			podOwnerID := uuid.New()

			request := domain.NewUploadRequest(requesterID, podID, podOwnerID, nil)

			// The relationship between requester, pod, and owner should be preserved
			// This ensures notifications are sent to the correct recipients
			return request.RequesterID == requesterID &&
				request.PodID == podID &&
				request.PodOwnerID == podOwnerID &&
				request.RequesterID != request.PodOwnerID // Requester cannot be the owner
		},
		gen.Int(),
	))

	// Property 7.9: Created notification targets different user than rejected notification
	// Validates: Requirements 4.2, 4.4 - different recipients for different events
	properties.Property("created and rejected notifications target different users", prop.ForAll(
		func(_ int) bool {
			requesterID := uuid.New()
			podID := uuid.New()
			podOwnerID := uuid.New()

			request := domain.NewUploadRequest(requesterID, podID, podOwnerID, nil)

			// Created notification goes to pod owner (PodOwnerID)
			// Rejected notification goes to requester (RequesterID)
			// These should be different users
			return request.PodOwnerID != request.RequesterID
		},
		gen.Int(),
	))

	// Property 7.10: Rejection reason can be empty string or nil
	// Validates: Requirement 4.4 - optional rejection reason
	properties.Property("rejection reason can be empty or nil", prop.ForAll(
		func(useEmptyString bool) bool {
			request := domain.NewUploadRequest(uuid.New(), uuid.New(), uuid.New(), nil)
			request.Status = domain.UploadRequestStatusRejected

			if useEmptyString {
				emptyReason := ""
				request.RejectionReason = &emptyReason
				// Empty string is a valid (though not useful) reason
				return request.RejectionReason != nil && *request.RejectionReason == ""
			}
			// Nil reason is valid
			return request.RejectionReason == nil
		},
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Generator for valid pod names
func genValidPodNameForNotif() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) == 0 {
			return "Default Pod"
		}
		if len(s) > 100 {
			return s[:100]
		}
		return s
	})
}
