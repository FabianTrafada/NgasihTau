package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/redis/go-redis/v9"

	"ngasihtau/internal/offline/domain"
)

// TestProperty4_ErrorResponseInformationHiding tests Property 4: Error Response Information Hiding.
// Property: Error responses must not leak internal implementation details.
func TestProperty4_ErrorResponseInformationHiding(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 50

	properties := gopter.NewProperties(parameters)

	sanitizer := NewErrorResponseSanitizer()

	// Property 4.1: Internal errors never expose stack traces
	properties.Property("internal errors never expose stack traces", prop.ForAll(
		func(errorCode int) bool {
			// Create various error types
			var err error
			switch errorCode % 5 {
			case 0:
				err = domain.ErrDatabaseError
			case 1:
				err = domain.ErrStorageError
			case 2:
				err = domain.ErrCacheError
			case 3:
				err = domain.ErrInternalError
			case 4:
				err = domain.ErrKeyGenerationFailed
			}

			_, _, msg := sanitizer.SanitizeError(err)

			// Message should not contain stack trace indicators
			forbidden := []string{"goroutine", "panic", "runtime", ".go:", "stack"}
			for _, f := range forbidden {
				if strings.Contains(strings.ToLower(msg), f) {
					return false
				}
			}

			return true
		},
		gen.IntRange(0, 100),
	))

	// Property 4.2: Security errors use generic messages
	properties.Property("security errors use generic messages", prop.ForAll(
		func(errorType int) bool {
			var err error
			switch errorType % 3 {
			case 0:
				err = domain.ErrReplayAttack
			case 1:
				err = domain.ErrInvalidSignature
			case 2:
				err = domain.ErrInvalidRequest
			}

			_, _, msg := sanitizer.SanitizeError(err)

			// Security error messages should be generic
			genericMessages := []string{"Request rejected", "Invalid request"}
			for _, gm := range genericMessages {
				if msg == gm {
					return true
				}
			}

			return false
		},
		gen.IntRange(0, 100),
	))

	// Property 4.3: Error codes are always valid
	properties.Property("error codes are always valid", prop.ForAll(
		func(errorType int) bool {
			errors := []*domain.OfflineError{
				domain.ErrDeviceNotFound,
				domain.ErrLicenseExpired,
				domain.ErrMaterialAccessDenied,
				domain.ErrRateLimitExceeded,
				domain.ErrInternalError,
			}

			err := errors[errorType%len(errors)]
			status, code, _ := sanitizer.SanitizeError(err)

			// Status should be valid HTTP status
			if status < 100 || status > 599 {
				return false
			}

			// Code should not be empty
			if code == "" {
				return false
			}

			return true
		},
		gen.IntRange(0, 100),
	))

	// Property 4.4: Messages never contain UUIDs or IDs
	properties.Property("messages never contain UUIDs or IDs", prop.ForAll(
		func(dummy int) bool {
			errors := []*domain.OfflineError{
				domain.ErrDeviceNotFound.WithDeviceID(uuid.New()),
				domain.ErrLicenseNotFound.WithLicenseID(uuid.New()),
				domain.ErrMaterialNotFound.WithMaterialID(uuid.New()),
			}

			for _, err := range errors {
				_, _, msg := sanitizer.SanitizeError(err)

				// Check for UUID patterns
				if strings.Contains(msg, "-") && len(msg) > 36 {
					// Might contain UUID
					parts := strings.Split(msg, "-")
					if len(parts) == 5 {
						return false
					}
				}
			}

			return true
		},
		gen.IntRange(0, 10),
	))

	properties.TestingRun(t)
}

