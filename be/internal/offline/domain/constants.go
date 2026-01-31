// Package domain contains the core business entities and constants
// for the Offline Material Service.
package domain

import "time"

// Device limits and constraints.
const (
	// MaxDevicesPerUser is the maximum number of devices a user can register.
	MaxDevicesPerUser = 5

	// MinFingerprintLength is the minimum length for a device fingerprint.
	MinFingerprintLength = 32

	// MaxFingerprintLength is the maximum length for a device fingerprint.
	MaxFingerprintLength = 512

	// MaxDeviceNameLength is the maximum length for a device name.
	MaxDeviceNameLength = 255
)

// License configuration constants.
const (
	// DefaultLicenseExpirationDays is the default license expiration in days.
	DefaultLicenseExpirationDays = 30

	// DefaultOfflineGracePeriodHours is the default offline grace period in hours.
	DefaultOfflineGracePeriodHours = 72

	// NonceLength is the length of the license nonce in bytes.
	NonceLength = 32

	// NonceHexLength is the length of the hex-encoded nonce string.
	NonceHexLength = 64
)

// Encryption configuration constants.
const (
	// CEKSize is the size of the Content Encryption Key in bytes (256 bits).
	CEKSize = 32

	// IVSize is the size of the Initialization Vector in bytes (96 bits for GCM).
	IVSize = 12

	// AuthTagSize is the size of the authentication tag in bytes (128 bits).
	AuthTagSize = 16

	// KEKSize is the size of the Key Encryption Key in bytes (256 bits).
	KEKSize = 32

	// SaltSize is the size of the salt for HKDF in bytes.
	SaltSize = 32

	// HKDFInfoPrefix is the prefix for HKDF info parameter.
	HKDFInfoPrefix = "ngasihtau-cek-v1"
)

// Rate limiting constants.
const (
	// MaxDownloadsPerHour is the maximum downloads per user per hour.
	MaxDownloadsPerHour = 10

	// MaxValidationFailuresPerHour is the max validation failures before blocking.
	MaxValidationFailuresPerHour = 5

	// RateLimitWindow is the time window for rate limiting.
	RateLimitWindow = 1 * time.Hour

	// ReplayProtectionWindow is the time window for replay attack prevention.
	ReplayProtectionWindow = 5 * time.Minute
)

// Cache TTL constants.
const (
	// LicenseCacheTTL is the TTL for cached licenses.
	LicenseCacheTTL = 5 * time.Minute

	// DeviceCacheTTL is the TTL for cached devices.
	DeviceCacheTTL = 10 * time.Minute

	// CEKCacheTTL is the TTL for cached CEKs (short for security).
	CEKCacheTTL = 1 * time.Minute

	// RateLimitCacheTTL is the TTL for rate limit counters.
	RateLimitCacheTTL = 1 * time.Hour
)

// Job processing constants.
const (
	// DefaultJobTimeout is the default timeout for encryption jobs.
	DefaultJobTimeout = 30 * time.Minute

	// DefaultJobPollInterval is the default interval for polling job queue.
	DefaultJobPollInterval = 1 * time.Second

	// DefaultWorkerConcurrency is the default number of concurrent workers.
	DefaultWorkerConcurrency = 2

	// DefaultShutdownTimeout is the default timeout for graceful shutdown.
	DefaultShutdownTimeout = 60 * time.Second
)

// Supported file types for offline encryption.
var SupportedFileTypes = []string{"pdf", "docx", "pptx"}

// IsSupportedFileType checks if a file type is supported for offline encryption.
func IsSupportedFileType(fileType string) bool {
	for _, ft := range SupportedFileTypes {
		if ft == fileType {
			return true
		}
	}
	return false
}

// ValidPlatforms contains all valid device platforms.
var ValidPlatforms = []Platform{PlatformIOS, PlatformAndroid, PlatformDesktop}

// IsValidPlatform checks if a platform is valid.
func IsValidPlatform(platform Platform) bool {
	for _, p := range ValidPlatforms {
		if p == platform {
			return true
		}
	}
	return false
}

// NATS subjects for offline events.
const (
	// NATSSubjectKeyGenerated is the subject for key generation events.
	NATSSubjectKeyGenerated = "offline.key.generated"

	// NATSSubjectKeyRetrieved is the subject for key retrieval events.
	NATSSubjectKeyRetrieved = "offline.key.retrieved"

	// NATSSubjectLicenseIssued is the subject for license issuance events.
	NATSSubjectLicenseIssued = "offline.license.issued"

	// NATSSubjectLicenseValidated is the subject for license validation events.
	NATSSubjectLicenseValidated = "offline.license.validated"

	// NATSSubjectLicenseRevoked is the subject for license revocation events.
	NATSSubjectLicenseRevoked = "offline.license.revoked"

	// NATSSubjectLicenseRenewed is the subject for license renewal events.
	NATSSubjectLicenseRenewed = "offline.license.renewed"

	// NATSSubjectDeviceRegistered is the subject for device registration events.
	NATSSubjectDeviceRegistered = "offline.device.registered"

	// NATSSubjectDeviceDeregistered is the subject for device deregistration events.
	NATSSubjectDeviceDeregistered = "offline.device.deregistered"

	// NATSSubjectEncryptionRequested is the subject for encryption job requests.
	NATSSubjectEncryptionRequested = "offline.encryption.requested"

	// NATSSubjectEncryptionCompleted is the subject for encryption completion events.
	NATSSubjectEncryptionCompleted = "offline.encryption.completed"

	// NATSSubjectEncryptionFailed is the subject for encryption failure events.
	NATSSubjectEncryptionFailed = "offline.encryption.failed"
)

// JetStream stream and consumer names.
const (
	// JetStreamEncryptionStream is the name of the encryption job stream.
	JetStreamEncryptionStream = "OFFLINE_ENCRYPTION"

	// JetStreamEncryptionConsumer is the name of the encryption job consumer.
	JetStreamEncryptionConsumer = "encryption-worker"
)

// Redis key prefixes for caching.
const (
	// RedisKeyPrefixLicense is the prefix for license cache keys.
	RedisKeyPrefixLicense = "offline:license:"

	// RedisKeyPrefixDevice is the prefix for device cache keys.
	RedisKeyPrefixDevice = "offline:device:"

	// RedisKeyPrefixUserDevices is the prefix for user device list cache keys.
	RedisKeyPrefixUserDevices = "offline:devices:user:"

	// RedisKeyPrefixRateLimit is the prefix for rate limit cache keys.
	RedisKeyPrefixRateLimit = "offline:ratelimit:"

	// RedisKeyPrefixDownloadCount is the prefix for download count keys.
	RedisKeyPrefixDownloadCount = "offline:ratelimit:download:"

	// RedisKeyPrefixValidationFailure is the prefix for validation failure keys.
	RedisKeyPrefixValidationFailure = "offline:ratelimit:validation:"

	// RedisKeyPrefixCEK is the prefix for CEK cache keys.
	RedisKeyPrefixCEK = "offline:cek:"
)
