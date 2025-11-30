package validator

import (
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Regular expressions for custom validations
var (
	// slugRegex matches valid URL slugs: lowercase letters, numbers, and hyphens
	slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

	// usernameRegex matches valid usernames: 3-30 chars, alphanumeric with underscores
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

	// safeStringRegex matches strings without dangerous characters
	safeStringRegex = regexp.MustCompile(`^[^<>{}|\\\^~\[\]` + "`" + `]*$`)
)

// registerCustomRules registers all custom validation rules.
func registerCustomRules(v *validator.Validate) {
	// Register custom validators
	v.RegisterValidation("slug", validateSlug)
	v.RegisterValidation("password", validatePassword)
	v.RegisterValidation("username", validateUsername)
	v.RegisterValidation("safe_string", validateSafeString)
	v.RegisterValidation("uuid_or_empty", validateUUIDOrEmpty)
}

// validateSlug validates that a string is a valid URL slug.
// Valid slugs contain only lowercase letters, numbers, and hyphens.
// They cannot start or end with a hyphen, and cannot have consecutive hyphens.
func validateSlug(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Empty is handled by 'required' tag
	}
	return slugRegex.MatchString(value)
}

// validatePassword validates password strength requirements.
// Password must be at least 8 characters and contain:
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one digit
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}

// validateUsername validates username format.
// Username must be 3-30 characters, containing only alphanumeric characters and underscores.
func validateUsername(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Empty is handled by 'required' tag
	}
	return usernameRegex.MatchString(value)
}

// validateSafeString validates that a string doesn't contain potentially dangerous characters.
// This helps prevent XSS and injection attacks.
func validateSafeString(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}
	return safeStringRegex.MatchString(value)
}

// validateUUIDOrEmpty validates that a string is either empty or a valid UUID.
// Useful for optional UUID fields.
func validateUUIDOrEmpty(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}
	_, err := uuid.Parse(value)
	return err == nil
}
