package validator

import (
	"testing"
)

type testStruct struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,password"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

func TestStruct_Valid(t *testing.T) {
	v := Get()
	input := testStruct{
		Email:    "test@example.com",
		Password: "SecurePass1",
		Name:     "John Doe",
	}

	err := v.Struct(&input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestStruct_MissingRequired(t *testing.T) {
	v := Get()
	input := testStruct{
		Email: "test@example.com",
		// Missing password and name
	}

	err := v.Struct(&input)
	if err == nil {
		t.Error("Expected validation error")
	}
}

func TestStruct_InvalidEmail(t *testing.T) {
	v := Get()
	input := testStruct{
		Email:    "not-an-email",
		Password: "SecurePass1",
		Name:     "John",
	}

	err := v.Struct(&input)
	if err == nil {
		t.Error("Expected validation error for invalid email")
	}
}

func TestPassword_Valid(t *testing.T) {
	tests := []string{
		"SecurePass1",
		"MyP@ssw0rd",
		"Test1234",
		"ABCdef123",
	}

	v := Get()
	for _, pw := range tests {
		err := v.Var(pw, "password")
		if err != nil {
			t.Errorf("Password %q should be valid, got error: %v", pw, err)
		}
	}
}

func TestPassword_Invalid(t *testing.T) {
	tests := []struct {
		password string
		reason   string
	}{
		{"short1A", "too short"},
		{"alllowercase1", "no uppercase"},
		{"ALLUPPERCASE1", "no lowercase"},
		{"NoNumbers!", "no digit"},
	}

	v := Get()
	for _, tc := range tests {
		err := v.Var(tc.password, "password")
		if err == nil {
			t.Errorf("Password %q should be invalid (%s)", tc.password, tc.reason)
		}
	}
}

func TestSlug_Valid(t *testing.T) {
	tests := []string{
		"hello",
		"hello-world",
		"my-cool-slug",
		"test123",
		"a1b2c3",
	}

	v := Get()
	for _, slug := range tests {
		err := v.Var(slug, "slug")
		if err != nil {
			t.Errorf("Slug %q should be valid, got error: %v", slug, err)
		}
	}
}

func TestSlug_Invalid(t *testing.T) {
	tests := []string{
		"Hello",       // uppercase
		"hello_world", // underscore
		"hello world", // space
		"-hello",      // starts with hyphen
		"hello-",      // ends with hyphen
	}

	v := Get()
	for _, slug := range tests {
		err := v.Var(slug, "slug")
		if err == nil {
			t.Errorf("Slug %q should be invalid", slug)
		}
	}
}

func TestUsername_Valid(t *testing.T) {
	tests := []string{
		"john",
		"john_doe",
		"User123",
		"test_user_123",
	}

	v := Get()
	for _, username := range tests {
		err := v.Var(username, "username")
		if err != nil {
			t.Errorf("Username %q should be valid, got error: %v", username, err)
		}
	}
}

func TestUsername_Invalid(t *testing.T) {
	tests := []string{
		"ab",       // too short
		"john-doe", // hyphen not allowed
		"john.doe", // dot not allowed
		"this_username_is_way_too_long_for_validation", // too long
	}

	v := Get()
	for _, username := range tests {
		err := v.Var(username, "username")
		if err == nil {
			t.Errorf("Username %q should be invalid", username)
		}
	}
}

func TestSafeString_Valid(t *testing.T) {
	tests := []string{
		"Hello World",
		"This is a normal string!",
		"Numbers 123 and symbols @#$%",
	}

	v := Get()
	for _, s := range tests {
		err := v.Var(s, "safe_string")
		if err != nil {
			t.Errorf("String %q should be valid, got error: %v", s, err)
		}
	}
}

func TestSafeString_Invalid(t *testing.T) {
	tests := []string{
		"<script>alert('xss')</script>",
		"Hello {world}",
		"Test|pipe",
		"Back\\slash",
	}

	v := Get()
	for _, s := range tests {
		err := v.Var(s, "safe_string")
		if err == nil {
			t.Errorf("String %q should be invalid (contains dangerous chars)", s)
		}
	}
}

func TestUUIDOrEmpty_Valid(t *testing.T) {
	tests := []string{
		"",
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	v := Get()
	for _, s := range tests {
		err := v.Var(s, "uuid_or_empty")
		if err != nil {
			t.Errorf("Value %q should be valid, got error: %v", s, err)
		}
	}
}

func TestUUIDOrEmpty_Invalid(t *testing.T) {
	tests := []string{
		"not-a-uuid",
		"12345",
		"550e8400-e29b-41d4-a716", // incomplete
	}

	v := Get()
	for _, s := range tests {
		err := v.Var(s, "uuid_or_empty")
		if err == nil {
			t.Errorf("Value %q should be invalid", s)
		}
	}
}

func TestVarWithName(t *testing.T) {
	v := Get()
	err := v.VarWithName("ab", "min=5", "custom_field")
	if err == nil {
		t.Error("Expected validation error")
	}
	// Check that custom field name is used
	found := false
	for _, detail := range err.Details {
		if detail.Field == "custom_field" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected custom field name in error details")
	}
}
