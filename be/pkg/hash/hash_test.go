package hash

import (
	"testing"
)

func TestPassword_Success(t *testing.T) {
	password := "SecurePass123!"
	hashed, err := Password(password)
	if err != nil {
		t.Fatalf("Password() error = %v", err)
	}
	if hashed == "" {
		t.Error("Expected non-empty hash")
	}
	if hashed == password {
		t.Error("Hash should not equal plaintext password")
	}
}

func TestPassword_Empty(t *testing.T) {
	_, err := Password("")
	if err != ErrEmptyPassword {
		t.Errorf("Expected ErrEmptyPassword, got %v", err)
	}
}

func TestVerifyPassword_Success(t *testing.T) {
	password := "SecurePass123!"
	hashed, _ := Password(password)

	err := VerifyPassword(password, hashed)
	if err != nil {
		t.Errorf("VerifyPassword() error = %v", err)
	}
}

func TestVerifyPassword_Mismatch(t *testing.T) {
	password := "SecurePass123!"
	hashed, _ := Password(password)

	err := VerifyPassword("WrongPassword!", hashed)
	if err != ErrPasswordMismatch {
		t.Errorf("Expected ErrPasswordMismatch, got %v", err)
	}
}

func TestVerifyPassword_EmptyPassword(t *testing.T) {
	err := VerifyPassword("", "somehash")
	if err != ErrEmptyPassword {
		t.Errorf("Expected ErrEmptyPassword, got %v", err)
	}
}

func TestSHA256_Success(t *testing.T) {
	input := "test-refresh-token"
	hash1, err := SHA256(input)
	if err != nil {
		t.Fatalf("SHA256() error = %v", err)
	}
	if hash1 == "" {
		t.Error("Expected non-empty hash")
	}
	if len(hash1) != 64 { // SHA256 hex is 64 chars
		t.Errorf("Expected 64 char hash, got %d", len(hash1))
	}

	// Same input should produce same hash
	hash2, _ := SHA256(input)
	if hash1 != hash2 {
		t.Error("Same input should produce same hash")
	}
}

func TestSHA256_Empty(t *testing.T) {
	_, err := SHA256("")
	if err != ErrEmptyInput {
		t.Errorf("Expected ErrEmptyInput, got %v", err)
	}
}

func TestSHA256Bytes(t *testing.T) {
	input := []byte("test-data")
	hash := SHA256Bytes(input)
	if hash == "" {
		t.Error("Expected non-empty hash")
	}
	if len(hash) != 64 {
		t.Errorf("Expected 64 char hash, got %d", len(hash))
	}
}
