// Package application contains property-based tests for the User Service.
package application

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"ngasihtau/internal/user/domain"
	"ngasihtau/pkg/jwt"
)

// **Feature: student-teacher-roles, Property 3: Teacher Verification Workflow**
// **Validates: Requirements 2.1, 2.2, 2.5**
//
// Property 3: Teacher Verification Workflow
// *For any* teacher verification request with valid data:
// - A pending verification record SHALL be created
// - When approved, the user's role SHALL change from "student" to "teacher"
// - The verification record SHALL store verification date and credential type

func TestProperty_TeacherVerificationWorkflow(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 3.1: Submitting verification creates a pending record
	// Validates: Requirement 2.1 - WHEN a user submits teacher verification request,
	// THE User Service SHALL create a pending verification record.
	properties.Property("submitting verification creates pending record", prop.ForAll(
		func(fullName, idNumber, documentRef string, credType domain.CredentialType) bool {
			svc, userRepo, _ := newTestServiceWithTeacherVerification()
			ctx := context.Background()

			// Create a student user
			user := domain.NewUser("test@example.com", "hash", "Test User")
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			input := TeacherVerificationInput{
				FullName:       fullName,
				IDNumber:       idNumber,
				CredentialType: credType,
				DocumentRef:    documentRef,
			}

			verification, err := svc.SubmitTeacherVerification(ctx, user.ID, input)
			if err != nil {
				return false
			}

			// Verify pending status
			return verification != nil &&
				verification.Status == domain.VerificationStatusPending &&
				verification.IsPending()
		},
		genValidFullName(),
		genValidIDNumber(),
		genValidDocumentRef(),
		genCredentialType(),
	))

	// Property 3.2: Verification stores credential type correctly
	// Validates: Requirement 2.5 - THE User Service SHALL store teacher verification
	// metadata including verification date and credential type.
	properties.Property("verification stores credential type correctly", prop.ForAll(
		func(fullName, idNumber, documentRef string, credType domain.CredentialType) bool {
			svc, userRepo, _ := newTestServiceWithTeacherVerification()
			ctx := context.Background()

			// Create a student user
			user := domain.NewUser("test@example.com", "hash", "Test User")
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			input := TeacherVerificationInput{
				FullName:       fullName,
				IDNumber:       idNumber,
				CredentialType: credType,
				DocumentRef:    documentRef,
			}

			verification, err := svc.SubmitTeacherVerification(ctx, user.ID, input)
			if err != nil {
				return false
			}

			// Verify credential type is stored correctly
			return verification.CredentialType == credType
		},
		genValidFullName(),
		genValidIDNumber(),
		genValidDocumentRef(),
		genCredentialType(),
	))

	// Property 3.3: Verification stores all required fields correctly
	// Validates: Requirement 2.5 - THE User Service SHALL store teacher verification
	// metadata including verification date and credential type.
	properties.Property("verification stores all required fields", prop.ForAll(
		func(fullName, idNumber, documentRef string, credType domain.CredentialType) bool {
			svc, userRepo, _ := newTestServiceWithTeacherVerification()
			ctx := context.Background()

			// Create a student user
			user := domain.NewUser("test@example.com", "hash", "Test User")
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			input := TeacherVerificationInput{
				FullName:       fullName,
				IDNumber:       idNumber,
				CredentialType: credType,
				DocumentRef:    documentRef,
			}

			verification, err := svc.SubmitTeacherVerification(ctx, user.ID, input)
			if err != nil {
				return false
			}

			// Verify all fields are stored correctly
			return verification.FullName == fullName &&
				verification.IDNumber == idNumber &&
				verification.CredentialType == credType &&
				verification.DocumentRef == documentRef &&
				verification.UserID == user.ID &&
				verification.ID != uuid.Nil
		},
		genValidFullName(),
		genValidIDNumber(),
		genValidDocumentRef(),
		genCredentialType(),
	))

	// Property 3.4: Verification stores creation timestamp
	// Validates: Requirement 2.5 - THE User Service SHALL store teacher verification
	// metadata including verification date and credential type.
	properties.Property("verification stores creation timestamp", prop.ForAll(
		func(fullName, idNumber, documentRef string, credType domain.CredentialType) bool {
			svc, userRepo, _ := newTestServiceWithTeacherVerification()
			ctx := context.Background()

			// Create a student user
			user := domain.NewUser("test@example.com", "hash", "Test User")
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			beforeSubmit := time.Now().Add(-1 * time.Second)

			input := TeacherVerificationInput{
				FullName:       fullName,
				IDNumber:       idNumber,
				CredentialType: credType,
				DocumentRef:    documentRef,
			}

			verification, err := svc.SubmitTeacherVerification(ctx, user.ID, input)
			if err != nil {
				return false
			}

			afterSubmit := time.Now().Add(1 * time.Second)

			// Verify timestamps are set and within expected range
			return !verification.CreatedAt.IsZero() &&
				!verification.UpdatedAt.IsZero() &&
				verification.CreatedAt.After(beforeSubmit) &&
				verification.CreatedAt.Before(afterSubmit) &&
				verification.ReviewedAt == nil &&
				verification.ReviewedBy == nil
		},
		genValidFullName(),
		genValidIDNumber(),
		genValidDocumentRef(),
		genCredentialType(),
	))

	// Property 3.5: Teachers cannot submit verification (already verified)
	// Validates: Requirement 2.1 - Only students can submit verification requests.
	properties.Property("teachers cannot submit verification", prop.ForAll(
		func(fullName, idNumber, documentRef string, credType domain.CredentialType) bool {
			svc, userRepo, _ := newTestServiceWithTeacherVerification()
			ctx := context.Background()

			// Create a teacher user (already verified)
			user := domain.NewUser("teacher@example.com", "hash", "Teacher User")
			user.Role = domain.RoleTeacher
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			input := TeacherVerificationInput{
				FullName:       fullName,
				IDNumber:       idNumber,
				CredentialType: credType,
				DocumentRef:    documentRef,
			}

			_, err := svc.SubmitTeacherVerification(ctx, user.ID, input)

			// Should return error for teachers
			return err != nil
		},
		genValidFullName(),
		genValidIDNumber(),
		genValidDocumentRef(),
		genCredentialType(),
	))

	// Property 3.6: Duplicate pending verification is rejected
	// Validates: Requirement 2.1 - Only one pending verification per user.
	properties.Property("duplicate pending verification is rejected", prop.ForAll(
		func(fullName, idNumber, documentRef string, credType domain.CredentialType) bool {
			svc, userRepo, teacherVerificationRepo := newTestServiceWithTeacherVerification()
			ctx := context.Background()

			// Create a student user
			user := domain.NewUser("test@example.com", "hash", "Test User")
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			input := TeacherVerificationInput{
				FullName:       fullName,
				IDNumber:       idNumber,
				CredentialType: credType,
				DocumentRef:    documentRef,
			}

			// First submission should succeed
			_, err := svc.SubmitTeacherVerification(ctx, user.ID, input)
			if err != nil {
				return false
			}

			// Verify there's one verification in the repo
			if len(teacherVerificationRepo.verifications) != 1 {
				return false
			}

			// Second submission should fail
			_, err = svc.SubmitTeacherVerification(ctx, user.ID, input)

			// Should return error for duplicate
			return err != nil
		},
		genValidFullName(),
		genValidIDNumber(),
		genValidDocumentRef(),
		genCredentialType(),
	))

	properties.TestingRun(t)
}

