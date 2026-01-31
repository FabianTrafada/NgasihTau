// Package domain contains the core business entities and error definitions
// for the Offline Material Service.
package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// ErrorCode represents offline-specific error codes.
type ErrorCode string

// Offline-specific error codes as defined in the design document.
const (
	// Device errors
	ErrCodeDeviceLimitExceeded       ErrorCode = "DEVICE_LIMIT_EXCEEDED"
	ErrCodeDeviceNotFound            ErrorCode = "DEVICE_NOT_FOUND"
	ErrCodeDeviceFingerprintMismatch ErrorCode = "DEVICE_FINGERPRINT_MISMATCH"
	ErrCodeDeviceBlocked             ErrorCode = "DEVICE_BLOCKED"
	ErrCodeDeviceAlreadyRegistered   ErrorCode = "DEVICE_ALREADY_REGISTERED"
	ErrCodeInvalidFingerprint        ErrorCode = "INVALID_FINGERPRINT"
	ErrCodeInvalidPlatform           ErrorCode = "INVALID_PLATFORM"

	// License errors
	ErrCodeLicenseNotFound       ErrorCode = "LICENSE_NOT_FOUND"
	ErrCodeLicenseExpired        ErrorCode = "LICENSE_EXPIRED"
	ErrCodeLicenseRevoked        ErrorCode = "LICENSE_REVOKED"
	ErrCodeLicenseOfflineExpired ErrorCode = "LICENSE_OFFLINE_EXPIRED"
	ErrCodeInvalidNonce          ErrorCode = "INVALID_NONCE"
	ErrCodeLicenseAlreadyExists  ErrorCode = "LICENSE_ALREADY_EXISTS"

	// Material errors
	ErrCodeMaterialAccessDenied ErrorCode = "MATERIAL_ACCESS_DENIED"
	ErrCodeMaterialNotFound     ErrorCode = "MATERIAL_NOT_FOUND"
	ErrCodeUnsupportedFileType  ErrorCode = "UNSUPPORTED_FILE_TYPE"

	// Encryption errors
	ErrCodeEncryptionFailed    ErrorCode = "ENCRYPTION_FAILED"
	ErrCodeDecryptionFailed    ErrorCode = "DECRYPTION_FAILED"
	ErrCodeKeyGenerationFailed ErrorCode = "KEY_GENERATION_FAILED"
	ErrCodeKeyNotFound         ErrorCode = "KEY_NOT_FOUND"
	ErrCodeCEKNotFound         ErrorCode = "CEK_NOT_FOUND"
	ErrCodeInvalidKey          ErrorCode = "INVALID_KEY"

	// Rate limiting errors
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Job errors
	ErrCodeJobNotFound   ErrorCode = "JOB_NOT_FOUND"
	ErrCodeJobFailed     ErrorCode = "JOB_FAILED"
	ErrCodeJobInProgress ErrorCode = "JOB_IN_PROGRESS"

	// Security errors
	ErrCodeReplayAttack     ErrorCode = "REPLAY_ATTACK"
	ErrCodeInvalidSignature ErrorCode = "INVALID_SIGNATURE"
	ErrCodeInvalidRequest   ErrorCode = "INVALID_REQUEST"

	// Internal errors
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrCodeStorageError       ErrorCode = "STORAGE_ERROR"
	ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrCodeCacheError         ErrorCode = "CACHE_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// HTTPStatus returns the HTTP status code for an error code.
func (c ErrorCode) HTTPStatus() int {
	switch c {
	// 400 Bad Request
	case ErrCodeInvalidFingerprint, ErrCodeInvalidPlatform, ErrCodeInvalidRequest,
		ErrCodeInvalidKey, ErrCodeUnsupportedFileType:
		return 400

	// 403 Forbidden
	case ErrCodeDeviceLimitExceeded, ErrCodeDeviceFingerprintMismatch,
		ErrCodeLicenseExpired, ErrCodeLicenseRevoked, ErrCodeLicenseOfflineExpired,
		ErrCodeMaterialAccessDenied, ErrCodeInvalidNonce, ErrCodeReplayAttack,
		ErrCodeInvalidSignature:
		return 403

	// 404 Not Found
	case ErrCodeDeviceNotFound, ErrCodeLicenseNotFound, ErrCodeMaterialNotFound,
		ErrCodeKeyNotFound, ErrCodeCEKNotFound, ErrCodeJobNotFound:
		return 404

	// 409 Conflict
	case ErrCodeDeviceAlreadyRegistered, ErrCodeLicenseAlreadyExists, ErrCodeJobInProgress:
		return 409

	// 429 Too Many Requests
	case ErrCodeDeviceBlocked, ErrCodeRateLimitExceeded:
		return 429

	// 500 Internal Server Error
	case ErrCodeEncryptionFailed, ErrCodeDecryptionFailed, ErrCodeKeyGenerationFailed,
		ErrCodeInternalError, ErrCodeStorageError, ErrCodeDatabaseError,
		ErrCodeCacheError, ErrCodeJobFailed:
		return 500

	// 503 Service Unavailable
	case ErrCodeServiceUnavailable:
		return 503

	default:
		return 500
	}
}

// String returns the string representation of the error code.
func (c ErrorCode) String() string {
	return string(c)
}

// OfflineError represents an offline-specific error with code and context.
type OfflineError struct {
	Code       ErrorCode
	Message    string
	DeviceID   *uuid.UUID
	MaterialID *uuid.UUID
	LicenseID  *uuid.UUID
	UserID     *uuid.UUID
	Err        error
}

// Error implements the error interface.
func (e *OfflineError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *OfflineError) Unwrap() error {
	return e.Err
}

// HTTPStatus returns the HTTP status code for this error.
func (e *OfflineError) HTTPStatus() int {
	return e.Code.HTTPStatus()
}

// WithDeviceID adds device ID context to the error.
func (e *OfflineError) WithDeviceID(id uuid.UUID) *OfflineError {
	e.DeviceID = &id
	return e
}

// WithMaterialID adds material ID context to the error.
func (e *OfflineError) WithMaterialID(id uuid.UUID) *OfflineError {
	e.MaterialID = &id
	return e
}

// WithLicenseID adds license ID context to the error.
func (e *OfflineError) WithLicenseID(id uuid.UUID) *OfflineError {
	e.LicenseID = &id
	return e
}

// WithUserID adds user ID context to the error.
func (e *OfflineError) WithUserID(id uuid.UUID) *OfflineError {
	e.UserID = &id
	return e
}

// WithError wraps an underlying error.
func (e *OfflineError) WithError(err error) *OfflineError {
	e.Err = err
	return e
}

// NewOfflineError creates a new OfflineError.
func NewOfflineError(code ErrorCode, message string) *OfflineError {
	return &OfflineError{
		Code:    code,
		Message: message,
	}
}

// WrapOfflineError wraps an existing error with an OfflineError.
func WrapOfflineError(code ErrorCode, message string, err error) *OfflineError {
	return &OfflineError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Predefined errors for common scenarios.
var (
	// Device errors
	ErrDeviceLimitExceeded = NewOfflineError(
		ErrCodeDeviceLimitExceeded,
		fmt.Sprintf("maximum device limit of %d reached", MaxDevicesPerUser),
	)
	ErrDeviceNotFound = NewOfflineError(
		ErrCodeDeviceNotFound,
		"device not found or has been revoked",
	)
	ErrDeviceFingerprintMismatch = NewOfflineError(
		ErrCodeDeviceFingerprintMismatch,
		"device fingerprint does not match registered fingerprint",
	)
	ErrDeviceBlocked = NewOfflineError(
		ErrCodeDeviceBlocked,
		"device is temporarily blocked due to too many failed attempts",
	)
	ErrDeviceAlreadyRegistered = NewOfflineError(
		ErrCodeDeviceAlreadyRegistered,
		"device with this fingerprint is already registered",
	)
	ErrInvalidFingerprint = NewOfflineError(
		ErrCodeInvalidFingerprint,
		fmt.Sprintf("fingerprint must be between %d and %d characters", MinFingerprintLength, MaxFingerprintLength),
	)
	ErrInvalidPlatform = NewOfflineError(
		ErrCodeInvalidPlatform,
		"platform must be one of: ios, android, desktop",
	)

	// License errors
	ErrLicenseNotFound = NewOfflineError(
		ErrCodeLicenseNotFound,
		"license not found",
	)
	ErrLicenseExpired = NewOfflineError(
		ErrCodeLicenseExpired,
		"license has expired",
	)
	ErrLicenseRevoked = NewOfflineError(
		ErrCodeLicenseRevoked,
		"license has been revoked",
	)
	ErrLicenseOfflineExpired = NewOfflineError(
		ErrCodeLicenseOfflineExpired,
		"offline grace period has expired, online validation required",
	)
	ErrInvalidNonce = NewOfflineError(
		ErrCodeInvalidNonce,
		"license nonce validation failed",
	)
	ErrLicenseAlreadyExists = NewOfflineError(
		ErrCodeLicenseAlreadyExists,
		"an active license already exists for this material and device",
	)

	// Material errors
	ErrMaterialAccessDenied = NewOfflineError(
		ErrCodeMaterialAccessDenied,
		"user does not have access to this material",
	)
	ErrMaterialNotFound = NewOfflineError(
		ErrCodeMaterialNotFound,
		"material not found",
	)
	ErrUnsupportedFileType = NewOfflineError(
		ErrCodeUnsupportedFileType,
		"file type is not supported for offline access",
	)

	// Encryption errors
	ErrEncryptionFailed = NewOfflineError(
		ErrCodeEncryptionFailed,
		"failed to encrypt material",
	)
	ErrDecryptionFailed = NewOfflineError(
		ErrCodeDecryptionFailed,
		"failed to decrypt content",
	)
	ErrKeyGenerationFailed = NewOfflineError(
		ErrCodeKeyGenerationFailed,
		"failed to generate encryption key",
	)
	ErrKeyNotFound = NewOfflineError(
		ErrCodeKeyNotFound,
		"encryption key not found",
	)
	ErrInvalidKey = NewOfflineError(
		ErrCodeInvalidKey,
		"invalid encryption key",
	)

	// Rate limiting errors
	ErrRateLimitExceeded = NewOfflineError(
		ErrCodeRateLimitExceeded,
		fmt.Sprintf("download rate limit of %d per hour exceeded", MaxDownloadsPerHour),
	)

	// Job errors
	ErrJobNotFound = NewOfflineError(
		ErrCodeJobNotFound,
		"encryption job not found",
	)
	ErrJobFailed = NewOfflineError(
		ErrCodeJobFailed,
		"encryption job failed",
	)
	ErrJobInProgress = NewOfflineError(
		ErrCodeJobInProgress,
		"encryption job is already in progress",
	)

	// Security errors
	ErrReplayAttack = NewOfflineError(
		ErrCodeReplayAttack,
		"request replay detected",
	)
	ErrInvalidSignature = NewOfflineError(
		ErrCodeInvalidSignature,
		"request signature validation failed",
	)
	ErrInvalidRequest = NewOfflineError(
		ErrCodeInvalidRequest,
		"invalid request format",
	)

	// Internal errors
	ErrInternalError = NewOfflineError(
		ErrCodeInternalError,
		"an internal error occurred",
	)
	ErrStorageError = NewOfflineError(
		ErrCodeStorageError,
		"storage operation failed",
	)
	ErrDatabaseError = NewOfflineError(
		ErrCodeDatabaseError,
		"database operation failed",
	)
	ErrCacheError = NewOfflineError(
		ErrCodeCacheError,
		"cache operation failed",
	)
	ErrServiceUnavailable = NewOfflineError(
		ErrCodeServiceUnavailable,
		"service is temporarily unavailable",
	)
)

// IsOfflineError checks if an error is an OfflineError.
func IsOfflineError(err error) bool {
	_, ok := err.(*OfflineError)
	return ok
}

// GetOfflineError extracts an OfflineError from an error chain.
func GetOfflineError(err error) (*OfflineError, bool) {
	var offlineErr *OfflineError
	if err == nil {
		return nil, false
	}
	for err != nil {
		if oe, ok := err.(*OfflineError); ok {
			offlineErr = oe
			return offlineErr, true
		}
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}
	return nil, false
}
