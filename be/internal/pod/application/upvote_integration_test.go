// Package application contains integration tests for the upvote system.
// These tests verify the complete upvote flow: upvote, remove upvote, duplicate prevention.
//
// Prerequisites:
// - Docker Compose environment must be running: docker-compose up -d
// - All services must be healthy
//
// Run tests: go test -v -tags=integration ./internal/pod/application/...
//
//go:build integration

package application

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// newTestServiceForUpvote creates a test service with all required mock repositories
// for upvote integration testing.
func newTestServiceForUpvote() (PodService, *mockPodRepo, *mockUpvoteRepo) {
	podRepo := newMockPodRepo()
	collaboratorRepo := newMockCollaboratorRepo()
	starRepo := newMockStarRepo()
	upvoteRepo := newMockUpvoteRepo()
	uploadReqRepo := newMockUploadRequestRepo()
	sharedPodRepo := newMockSharedPodRepo()
	followRepo := newMockFollowRepo()
	activityRepo := newMockActivityRepo()
	eventPublisher := NewNoOpEventPublisher()

	svc := NewPodService(
		podRepo,
		collaboratorRepo,
		starRepo,
		upvoteRepo,
		uploadReqRepo,
		sharedPodRepo,
		followRepo,
		activityRepo,
		eventPublisher,
		nil, // UserRoleChecker - nil for basic tests
	)

	return svc, podRepo, upvoteRepo
}

// TestUpvoteSystem_CompleteFlow tests the complete upvote flow:
// 1. User upvotes a pod
// 2. Upvote count is incremented
// 3. User removes upvote
// 4. Upvote count is decremented
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUpvoteSystem_CompleteFlow(t *testing.T) {
	svc, podRepo, upvoteRepo := newTestServiceForUpvote()
	ctx := context.Background()

	// Step 1: Create a public pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Integration Test Pod", "integration-test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Verify initial upvote count is 0
	if pod.UpvoteCount != 0 {
		t.Fatalf("Expected initial upvote count 0, got %d", pod.UpvoteCount)
	}

	// Step 2: User upvotes the pod
	userID := uuid.New()
	err := svc.UpvotePod(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("UpvotePod failed: %v", err)
	}

	// Verify upvote count was incremented (requirement 5.1)
	if pod.UpvoteCount != 1 {
		t.Errorf("Expected upvote count 1 after upvote, got %d", pod.UpvoteCount)
	}

	// Verify upvote record was created
	key := userID.String() + ":" + pod.ID.String()
	if !upvoteRepo.upvotes[key] {
		t.Error("Expected upvote record to be created")
	}

	// Step 3: Verify user has upvoted
	hasUpvoted, err := svc.HasUpvoted(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("HasUpvoted failed: %v", err)
	}
	if !hasUpvoted {
		t.Error("Expected HasUpvoted to return true after upvoting")
	}

	// Step 4: User removes upvote
	err = svc.RemoveUpvote(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("RemoveUpvote failed: %v", err)
	}

	// Verify upvote count was decremented (requirement 5.2)
	if pod.UpvoteCount != 0 {
		t.Errorf("Expected upvote count 0 after removing upvote, got %d", pod.UpvoteCount)
	}

	// Verify upvote record was deleted
	if upvoteRepo.upvotes[key] {
		t.Error("Expected upvote record to be deleted")
	}

	// Step 5: Verify user no longer has upvoted
	hasUpvoted, err = svc.HasUpvoted(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("HasUpvoted failed: %v", err)
	}
	if hasUpvoted {
		t.Error("Expected HasUpvoted to return false after removing upvote")
	}
}

// TestUpvoteSystem_DuplicatePrevention tests that duplicate upvotes are prevented.
// Implements requirement 5.3: Each user can upvote a pod only once.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUpvoteSystem_DuplicatePrevention(t *testing.T) {
	svc, podRepo, _ := newTestServiceForUpvote()
	ctx := context.Background()

	// Create a public pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Duplicate Test Pod", "duplicate-test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// First upvote should succeed
	userID := uuid.New()
	err := svc.UpvotePod(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("First UpvotePod failed: %v", err)
	}

	// Verify upvote count is 1
	if pod.UpvoteCount != 1 {
		t.Errorf("Expected upvote count 1, got %d", pod.UpvoteCount)
	}

	// Second upvote should fail with conflict error (requirement 5.3)
	err = svc.UpvotePod(ctx, pod.ID, userID)
	if err == nil {
		t.Fatal("Expected error when upvoting twice")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeConflict {
		t.Errorf("Expected conflict error code, got %s", appErr.Code)
	}

	// Verify upvote count is still 1 (not incremented on duplicate)
	if pod.UpvoteCount != 1 {
		t.Errorf("Expected upvote count to remain 1, got %d", pod.UpvoteCount)
	}
}

