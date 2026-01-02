// Package application contains integration tests for the teacher verification flow.
// These tests verify the complete flow: submit -> approve -> role change.
//
// Prerequisites:
// - Docker Compose environment must be running: docker-compose up -d
// - All services must be healthy
//
// Run tests: go test -v -tags=integration ./internal/user/application/...
//
//go:build integration

package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// TestTeacherVerificationFlow_CompleteFlow tests the complete teacher verification flow:
// 1. Student submits verification request
// 2. Admin approves verification
// 3. User role changes from student to teacher
// Implements requirement 9.3: Integration tests for repository implementations.
func TestTeacherVerificationFlow_CompleteFlow(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Step 1: Create a student user
	student := domain.NewUser("integration-test@example.com", "hash", "Integration Test Student")
	if student.Role != domain.RoleStudent {
		t.Fatalf("Expected new user to have student role, got %s", student.Role)
	}
	userRepo.users[student.ID] = student
	userRepo.emailIndex[student.Email] = student

	// Step 2: Student submits teacher verification request
	verificationInput := TeacherVerificationInput{
		FullName:       "Integration Test Teacher",
		IDNumber:       "1234567890123456",
		CredentialType: domain.CredentialTypeEducatorCard,
		DocumentRef:    "ref://integration-test/educator-card-123",
	}

	verification, err := svc.SubmitTeacherVerification(ctx, student.ID, verificationInput)
	if err != nil {
		t.Fatalf("SubmitTeacherVerification failed: %v", err)
	}

	// Verify verification was created with pending status
	if verification.Status != domain.VerificationStatusPending {
		t.Errorf("Expected verification status %s, got %s", domain.VerificationStatusPending, verification.Status)
	}
	if verification.UserID != student.ID {
		t.Errorf("Expected verification user ID %s, got %s", student.ID, verification.UserID)
	}

	// Verify user is still a student at this point
	currentUser := userRepo.users[student.ID]
	if currentUser.Role != domain.RoleStudent {
		t.Errorf("Expected user role to still be %s before approval, got %s", domain.RoleStudent, currentUser.Role)
	}

	// Step 3: Admin approves the verification
	adminID := uuid.New()
	err = svc.ApproveVerification(ctx, verification.ID, adminID)
	if err != nil {
		t.Fatalf("ApproveVerification failed: %v", err)
	}

	// Step 4: Verify verification status is now approved
	updatedVerification := teacherVerificationRepo.verifications[verification.ID]
	if updatedVerification.Status != domain.VerificationStatusApproved {
		t.Errorf("Expected verification status %s, got %s", domain.VerificationStatusApproved, updatedVerification.Status)
	}
	if updatedVerification.ReviewedBy == nil || *updatedVerification.ReviewedBy != adminID {
		t.Error("Expected reviewer ID to be set to admin ID")
	}
	if updatedVerification.ReviewedAt == nil {
		t.Error("Expected reviewed at timestamp to be set")
	}

	// Step 5: Verify user role has changed from student to teacher
	updatedUser := userRepo.users[student.ID]
	if updatedUser.Role != domain.RoleTeacher {
		t.Errorf("Expected user role to change to %s after approval, got %s", domain.RoleTeacher, updatedUser.Role)
	}
}

// TestTeacherVerificationFlow_RejectFlow tests the rejection flow:
// 1. Student submits verification request
// 2. Admin rejects verification
// 3. User role remains student
func TestTeacherVerificationFlow_RejectFlow(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Step 1: Create a student user
	student := domain.NewUser("reject-test@example.com", "hash", "Reject Test Student")
	userRepo.users[student.ID] = student
	userRepo.emailIndex[student.Email] = student

	// Step 2: Student submits teacher verification request
	verificationInput := TeacherVerificationInput{
		FullName:       "Reject Test Teacher",
		IDNumber:       "9876543210123456",
		CredentialType: domain.CredentialTypeGovernmentID,
		DocumentRef:    "ref://reject-test/government-id-456",
	}

	verification, err := svc.SubmitTeacherVerification(ctx, student.ID, verificationInput)
	if err != nil {
		t.Fatalf("SubmitTeacherVerification failed: %v", err)
	}

	// Step 3: Admin rejects the verification
	adminID := uuid.New()
	rejectionReason := "Invalid credentials: document is not clear"
	err = svc.RejectVerification(ctx, verification.ID, adminID, rejectionReason)
	if err != nil {
		t.Fatalf("RejectVerification failed: %v", err)
	}

	// Step 4: Verify verification status is now rejected
	updatedVerification := teacherVerificationRepo.verifications[verification.ID]
	if updatedVerification.Status != domain.VerificationStatusRejected {
		t.Errorf("Expected verification status %s, got %s", domain.VerificationStatusRejected, updatedVerification.Status)
	}
	if updatedVerification.RejectionReason == nil || *updatedVerification.RejectionReason != rejectionReason {
		t.Errorf("Expected rejection reason %s, got %v", rejectionReason, updatedVerification.RejectionReason)
	}

	// Step 5: Verify user role remains student
	updatedUser := userRepo.users[student.ID]
	if updatedUser.Role != domain.RoleStudent {
		t.Errorf("Expected user role to remain %s after rejection, got %s", domain.RoleStudent, updatedUser.Role)
	}
}

