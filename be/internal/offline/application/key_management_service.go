// Package application contains the business logic and use cases for the Offline Material Service.
package application

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/hkdf"

	"ngasihtau/internal/offline/domain"
)

// OfflineEventPublisher defines the interface for publishing offline-related events.
type OfflineEventPublisher interface {
	// PublishKeyGenerated publishes a key generation event.
	PublishKeyGenerated(ctx context.Context, event KeyEvent) error

	// PublishKeyRetrieved publishes a key retrieval event.
	PublishKeyRetrieved(ctx context.Context, event KeyEvent) error

	// PublishLicenseIssued publishes a license issuance event.
	PublishLicenseIssued(ctx context.Context, event LicenseEvent) error

	// PublishLicenseValidated publishes a license validation event.
	PublishLicenseValidated(ctx context.Context, event LicenseEvent) error

	// PublishLicenseRevoked publishes a license revocation event.
	PublishLicenseRevoked(ctx context.Context, event LicenseEvent) error

	// PublishLicenseRenewed publishes a license renewal event.
	PublishLicenseRenewed(ctx context.Context, event LicenseEvent) error

	// PublishDeviceRegistered publishes a device registration event.
	PublishDeviceRegistered(ctx context.Context, event DeviceEvent) error

	// PublishDeviceDeregistered publishes a device deregistration event.
	PublishDeviceDeregistered(ctx context.Context, event DeviceEvent) error

	// PublishEncryptionRequested publishes an encryption job request event.
	PublishEncryptionRequested(ctx context.Context, event EncryptionJobEvent) error

	// PublishEncryptionCompleted publishes an encryption completion event.
	PublishEncryptionCompleted(ctx context.Context, event EncryptionJobEvent) error

	// PublishEncryptionFailed publishes an encryption failure event.
	PublishEncryptionFailed(ctx context.Context, event EncryptionJobEvent) error

	// PublishMaterialDownloaded publishes a material download event.
	PublishMaterialDownloaded(ctx context.Context, event MaterialDownloadEvent) error
}

// KeyManagementService handles CEK generation, storage, and management.
// Implements Requirement 1: Key Management Service.
type KeyManagementService struct {
	cekRepo        domain.CEKRepository
	auditRepo      domain.AuditLogRepository
	eventPublisher OfflineEventPublisher
	masterSecret   []byte
	kek            []byte // Key Encryption Key for encrypting CEKs at rest
	currentVersion int
}

// KeyManagementConfig holds configuration for the Key Management Service.
type KeyManagementConfig struct {
	MasterSecret   []byte
	KEK            []byte
	CurrentVersion int
}

// NewKeyManagementService creates a new Key Management Service.
func NewKeyManagementService(
	cekRepo domain.CEKRepository,
	auditRepo domain.AuditLogRepository,
	eventPublisher OfflineEventPublisher,
	config KeyManagementConfig,
) *KeyManagementService {
	return &KeyManagementService{
		cekRepo:        cekRepo,
		auditRepo:      auditRepo,
		eventPublisher: eventPublisher,
		masterSecret:   config.MasterSecret,
		kek:            config.KEK,
		currentVersion: config.CurrentVersion,
	}
}

