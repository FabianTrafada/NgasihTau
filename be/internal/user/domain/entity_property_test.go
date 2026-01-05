// Package domain contains property-based tests for the User domain entities.
package domain

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: student-teacher-roles, Property 1: Default Student Role Assignment**
//
// Property 1: Default Student Role Assignment
// *For any* new user registration with valid data, the resulting user SHALL always
// have the "student" role assigned by default.

func TestProperty_DefaultStudentRoleAssignment(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 1.1: NewUser always assigns RoleStudent
	properties.Property("NewUser always assigns student role", prop.ForAll(
		func(email, passwordHash, name string) bool {
			user := NewUser(email, passwordHash, name)
			return user.Role == RoleStudent
		},
		genEmail(),
		genPasswordHash(),
		genUserName(),
	))

	// Property 1.2: NewUser never assigns teacher or verified_student role
	properties.Property("NewUser never assigns teacher or verified_student role", prop.ForAll(
		func(email, passwordHash, name string) bool {
			user := NewUser(email, passwordHash, name)
			return user.Role != RoleTeacher && user.Role != RoleVerifiedStudent
		},
		genEmail(),
		genPasswordHash(),
		genUserName(),
	))

	// Property 1.3: NewUser sets all required default fields correctly
	properties.Property("NewUser sets all default fields correctly", prop.ForAll(
		func(email, passwordHash, name string) bool {
			user := NewUser(email, passwordHash, name)

			return user.Email == email &&
				user.PasswordHash == passwordHash &&
				user.Name == name &&
				user.Role == RoleStudent &&
				user.EmailVerified == false &&
				user.TwoFactorEnabled == false &&
				user.Language == "id" &&
				user.OnboardingCompleted == false &&
				!user.CreatedAt.IsZero() &&
				!user.UpdatedAt.IsZero()
		},
		genEmail(),
		genPasswordHash(),
		genUserName(),
	))

	properties.TestingRun(t)
}

// Generator for email addresses
func genEmail() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	}).Map(func(s string) string {
		if len(s) == 0 {
			s = "user"
		}
		return s + "@example.com"
	})
}

// Generator for password hashes (simulating bcrypt hashes)
func genPasswordHash() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) == 0 {
			return "$2a$10$defaulthash"
		}
		return "$2a$10$" + s
	})
}

// Generator for user names
func genUserName() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	}).Map(func(s string) string {
		if len(s) == 0 {
			return "Default User"
		}
		return s
	})
}

// **Feature: student-teacher-roles, Property 5: Verification Data Validation**
//
// Property 5: Verification Data Validation
// *For any* teacher verification submission:
// - Document references SHALL be stored without actual file content
// - Required fields (full_name, id_number, credential_type) SHALL be validated
// - Invalid or missing required fields SHALL result in validation error