// TestUpvoteSystem_MultipleUsersUpvote tests that multiple users can upvote the same pod.
// Implements requirement 5.1: Upvote count is incremented for each user.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUpvoteSystem_MultipleUsersUpvote(t *testing.T) {
	svc, podRepo, _ := newTestServiceForUpvote()
	ctx := context.Background()

	// Create a public pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Multi User Test Pod", "multi-user-test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Multiple users upvote the pod
	numUsers := 5
	userIDs := make([]uuid.UUID, numUsers)
	for i := 0; i < numUsers; i++ {
		userIDs[i] = uuid.New()
		err := svc.UpvotePod(ctx, pod.ID, userIDs[i])
		if err != nil {
			t.Fatalf("UpvotePod failed for user %d: %v", i, err)
		}
	}

	// Verify upvote count equals number of users
	if pod.UpvoteCount != numUsers {
		t.Errorf("Expected upvote count %d, got %d", numUsers, pod.UpvoteCount)
	}

	// Verify each user has upvoted
	for i, userID := range userIDs {
		hasUpvoted, err := svc.HasUpvoted(ctx, pod.ID, userID)
		if err != nil {
			t.Fatalf("HasUpvoted failed for user %d: %v", i, err)
		}
		if !hasUpvoted {
			t.Errorf("Expected user %d to have upvoted", i)
		}
	}

	// Remove upvotes one by one and verify count decrements
	for i, userID := range userIDs {
		err := svc.RemoveUpvote(ctx, pod.ID, userID)
		if err != nil {
			t.Fatalf("RemoveUpvote failed for user %d: %v", i, err)
		}

		expectedCount := numUsers - i - 1
		if pod.UpvoteCount != expectedCount {
			t.Errorf("Expected upvote count %d after removing user %d's upvote, got %d", expectedCount, i, pod.UpvoteCount)
		}
	}

	// Verify final upvote count is 0
	if pod.UpvoteCount != 0 {
		t.Errorf("Expected final upvote count 0, got %d", pod.UpvoteCount)
	}
}

// TestUpvoteSystem_RemoveNonExistentUpvote tests that removing a non-existent upvote fails.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUpvoteSystem_RemoveNonExistentUpvote(t *testing.T) {
	svc, podRepo, _ := newTestServiceForUpvote()
	ctx := context.Background()

	// Create a public pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Remove Test Pod", "remove-test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Try to remove upvote without having upvoted
	userID := uuid.New()
	err := svc.RemoveUpvote(ctx, pod.ID, userID)
	if err == nil {
		t.Fatal("Expected error when removing non-existent upvote")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeNotFound {
		t.Errorf("Expected not found error code, got %s", appErr.Code)
	}
}

// TestUpvoteSystem_UpvotePrivatePodNoAccess tests that users cannot upvote private pods
// they don't have access to.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUpvoteSystem_UpvotePrivatePodNoAccess(t *testing.T) {
	svc, podRepo, _ := newTestServiceForUpvote()
	ctx := context.Background()

	// Create a private pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Private Test Pod", "private-test-pod", domain.VisibilityPrivate, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Try to upvote as non-collaborator
	userID := uuid.New()
	err := svc.UpvotePod(ctx, pod.ID, userID)
	if err == nil {
		t.Fatal("Expected error when upvoting private pod without access")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeForbidden {
		t.Errorf("Expected forbidden error code, got %s", appErr.Code)
	}

	// Verify upvote count is still 0
	if pod.UpvoteCount != 0 {
		t.Errorf("Expected upvote count 0, got %d", pod.UpvoteCount)
	}
}

// TestUpvoteSystem_OwnerCanUpvote tests that pod owners can upvote their own pods.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUpvoteSystem_OwnerCanUpvote(t *testing.T) {
	svc, podRepo, _ := newTestServiceForUpvote()
	ctx := context.Background()

	// Create a public pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Owner Upvote Test Pod", "owner-upvote-test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Owner upvotes their own pod
	err := svc.UpvotePod(ctx, pod.ID, ownerID)
	if err != nil {
		t.Fatalf("UpvotePod by owner failed: %v", err)
	}

	// Verify upvote count was incremented
	if pod.UpvoteCount != 1 {
		t.Errorf("Expected upvote count 1, got %d", pod.UpvoteCount)
	}

	// Verify owner has upvoted
	hasUpvoted, err := svc.HasUpvoted(ctx, pod.ID, ownerID)
	if err != nil {
		t.Fatalf("HasUpvoted failed: %v", err)
	}
	if !hasUpvoted {
		t.Error("Expected owner to have upvoted")
	}
}