// TestProperty5_AuditEventCompleteness tests Property 5: Audit Event Completeness.
// Property: All security-relevant operations must generate complete audit events.
func TestProperty5_AuditEventCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	// Property 5.1: Audit events have all required fields
	properties.Property("audit events have all required fields", prop.ForAll(
		func(actionIndex int) bool {
			actions := []string{
				domain.AuditActionKeyGenerate,
				domain.AuditActionLicenseIssue,
				domain.AuditActionDeviceRegister,
				domain.AuditActionMaterialDownload,
			}

			action := actions[actionIndex%len(actions)]
			userID := uuid.New()
			resourceID := uuid.New()

			event := domain.NewAuditLog(
				userID,
				nil,
				action,
				domain.AuditResourceLicense,
				resourceID,
				"192.168.1.1",
				"TestAgent/1.0",
				true,
				nil,
			)

			// Check required fields
			if event.ID == uuid.Nil {
				return false
			}
			if event.UserID == uuid.Nil {
				return false
			}
			if event.Action == "" {
				return false
			}
			if event.Resource == "" {
				return false
			}
			if event.ResourceID == uuid.Nil {
				return false
			}
			if event.CreatedAt.IsZero() {
				return false
			}

			return true
		},
		gen.IntRange(0, 100),
	))

	// Property 5.2: Failed operations include error codes
	properties.Property("failed operations include error codes", prop.ForAll(
		func(errorCodeIndex int) bool {
			errorCodes := []string{
				"DEVICE_NOT_FOUND",
				"LICENSE_EXPIRED",
				"RATE_LIMIT_EXCEEDED",
			}

			errorCode := errorCodes[errorCodeIndex%len(errorCodes)]

			event := domain.NewAuditLog(
				uuid.New(),
				nil,
				domain.AuditActionLicenseValidate,
				domain.AuditResourceLicense,
				uuid.New(),
				"192.168.1.1",
				"TestAgent/1.0",
				false,
				&errorCode,
			)

			// Failed events should have error code
			if !event.Success && event.ErrorCode == nil {
				return false
			}

			return true
		},
		gen.IntRange(0, 100),
	))

	// Property 5.3: Timestamps are always set
	properties.Property("timestamps are always set", prop.ForAll(
		func(dummy int) bool {
			event := domain.NewAuditLog(
				uuid.New(),
				nil,
				domain.AuditActionDeviceRegister,
				domain.AuditResourceDevice,
				uuid.New(),
				"192.168.1.1",
				"TestAgent/1.0",
				true,
				nil,
			)

			return !event.CreatedAt.IsZero() && event.CreatedAt.Before(time.Now().Add(time.Second))
		},
		gen.IntRange(0, 50),
	))

	properties.TestingRun(t)
}