// Helper function to create test service with teacher verification repo access
func newTestServiceWithTeacherVerification() (UserService, *mockUserRepo, *mockTeacherVerificationRepo) {
	userRepo := newMockUserRepo()
	oauthRepo := &mockOAuthRepo{}
	refreshTokenRepo := newMockRefreshTokenRepo()
	backupCodeRepo := newMockBackupCodeRepo()
	followRepo := newMockFollowRepo()
	verificationTokenRepo := newMockVerificationTokenRepo()
	teacherVerificationRepo := newMockTeacherVerificationRepo()

	jwtManager := jwt.NewManager(jwt.Config{
		Secret:             "test-secret-key-for-testing-only",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "ngasihtau-test",
	})

	svc := NewUserService(
		userRepo,
		oauthRepo,
		refreshTokenRepo,
		backupCodeRepo,
		followRepo,
		verificationTokenRepo,
		teacherVerificationRepo,
		jwtManager,
		nil, // No Google client for tests
		nil, // No event publisher for tests
	)

	return svc, userRepo, teacherVerificationRepo
}

// Generator for valid full names (min 3, max 255 characters)
func genValidFullName() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		trimmed := strings.TrimSpace(s)
		return len(trimmed) >= 3 && len(trimmed) <= 255
	}).Map(func(s string) string {
		if len(s) < 3 {
			return "Default Name"
		}
		if len(s) > 255 {
			return s[:255]
		}
		return s
	})
}

// Generator for valid ID numbers (min 10, max 100 characters)
func genValidIDNumber() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) >= 10 && len(s) <= 100
	}).Map(func(s string) string {
		if len(s) < 10 {
			return "1234567890"
		}
		if len(s) > 100 {
			return s[:100]
		}
		return s
	})
}

// Generator for valid document references (max 500 characters)
func genValidDocumentRef() gopter.Gen {
	prefixes := []string{"/documents/", "https://storage.example.com/docs/", "ref-"}
	suffixes := []string{".pdf", "", "-123"}

	return gen.IntRange(0, 2).FlatMap(func(idx interface{}) gopter.Gen {
		i := idx.(int)
		return gen.AlphaString().Map(func(s string) string {
			if len(s) == 0 {
				s = "default"
			}
			result := prefixes[i] + s + suffixes[i]
			if len(result) > 500 {
				return result[:500]
			}
			return result
		})
	}, reflect.TypeOf(""))
}

// Generator for valid credential types
func genCredentialType() gopter.Gen {
	return gen.OneConstOf(
		domain.CredentialTypeGovernmentID,
		domain.CredentialTypeEducatorCard,
		domain.CredentialTypeProfessionalCert,
	)
}