// GetOrCreateCEK returns an existing CEK or generates a new one.
// Implements Requirement 1.1, 1.3: CEK generation and retrieval.
func (s *KeyManagementService) GetOrCreateCEK(ctx context.Context, userID, materialID, deviceID uuid.UUID) (*domain.ContentEncryptionKey, error) {
	// Try to find existing CEK
	existing, err := s.cekRepo.FindByComposite(ctx, userID, materialID, deviceID)
	if err == nil && existing != nil {
		// Log retrieval
		s.logAuditEvent(ctx, userID, &deviceID, domain.AuditActionKeyRetrieve, domain.AuditResourceCEK, existing.ID, true, nil)

		// Publish event
		s.publishKeyEvent(ctx, domain.NATSSubjectKeyRetrieved, existing.ID, userID, materialID, deviceID)

		return existing, nil
	}

	// Generate new CEK
	cek, err := s.generateCEK(userID, materialID, deviceID)
	if err != nil {
		errCode := domain.ErrCodeKeyGenerationFailed.String()
		s.logAuditEvent(ctx, userID, &deviceID, domain.AuditActionKeyGenerate, domain.AuditResourceCEK, uuid.Nil, false, &errCode)
		return nil, domain.WrapOfflineError(domain.ErrCodeKeyGenerationFailed, "failed to generate CEK", err)
	}

	// Encrypt CEK for storage
	encryptedKey, err := s.encryptCEKForStorage(cek)
	if err != nil {
		errCode := domain.ErrCodeEncryptionFailed.String()
		s.logAuditEvent(ctx, userID, &deviceID, domain.AuditActionKeyGenerate, domain.AuditResourceCEK, uuid.Nil, false, &errCode)
		return nil, domain.WrapOfflineError(domain.ErrCodeEncryptionFailed, "failed to encrypt CEK for storage", err)
	}

	// Create CEK record
	cekRecord := domain.NewContentEncryptionKey(userID, materialID, deviceID, encryptedKey, s.currentVersion)

	if err := s.cekRepo.Create(ctx, cekRecord); err != nil {
		errCode := domain.ErrCodeDatabaseError.String()
		s.logAuditEvent(ctx, userID, &deviceID, domain.AuditActionKeyGenerate, domain.AuditResourceCEK, uuid.Nil, false, &errCode)
		return nil, domain.WrapOfflineError(domain.ErrCodeDatabaseError, "failed to store CEK", err)
	}

	// Log successful generation
	s.logAuditEvent(ctx, userID, &deviceID, domain.AuditActionKeyGenerate, domain.AuditResourceCEK, cekRecord.ID, true, nil)

	// Publish event
	s.publishKeyEvent(ctx, domain.NATSSubjectKeyGenerated, cekRecord.ID, userID, materialID, deviceID)

	log.Info().
		Str("cek_id", cekRecord.ID.String()).
		Str("user_id", userID.String()).
		Str("material_id", materialID.String()).
		Str("device_id", deviceID.String()).
		Int("key_version", s.currentVersion).
		Msg("CEK generated successfully")

	return cekRecord, nil
}

// DecryptCEK decrypts a stored CEK for use.
func (s *KeyManagementService) DecryptCEK(ctx context.Context, cekRecord *domain.ContentEncryptionKey) ([]byte, error) {
	return s.decryptCEKFromStorage(cekRecord.EncryptedKey)
}

// EncryptCEKForTransport encrypts a CEK with a device-specific transport key.
// Implements Requirement 8.1: CEK transport encryption.
func (s *KeyManagementService) EncryptCEKForTransport(ctx context.Context, cek []byte, deviceID uuid.UUID) ([]byte, error) {
	// Derive device-specific transport key using HKDF
	transportKey, err := s.deriveTransportKey(deviceID)
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeKeyGenerationFailed, "failed to derive transport key", err)
	}

	// Encrypt CEK with transport key using AES-256-GCM
	block, err := aes.NewCipher(transportKey)
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeEncryptionFailed, "failed to create cipher", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeEncryptionFailed, "failed to create GCM", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeEncryptionFailed, "failed to generate nonce", err)
	}

	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, cek, nil)

	return ciphertext, nil
}

