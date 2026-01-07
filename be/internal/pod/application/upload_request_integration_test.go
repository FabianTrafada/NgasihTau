// Package application contains integration tests for the upload request flow.
// These tests verify the complete flow: request -> approve/reject -> upload permission.
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

// newTestServiceForUploadRequest creates a test service with all required mock repositories
// for upload request integration testing.
func newTestServiceForUploadRequest() (PodService, *mockPodRepo, *mockUploadRequestRepo, *mockUserRoleChecker) {
	podRepo := newMockPodRepo()
	collaboratorRepo := newMockCollaboratorRepo()
	starRepo := newMockStarRepo()
	upvoteRepo := newMockUpvoteRepo()
	downvoteRepo := newMockDownvoteRepo()
	uploadReqRepo := newMockUploadRequestRepo()
	sharedPodRepo := newMockSharedPodRepo()
	followRepo := newMockFollowRepo()
	activityRepo := newMockActivityRepo()
	eventPublisher := NewNoOpEventPublisher()
	roleChecker := newMockUserRoleChecker()

	svc := NewPodService(
		podRepo,
		collaboratorRepo,
		starRepo,
		upvoteRepo,
		downvoteRepo,
		uploadReqRepo,
		sharedPodRepo,
		followRepo,
		activityRepo,
		eventPublisher,
		roleChecker,
	)

	return svc, podRepo, uploadReqRepo, roleChecker
}

