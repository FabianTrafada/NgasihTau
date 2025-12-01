// Package totp provides TOTP (Time-based One-Time Password) functionality
// for two-factor authentication using the pquerna/otp library.
// Implements requirement 1.1: Two-Factor Authentication.
package totp

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const (
	// Issuer is the name shown in authenticator apps.
	Issuer = "NgasihTau"
	// SecretSize is the size of the TOTP secret in bytes.
	SecretSize = 20
	// BackupCodeCount is the number of backup codes to generate.
	BackupCodeCount = 10
	// BackupCodeLength is the length of each backup code.
	BackupCodeLength = 8
)

// GenerateSecret generates a new TOTP secret for a user.
// Returns the secret key and the provisioning URI for QR code generation.
func GenerateSecret(accountName string) (secret string, uri string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      Issuer,
		AccountName: accountName,
		SecretSize:  SecretSize,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	return key.Secret(), key.URL(), nil
}

// ValidateCode validates a TOTP code against a secret.
// Returns true if the code is valid.
func ValidateCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// GenerateBackupCodes generates a set of backup codes for account recovery.
// Returns the plaintext codes that should be shown to the user once.
func GenerateBackupCodes() ([]string, error) {
	codes := make([]string, BackupCodeCount)
	for i := 0; i < BackupCodeCount; i++ {
		code, err := generateRandomCode(BackupCodeLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		codes[i] = code
	}
	return codes, nil
}

// generateRandomCode generates a random alphanumeric code of the specified length.
func generateRandomCode(length int) (string, error) {
	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode to base32 and take the first 'length' characters
	encoded := base32.StdEncoding.EncodeToString(bytes)
	// Remove padding and convert to lowercase for readability
	code := strings.TrimRight(encoded, "=")
	if len(code) > length {
		code = code[:length]
	}

	// Format as XXXX-XXXX for readability
	if length == 8 {
		return fmt.Sprintf("%s-%s", code[:4], code[4:]), nil
	}

	return code, nil
}

// NormalizeCode removes dashes and spaces from a backup code for comparison.
func NormalizeCode(code string) string {
	code = strings.ReplaceAll(code, "-", "")
	code = strings.ReplaceAll(code, " ", "")
	return strings.ToUpper(code)
}