// TestProperty25_CEKTransportEncryption tests Property 25: CEK Transport Encryption.
// Property: CEKs must be encrypted for transport and decryptable only by the intended device.
func TestProperty25_CEKTransportEncryption(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 50

	properties := gopter.NewProperties(parameters)

	encryption := NewCEKTransportEncryption()

	// Property 25.1: Encryption is reversible with correct fingerprint
	properties.Property("encryption is reversible with correct fingerprint", prop.ForAll(
		func(cekBytes []byte, fingerprint string) bool {
			if len(cekBytes) == 0 || len(fingerprint) == 0 {
				return true // Skip empty inputs
			}

			// Pad or truncate to 32 bytes
			cek := make([]byte, 32)
			copy(cek, cekBytes)

			encrypted, err := encryption.EncryptForTransport(cek, fingerprint)
			if err != nil {
				return false
			}

			decrypted, err := encryption.DecryptFromTransport(encrypted, fingerprint)
			if err != nil {
				return false
			}

			// Compare
			for i := range cek {
				if cek[i] != decrypted[i] {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(32, gen.UInt8()),
		gen.AnyString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	// Property 25.2: Different fingerprints produce different ciphertext
	properties.Property("different fingerprints produce different ciphertext", prop.ForAll(
		func(cekBytes []byte, fp1, fp2 string) bool {
			if len(cekBytes) == 0 || fp1 == fp2 || len(fp1) == 0 || len(fp2) == 0 {
				return true // Skip invalid inputs
			}

			cek := make([]byte, 32)
			copy(cek, cekBytes)

			encrypted1, _ := encryption.EncryptForTransport(cek, fp1)
			encrypted2, _ := encryption.EncryptForTransport(cek, fp2)

			// Should be different
			different := false
			for i := range encrypted1 {
				if encrypted1[i] != encrypted2[i] {
					different = true
					break
				}
			}

			return different
		},
		gen.SliceOfN(32, gen.UInt8()),
		gen.AnyString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AnyString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	// Property 25.3: Wrong fingerprint cannot decrypt
	properties.Property("wrong fingerprint cannot decrypt correctly", prop.ForAll(
		func(cekBytes []byte, correctFP, wrongFP string) bool {
			if len(cekBytes) == 0 || correctFP == wrongFP || len(correctFP) == 0 || len(wrongFP) == 0 {
				return true
			}

			cek := make([]byte, 32)
			copy(cek, cekBytes)

			encrypted, _ := encryption.EncryptForTransport(cek, correctFP)
			decrypted, _ := encryption.DecryptFromTransport(encrypted, wrongFP)

			// Should not match original
			matches := true
			for i := range cek {
				if cek[i] != decrypted[i] {
					matches = false
					break
				}
			}

			return !matches
		},
		gen.SliceOfN(32, gen.UInt8()),
		gen.AnyString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AnyString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t)
}

// TestProperty26_RequestReplayProtection tests Property 26: Request Replay Protection.
// Property: Requests with reused nonces or expired timestamps must be rejected.
func TestProperty26_RequestReplayProtection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	// Property 26.1: Fresh requests are accepted
	properties.Property("fresh requests are accepted", prop.ForAll(
		func(dummy int) bool {
			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			service := NewSecurityService(client, nil, nil, "secret")
			ctx := context.Background()

			nonce := uuid.New().String()
			timestamp := time.Now().Unix()

			err := service.CheckReplayAttack(ctx, nonce, timestamp)
			return err == nil
		},
		gen.IntRange(0, 50),
	))

	// Property 26.2: Reused nonces are rejected
	properties.Property("reused nonces are rejected", prop.ForAll(
		func(dummy int) bool {
			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			service := NewSecurityService(client, nil, nil, "secret")
			ctx := context.Background()

			nonce := uuid.New().String()
			timestamp := time.Now().Unix()

			// First request
			err := service.CheckReplayAttack(ctx, nonce, timestamp)
			if err != nil {
				return false
			}

			// Record nonce
			err = service.RecordRequest(ctx, nonce)
			if err != nil {
				return false
			}

			// Second request with same nonce
			err = service.CheckReplayAttack(ctx, nonce, timestamp)
			return err == domain.ErrReplayAttack
		},
		gen.IntRange(0, 50),
	))

	// Property 26.3: Old timestamps are rejected
	properties.Property("old timestamps are rejected", prop.ForAll(
		func(minutesOld int) bool {
			if minutesOld < 6 {
				return true // Only test timestamps older than 5 min window
			}

			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			service := NewSecurityService(client, nil, nil, "secret")
			ctx := context.Background()

			nonce := uuid.New().String()
			timestamp := time.Now().Add(-time.Duration(minutesOld) * time.Minute).Unix()

			err := service.CheckReplayAttack(ctx, nonce, timestamp)
			return err == domain.ErrReplayAttack
		},
		gen.IntRange(6, 60),
	))

	// Property 26.4: Future timestamps beyond tolerance are rejected
	properties.Property("future timestamps beyond tolerance are rejected", prop.ForAll(
		func(minutesFuture int) bool {
			if minutesFuture < 2 {
				return true // Allow 60s clock skew
			}

			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			service := NewSecurityService(client, nil, nil, "secret")
			ctx := context.Background()

			nonce := uuid.New().String()
			timestamp := time.Now().Add(time.Duration(minutesFuture) * time.Minute).Unix()

			err := service.CheckReplayAttack(ctx, nonce, timestamp)
			return err == domain.ErrReplayAttack
		},
		gen.IntRange(2, 30),
	))

	properties.TestingRun(t)
}

// TestProperty29_InputSanitization tests Property 29: Input Sanitization.
// Property: All user inputs must be sanitized to prevent injection attacks.
func TestProperty29_InputSanitization(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 100

	properties := gopter.NewProperties(parameters)

	mr, _ := miniredis.Run()
	defer mr.Close()
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	service := NewSecurityService(client, nil, nil, "secret")

	// Property 29.1: Null bytes are removed
	properties.Property("null bytes are removed", prop.ForAll(
		func(input string) bool {
			// Add null bytes
			withNulls := input + "\x00" + input
			sanitized := service.SanitizeInput(withNulls)

			return !strings.Contains(sanitized, "\x00")
		},
		gen.AnyString(),
	))

	// Property 29.2: Control characters are removed (except newline/tab)
	properties.Property("control characters are removed", prop.ForAll(
		func(input string) bool {
			// Add control characters
			withControl := input + "\x01\x02\x03\x04\x05"
			sanitized := service.SanitizeInput(withControl)

			for _, r := range sanitized {
				if r < 32 && r != '\n' && r != '\t' {
					return false
				}
			}

			return true
		},
		gen.AnyString(),
	))

	// Property 29.3: Whitespace is trimmed
	properties.Property("whitespace is trimmed", prop.ForAll(
		func(input string) bool {
			withWhitespace := "   " + input + "   "
			sanitized := service.SanitizeInput(withWhitespace)

			if len(sanitized) == 0 {
				return true
			}

			return sanitized[0] != ' ' && sanitized[len(sanitized)-1] != ' '
		},
		gen.AnyString(),
	))

	// Property 29.4: Valid content is preserved
	properties.Property("valid content is preserved", prop.ForAll(
		func(input string) bool {
			// Filter to only valid characters
			var valid strings.Builder
			for _, r := range input {
				if r == '\n' || r == '\t' || r >= 32 {
					valid.WriteRune(r)
				}
			}
			validInput := strings.TrimSpace(valid.String())

			sanitized := service.SanitizeInput(validInput)

			return sanitized == validInput
		},
		gen.AnyString(),
	))

	// Property 29.5: Newlines and tabs are preserved
	properties.Property("newlines and tabs are preserved", prop.ForAll(
		func(input string) bool {
			withNewlines := "line1\nline2\ttabbed"
			sanitized := service.SanitizeInput(withNewlines)

			return strings.Contains(sanitized, "\n") && strings.Contains(sanitized, "\t")
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}
