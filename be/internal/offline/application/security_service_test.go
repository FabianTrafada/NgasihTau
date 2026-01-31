package application

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ngasihtau/internal/offline/domain"
)

func TestSecurityService_ValidateRequestSignature(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	secret := "test-signing-secret"
	service := NewSecurityService(client, nil, nil, secret)
	builder := NewRequestSignatureBuilder(secret)

	ctx := context.Background()

	t.Run("valid signature passes", func(t *testing.T) {
		method := "POST"
		path := "/api/v1/offline/licenses/validate"
		body := []byte(`{"nonce":"test-nonce"}`)
		timestamp := time.Now().Unix()
		nonce := uuid.New().String()

		signature := builder.BuildSignature(method, path, body, timestamp, nonce)

		input := ValidateSignatureInput{
			Method:    method,
			Path:      path,
			Body:      body,
			Timestamp: timestamp,
			Nonce:     nonce,
			Signature: signature,
			DeviceID:  uuid.New().String(),
		}

		err := service.ValidateRequestSignature(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("invalid signature fails", func(t *testing.T) {
		input := ValidateSignatureInput{
			Method:    "POST",
			Path:      "/api/v1/offline/licenses/validate",
			Body:      []byte(`{"nonce":"test-nonce"}`),
			Timestamp: time.Now().Unix(),
			Nonce:     uuid.New().String(),
			Signature: "invalid-signature",
			DeviceID:  uuid.New().String(),
		}

		err := service.ValidateRequestSignature(ctx, input)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInvalidSignature, err)
	})

	t.Run("tampered body fails", func(t *testing.T) {
		method := "POST"
		path := "/api/v1/offline/licenses/validate"
		originalBody := []byte(`{"nonce":"test-nonce"}`)
		tamperedBody := []byte(`{"nonce":"tampered-nonce"}`)
		timestamp := time.Now().Unix()
		nonce := uuid.New().String()

		// Sign with original body
		signature := builder.BuildSignature(method, path, originalBody, timestamp, nonce)

		// Validate with tampered body
		input := ValidateSignatureInput{
			Method:    method,
			Path:      path,
			Body:      tamperedBody,
			Timestamp: timestamp,
			Nonce:     nonce,
			Signature: signature,
			DeviceID:  uuid.New().String(),
		}

		err := service.ValidateRequestSignature(ctx, input)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInvalidSignature, err)
	})
}

func TestSecurityService_CheckReplayAttack(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	service := NewSecurityService(client, nil, nil, "secret")
	ctx := context.Background()

	t.Run("valid timestamp and new nonce passes", func(t *testing.T) {
		nonce := uuid.New().String()
		timestamp := time.Now().Unix()

		err := service.CheckReplayAttack(ctx, nonce, timestamp)
		assert.NoError(t, err)
	})

	t.Run("old timestamp fails", func(t *testing.T) {
		nonce := uuid.New().String()
		timestamp := time.Now().Add(-10 * time.Minute).Unix() // 10 minutes ago

		err := service.CheckReplayAttack(ctx, nonce, timestamp)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrReplayAttack, err)
	})

	t.Run("future timestamp fails", func(t *testing.T) {
		nonce := uuid.New().String()
		timestamp := time.Now().Add(10 * time.Minute).Unix() // 10 minutes in future

		err := service.CheckReplayAttack(ctx, nonce, timestamp)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrReplayAttack, err)
	})

	t.Run("reused nonce fails", func(t *testing.T) {
		nonce := uuid.New().String()
		timestamp := time.Now().Unix()

		// First request should pass
		err := service.CheckReplayAttack(ctx, nonce, timestamp)
		assert.NoError(t, err)

		// Record the nonce
		err = service.RecordRequest(ctx, nonce)
		assert.NoError(t, err)

		// Second request with same nonce should fail
		err = service.CheckReplayAttack(ctx, nonce, timestamp)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrReplayAttack, err)
	})
}