// TestUploadRequestFlow_CompleteApprovalFlow tests the complete upload request approval flow:
// 1. Teacher A creates a pod
// 2. Teacher B requests upload permission
// 3. Teacher A approves the request
// 4. Teacher B can now upload to the pod
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_CompleteApprovalFlow(t *testing.T) {
	svc, podRepo, uploadReqRepo, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Step 1: Create a pod owned by Teacher A
	teacherA := uuid.New()
	roleChecker.teacherIDs[teacherA] = true
	pod := domain.NewPod(teacherA, "Teacher A Pod", "teacher-a-pod", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Step 2: Teacher B requests upload permission
	teacherB := uuid.New()
	roleChecker.teacherIDs[teacherB] = true
	message := "I would like to contribute materials on advanced topics"

	request, err := svc.CreateUploadRequest(ctx, teacherB, pod.ID, &message)
	if err != nil {
		t.Fatalf("CreateUploadRequest failed: %v", err)
	}

	// Verify request was created with pending status
	if request.Status != domain.UploadRequestStatusPending {
		t.Errorf("Expected status %s, got %s", domain.UploadRequestStatusPending, request.Status)
	}
	if request.RequesterID != teacherB {
		t.Errorf("Expected requester ID %s, got %s", teacherB, request.RequesterID)
	}
	if request.PodOwnerID != teacherA {
		t.Errorf("Expected pod owner ID %s, got %s", teacherA, request.PodOwnerID)
	}

	// Step 3: Verify Teacher B cannot upload yet (pending request)
	canUpload, err := svc.CanUserUploadToPod(ctx, pod.ID, teacherB)
	if err != nil {
		t.Fatalf("CanUserUploadToPod failed: %v", err)
	}
	if canUpload {
		t.Error("Expected Teacher B to NOT be able to upload while request is pending")
	}

	// Step 4: Teacher A approves the request
	err = svc.ApproveUploadRequest(ctx, request.ID, teacherA)
	if err != nil {
		t.Fatalf("ApproveUploadRequest failed: %v", err)
	}

	// Verify request status is now approved
	updatedRequest := uploadReqRepo.requests[request.ID]
	if updatedRequest.Status != domain.UploadRequestStatusApproved {
		t.Errorf("Expected status %s, got %s", domain.UploadRequestStatusApproved, updatedRequest.Status)
	}

	// Step 5: Verify Teacher B can now upload
	canUpload, err = svc.CanUserUploadToPod(ctx, pod.ID, teacherB)
	if err != nil {
		t.Fatalf("CanUserUploadToPod failed: %v", err)
	}
	if !canUpload {
		t.Error("Expected Teacher B to be able to upload after approval")
	}
}

// TestUploadRequestFlow_CompleteRejectionFlow tests the complete upload request rejection flow:
// 1. Teacher A creates a pod
// 2. Teacher B requests upload permission
// 3. Teacher A rejects the request with reason
// 4. Teacher B still cannot upload to the pod
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_CompleteRejectionFlow(t *testing.T) {
	svc, podRepo, uploadReqRepo, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Step 1: Create a pod owned by Teacher A
	teacherA := uuid.New()
	roleChecker.teacherIDs[teacherA] = true
	pod := domain.NewPod(teacherA, "Teacher A Pod", "teacher-a-pod-reject", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Step 2: Teacher B requests upload permission
	teacherB := uuid.New()
	roleChecker.teacherIDs[teacherB] = true
	message := "I would like to contribute materials"

	request, err := svc.CreateUploadRequest(ctx, teacherB, pod.ID, &message)
	if err != nil {
		t.Fatalf("CreateUploadRequest failed: %v", err)
	}

	// Step 3: Teacher A rejects the request with reason
	rejectionReason := "Not accepting contributions at this time"
	err = svc.RejectUploadRequest(ctx, request.ID, teacherA, &rejectionReason)
	if err != nil {
		t.Fatalf("RejectUploadRequest failed: %v", err)
	}

	// Verify request status is now rejected with reason
	updatedRequest := uploadReqRepo.requests[request.ID]
	if updatedRequest.Status != domain.UploadRequestStatusRejected {
		t.Errorf("Expected status %s, got %s", domain.UploadRequestStatusRejected, updatedRequest.Status)
	}
	if updatedRequest.RejectionReason == nil || *updatedRequest.RejectionReason != rejectionReason {
		t.Errorf("Expected rejection reason %s, got %v", rejectionReason, updatedRequest.RejectionReason)
	}

	// Step 4: Verify Teacher B still cannot upload
	canUpload, err := svc.CanUserUploadToPod(ctx, pod.ID, teacherB)
	if err != nil {
		t.Fatalf("CanUserUploadToPod failed: %v", err)
	}
	if canUpload {
		t.Error("Expected Teacher B to NOT be able to upload after rejection")
	}
}

// TestUploadRequestFlow_RevokePermission tests the revocation flow:
// 1. Teacher A creates a pod
// 2. Teacher B requests and gets approved
// 3. Teacher A revokes the permission
// 4. Teacher B can no longer upload
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_RevokePermission(t *testing.T) {
	svc, podRepo, uploadReqRepo, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Step 1: Create a pod owned by Teacher A
	teacherA := uuid.New()
	roleChecker.teacherIDs[teacherA] = true
	pod := domain.NewPod(teacherA, "Teacher A Pod", "teacher-a-pod-revoke", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Step 2: Teacher B requests and gets approved
	teacherB := uuid.New()
	roleChecker.teacherIDs[teacherB] = true

	request, err := svc.CreateUploadRequest(ctx, teacherB, pod.ID, nil)
	if err != nil {
		t.Fatalf("CreateUploadRequest failed: %v", err)
	}

	err = svc.ApproveUploadRequest(ctx, request.ID, teacherA)
	if err != nil {
		t.Fatalf("ApproveUploadRequest failed: %v", err)
	}

	// Verify Teacher B can upload
	canUpload, err := svc.CanUserUploadToPod(ctx, pod.ID, teacherB)
	if err != nil {
		t.Fatalf("CanUserUploadToPod failed: %v", err)
	}
	if !canUpload {
		t.Error("Expected Teacher B to be able to upload after approval")
	}

	// Step 3: Teacher A revokes the permission
	err = svc.RevokeUploadPermission(ctx, request.ID, teacherA)
	if err != nil {
		t.Fatalf("RevokeUploadPermission failed: %v", err)
	}

	// Verify request status is now revoked
	updatedRequest := uploadReqRepo.requests[request.ID]
	if updatedRequest.Status != domain.UploadRequestStatusRevoked {
		t.Errorf("Expected status %s, got %s", domain.UploadRequestStatusRevoked, updatedRequest.Status)
	}

	// Step 4: Verify Teacher B can no longer upload
	canUpload, err = svc.CanUserUploadToPod(ctx, pod.ID, teacherB)
	if err != nil {
		t.Fatalf("CanUserUploadToPod failed: %v", err)
	}
	if canUpload {
		t.Error("Expected Teacher B to NOT be able to upload after revocation")
	}
}

// TestUploadRequestFlow_NonTeacherCannotRequest tests that non-teachers cannot create upload requests.
// Implements requirement 4.1: Only teachers can request upload permission.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_NonTeacherCannotRequest(t *testing.T) {
	svc, podRepo, _, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Create a pod owned by a teacher
	teacherA := uuid.New()
	roleChecker.teacherIDs[teacherA] = true
	pod := domain.NewPod(teacherA, "Teacher Pod", "teacher-pod-nonteacher", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Student (non-teacher) tries to request upload permission
	studentID := uuid.New()
	// Note: studentID is NOT added to roleChecker.teacherIDs

	_, err := svc.CreateUploadRequest(ctx, studentID, pod.ID, nil)
	if err == nil {
		t.Fatal("Expected error when non-teacher requests upload permission")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeForbidden {
		t.Errorf("Expected forbidden error code, got %s", appErr.Code)
	}
}

// TestUploadRequestFlow_CannotRequestToStudentPod tests that upload requests can only be made to teacher pods.
// Implements requirement 4.1: Upload requests can only be made to pods owned by teachers.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_CannotRequestToStudentPod(t *testing.T) {
	svc, podRepo, _, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Create a pod owned by a student (non-teacher)
	studentOwner := uuid.New()
	// Note: studentOwner is NOT added to roleChecker.teacherIDs
	pod := domain.NewPod(studentOwner, "Student Pod", "student-pod", domain.VisibilityPublic, false)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Teacher tries to request upload permission to student's pod
	teacherID := uuid.New()
	roleChecker.teacherIDs[teacherID] = true

	_, err := svc.CreateUploadRequest(ctx, teacherID, pod.ID, nil)
	if err == nil {
		t.Fatal("Expected error when requesting upload to student's pod")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

// TestUploadRequestFlow_CannotRequestOwnPod tests that teachers cannot request upload to their own pods.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_CannotRequestOwnPod(t *testing.T) {
	svc, podRepo, _, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Create a pod owned by a teacher
	teacherID := uuid.New()
	roleChecker.teacherIDs[teacherID] = true
	pod := domain.NewPod(teacherID, "My Pod", "my-pod", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Teacher tries to request upload permission to their own pod
	_, err := svc.CreateUploadRequest(ctx, teacherID, pod.ID, nil)
	if err == nil {
		t.Fatal("Expected error when requesting upload to own pod")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

// TestUploadRequestFlow_DuplicatePendingRequest tests that duplicate pending requests are prevented.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_DuplicatePendingRequest(t *testing.T) {
	svc, podRepo, _, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Create a pod owned by Teacher A
	teacherA := uuid.New()
	roleChecker.teacherIDs[teacherA] = true
	pod := domain.NewPod(teacherA, "Teacher A Pod", "teacher-a-pod-dup", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Teacher B creates first request
	teacherB := uuid.New()
	roleChecker.teacherIDs[teacherB] = true

	_, err := svc.CreateUploadRequest(ctx, teacherB, pod.ID, nil)
	if err != nil {
		t.Fatalf("First CreateUploadRequest failed: %v", err)
	}

	// Teacher B tries to create second request while first is pending
	_, err = svc.CreateUploadRequest(ctx, teacherB, pod.ID, nil)
	if err == nil {
		t.Fatal("Expected error when creating duplicate pending request")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeConflict {
		t.Errorf("Expected conflict error code, got %s", appErr.Code)
	}
}

// TestUploadRequestFlow_NonOwnerCannotApprove tests that only pod owners can approve requests.
// Implements requirement 4.3: Only pod owner can approve upload requests.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_NonOwnerCannotApprove(t *testing.T) {
	svc, podRepo, _, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Create a pod owned by Teacher A
	teacherA := uuid.New()
	roleChecker.teacherIDs[teacherA] = true
	pod := domain.NewPod(teacherA, "Teacher A Pod", "teacher-a-pod-nonowner", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Teacher B creates request
	teacherB := uuid.New()
	roleChecker.teacherIDs[teacherB] = true

	request, err := svc.CreateUploadRequest(ctx, teacherB, pod.ID, nil)
	if err != nil {
		t.Fatalf("CreateUploadRequest failed: %v", err)
	}

	// Teacher C (not the owner) tries to approve
	teacherC := uuid.New()
	roleChecker.teacherIDs[teacherC] = true

	err = svc.ApproveUploadRequest(ctx, request.ID, teacherC)
	if err == nil {
		t.Fatal("Expected error when non-owner approves request")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeForbidden {
		t.Errorf("Expected forbidden error code, got %s", appErr.Code)
	}
}

// TestUploadRequestFlow_NonOwnerCannotReject tests that only pod owners can reject requests.
// Implements requirement 4.4: Only pod owner can reject upload requests.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_NonOwnerCannotReject(t *testing.T) {
	svc, podRepo, _, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Create a pod owned by Teacher A
	teacherA := uuid.New()
	roleChecker.teacherIDs[teacherA] = true
	pod := domain.NewPod(teacherA, "Teacher A Pod", "teacher-a-pod-nonowner-reject", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Teacher B creates request
	teacherB := uuid.New()
	roleChecker.teacherIDs[teacherB] = true

	request, err := svc.CreateUploadRequest(ctx, teacherB, pod.ID, nil)
	if err != nil {
		t.Fatalf("CreateUploadRequest failed: %v", err)
	}

	// Teacher C (not the owner) tries to reject
	teacherC := uuid.New()
	roleChecker.teacherIDs[teacherC] = true

	err = svc.RejectUploadRequest(ctx, request.ID, teacherC, nil)
	if err == nil {
		t.Fatal("Expected error when non-owner rejects request")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeForbidden {
		t.Errorf("Expected forbidden error code, got %s", appErr.Code)
	}
}

// TestUploadRequestFlow_CannotApproveAlreadyApproved tests that already approved requests cannot be approved again.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_CannotApproveAlreadyApproved(t *testing.T) {
	svc, podRepo, _, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Create a pod owned by Teacher A
	teacherA := uuid.New()
	roleChecker.teacherIDs[teacherA] = true
	pod := domain.NewPod(teacherA, "Teacher A Pod", "teacher-a-pod-already-approved", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Teacher B creates request
	teacherB := uuid.New()
	roleChecker.teacherIDs[teacherB] = true

	request, err := svc.CreateUploadRequest(ctx, teacherB, pod.ID, nil)
	if err != nil {
		t.Fatalf("CreateUploadRequest failed: %v", err)
	}

	// First approval
	err = svc.ApproveUploadRequest(ctx, request.ID, teacherA)
	if err != nil {
		t.Fatalf("First ApproveUploadRequest failed: %v", err)
	}

	// Try to approve again
	err = svc.ApproveUploadRequest(ctx, request.ID, teacherA)
	if err == nil {
		t.Fatal("Expected error when approving already approved request")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

// TestUploadRequestFlow_CannotRevokeNonApproved tests that only approved requests can be revoked.
// Implements requirement 4.6: Only approved permissions can be revoked.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_CannotRevokeNonApproved(t *testing.T) {
	svc, podRepo, _, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Create a pod owned by Teacher A
	teacherA := uuid.New()
	roleChecker.teacherIDs[teacherA] = true
	pod := domain.NewPod(teacherA, "Teacher A Pod", "teacher-a-pod-revoke-pending", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Teacher B creates request (still pending)
	teacherB := uuid.New()
	roleChecker.teacherIDs[teacherB] = true

	request, err := svc.CreateUploadRequest(ctx, teacherB, pod.ID, nil)
	if err != nil {
		t.Fatalf("CreateUploadRequest failed: %v", err)
	}

	// Try to revoke pending request
	err = svc.RevokeUploadPermission(ctx, request.ID, teacherA)
	if err == nil {
		t.Fatal("Expected error when revoking pending request")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

// TestUploadRequestFlow_GetUploadRequestsForOwner tests retrieving upload requests for a pod owner.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_GetUploadRequestsForOwner(t *testing.T) {
	svc, podRepo, _, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Create a pod owned by Teacher A
	teacherA := uuid.New()
	roleChecker.teacherIDs[teacherA] = true
	pod := domain.NewPod(teacherA, "Teacher A Pod", "teacher-a-pod-list", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Multiple teachers create requests
	numRequests := 3
	for i := 0; i < numRequests; i++ {
		teacherID := uuid.New()
		roleChecker.teacherIDs[teacherID] = true

		_, err := svc.CreateUploadRequest(ctx, teacherID, pod.ID, nil)
		if err != nil {
			t.Fatalf("CreateUploadRequest %d failed: %v", i, err)
		}
	}

	// Get upload requests for owner
	result, err := svc.GetUploadRequestsForOwner(ctx, teacherA, nil, 1, 10)
	if err != nil {
		t.Fatalf("GetUploadRequestsForOwner failed: %v", err)
	}

	if result.Total != numRequests {
		t.Errorf("Expected %d requests, got %d", numRequests, result.Total)
	}
	if len(result.UploadRequests) != numRequests {
		t.Errorf("Expected %d requests in result, got %d", numRequests, len(result.UploadRequests))
	}

	// Verify all are pending
	for _, req := range result.UploadRequests {
		if req.Status != domain.UploadRequestStatusPending {
			t.Errorf("Expected all requests to be pending, got %s", req.Status)
		}
	}
}

// TestUploadRequestFlow_GetUploadRequestsByRequester tests retrieving upload requests made by a requester.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_GetUploadRequestsByRequester(t *testing.T) {
	svc, podRepo, _, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Teacher B will make requests to multiple pods
	teacherB := uuid.New()
	roleChecker.teacherIDs[teacherB] = true

	// Create multiple pods owned by different teachers
	numPods := 3
	for i := 0; i < numPods; i++ {
		teacherID := uuid.New()
		roleChecker.teacherIDs[teacherID] = true
		pod := domain.NewPod(teacherID, "Teacher Pod", "teacher-pod-"+uuid.New().String()[:8], domain.VisibilityPublic, true)
		podRepo.pods[pod.ID] = pod
		podRepo.slugIndex[pod.Slug] = pod

		_, err := svc.CreateUploadRequest(ctx, teacherB, pod.ID, nil)
		if err != nil {
			t.Fatalf("CreateUploadRequest %d failed: %v", i, err)
		}
	}

	// Get upload requests by requester
	result, err := svc.GetUploadRequestsByRequester(ctx, teacherB, 1, 10)
	if err != nil {
		t.Fatalf("GetUploadRequestsByRequester failed: %v", err)
	}

	if result.Total != numPods {
		t.Errorf("Expected %d requests, got %d", numPods, result.Total)
	}
	if len(result.UploadRequests) != numPods {
		t.Errorf("Expected %d requests in result, got %d", numPods, len(result.UploadRequests))
	}

	// Verify all belong to Teacher B
	for _, req := range result.UploadRequests {
		if req.RequesterID != teacherB {
			t.Errorf("Expected requester ID %s, got %s", teacherB, req.RequesterID)
		}
	}
}

// TestUploadRequestFlow_OwnerAlwaysCanUpload tests that pod owners can always upload regardless of requests.
// Implements requirement 9.3: Integration tests for repository implementations.
func TestUploadRequestFlow_OwnerAlwaysCanUpload(t *testing.T) {
	svc, podRepo, _, roleChecker := newTestServiceForUploadRequest()
	ctx := context.Background()

	// Create a pod owned by Teacher A
	teacherA := uuid.New()
	roleChecker.teacherIDs[teacherA] = true
	pod := domain.NewPod(teacherA, "Teacher A Pod", "teacher-a-pod-owner", domain.VisibilityPublic, true)
	podRepo.pods[pod.ID] = pod
	podRepo.slugIndex[pod.Slug] = pod

	// Owner should always be able to upload
	canUpload, err := svc.CanUserUploadToPod(ctx, pod.ID, teacherA)
	if err != nil {
		t.Fatalf("CanUserUploadToPod failed: %v", err)
	}
	if !canUpload {
		t.Error("Expected owner to always be able to upload")
	}
}
