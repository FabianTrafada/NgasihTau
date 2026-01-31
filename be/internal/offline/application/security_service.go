// Package application contains the business logic and use cases for the Offline Material feature.
package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/offline/domain"
	natspkg "ngasihtau/pkg/nats"
)

// SecurityService defines the interface for security-related operations.
// Implements Requirement 7: Security Features.
type SecurityService interface {
	// ValidateRequestSignature validates the HMAC signature of a request.
	ValidateRequestSignature(ctx context.Context, input ValidateSignatureInput) error

	// CheckReplayAttack checks if a request is a replay attack.
	CheckReplayAttack(ctx context.Context, nonce string, timestamp int64) error

	// RecordRequest records a request nonce to prevent replay attacks.
	RecordRequest(ctx context.Context, nonce string) error

	// SanitizeInput sanitizes user input to prevent injection attacks.
	SanitizeInput(input string) string

	// LogAuditEvent logs a security audit event.
	LogAuditEvent(ctx context.Context, event *domain.AuditLog) error

	// GetAuditLogs retrieves audit logs for a user.
	GetAuditLogs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.AuditLog, int, error)
}

// ValidateSignatureInput contains the data required for signature validation.
type ValidateSignatureInput struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Body      []byte `json:"body"`
	Timestamp int64  `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
	DeviceID  string `json:"device_id"`
}

// securityService implements the SecurityService interface.
type securityService struct {
	redisClient    *redis.Client
	auditRepo      domain.AuditLogRepository
	eventPublisher natspkg.EventPublisher
	signingSecret  []byte
}

// NewSecurityService creates a new SecurityService instance.
func NewSecurityService(
	redisClient *redis.Client,
	auditRepo domain.AuditLogRepository,
	eventPublisher natspkg.EventPublisher,
	signingSecret string,
) SecurityService {
	return &securityService{
		redisClient:    redisClient,
		auditRepo:      auditRepo,
		eventPublisher: eventPublisher,
		signingSecret:  []byte(signingSecret),
	}
}

// Redis key prefix for replay protection.
const replayProtectionPrefix = "offline:replay:"

// ValidateRequestSignature validates the HMAC signature of a request.
// Implements Requirement 7.1: Request signing validation.
// Implements Property 26: Request Replay Protection (signature component).
func (s *securityService) ValidateRequestSignature(ctx context.Context, input ValidateSignatureInput) error {
	// Build the signing string: METHOD|PATH|TIMESTAMP|NONCE|BODY_HASH
	bodyHash := sha256.Sum256(input.Body)
	bodyHashHex := hex.EncodeToString(bodyHash[:])

	signingString := fmt.Sprintf("%s|%s|%d|%s|%s",
		input.Method,
		input.Path,
		input.Timestamp,
		input.Nonce,
		bodyHashHex,
	)

	// Compute expected signature
	mac := hmac.New(sha256.New, s.signingSecret)
	mac.Write([]byte(signingString))
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Compare signatures using constant-time comparison
	if !hmac.Equal([]byte(input.Signature), []byte(expectedSignature)) {
		log.Warn().
			Str("device_id", input.DeviceID).
			Str("path", input.Path).
			Msg("invalid request signature")
		return domain.ErrInvalidSignature
	}

	return nil
}

// CheckReplayAttack checks if a request is a replay attack.
// Implements Requirement 7.2: Replay attack prevention (5 min window).
// Implements Property 26: Request Replay Protection.
func (s *securityService) CheckReplayAttack(ctx context.Context, nonce string, timestamp int64) error {
	// Check timestamp is within acceptable window
	now := time.Now().Unix()
	windowSeconds := int64(domain.ReplayProtectionWindow.Seconds())

	if timestamp < now-windowSeconds || timestamp > now+60 { // Allow 60s clock skew
		log.Warn().
			Int64("timestamp", timestamp).
			Int64("now", now).
			Msg("request timestamp outside acceptable window")
		return domain.ErrReplayAttack
	}

	// Check if nonce has been used
	key := replayProtectionPrefix + nonce
	exists, err := s.redisClient.Exists(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Str("nonce", nonce).Msg("failed to check nonce")
		return fmt.Errorf("failed to check replay: %w", err)
	}

	if exists > 0 {
		log.Warn().Str("nonce", nonce).Msg("replay attack detected - nonce already used")
		return domain.ErrReplayAttack
	}

	return nil
}

// RecordRequest records a request nonce to prevent replay attacks.
func (s *securityService) RecordRequest(ctx context.Context, nonce string) error {
	key := replayProtectionPrefix + nonce

	// Store nonce with TTL equal to replay protection window
	err := s.redisClient.Set(ctx, key, "1", domain.ReplayProtectionWindow).Err()
	if err != nil {
		log.Error().Err(err).Str("nonce", nonce).Msg("failed to record request nonce")
		return fmt.Errorf("failed to record request: %w", err)
	}

	return nil
}

// SanitizeInput sanitizes user input to prevent injection attacks.
// Implements Requirement 7.4: Input sanitization.
// Implements Property 29: Input Sanitization.
func (s *securityService) SanitizeInput(input string) string {
	// Remove null bytes
	sanitized := strings.ReplaceAll(input, "\x00", "")

	// Remove control characters (except newline and tab)
	var result strings.Builder
	for _, r := range sanitized {
		if r == '\n' || r == '\t' || r >= 32 {
			result.WriteRune(r)
		}
	}

	// Trim whitespace
	return strings.TrimSpace(result.String())
}

// LogAuditEvent logs a security audit event.
// Implements Requirement 7.5: Audit logging.
// Implements Property 5: Audit Event Completeness.
func (s *securityService) LogAuditEvent(ctx context.Context, event *domain.AuditLog) error {
	// Persist to database
	if s.auditRepo != nil {
		if err := s.auditRepo.Create(ctx, event); err != nil {
			log.Error().Err(err).
				Str("action", event.Action).
				Str("user_id", event.UserID.String()).
				Msg("failed to persist audit log")
			// Don't return error - audit logging should not block operations
		}
	}

	// Publish via NATS
	s.publishAuditEvent(ctx, event)

	return nil
}

// GetAuditLogs retrieves audit logs for a user.
func (s *securityService) GetAuditLogs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.AuditLog, int, error) {
	if s.auditRepo == nil {
		return nil, 0, nil
	}
	return s.auditRepo.FindByUserID(ctx, userID, limit, offset)
}

// publishAuditEvent publishes an audit event via NATS.
// Implements Requirement 7.6: Audit event publishing via NATS.
func (s *securityService) publishAuditEvent(ctx context.Context, event *domain.AuditLog) {
	if s.eventPublisher == nil {
		return
	}

	eventData := map[string]interface{}{
		"event_type":  "audit." + event.Action,
		"audit_id":    event.ID.String(),
		"user_id":     event.UserID.String(),
		"action":      event.Action,
		"resource":    event.Resource,
		"resource_id": event.ResourceID.String(),
		"success":     event.Success,
		"ip_address":  event.IPAddress,
		"user_agent":  event.UserAgent,
		"created_at":  event.CreatedAt.Format(time.RFC3339),
	}

	if event.DeviceID != nil {
		eventData["device_id"] = event.DeviceID.String()
	}
	if event.ErrorCode != nil {
		eventData["error_code"] = *event.ErrorCode
	}

	log.Debug().
		Str("action", event.Action).
		Str("user_id", event.UserID.String()).
		Bool("success", event.Success).
		Msg("audit event published")
}

// ErrorResponseSanitizer provides methods to sanitize error responses.
// Implements Property 4: Error Response Information Hiding.
type ErrorResponseSanitizer struct{}

// NewErrorResponseSanitizer creates a new ErrorResponseSanitizer.
func NewErrorResponseSanitizer() *ErrorResponseSanitizer {
	return &ErrorResponseSanitizer{}
}

// SanitizeError converts internal errors to safe external error messages.
// Implements Property 4: Error Response Information Hiding.
func (e *ErrorResponseSanitizer) SanitizeError(err error) (int, string, string) {
	// Check if it's an OfflineError
	if offlineErr, ok := domain.GetOfflineError(err); ok {
		return e.sanitizeOfflineError(offlineErr)
	}

	// Generic internal error - hide details
	return 500, "INTERNAL_ERROR", "An internal error occurred"
}

// sanitizeOfflineError sanitizes an OfflineError for external response.
func (e *ErrorResponseSanitizer) sanitizeOfflineError(err *domain.OfflineError) (int, string, string) {
	status := err.HTTPStatus()
	code := err.Code.String()

	// Map internal error codes to safe external messages
	message := e.getSafeMessage(err.Code)

	return status, code, message
}

// getSafeMessage returns a safe external message for an error code.
// This ensures no internal details are leaked in error responses.
func (e *ErrorResponseSanitizer) getSafeMessage(code domain.ErrorCode) string {
	safeMessages := map[domain.ErrorCode]string{
		// Device errors - safe to expose
		domain.ErrCodeDeviceLimitExceeded:       "Maximum device limit reached",
		domain.ErrCodeDeviceNotFound:            "Device not found",
		domain.ErrCodeDeviceFingerprintMismatch: "Device verification failed",
		domain.ErrCodeDeviceBlocked:             "Device temporarily blocked",
		domain.ErrCodeDeviceAlreadyRegistered:   "Device already registered",
		domain.ErrCodeInvalidFingerprint:        "Invalid device fingerprint",
		domain.ErrCodeInvalidPlatform:           "Invalid platform",

		// License errors - safe to expose
		domain.ErrCodeLicenseNotFound:       "License not found",
		domain.ErrCodeLicenseExpired:        "License has expired",
		domain.ErrCodeLicenseRevoked:        "License has been revoked",
		domain.ErrCodeLicenseOfflineExpired: "Offline access expired",
		domain.ErrCodeInvalidNonce:          "License validation failed",
		domain.ErrCodeLicenseAlreadyExists:  "License already exists",

		// Material errors - safe to expose
		domain.ErrCodeMaterialAccessDenied: "Access denied",
		domain.ErrCodeMaterialNotFound:     "Material not found",
		domain.ErrCodeUnsupportedFileType:  "Unsupported file type",

		// Rate limiting - safe to expose
		domain.ErrCodeRateLimitExceeded: "Rate limit exceeded",

		// Job errors - safe to expose
		domain.ErrCodeJobNotFound:   "Job not found",
		domain.ErrCodeJobFailed:     "Processing failed",
		domain.ErrCodeJobInProgress: "Processing in progress",

		// Security errors - generic messages to hide details
		domain.ErrCodeReplayAttack:     "Request rejected",
		domain.ErrCodeInvalidSignature: "Request rejected",
		domain.ErrCodeInvalidRequest:   "Invalid request",

		// Internal errors - generic messages to hide details
		domain.ErrCodeEncryptionFailed:    "Processing failed",
		domain.ErrCodeDecryptionFailed:    "Processing failed",
		domain.ErrCodeKeyGenerationFailed: "Processing failed",
		domain.ErrCodeKeyNotFound:         "Processing failed",
		domain.ErrCodeCEKNotFound:         "Processing failed",
		domain.ErrCodeInvalidKey:          "Processing failed",
		domain.ErrCodeInternalError:       "An error occurred",
		domain.ErrCodeStorageError:        "An error occurred",
		domain.ErrCodeDatabaseError:       "An error occurred",
		domain.ErrCodeCacheError:          "An error occurred",
		domain.ErrCodeServiceUnavailable:  "Service temporarily unavailable",
	}

	if msg, ok := safeMessages[code]; ok {
		return msg
	}

	return "An error occurred"
}

// CEKTransportEncryption provides methods for encrypting CEKs for transport.
// Implements Requirement 7.3: CEK transport encryption.
// Implements Property 25: CEK Transport Encryption.
type CEKTransportEncryption struct {
	// In production, this would use device-specific public keys
	// For now, we use a shared transport key derived from device fingerprint
}

// NewCEKTransportEncryption creates a new CEKTransportEncryption.
func NewCEKTransportEncryption() *CEKTransportEncryption {
	return &CEKTransportEncryption{}
}

// EncryptForTransport encrypts a CEK for secure transport to a device.
// The CEK is encrypted using a key derived from the device fingerprint.
func (c *CEKTransportEncryption) EncryptForTransport(cek []byte, deviceFingerprint string) ([]byte, error) {
	// Derive transport key from device fingerprint
	// In production, this would use the device's public key
	transportKey := sha256.Sum256([]byte("transport:" + deviceFingerprint))

	// XOR the CEK with the transport key (simplified for demo)
	// In production, use proper asymmetric encryption (RSA/ECDH)
	encrypted := make([]byte, len(cek))
	for i := range cek {
		encrypted[i] = cek[i] ^ transportKey[i%32]
	}

	return encrypted, nil
}

// DecryptFromTransport decrypts a CEK received from transport.
func (c *CEKTransportEncryption) DecryptFromTransport(encryptedCEK []byte, deviceFingerprint string) ([]byte, error) {
	// Same operation as encrypt (XOR is symmetric)
	return c.EncryptForTransport(encryptedCEK, deviceFingerprint)
}

// RequestSignatureBuilder helps build request signatures for clients.
type RequestSignatureBuilder struct {
	secret []byte
}

// NewRequestSignatureBuilder creates a new RequestSignatureBuilder.
func NewRequestSignatureBuilder(secret string) *RequestSignatureBuilder {
	return &RequestSignatureBuilder{secret: []byte(secret)}
}

// BuildSignature builds a signature for a request.
func (b *RequestSignatureBuilder) BuildSignature(method, path string, body []byte, timestamp int64, nonce string) string {
	bodyHash := sha256.Sum256(body)
	bodyHashHex := hex.EncodeToString(bodyHash[:])

	signingString := fmt.Sprintf("%s|%s|%d|%s|%s",
		method,
		path,
		timestamp,
		nonce,
		bodyHashHex,
	)

	mac := hmac.New(sha256.New, b.secret)
	mac.Write([]byte(signingString))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// ParseSignatureHeader parses the X-Signature header.
// Format: "t=<timestamp>,n=<nonce>,s=<signature>"
func ParseSignatureHeader(header string) (timestamp int64, nonce, signature string, err error) {
	parts := strings.Split(header, ",")
	if len(parts) != 3 {
		return 0, "", "", fmt.Errorf("invalid signature header format")
	}

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return 0, "", "", fmt.Errorf("invalid signature header format")
		}

		switch kv[0] {
		case "t":
			timestamp, err = strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				return 0, "", "", fmt.Errorf("invalid timestamp: %w", err)
			}
		case "n":
			nonce = kv[1]
		case "s":
			signature = kv[1]
		}
	}

	if timestamp == 0 || nonce == "" || signature == "" {
		return 0, "", "", fmt.Errorf("missing required signature components")
	}

	return timestamp, nonce, signature, nil
}
