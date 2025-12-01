// Package hash provides password hashing and verification using bcrypt,
// as well as SHA-256 hashing for refresh tokens.
package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the bcrypt cost factor used for password hashing.
// Cost 12 provides a good balance between security and performance.
const DefaultCost = 12

// Common errors returned by the hash package.
var (
	ErrEmptyPassword    = errors.New("password cannot be empty")
	ErrEmptyInput       = errors.New("input cannot be empty")
	ErrPasswordMismatch = errors.New("password does not match")
)

// Password hashes a password using bcrypt with the default cost factor (12).
// Returns the hashed password as a string or an error if hashing fails.
func Password(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// VerifyPassword compares a plaintext password with a bcrypt hash.
// Returns nil if the password matches, ErrPasswordMismatch if it doesn't,
// or another error if verification fails.
func VerifyPassword(password, hash string) error {
	if password == "" {
		return ErrEmptyPassword
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordMismatch
		}
		return err
	}

	return nil
}

// SHA256 generates a SHA-256 hash of the input string.
// Used for hashing refresh tokens before storage.
// Returns the hex-encoded hash string.
func SHA256(input string) (string, error) {
	if input == "" {
		return "", ErrEmptyInput
	}

	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:]), nil
}

// SHA256Bytes generates a SHA-256 hash of the input bytes.
// Returns the hex-encoded hash string.
func SHA256Bytes(input []byte) string {
	hash := sha256.Sum256(input)
	return hex.EncodeToString(hash[:])
}