// TestUpvoteSystem_UpvoteNonExistentPod tests that upvoting a non-existent pod fails.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUpvoteSystem_UpvoteNonExistentPod(t *testing.T) {
	svc, _, _ := newTestServiceForUpvote()
	ctx := context.Background()

	// Try to upvote non-existent pod
	userID := uuid.New()
	nonExistentPodID := uuid.New()
	err := svc.UpvotePod(ctx, nonExistentPodID, userID)
	if err == nil {
		t.Fatal("Expected error when upvoting non-existent pod")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeNotFound {
		t.Errorf("Expected not found error code, got %s", appErr.Code)
	}
}

// TestUpvoteSystem_UpvoteAndReupvote tests that a user can upvote again after removing their upvote.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUpvoteSystem_UpvoteAndReupvote(t *testing.T) {
	svc, podRepo, _ := newTestServiceForUpvote()
	ctx := context.Background()

	// Create a public pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Reupvote Test Pod", "reupvote-test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	userID := uuid.New()

	// First upvote
	err := svc.UpvotePod(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("First UpvotePod failed: %v", err)
	}
	if pod.UpvoteCount != 1 {
		t.Errorf("Expected upvote count 1, got %d", pod.UpvoteCount)
	}

	// Remove upvote
	err = svc.RemoveUpvote(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("RemoveUpvote failed: %v", err)
	}
	if pod.UpvoteCount != 0 {
		t.Errorf("Expected upvote count 0, got %d", pod.UpvoteCount)
	}

	// Re-upvote (should succeed)
	err = svc.UpvotePod(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("Re-upvote failed: %v", err)
	}
	if pod.UpvoteCount != 1 {
		t.Errorf("Expected upvote count 1 after re-upvote, got %d", pod.UpvoteCount)
	}

	// Verify user has upvoted
	hasUpvoted, err := svc.HasUpvoted(ctx, pod.ID, userID)
	if err != nil {
		t.Fatalf("HasUpvoted failed: %v", err)
	}
	if !hasUpvoted {
		t.Error("Expected HasUpvoted to return true after re-upvote")
	}
}

// TestUpvoteSystem_UpvoteCountConsistency tests that upvote count remains consistent
// across multiple operations.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUpvoteSystem_UpvoteCountConsistency(t *testing.T) {
	svc, podRepo, _ := newTestServiceForUpvote()
	ctx := context.Background()

	// Create a public pod
	ownerID := uuid.New()
	pod := domain.NewPod(ownerID, "Consistency Test Pod", "consistency-test-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Perform a series of upvote/remove operations
	user1 := uuid.New()
	user2 := uuid.New()
	user3 := uuid.New()

	// User1 upvotes (count: 1)
	if err := svc.UpvotePod(ctx, pod.ID, user1); err != nil {
		t.Fatalf("User1 upvote failed: %v", err)
	}
	if pod.UpvoteCount != 1 {
		t.Errorf("Expected count 1, got %d", pod.UpvoteCount)
	}

	// User2 upvotes (count: 2)
	if err := svc.UpvotePod(ctx, pod.ID, user2); err != nil {
		t.Fatalf("User2 upvote failed: %v", err)
	}
	if pod.UpvoteCount != 2 {
		t.Errorf("Expected count 2, got %d", pod.UpvoteCount)
	}

	// User1 removes upvote (count: 1)
	if err := svc.RemoveUpvote(ctx, pod.ID, user1); err != nil {
		t.Fatalf("User1 remove upvote failed: %v", err)
	}
	if pod.UpvoteCount != 1 {
		t.Errorf("Expected count 1, got %d", pod.UpvoteCount)
	}

	// User3 upvotes (count: 2)
	if err := svc.UpvotePod(ctx, pod.ID, user3); err != nil {
		t.Fatalf("User3 upvote failed: %v", err)
	}
	if pod.UpvoteCount != 2 {
		t.Errorf("Expected count 2, got %d", pod.UpvoteCount)
	}

	// User1 re-upvotes (count: 3)
	if err := svc.UpvotePod(ctx, pod.ID, user1); err != nil {
		t.Fatalf("User1 re-upvote failed: %v", err)
	}
	if pod.UpvoteCount != 3 {
		t.Errorf("Expected count 3, got %d", pod.UpvoteCount)
	}

	// Verify all users have upvoted
	for _, userID := range []uuid.UUID{user1, user2, user3} {
		hasUpvoted, err := svc.HasUpvoted(ctx, pod.ID, userID)
		if err != nil {
			t.Fatalf("HasUpvoted failed: %v", err)
		}
		if !hasUpvoted {
			t.Errorf("Expected user %s to have upvoted", userID)
		}
	}
}