// RotateKeys initiates key rotation from one version to another.
// Implements Requirement 1.4: Key rotation support.
func (s *KeyManagementService) RotateKeys(ctx context.Context, fromVersion, toVersion int, newKEK []byte) error {
	// Find all CEKs with the old version
	ceks, err := s.cekRepo.FindByKeyVersion(ctx, fromVersion)
	if err != nil {
		return domain.WrapOfflineError(domain.ErrCodeDatabaseError, "failed to find CEKs for rotation", err)
	}

	log.Info().
		Int("from_version", fromVersion).
		Int("to_version", toVersion).
		Int("cek_count", len(ceks)).
		Msg("starting key rotation")

	// Re-encrypt each CEK with the new KEK
	for _, cek := range ceks {
		// Decrypt with old KEK
		plainCEK, err := s.decryptCEKFromStorage(cek.EncryptedKey)
		if err != nil {
			log.Error().Err(err).Str("cek_id", cek.ID.String()).Msg("failed to decrypt CEK during rotation")
			continue
		}

		// Encrypt with new KEK
		newEncrypted, err := s.encryptCEKWithKEK(plainCEK, newKEK)
		if err != nil {
			log.Error().Err(err).Str("cek_id", cek.ID.String()).Msg("failed to re-encrypt CEK during rotation")
			continue
		}

		// Update in database
		if err := s.cekRepo.UpdateKeyVersion(ctx, cek.ID, newEncrypted, toVersion); err != nil {
			log.Error().Err(err).Str("cek_id", cek.ID.String()).Msg("failed to update CEK version")
			continue
		}

		log.Debug().Str("cek_id", cek.ID.String()).Msg("CEK rotated successfully")
	}

	log.Info().
		Int("from_version", fromVersion).
		Int("to_version", toVersion).
		Msg("key rotation completed")

	return nil
}

// generateCEK generates a new CEK using HKDF.
// Implements Requirement 1.1: HKDF-based CEK generation.
func (s *KeyManagementService) generateCEK(userID, materialID, deviceID uuid.UUID) ([]byte, error) {
	// Build info string: prefix || user_id || material_id || device_id
	info := buildHKDFInfo(userID, materialID, deviceID)

	// Use HKDF-SHA256 to derive the CEK
	hkdfReader := hkdf.New(sha256.New, s.masterSecret, nil, info)

	cek := make([]byte, domain.CEKSize)
	if _, err := io.ReadFull(hkdfReader, cek); err != nil {
		return nil, fmt.Errorf("HKDF expansion failed: %w", err)
	}

	return cek, nil
}

// buildHKDFInfo builds the info parameter for HKDF.
func buildHKDFInfo(userID, materialID, deviceID uuid.UUID) []byte {
	info := make([]byte, 0, len(domain.HKDFInfoPrefix)+48) // prefix + 3 UUIDs (16 bytes each)
	info = append(info, []byte(domain.HKDFInfoPrefix)...)
	info = append(info, userID[:]...)
	info = append(info, materialID[:]...)
	info = append(info, deviceID[:]...)
	return info
}

// encryptCEKForStorage encrypts a CEK for storage using the KEK.
// Implements Requirement 1.2: CEK encryption at rest.
func (s *KeyManagementService) encryptCEKForStorage(cek []byte) ([]byte, error) {
	return s.encryptCEKWithKEK(cek, s.kek)
}

// encryptCEKWithKEK encrypts a CEK with a specific KEK.
func (s *KeyManagementService) encryptCEKWithKEK(cek, kek []byte) ([]byte, error) {
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, cek, nil)

	return ciphertext, nil
}

// decryptCEKFromStorage decrypts a stored CEK using the KEK.
func (s *KeyManagementService) decryptCEKFromStorage(encryptedCEK []byte) ([]byte, error) {
	return s.decryptCEKWithKEK(encryptedCEK, s.kek)
}