func TestSecurityService_SanitizeInput(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	service := NewSecurityService(client, nil, nil, "secret")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal input unchanged",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "removes null bytes",
			input:    "Hello\x00World",
			expected: "HelloWorld",
		},
		{
			name:     "removes control characters",
			input:    "Hello\x01\x02\x03World",
			expected: "HelloWorld",
		},
		{
			name:     "preserves newlines",
			input:    "Hello\nWorld",
			expected: "Hello\nWorld",
		},
		{
			name:     "preserves tabs",
			input:    "Hello\tWorld",
			expected: "Hello\tWorld",
		},
		{
			name:     "trims whitespace",
			input:    "  Hello World  ",
			expected: "Hello World",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles unicode",
			input:    "Hello 世界 🌍",
			expected: "Hello 世界 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorResponseSanitizer(t *testing.T) {
	sanitizer := NewErrorResponseSanitizer()

	t.Run("sanitizes device errors", func(t *testing.T) {
		status, code, msg := sanitizer.SanitizeError(domain.ErrDeviceNotFound)
		assert.Equal(t, 404, status)
		assert.Equal(t, "DEVICE_NOT_FOUND", code)
		assert.Equal(t, "Device not found", msg)
	})

	t.Run("sanitizes license errors", func(t *testing.T) {
		status, code, msg := sanitizer.SanitizeError(domain.ErrLicenseExpired)
		assert.Equal(t, 403, status)
		assert.Equal(t, "LICENSE_EXPIRED", code)
		assert.Equal(t, "License has expired", msg)
	})

	t.Run("hides internal error details", func(t *testing.T) {
		status, code, msg := sanitizer.SanitizeError(domain.ErrDatabaseError)
		assert.Equal(t, 500, status)
		assert.Equal(t, "DATABASE_ERROR", code)
		assert.Equal(t, "An error occurred", msg)
	})

	t.Run("hides security error details", func(t *testing.T) {
		status, code, msg := sanitizer.SanitizeError(domain.ErrReplayAttack)
		assert.Equal(t, 403, status)
		assert.Equal(t, "REPLAY_ATTACK", code)
		assert.Equal(t, "Request rejected", msg)
	})

	t.Run("handles generic errors", func(t *testing.T) {
		status, code, msg := sanitizer.SanitizeError(assert.AnError)
		assert.Equal(t, 500, status)
		assert.Equal(t, "INTERNAL_ERROR", code)
		assert.Equal(t, "An internal error occurred", msg)
	})
}

func TestCEKTransportEncryption(t *testing.T) {
	encryption := NewCEKTransportEncryption()

	t.Run("encrypt and decrypt round trip", func(t *testing.T) {
		originalCEK := make([]byte, 32)
		for i := range originalCEK {
			originalCEK[i] = byte(i)
		}
		fingerprint := "test-device-fingerprint"

		// Encrypt
		encrypted, err := encryption.EncryptForTransport(originalCEK, fingerprint)
		require.NoError(t, err)
		assert.NotEqual(t, originalCEK, encrypted)

		// Decrypt
		decrypted, err := encryption.DecryptFromTransport(encrypted, fingerprint)
		require.NoError(t, err)
		assert.Equal(t, originalCEK, decrypted)
	})

	t.Run("different fingerprints produce different ciphertext", func(t *testing.T) {
		cek := make([]byte, 32)
		for i := range cek {
			cek[i] = byte(i)
		}

		encrypted1, _ := encryption.EncryptForTransport(cek, "fingerprint1")
		encrypted2, _ := encryption.EncryptForTransport(cek, "fingerprint2")

		assert.NotEqual(t, encrypted1, encrypted2)
	})
}

func TestParseSignatureHeader(t *testing.T) {
	t.Run("valid header", func(t *testing.T) {
		header := "t=1234567890,n=test-nonce,s=test-signature"
		timestamp, nonce, signature, err := ParseSignatureHeader(header)

		require.NoError(t, err)
		assert.Equal(t, int64(1234567890), timestamp)
		assert.Equal(t, "test-nonce", nonce)
		assert.Equal(t, "test-signature", signature)
	})

	t.Run("invalid format - missing parts", func(t *testing.T) {
		header := "t=1234567890,n=test-nonce"
		_, _, _, err := ParseSignatureHeader(header)
		assert.Error(t, err)
	})

	t.Run("invalid format - bad timestamp", func(t *testing.T) {
		header := "t=invalid,n=test-nonce,s=test-signature"
		_, _, _, err := ParseSignatureHeader(header)
		assert.Error(t, err)
	})

	t.Run("invalid format - missing values", func(t *testing.T) {
		header := "t=,n=,s="
		_, _, _, err := ParseSignatureHeader(header)
		assert.Error(t, err)
	})
}

func TestRequestSignatureBuilder(t *testing.T) {
	secret := "test-secret"
	builder := NewRequestSignatureBuilder(secret)

	t.Run("builds consistent signatures", func(t *testing.T) {
		method := "POST"
		path := "/api/test"
		body := []byte(`{"test":"data"}`)
		timestamp := int64(1234567890)
		nonce := "test-nonce"

		sig1 := builder.BuildSignature(method, path, body, timestamp, nonce)
		sig2 := builder.BuildSignature(method, path, body, timestamp, nonce)

		assert.Equal(t, sig1, sig2)
	})

	t.Run("different inputs produce different signatures", func(t *testing.T) {
		body := []byte(`{"test":"data"}`)
		timestamp := int64(1234567890)
		nonce := "test-nonce"

		sig1 := builder.BuildSignature("POST", "/api/test", body, timestamp, nonce)
		sig2 := builder.BuildSignature("GET", "/api/test", body, timestamp, nonce)

		assert.NotEqual(t, sig1, sig2)
	})
}
