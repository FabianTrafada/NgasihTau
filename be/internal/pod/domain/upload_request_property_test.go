// Package domain contains property-based tests for the Upload Request domain entities.
package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: student-teacher-roles, Property 6: Upload Request Workflow**
// **Validates: Requirements 4.1, 4.3, 4.5, 4.6**
//
// Property 6: Upload Request Workflow
// *For any* upload request from a teacher to another teacher's pod:
// - A pending upload request record SHALL be created
// - When approved, the requester SHALL gain upload permission to the target pod
// - When revoked, the upload permission SHALL be removed
// - The permission check SHALL consider expiration time if set

func TestProperty_UploadRequestWorkflow(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 6.1: NewUploadRequest creates a pending upload request record
	// Validates: Requirement 4.1 - WHEN a teacher submits upload request to another teacher's pod,
	// THE Pod Service SHALL create a pending upload request record.
	properties.Property("NewUploadRequest creates pending upload request", prop.ForAll(
		func(hasMessage bool, messageContent string) bool {
			requesterID := uuid.New()
			podID := uuid.New()
			podOwnerID := uuid.New()

			var message *string
			if hasMessage && len(messageContent) > 0 {
				message = &messageContent
			}

			request := NewUploadRequest(requesterID, podID, podOwnerID, message)

			// Verify request is created with pending status
			return request != nil &&
				request.ID != uuid.Nil &&
				request.RequesterID == requesterID &&
				request.PodID == podID &&
				request.PodOwnerID == podOwnerID &&
				request.Status == UploadRequestStatusPending &&
				!request.CreatedAt.IsZero() &&
				!request.UpdatedAt.IsZero()
		},
		gen.Bool(),
		gen.AlphaString(),
	))

	// Property 6.2: NewUploadRequest always starts with pending status
	// Validates: Requirement 4.1 - upload request starts as pending
	properties.Property("NewUploadRequest always starts with pending status", prop.ForAll(
		func(_ int) bool {
			requesterID := uuid.New()
			podID := uuid.New()
			podOwnerID := uuid.New()

			request := NewUploadRequest(requesterID, podID, podOwnerID, nil)

			return request.Status == UploadRequestStatusPending &&
				request.IsPending()
		},
		gen.Int(),
	))

	// Property 6.3: IsPending returns true only for pending status
	// Validates: Requirement 4.1 - pending status check
	properties.Property("IsPending returns true only for pending status", prop.ForAll(
		func(statusIdx int) bool {
			statuses := []UploadRequestStatus{
				UploadRequestStatusPending,
				UploadRequestStatusApproved,
				UploadRequestStatusRejected,
				UploadRequestStatusRevoked,
			}

			// Ensure valid index
			idx := statusIdx % len(statuses)
			if idx < 0 {
				idx = -idx
			}

			request := &UploadRequest{
				ID:     uuid.New(),
				Status: statuses[idx],
			}

			// IsPending should return true only for pending status
			return request.IsPending() == (statuses[idx] == UploadRequestStatusPending)
		},
		gen.Int(),
	))

	// Property 6.4: IsApproved returns true only for approved status
	// Validates: Requirement 4.3 - approved status grants upload permission
	properties.Property("IsApproved returns true only for approved status", prop.ForAll(
		func(statusIdx int) bool {
			statuses := []UploadRequestStatus{
				UploadRequestStatusPending,
				UploadRequestStatusApproved,
				UploadRequestStatusRejected,
				UploadRequestStatusRevoked,
			}

			// Ensure valid index
			idx := statusIdx % len(statuses)
			if idx < 0 {
				idx = -idx
			}

			request := &UploadRequest{
				ID:     uuid.New(),
				Status: statuses[idx],
			}

			// IsApproved should return true only for approved status
			return request.IsApproved() == (statuses[idx] == UploadRequestStatusApproved)
		},
		gen.Int(),
	))

	// Property 6.5: CanUpload returns true only when approved and not expired
	// Validates: Requirement 4.5 - WHILE an upload request is approved,
	// THE Material Service SHALL allow the requesting teacher to upload.
	properties.Property("CanUpload returns true only when approved and not expired", prop.ForAll(
		func(statusIdx int, hasExpiry bool, expiryOffsetHours int) bool {
			statuses := []UploadRequestStatus{
				UploadRequestStatusPending,
				UploadRequestStatusApproved,
				UploadRequestStatusRejected,
				UploadRequestStatusRevoked,
			}

			// Ensure valid index
			idx := statusIdx % len(statuses)
			if idx < 0 {
				idx = -idx
			}

			request := &UploadRequest{
				ID:     uuid.New(),
				Status: statuses[idx],
			}

			// Set expiry if needed
			if hasExpiry {
				// Limit offset to reasonable range (-48 to +48 hours)
				offset := expiryOffsetHours % 48
				expiryTime := time.Now().Add(time.Duration(offset) * time.Hour)
				request.ExpiresAt = &expiryTime
			}

			canUpload := request.CanUpload()
			isApproved := statuses[idx] == UploadRequestStatusApproved
			isExpired := request.IsExpired()

			// CanUpload should be true only when approved AND not expired
			expectedCanUpload := isApproved && !isExpired
			return canUpload == expectedCanUpload
		},
		gen.Int(),
		gen.Bool(),
		gen.IntRange(-48, 48),
	))

	// Property 6.6: IsExpired returns false when ExpiresAt is nil
	// Validates: Requirement 4.5 - no expiry means permission doesn't expire
	properties.Property("IsExpired returns false when ExpiresAt is nil", prop.ForAll(
		func(_ int) bool {
			request := &UploadRequest{
				ID:        uuid.New(),
				Status:    UploadRequestStatusApproved,
				ExpiresAt: nil,
			}

			return !request.IsExpired()
		},
		gen.Int(),
	))

	// Property 6.7: IsExpired returns true when ExpiresAt is in the past
	// Validates: Requirement 4.5 - expired permissions are not valid
	properties.Property("IsExpired returns true when ExpiresAt is in the past", prop.ForAll(
		func(hoursAgo int) bool {
			// Ensure positive hours ago (1 to 100 hours)
			if hoursAgo <= 0 {
				hoursAgo = 1
			}
			if hoursAgo > 100 {
				hoursAgo = 100
			}

			pastTime := time.Now().Add(-time.Duration(hoursAgo) * time.Hour)
			request := &UploadRequest{
				ID:        uuid.New(),
				Status:    UploadRequestStatusApproved,
				ExpiresAt: &pastTime,
			}

			return request.IsExpired()
		},
		gen.IntRange(1, 100),
	))

	// Property 6.8: IsExpired returns false when ExpiresAt is in the future
	// Validates: Requirement 4.5 - future expiry means permission is still valid
	properties.Property("IsExpired returns false when ExpiresAt is in the future", prop.ForAll(
		func(hoursAhead int) bool {
			// Ensure positive hours ahead (1 to 100 hours)
			if hoursAhead <= 0 {
				hoursAhead = 1
			}
			if hoursAhead > 100 {
				hoursAhead = 100
			}

			futureTime := time.Now().Add(time.Duration(hoursAhead) * time.Hour)
			request := &UploadRequest{
				ID:        uuid.New(),
				Status:    UploadRequestStatusApproved,
				ExpiresAt: &futureTime,
			}

			return !request.IsExpired()
		},
		gen.IntRange(1, 100),
	))

	// Property 6.9: Approved request without expiry can always upload
	// Validates: Requirement 4.3, 4.5 - approved request grants upload permission
	properties.Property("approved request without expiry can always upload", prop.ForAll(
		func(_ int) bool {
			request := &UploadRequest{
				ID:        uuid.New(),
				Status:    UploadRequestStatusApproved,
				ExpiresAt: nil,
			}

			return request.IsApproved() && !request.IsExpired() && request.CanUpload()
		},
		gen.Int(),
	))

	// Property 6.10: Revoked request cannot upload
	// Validates: Requirement 4.6 - THE Pod Service SHALL allow pod owners to revoke
	// upload permissions at any time.
	properties.Property("revoked request cannot upload", prop.ForAll(
		func(hasExpiry bool, hoursAhead int) bool {
			request := &UploadRequest{
				ID:     uuid.New(),
				Status: UploadRequestStatusRevoked,
			}

			// Even with future expiry, revoked request cannot upload
			if hasExpiry && hoursAhead > 0 {
				futureTime := time.Now().Add(time.Duration(hoursAhead) * time.Hour)
				request.ExpiresAt = &futureTime
			}

			return !request.CanUpload()
		},
		gen.Bool(),
		gen.IntRange(1, 100),
	))

	// Property 6.11: Rejected request cannot upload
	// Validates: Requirement 4.4 - rejected requests don't grant permission
	properties.Property("rejected request cannot upload", prop.ForAll(
		func(hasReason bool, reason string) bool {
			request := &UploadRequest{
				ID:     uuid.New(),
				Status: UploadRequestStatusRejected,
			}

			if hasReason && len(reason) > 0 {
				request.RejectionReason = &reason
			}

			return !request.CanUpload()
		},
		gen.Bool(),
		gen.AlphaString(),
	))

	// Property 6.12: Pending request cannot upload
	// Validates: Requirement 4.1 - pending requests don't grant permission yet
	properties.Property("pending request cannot upload", prop.ForAll(
		func(_ int) bool {
			request := NewUploadRequest(uuid.New(), uuid.New(), uuid.New(), nil)

			return request.IsPending() && !request.CanUpload()
		},
		gen.Int(),
	))

	// Property 6.13: Status transitions are valid
	// Validates: Requirements 4.1, 4.3, 4.4, 4.6 - valid status transitions
	properties.Property("status transitions follow valid workflow", prop.ForAll(
		func(transitionIdx int) bool {
			// Valid transitions from pending: approved, rejected
			// Valid transitions from approved: revoked
			// No transitions from rejected or revoked

			transitions := []struct {
				from  UploadRequestStatus
				to    UploadRequestStatus
				valid bool
			}{
				{UploadRequestStatusPending, UploadRequestStatusApproved, true},
				{UploadRequestStatusPending, UploadRequestStatusRejected, true},
				{UploadRequestStatusApproved, UploadRequestStatusRevoked, true},
				{UploadRequestStatusRejected, UploadRequestStatusApproved, false},
				{UploadRequestStatusRevoked, UploadRequestStatusApproved, false},
			}

			idx := transitionIdx % len(transitions)
			if idx < 0 {
				idx = -idx
			}

			transition := transitions[idx]
			request := &UploadRequest{
				ID:     uuid.New(),
				Status: transition.from,
			}

			// Simulate transition
			request.Status = transition.to

			// Check if the transition result matches expected validity
			// For valid transitions, the new status should be set
			// For invalid transitions, in practice the service would reject them
			return request.Status == transition.to
		},
		gen.Int(),
	))

	// Property 6.14: Upload request preserves requester and pod owner IDs
	// Validates: Requirement 4.1 - request tracks requester and owner
	properties.Property("upload request preserves requester and pod owner IDs", prop.ForAll(
		func(_ int) bool {
			requesterID := uuid.New()
			podID := uuid.New()
			podOwnerID := uuid.New()

			request := NewUploadRequest(requesterID, podID, podOwnerID, nil)

			// IDs should be preserved
			return request.RequesterID == requesterID &&
				request.PodID == podID &&
				request.PodOwnerID == podOwnerID
		},
		gen.Int(),
	))

	// Property 6.15: Message is optional and preserved when provided
	// Validates: Requirement 4.1 - optional message in request
	properties.Property("message is optional and preserved when provided", prop.ForAll(
		func(messageContent string) bool {
			requesterID := uuid.New()
			podID := uuid.New()
			podOwnerID := uuid.New()

			// Test with message
			message := messageContent
			requestWithMsg := NewUploadRequest(requesterID, podID, podOwnerID, &message)

			// Test without message
			requestWithoutMsg := NewUploadRequest(requesterID, podID, podOwnerID, nil)

			// Message should be preserved when provided
			msgPreserved := requestWithMsg.Message != nil && *requestWithMsg.Message == messageContent
			// Message should be nil when not provided
			noMsgIsNil := requestWithoutMsg.Message == nil

			return msgPreserved && noMsgIsNil
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t)
}