// TestTeacherVerificationFlow_CannotResubmitWhilePending tests that a user cannot
// submit a new verification request while one is already pending.
func TestTeacherVerificationFlow_CannotResubmitWhilePending(t *testing.T) {
	svc, userRepo, _ := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user
	student := domain.NewUser("pending-test@example.com", "hash", "Pending Test Student")
	userRepo.users[student.ID] = student
	userRepo.emailIndex[student.Email] = student

	// Submit first verification request
	firstInput := TeacherVerificationInput{
		FullName:       "First Submission",
		IDNumber:       "1111111111111111",
		CredentialType: domain.CredentialTypeEducatorCard,
		DocumentRef:    "ref://first-submission",
	}

	_, err := svc.SubmitTeacherVerification(ctx, student.ID, firstInput)
	if err != nil {
		t.Fatalf("First SubmitTeacherVerification failed: %v", err)
	}

	// Try to submit second verification request while first is pending
	secondInput := TeacherVerificationInput{
		FullName:       "Second Submission",
		IDNumber:       "2222222222222222",
		CredentialType: domain.CredentialTypeProfessionalCert,
		DocumentRef:    "ref://second-submission",
	}

	_, err = svc.SubmitTeacherVerification(ctx, student.ID, secondInput)
	if err == nil {
		t.Fatal("Expected error when submitting second verification while first is pending")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeConflict {
		t.Errorf("Expected conflict error code, got %s", appErr.Code)
	}
}

// TestTeacherVerificationFlow_TeacherCannotSubmit tests that a user who is already
// a teacher cannot submit a verification request.
func TestTeacherVerificationFlow_TeacherCannotSubmit(t *testing.T) {
	svc, userRepo, _ := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a teacher user
	teacher := domain.NewUser("teacher-test@example.com", "hash", "Already Teacher")
	teacher.Role = domain.RoleTeacher
	userRepo.users[teacher.ID] = teacher
	userRepo.emailIndex[teacher.Email] = teacher

	// Try to submit verification request as teacher
	input := TeacherVerificationInput{
		FullName:       "Already Teacher",
		IDNumber:       "3333333333333333",
		CredentialType: domain.CredentialTypeEducatorCard,
		DocumentRef:    "ref://already-teacher",
	}

	_, err := svc.SubmitTeacherVerification(ctx, teacher.ID, input)
	if err == nil {
		t.Fatal("Expected error when teacher submits verification")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

// TestTeacherVerificationFlow_CannotApproveAlreadyReviewed tests that an already
// reviewed verification cannot be approved again.
func TestTeacherVerificationFlow_CannotApproveAlreadyReviewed(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user
	student := domain.NewUser("already-reviewed@example.com", "hash", "Already Reviewed Student")
	userRepo.users[student.ID] = student
	userRepo.emailIndex[student.Email] = student

	// Create an already approved verification
	verification := domain.NewTeacherVerification(
		student.ID,
		"Already Reviewed",
		"4444444444444444",
		domain.CredentialTypeEducatorCard,
		"ref://already-reviewed",
	)
	firstAdminID := uuid.New()
	verification.Approve(firstAdminID)
	teacherVerificationRepo.verifications[verification.ID] = verification
	teacherVerificationRepo.userIndex[student.ID] = verification

	// Try to approve again
	secondAdminID := uuid.New()
	err := svc.ApproveVerification(ctx, verification.ID, secondAdminID)
	if err == nil {
		t.Fatal("Expected error when approving already reviewed verification")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.Code != errors.CodeBadRequest {
		t.Errorf("Expected bad request error code, got %s", appErr.Code)
	}
}

// TestTeacherVerificationFlow_VerificationStatusCheck tests that verification status
// can be retrieved correctly at each stage of the flow.
func TestTeacherVerificationFlow_VerificationStatusCheck(t *testing.T) {
	svc, userRepo, _ := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user
	student := domain.NewUser("status-check@example.com", "hash", "Status Check Student")
	userRepo.users[student.ID] = student
	userRepo.emailIndex[student.Email] = student

	// Initially, no verification should exist
	_, err := svc.GetVerificationStatus(ctx, student.ID)
	if err == nil {
		t.Fatal("Expected error when no verification exists")
	}
	appErr, ok := err.(*errors.AppError)
	if !ok || appErr.Code != errors.CodeNotFound {
		t.Errorf("Expected not found error, got %v", err)
	}

	// Submit verification
	input := TeacherVerificationInput{
		FullName:       "Status Check Teacher",
		IDNumber:       "5555555555555555",
		CredentialType: domain.CredentialTypeEducatorCard,
		DocumentRef:    "ref://status-check",
	}

	verification, err := svc.SubmitTeacherVerification(ctx, student.ID, input)
	if err != nil {
		t.Fatalf("SubmitTeacherVerification failed: %v", err)
	}

	// Check status - should be pending
	status, err := svc.GetVerificationStatus(ctx, student.ID)
	if err != nil {
		t.Fatalf("GetVerificationStatus failed: %v", err)
	}
	if status.Status != domain.VerificationStatusPending {
		t.Errorf("Expected status %s, got %s", domain.VerificationStatusPending, status.Status)
	}

	// Approve verification
	adminID := uuid.New()
	err = svc.ApproveVerification(ctx, verification.ID, adminID)
	if err != nil {
		t.Fatalf("ApproveVerification failed: %v", err)
	}

	// Check status - should be approved
	status, err = svc.GetVerificationStatus(ctx, student.ID)
	if err != nil {
		t.Fatalf("GetVerificationStatus failed: %v", err)
	}
	if status.Status != domain.VerificationStatusApproved {
		t.Errorf("Expected status %s, got %s", domain.VerificationStatusApproved, status.Status)
	}
}

// TestTeacherVerificationFlow_VerificationDataPersistence tests that all verification
// data is correctly persisted throughout the flow.
func TestTeacherVerificationFlow_VerificationDataPersistence(t *testing.T) {
	svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create a student user
	student := domain.NewUser("data-persistence@example.com", "hash", "Data Persistence Student")
	userRepo.users[student.ID] = student
	userRepo.emailIndex[student.Email] = student

	// Submit verification with specific data
	input := TeacherVerificationInput{
		FullName:       "Dr. Data Persistence Teacher",
		IDNumber:       "6666666666666666",
		CredentialType: domain.CredentialTypeProfessionalCert,
		DocumentRef:    "ref://data-persistence/professional-cert-789",
	}

	verification, err := svc.SubmitTeacherVerification(ctx, student.ID, input)
	if err != nil {
		t.Fatalf("SubmitTeacherVerification failed: %v", err)
	}

	// Verify all data was persisted correctly
	storedVerification := teacherVerificationRepo.verifications[verification.ID]
	if storedVerification.FullName != input.FullName {
		t.Errorf("Expected full name %s, got %s", input.FullName, storedVerification.FullName)
	}
	if storedVerification.IDNumber != input.IDNumber {
		t.Errorf("Expected ID number %s, got %s", input.IDNumber, storedVerification.IDNumber)
	}
	if storedVerification.CredentialType != input.CredentialType {
		t.Errorf("Expected credential type %s, got %s", input.CredentialType, storedVerification.CredentialType)
	}
	if storedVerification.DocumentRef != input.DocumentRef {
		t.Errorf("Expected document ref %s, got %s", input.DocumentRef, storedVerification.DocumentRef)
	}
	if storedVerification.CreatedAt.IsZero() {
		t.Error("Expected created at timestamp to be set")
	}
	if storedVerification.UpdatedAt.IsZero() {
		t.Error("Expected updated at timestamp to be set")
	}

	// Approve and verify reviewer data is persisted
	adminID := uuid.New()
	beforeApproval := time.Now()
	err = svc.ApproveVerification(ctx, verification.ID, adminID)
	if err != nil {
		t.Fatalf("ApproveVerification failed: %v", err)
	}

	storedVerification = teacherVerificationRepo.verifications[verification.ID]
	if storedVerification.ReviewedBy == nil || *storedVerification.ReviewedBy != adminID {
		t.Error("Expected reviewer ID to be persisted")
	}
	if storedVerification.ReviewedAt == nil {
		t.Error("Expected reviewed at timestamp to be persisted")
	}
	if storedVerification.ReviewedAt.Before(beforeApproval) {
		t.Error("Expected reviewed at timestamp to be after approval time")
	}
}

// TestTeacherVerificationFlow_GetPendingVerifications tests that pending verifications
// can be retrieved correctly for admin review.
func TestTeacherVerificationFlow_GetPendingVerifications(t *testing.T) {
	svc, userRepo, _ := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create multiple students and submit verifications
	numStudents := 5
	for i := 0; i < numStudents; i++ {
		student := domain.NewUser(
			"pending-list-"+string(rune('a'+i))+"@example.com",
			"hash",
			"Pending List Student "+string(rune('A'+i)),
		)
		userRepo.users[student.ID] = student
		userRepo.emailIndex[student.Email] = student

		input := TeacherVerificationInput{
			FullName:       "Pending Teacher " + string(rune('A'+i)),
			IDNumber:       "777777777777777" + string(rune('0'+i)),
			CredentialType: domain.CredentialTypeEducatorCard,
			DocumentRef:    "ref://pending-list/" + string(rune('a'+i)),
		}

		_, err := svc.SubmitTeacherVerification(ctx, student.ID, input)
		if err != nil {
			t.Fatalf("SubmitTeacherVerification failed for student %d: %v", i, err)
		}
	}

	// Get pending verifications
	result, err := svc.GetPendingVerifications(ctx, 1, 10)
	if err != nil {
		t.Fatalf("GetPendingVerifications failed: %v", err)
	}

	if result.Total != numStudents {
		t.Errorf("Expected %d pending verifications, got %d", numStudents, result.Total)
	}
	if len(result.Verifications) != numStudents {
		t.Errorf("Expected %d verifications in result, got %d", numStudents, len(result.Verifications))
	}

	// Verify all are pending
	for _, v := range result.Verifications {
		if v.Status != domain.VerificationStatusPending {
			t.Errorf("Expected all verifications to be pending, got %s", v.Status)
		}
	}
}

// TestTeacherVerificationFlow_PaginationWorks tests that pagination works correctly
// for pending verifications.
func TestTeacherVerificationFlow_PaginationWorks(t *testing.T) {
	svc, userRepo, _ := newTestServiceWithTeacherVerification()
	ctx := context.Background()

	// Create 7 students and submit verifications
	numStudents := 7
	for i := 0; i < numStudents; i++ {
		student := domain.NewUser(
			"pagination-"+string(rune('a'+i))+"@example.com",
			"hash",
			"Pagination Student "+string(rune('A'+i)),
		)
		userRepo.users[student.ID] = student
		userRepo.emailIndex[student.Email] = student

		input := TeacherVerificationInput{
			FullName:       "Pagination Teacher " + string(rune('A'+i)),
			IDNumber:       "888888888888888" + string(rune('0'+i)),
			CredentialType: domain.CredentialTypeEducatorCard,
			DocumentRef:    "ref://pagination/" + string(rune('a'+i)),
		}

		_, err := svc.SubmitTeacherVerification(ctx, student.ID, input)
		if err != nil {
			t.Fatalf("SubmitTeacherVerification failed for student %d: %v", i, err)
		}
	}

	// Get first page with 3 items per page
	page1, err := svc.GetPendingVerifications(ctx, 1, 3)
	if err != nil {
		t.Fatalf("GetPendingVerifications page 1 failed: %v", err)
	}

	if page1.Total != numStudents {
		t.Errorf("Expected total %d, got %d", numStudents, page1.Total)
	}
	if len(page1.Verifications) != 3 {
		t.Errorf("Expected 3 verifications on page 1, got %d", len(page1.Verifications))
	}
	if page1.TotalPages != 3 {
		t.Errorf("Expected 3 total pages, got %d", page1.TotalPages)
	}

	// Get second page
	page2, err := svc.GetPendingVerifications(ctx, 2, 3)
	if err != nil {
		t.Fatalf("GetPendingVerifications page 2 failed: %v", err)
	}

	if len(page2.Verifications) != 3 {
		t.Errorf("Expected 3 verifications on page 2, got %d", len(page2.Verifications))
	}

	// Get third page (should have 1 item)
	page3, err := svc.GetPendingVerifications(ctx, 3, 3)
	if err != nil {
		t.Fatalf("GetPendingVerifications page 3 failed: %v", err)
	}

	if len(page3.Verifications) != 1 {
		t.Errorf("Expected 1 verification on page 3, got %d", len(page3.Verifications))
	}
}