func TestProperty_VerificationDataValidation(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 5.1: NewTeacherVerification always creates verification with pending status
	properties.Property("new teacher verification always has pending status", prop.ForAll(
		func(fullName, idNumber, documentRef string, credType CredentialType) bool {
			userID := uuid.New()
			verification := NewTeacherVerification(userID, fullName, idNumber, credType, documentRef)

			return verification.Status == VerificationStatusPending &&
				verification.IsPending() &&
				!verification.IsApproved() &&
				!verification.IsRejected()
		},
		genNonEmptyString(),
		genNonEmptyString(),
		genNonEmptyString(),
		genCredentialType(),
	))

	// Property 5.2: Document reference is stored as-is (not modified/processed)
	properties.Property("document reference is stored without modification", prop.ForAll(
		func(fullName, idNumber, documentRef string, credType CredentialType) bool {
			userID := uuid.New()
			verification := NewTeacherVerification(userID, fullName, idNumber, credType, documentRef)

			// Document reference should be stored exactly as provided
			return verification.DocumentRef == documentRef
		},
		genNonEmptyString(),
		genNonEmptyString(),
		genDocumentRef(),
		genCredentialType(),
	))

	// Property 5.3: All required fields are stored correctly
	properties.Property("all required fields are stored correctly", prop.ForAll(
		func(fullName, idNumber, documentRef string, credType CredentialType) bool {
			userID := uuid.New()
			verification := NewTeacherVerification(userID, fullName, idNumber, credType, documentRef)

			return verification.FullName == fullName &&
				verification.IDNumber == idNumber &&
				verification.CredentialType == credType &&
				verification.UserID == userID &&
				verification.ID != uuid.Nil
		},
		genNonEmptyString(),
		genNonEmptyString(),
		genNonEmptyString(),
		genCredentialType(),
	))

	// Property 5.4: Credential type validation - only valid types are accepted
	properties.Property("IsValidCredentialType returns true only for valid types", prop.ForAll(
		func(credType CredentialType) bool {
			isValid := IsValidCredentialType(credType)
			validTypes := ValidCredentialTypes()

			// Check if credType is in validTypes
			found := false
			for _, vt := range validTypes {
				if vt == credType {
					found = true
					break
				}
			}

			return isValid == found
		},
		genCredentialType(),
	))

	// Property 5.5: Invalid credential types are rejected
	properties.Property("invalid credential types are rejected", prop.ForAll(
		func(invalidType string) bool {
			// Skip if the random string happens to be a valid type
			ct := CredentialType(invalidType)
			if ct == CredentialTypeGovernmentID ||
				ct == CredentialTypeEducatorCard ||
				ct == CredentialTypeProfessionalCert {
				return true // Skip this case
			}

			return !IsValidCredentialType(ct)
		},
		gen.AlphaString(),
	))

	// Property 5.6: Verification timestamps are set correctly
	properties.Property("verification timestamps are set on creation", prop.ForAll(
		func(fullName, idNumber, documentRef string, credType CredentialType) bool {
			userID := uuid.New()
			verification := NewTeacherVerification(userID, fullName, idNumber, credType, documentRef)

			return !verification.CreatedAt.IsZero() &&
				!verification.UpdatedAt.IsZero() &&
				verification.ReviewedAt == nil &&
				verification.ReviewedBy == nil &&
				verification.RejectionReason == nil
		},
		genNonEmptyString(),
		genNonEmptyString(),
		genNonEmptyString(),
		genCredentialType(),
	))

	// Property 5.7: Approve sets correct status and reviewer info
	properties.Property("approve sets correct status and reviewer info", prop.ForAll(
		func(fullName, idNumber, documentRef string, credType CredentialType) bool {
			userID := uuid.New()
			reviewerID := uuid.New()
			verification := NewTeacherVerification(userID, fullName, idNumber, credType, documentRef)

			verification.Approve(reviewerID)

			return verification.Status == VerificationStatusApproved &&
				verification.IsApproved() &&
				!verification.IsPending() &&
				!verification.IsRejected() &&
				verification.ReviewedBy != nil &&
				*verification.ReviewedBy == reviewerID &&
				verification.ReviewedAt != nil &&
				verification.RejectionReason == nil
		},
		genNonEmptyString(),
		genNonEmptyString(),
		genNonEmptyString(),
		genCredentialType(),
	))

	// Property 5.8: Reject sets correct status, reviewer info, and reason
	properties.Property("reject sets correct status, reviewer info, and reason", prop.ForAll(
		func(fullName, idNumber, documentRef, reason string, credType CredentialType) bool {
			userID := uuid.New()
			reviewerID := uuid.New()
			verification := NewTeacherVerification(userID, fullName, idNumber, credType, documentRef)

			verification.Reject(reviewerID, reason)

			return verification.Status == VerificationStatusRejected &&
				verification.IsRejected() &&
				!verification.IsPending() &&
				!verification.IsApproved() &&
				verification.ReviewedBy != nil &&
				*verification.ReviewedBy == reviewerID &&
				verification.ReviewedAt != nil &&
				verification.RejectionReason != nil &&
				*verification.RejectionReason == reason
		},
		genNonEmptyString(),
		genNonEmptyString(),
		genNonEmptyString(),
		genNonEmptyString(),
		genCredentialType(),
	))

	properties.TestingRun(t)
}

// Generator for non-empty strings (for required fields)
func genNonEmptyString() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(strings.TrimSpace(s)) > 0
	}).Map(func(s string) string {
		if len(s) == 0 {
			return "default"
		}
		return s
	})
}

// Generator for valid credential types
func genCredentialType() gopter.Gen {
	return gen.OneConstOf(
		CredentialTypeGovernmentID,
		CredentialTypeEducatorCard,
		CredentialTypeProfessionalCert,
	)
}

// Generator for document references (simulating file paths or URLs)
func genDocumentRef() gopter.Gen {
	prefixes := []string{"/documents/", "https://storage.example.com/docs/", "ref-"}
	suffixes := []string{".pdf", "", "-123"}

	return gen.IntRange(0, 2).FlatMap(func(idx interface{}) gopter.Gen {
		i := idx.(int)
		return gen.AlphaString().Map(func(s string) string {
			if len(s) == 0 {
				s = "default"
			}
			return prefixes[i] + s + suffixes[i]
		})
	}, reflect.TypeOf(""))
}