// decryptCEKWithKEK decrypts a CEK with a specific KEK.
func (s *KeyManagementService) decryptCEKWithKEK(encryptedCEK, kek []byte) ([]byte, error) {
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedCEK) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encryptedCEK[:nonceSize], encryptedCEK[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// deriveTransportKey derives a device-specific transport key.
func (s *KeyManagementService) deriveTransportKey(deviceID uuid.UUID) ([]byte, error) {
	info := make([]byte, 0, 32)
	info = append(info, []byte("ngasihtau-transport-v1")...)
	info = append(info, deviceID[:]...)

	hkdfReader := hkdf.New(sha256.New, s.masterSecret, nil, info)

	key := make([]byte, domain.KEKSize)
	if _, err := io.ReadFull(hkdfReader, key); err != nil {
		return nil, fmt.Errorf("HKDF expansion failed: %w", err)
	}

	return key, nil
}

// logAuditEvent logs an audit event.
func (s *KeyManagementService) logAuditEvent(ctx context.Context, userID uuid.UUID, deviceID *uuid.UUID, action, resource string, resourceID uuid.UUID, success bool, errorCode *string) {
	if s.auditRepo == nil {
		return
	}

	auditLog := domain.NewAuditLog(userID, deviceID, action, resource, resourceID, "", "", success, errorCode)
	if err := s.auditRepo.Create(ctx, auditLog); err != nil {
		log.Error().Err(err).Msg("failed to create audit log")
	}
}

// publishKeyEvent publishes a key-related event.
func (s *KeyManagementService) publishKeyEvent(ctx context.Context, subject string, cekID, userID, materialID, deviceID uuid.UUID) {
	if s.eventPublisher == nil {
		return
	}

	event := KeyEvent{
		CEKID:      cekID,
		UserID:     userID,
		MaterialID: materialID,
		DeviceID:   deviceID,
	}

	var err error
	switch subject {
	case domain.NATSSubjectKeyGenerated:
		err = s.eventPublisher.PublishKeyGenerated(ctx, event)
	case domain.NATSSubjectKeyRetrieved:
		err = s.eventPublisher.PublishKeyRetrieved(ctx, event)
	}

	if err != nil {
		log.Error().Err(err).Str("subject", subject).Msg("failed to publish key event")
	}
}

// KeyEvent represents a key-related event for NATS publishing.
type KeyEvent struct {
	CEKID      uuid.UUID `json:"cek_id"`
	UserID     uuid.UUID `json:"user_id"`
	MaterialID uuid.UUID `json:"material_id"`
	DeviceID   uuid.UUID `json:"device_id"`
}

// LicenseEvent represents a license-related event for NATS publishing.
type LicenseEvent struct {
	LicenseID  uuid.UUID `json:"license_id"`
	UserID     uuid.UUID `json:"user_id"`
	MaterialID uuid.UUID `json:"material_id"`
	DeviceID   uuid.UUID `json:"device_id"`
}

// DeviceEvent represents a device-related event for NATS publishing.
type DeviceEvent struct {
	DeviceID uuid.UUID `json:"device_id"`
	UserID   uuid.UUID `json:"user_id"`
	Platform string    `json:"platform"`
}

// EncryptionJobEvent represents an encryption job event for NATS publishing.
type EncryptionJobEvent struct {
	JobID      uuid.UUID `json:"job_id"`
	MaterialID uuid.UUID `json:"material_id"`
	UserID     uuid.UUID `json:"user_id"`
	DeviceID   uuid.UUID `json:"device_id"`
	LicenseID  uuid.UUID `json:"license_id"`
	Error      string    `json:"error,omitempty"`
}

// DeriveIVForChunk derives a unique IV for a specific chunk index.
// Implements Requirement 2.3: Unique IV per chunk.
func DeriveIVForChunk(baseSeed []byte, chunkIndex int) []byte {
	iv := make([]byte, domain.IVSize)

	// Use first 8 bytes from seed
	copy(iv, baseSeed[:8])

	// Use last 4 bytes for chunk index (big-endian)
	binary.BigEndian.PutUint32(iv[8:], uint32(chunkIndex))

	return iv
}

// GenerateNonce generates a random nonce for license creation.
// Implements Requirement 8.4: License nonce generation.
func GenerateNonce() (string, error) {
	nonce := make([]byte, domain.NonceLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return fmt.Sprintf("%x", nonce), nil
}
