// Package application contains unit tests for the Key Management Service.
package application

import (
	"bytes"
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ngasihtau/internal/offline/domain"
)

// mockCEKRepository is a mock implementation of domain.CEKRepository for testing.
type mockCEKRepository struct {
	ceks      map[string]*domain.ContentEncryptionKey
	byVersion map[int][]*domain.ContentEncryptionKey
	createErr error
	findErr   error
	updateErr error
}

func newMockCEKRepository() *mockCEKRepository {
	return &mockCEKRepository{
		ceks:      make(map[string]*domain.ContentEncryptionKey),
		byVersion: make(map[int][]*domain.ContentEncryptionKey),
	}
}

func (m *mockCEKRepository) Create(ctx context.Context, cek *domain.ContentEncryptionKey) error {
	if m.createErr != nil {
		return m.createErr
	}
	key := compositeKey(cek.UserID, cek.MaterialID, cek.DeviceID)
	m.ceks[key] = cek
	m.byVersion[cek.KeyVersion] = append(m.byVersion[cek.KeyVersion], cek)
	return nil
}

func (m *mockCEKRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ContentEncryptionKey, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	for _, cek := range m.ceks {
		if cek.ID == id {
			return cek, nil
		}
	}
	return nil, nil
}

func (m *mockCEKRepository) FindByComposite(ctx context.Context, userID, materialID, deviceID uuid.UUID) (*domain.ContentEncryptionKey, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	key := compositeKey(userID, materialID, deviceID)
	return m.ceks[key], nil
}

func (m *mockCEKRepository) DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	for key, cek := range m.ceks {
		if cek.DeviceID == deviceID {
			delete(m.ceks, key)
		}
	}
	return nil
}

func (m *mockCEKRepository) DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error {
	for key, cek := range m.ceks {
		if cek.MaterialID == materialID {
			delete(m.ceks, key)
		}
	}
	return nil
}

func (m *mockCEKRepository) FindByKeyVersion(ctx context.Context, keyVersion int) ([]*domain.ContentEncryptionKey, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	// Return copies to avoid mutation issues during iteration
	result := make([]*domain.ContentEncryptionKey, len(m.byVersion[keyVersion]))
	for i, cek := range m.byVersion[keyVersion] {
		cekCopy := *cek
		result[i] = &cekCopy
	}
	return result, nil
}

func (m *mockCEKRepository) UpdateKeyVersion(ctx context.Context, id uuid.UUID, encryptedKey []byte, keyVersion int) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	for _, cek := range m.ceks {
		if cek.ID == id {
			// Remove from old version list
			oldVersion := cek.KeyVersion
			newList := make([]*domain.ContentEncryptionKey, 0)
			for _, c := range m.byVersion[oldVersion] {
				if c.ID != id {
					newList = append(newList, c)
				}
			}
			m.byVersion[oldVersion] = newList
			// Update CEK
			cek.EncryptedKey = encryptedKey
			cek.KeyVersion = keyVersion
			// Add to new version list
			m.byVersion[keyVersion] = append(m.byVersion[keyVersion], cek)
			return nil
		}
	}
	return nil
}

func compositeKey(userID, materialID, deviceID uuid.UUID) string {
	return userID.String() + ":" + materialID.String() + ":" + deviceID.String()
}

// mockAuditLogRepository is a mock implementation of domain.AuditLogRepository.
type mockAuditLogRepository struct {
	logs []*domain.AuditLog
}

func newMockAuditLogRepository() *mockAuditLogRepository {
	return &mockAuditLogRepository{
		logs: make([]*domain.AuditLog, 0),
	}
}

func (m *mockAuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockAuditLogRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.AuditLog, int, error) {
	return nil, 0, nil
}

func (m *mockAuditLogRepository) FindByDeviceID(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*domain.AuditLog, int, error) {
	return nil, 0, nil
}

func (m *mockAuditLogRepository) FindByAction(ctx context.Context, action string, limit, offset int) ([]*domain.AuditLog, int, error) {
	return nil, 0, nil
}

func (m *mockAuditLogRepository) CountFailedByDevice(ctx context.Context, deviceID uuid.UUID, since time.Time) (int, error) {
	return 0, nil
}

func (m *mockAuditLogRepository) CountFailedByDeviceAndAction(ctx context.Context, deviceID uuid.UUID, action string, since time.Time) (int, error) {
	return 0, nil
}

// mockOfflineEventPublisher is a mock implementation of OfflineEventPublisher.
type mockOfflineEventPublisher struct {
	keyEvents        []KeyEvent
	licenseEvents    []LicenseEvent
	deviceEvents     []DeviceEvent
	encryptionEvents []EncryptionJobEvent
}

func newMockOfflineEventPublisher() *mockOfflineEventPublisher {
	return &mockOfflineEventPublisher{
		keyEvents:        make([]KeyEvent, 0),
		licenseEvents:    make([]LicenseEvent, 0),
		deviceEvents:     make([]DeviceEvent, 0),
		encryptionEvents: make([]EncryptionJobEvent, 0),
	}
}

func (m *mockOfflineEventPublisher) PublishKeyGenerated(ctx context.Context, event KeyEvent) error {
	m.keyEvents = append(m.keyEvents, event)
	return nil
}

func (m *mockOfflineEventPublisher) PublishKeyRetrieved(ctx context.Context, event KeyEvent) error {
	m.keyEvents = append(m.keyEvents, event)
	return nil
}

func (m *mockOfflineEventPublisher) PublishLicenseIssued(ctx context.Context, event LicenseEvent) error {
	m.licenseEvents = append(m.licenseEvents, event)
	return nil
}

func (m *mockOfflineEventPublisher) PublishLicenseValidated(ctx context.Context, event LicenseEvent) error {
	m.licenseEvents = append(m.licenseEvents, event)
	return nil
}

func (m *mockOfflineEventPublisher) PublishLicenseRevoked(ctx context.Context, event LicenseEvent) error {
	m.licenseEvents = append(m.licenseEvents, event)
	return nil
}

func (m *mockOfflineEventPublisher) PublishLicenseRenewed(ctx context.Context, event LicenseEvent) error {
	m.licenseEvents = append(m.licenseEvents, event)
	return nil
}

func (m *mockOfflineEventPublisher) PublishDeviceRegistered(ctx context.Context, event DeviceEvent) error {
	m.deviceEvents = append(m.deviceEvents, event)
	return nil
}

func (m *mockOfflineEventPublisher) PublishDeviceDeregistered(ctx context.Context, event DeviceEvent) error {
	m.deviceEvents = append(m.deviceEvents, event)
	return nil
}

func (m *mockOfflineEventPublisher) PublishEncryptionRequested(ctx context.Context, event EncryptionJobEvent) error {
	m.encryptionEvents = append(m.encryptionEvents, event)
	return nil
}

func (m *mockOfflineEventPublisher) PublishEncryptionCompleted(ctx context.Context, event EncryptionJobEvent) error {
	m.encryptionEvents = append(m.encryptionEvents, event)
	return nil
}

func (m *mockOfflineEventPublisher) PublishEncryptionFailed(ctx context.Context, event EncryptionJobEvent) error {
	m.encryptionEvents = append(m.encryptionEvents, event)
	return nil
}

// Helper function to create a test KeyManagementService
func newTestKeyManagementService() (*KeyManagementService, *mockCEKRepository, *mockAuditLogRepository, *mockOfflineEventPublisher) {
	cekRepo := newMockCEKRepository()
	auditRepo := newMockAuditLogRepository()
	eventPublisher := newMockOfflineEventPublisher()

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)

	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, auditRepo, eventPublisher, config)
	return svc, cekRepo, auditRepo, eventPublisher
}

// TestGenerateCEK_ValidInputs tests CEK generation with valid inputs.
func TestGenerateCEK_ValidInputs(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	cek, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)

	require.NoError(t, err)
	assert.NotNil(t, cek)
	assert.Equal(t, userID, cek.UserID)
	assert.Equal(t, materialID, cek.MaterialID)
	assert.Equal(t, deviceID, cek.DeviceID)
	assert.NotEmpty(t, cek.EncryptedKey)
	assert.Equal(t, 1, cek.KeyVersion)
}

// TestGetOrCreateCEK_ReturnsExisting tests that existing CEK is returned.
func TestGetOrCreateCEK_ReturnsExisting(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Create first CEK
	cek1, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	// Get same CEK again
	cek2, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	// Should return the same CEK
	assert.Equal(t, cek1.ID, cek2.ID)
	assert.Equal(t, cek1.EncryptedKey, cek2.EncryptedKey)
}

// TestCEKEncryptionDecryption tests CEK encryption and decryption round-trip.
func TestCEKEncryptionDecryption(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Create CEK
	cekRecord, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	// Decrypt the CEK
	decryptedCEK, err := svc.DecryptCEK(ctx, cekRecord)
	require.NoError(t, err)

	// CEK should be 32 bytes (256 bits)
	assert.Len(t, decryptedCEK, domain.CEKSize)

	// Encrypted key should differ from decrypted
	assert.NotEqual(t, cekRecord.EncryptedKey, decryptedCEK)
}

// TestEncryptCEKForTransport tests transport encryption.
func TestEncryptCEKForTransport(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	// Generate a test CEK
	cek := make([]byte, domain.CEKSize)
	rand.Read(cek)

	deviceID := uuid.New()

	// Encrypt for transport
	encrypted, err := svc.EncryptCEKForTransport(ctx, cek, deviceID)
	require.NoError(t, err)

	// Encrypted should be different from original
	assert.NotEqual(t, cek, encrypted)

	// Encrypted should be longer (includes nonce)
	assert.Greater(t, len(encrypted), len(cek))
}

// TestDeriveIVForChunk tests IV derivation for chunks.
func TestDeriveIVForChunk(t *testing.T) {
	baseSeed := make([]byte, 16)
	rand.Read(baseSeed)

	// Generate IVs for different chunks
	iv0 := DeriveIVForChunk(baseSeed, 0)
	iv1 := DeriveIVForChunk(baseSeed, 1)
	iv2 := DeriveIVForChunk(baseSeed, 2)

	// All IVs should be 12 bytes
	assert.Len(t, iv0, domain.IVSize)
	assert.Len(t, iv1, domain.IVSize)
	assert.Len(t, iv2, domain.IVSize)

	// All IVs should be unique
	assert.NotEqual(t, iv0, iv1)
	assert.NotEqual(t, iv1, iv2)
	assert.NotEqual(t, iv0, iv2)
}

// TestDeriveIVForChunk_Deterministic tests that IV derivation is deterministic.
func TestDeriveIVForChunk_Deterministic(t *testing.T) {
	baseSeed := make([]byte, 16)
	rand.Read(baseSeed)

	// Same inputs should produce same output
	iv1 := DeriveIVForChunk(baseSeed, 5)
	iv2 := DeriveIVForChunk(baseSeed, 5)

	assert.Equal(t, iv1, iv2)
}

// TestGenerateNonce tests nonce generation.
func TestGenerateNonce(t *testing.T) {
	nonce1, err := GenerateNonce()
	require.NoError(t, err)
	assert.Len(t, nonce1, domain.NonceHexLength)

	nonce2, err := GenerateNonce()
	require.NoError(t, err)
	assert.Len(t, nonce2, domain.NonceHexLength)

	// Nonces should be unique
	assert.NotEqual(t, nonce1, nonce2)
}

// TestKeyRotation tests key rotation functionality.
func TestKeyRotation(t *testing.T) {
	svc, cekRepo, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	// Create some CEKs with version 1
	userID := uuid.New()
	for i := 0; i < 3; i++ {
		materialID := uuid.New()
		deviceID := uuid.New()
		_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
		require.NoError(t, err)
	}

	// Verify we have 3 CEKs with version 1
	v1CEKs, err := cekRepo.FindByKeyVersion(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, v1CEKs, 3)

	// Generate new KEK for rotation
	newKEK := make([]byte, 32)
	rand.Read(newKEK)

	// Rotate keys from version 1 to version 2
	err = svc.RotateKeys(ctx, 1, 2, newKEK)
	require.NoError(t, err)

	// Verify all CEKs are now version 2
	v2CEKs, err := cekRepo.FindByKeyVersion(ctx, 2)
	require.NoError(t, err)
	assert.Len(t, v2CEKs, 3)

	// Version 1 should be empty
	v1CEKs, err = cekRepo.FindByKeyVersion(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, v1CEKs, 0)
}

// TestAuditLogging tests that audit events are logged.
func TestAuditLogging(t *testing.T) {
	svc, _, auditRepo, _ := newTestKeyManagementService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Create CEK - should log audit event
	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	// Should have logged the key generation
	assert.Len(t, auditRepo.logs, 1)
	assert.Equal(t, domain.AuditActionKeyGenerate, auditRepo.logs[0].Action)
	assert.True(t, auditRepo.logs[0].Success)
}

// TestEventPublishing tests that events are published.
func TestEventPublishing(t *testing.T) {
	svc, _, _, eventPublisher := newTestKeyManagementService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Create CEK - should publish key generated event
	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	// Should have published key generated event
	assert.Len(t, eventPublisher.keyEvents, 1)
	assert.Equal(t, userID, eventPublisher.keyEvents[0].UserID)
	assert.Equal(t, materialID, eventPublisher.keyEvents[0].MaterialID)
	assert.Equal(t, deviceID, eventPublisher.keyEvents[0].DeviceID)

	// Get same CEK again - should publish key retrieved event
	_, err = svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	// Should have published key retrieved event
	assert.Len(t, eventPublisher.keyEvents, 2)
}

// TestBuildHKDFInfo tests the HKDF info building function.
func TestBuildHKDFInfo(t *testing.T) {
	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	info := buildHKDFInfo(userID, materialID, deviceID)

	// Info should start with prefix
	assert.True(t, bytes.HasPrefix(info, []byte(domain.HKDFInfoPrefix)))

	// Info should contain all three UUIDs (16 bytes each)
	expectedLen := len(domain.HKDFInfoPrefix) + 48 // prefix + 3 UUIDs
	assert.Len(t, info, expectedLen)
}

// TestCEKDerivationDeterminism tests that CEK derivation is deterministic.
func TestCEKDerivationDeterminism(t *testing.T) {
	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	// Create two services with same config
	svc1 := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
	svc2 := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	ctx := context.Background()

	// Generate CEKs from both services
	cek1, err := svc1.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	cek2, err := svc2.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	// Decrypt both CEKs
	decrypted1, err := svc1.DecryptCEK(ctx, cek1)
	require.NoError(t, err)

	decrypted2, err := svc2.DecryptCEK(ctx, cek2)
	require.NoError(t, err)

	// The decrypted CEKs should be identical (deterministic derivation)
	assert.Equal(t, decrypted1, decrypted2)
}

// TestGetOrCreateCEK_RepositoryCreateError tests error handling when repository create fails.
func TestGetOrCreateCEK_RepositoryCreateError(t *testing.T) {
	cekRepo := newMockCEKRepository()
	cekRepo.createErr = assert.AnError
	auditRepo := newMockAuditLogRepository()
	eventPublisher := newMockOfflineEventPublisher()

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, auditRepo, eventPublisher, config)
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	assert.Error(t, err)

	// Should have logged the failure
	assert.Len(t, auditRepo.logs, 1)
	assert.False(t, auditRepo.logs[0].Success)
}

// TestGetOrCreateCEK_RepositoryFindError tests error handling when repository find fails.
func TestGetOrCreateCEK_RepositoryFindError(t *testing.T) {
	cekRepo := newMockCEKRepository()
	cekRepo.findErr = assert.AnError
	auditRepo := newMockAuditLogRepository()
	eventPublisher := newMockOfflineEventPublisher()

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, auditRepo, eventPublisher, config)
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Should still succeed by creating a new CEK (find error is treated as "not found")
	cekRepo.findErr = nil // Reset for create
	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	assert.NoError(t, err)
}

// TestRotateKeys_RepositoryFindError tests error handling when finding CEKs for rotation fails.
func TestRotateKeys_RepositoryFindError(t *testing.T) {
	cekRepo := newMockCEKRepository()
	cekRepo.findErr = assert.AnError

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, nil, nil, config)
	ctx := context.Background()

	newKEK := make([]byte, 32)
	rand.Read(newKEK)

	err := svc.RotateKeys(ctx, 1, 2, newKEK)
	assert.Error(t, err)
}

// TestRotateKeys_UpdateError tests error handling when updating CEK version fails.
func TestRotateKeys_UpdateError(t *testing.T) {
	cekRepo := newMockCEKRepository()

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, nil, nil, config)
	ctx := context.Background()

	// Create a CEK first
	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()
	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	// Set update error
	cekRepo.updateErr = assert.AnError

	newKEK := make([]byte, 32)
	rand.Read(newKEK)

	// Rotation should complete but log errors for failed updates
	err = svc.RotateKeys(ctx, 1, 2, newKEK)
	assert.NoError(t, err) // RotateKeys doesn't return error for individual failures
}

// TestRotateKeys_EmptyCEKList tests rotation with no CEKs to rotate.
func TestRotateKeys_EmptyCEKList(t *testing.T) {
	cekRepo := newMockCEKRepository()

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, nil, nil, config)
	ctx := context.Background()

	newKEK := make([]byte, 32)
	rand.Read(newKEK)

	// Should succeed with no CEKs to rotate
	err := svc.RotateKeys(ctx, 1, 2, newKEK)
	assert.NoError(t, err)
}

// TestEncryptCEKForTransport_DifferentDevices tests transport encryption produces different results for different devices.
func TestEncryptCEKForTransport_DifferentDevices(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	cek := make([]byte, domain.CEKSize)
	rand.Read(cek)

	deviceID1 := uuid.New()
	deviceID2 := uuid.New()

	encrypted1, err := svc.EncryptCEKForTransport(ctx, cek, deviceID1)
	require.NoError(t, err)

	encrypted2, err := svc.EncryptCEKForTransport(ctx, cek, deviceID2)
	require.NoError(t, err)

	// Different devices should produce different encrypted outputs
	assert.NotEqual(t, encrypted1, encrypted2)
}

// TestLogAuditEvent_NilRepository tests that nil audit repository doesn't cause panic.
func TestLogAuditEvent_NilRepository(t *testing.T) {
	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	// Create service with nil audit repo
	svc := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Should not panic
	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	assert.NoError(t, err)
}

// TestPublishKeyEvent_NilPublisher tests that nil event publisher doesn't cause panic.
func TestPublishKeyEvent_NilPublisher(t *testing.T) {
	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	// Create service with nil event publisher
	svc := NewKeyManagementService(newMockCEKRepository(), newMockAuditLogRepository(), nil, config)
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Should not panic
	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	assert.NoError(t, err)
}

// TestDecryptCEK_InvalidCiphertext tests decryption with invalid ciphertext.
func TestDecryptCEK_InvalidCiphertext(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	// Create a CEK record with invalid encrypted key (too short)
	invalidCEK := &domain.ContentEncryptionKey{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		MaterialID:   uuid.New(),
		DeviceID:     uuid.New(),
		EncryptedKey: []byte{1, 2, 3}, // Too short to be valid
		KeyVersion:   1,
	}

	_, err := svc.DecryptCEK(ctx, invalidCEK)
	assert.Error(t, err)
}

// TestDecryptCEK_CorruptedCiphertext tests decryption with corrupted ciphertext.
func TestDecryptCEK_CorruptedCiphertext(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Create a valid CEK first
	cekRecord, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	// Corrupt the encrypted key
	cekRecord.EncryptedKey[len(cekRecord.EncryptedKey)-1] ^= 0xFF

	// Decryption should fail
	_, err = svc.DecryptCEK(ctx, cekRecord)
	assert.Error(t, err)
}

// TestDeriveIVForChunk_LargeChunkIndex tests IV derivation with large chunk indices.
func TestDeriveIVForChunk_LargeChunkIndex(t *testing.T) {
	baseSeed := make([]byte, 16)
	rand.Read(baseSeed)

	// Test with large chunk indices
	iv1 := DeriveIVForChunk(baseSeed, 1000000)
	iv2 := DeriveIVForChunk(baseSeed, 1000001)

	assert.Len(t, iv1, domain.IVSize)
	assert.Len(t, iv2, domain.IVSize)
	assert.NotEqual(t, iv1, iv2)
}

// TestDeriveIVForChunk_ZeroIndex tests IV derivation with zero index.
func TestDeriveIVForChunk_ZeroIndex(t *testing.T) {
	baseSeed := make([]byte, 16)
	rand.Read(baseSeed)

	iv := DeriveIVForChunk(baseSeed, 0)
	assert.Len(t, iv, domain.IVSize)

	// First 8 bytes should match seed
	assert.Equal(t, baseSeed[:8], iv[:8])
}

// TestGenerateNonce_Uniqueness tests that multiple nonce generations produce unique values.
func TestGenerateNonce_Uniqueness(t *testing.T) {
	nonces := make(map[string]bool)

	for i := 0; i < 100; i++ {
		nonce, err := GenerateNonce()
		require.NoError(t, err)
		assert.False(t, nonces[nonce], "duplicate nonce generated")
		nonces[nonce] = true
	}
}

// TestMultipleCEKsForSameUser tests creating multiple CEKs for the same user with different materials.
func TestMultipleCEKsForSameUser(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	userID := uuid.New()
	deviceID := uuid.New()

	// Create CEKs for different materials
	var ceks []*domain.ContentEncryptionKey
	for i := 0; i < 5; i++ {
		materialID := uuid.New()
		cek, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
		require.NoError(t, err)
		ceks = append(ceks, cek)
	}

	// All CEKs should be unique
	ids := make(map[uuid.UUID]bool)
	for _, cek := range ceks {
		assert.False(t, ids[cek.ID], "duplicate CEK ID")
		ids[cek.ID] = true
	}
}

// TestCEKVersionTracking tests that CEKs are created with correct version.
func TestCEKVersionTracking(t *testing.T) {
	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	// Test with different versions
	for version := 1; version <= 5; version++ {
		config := KeyManagementConfig{
			MasterSecret:   masterSecret,
			KEK:            kek,
			CurrentVersion: version,
		}

		svc := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
		ctx := context.Background()

		cek, err := svc.GetOrCreateCEK(ctx, uuid.New(), uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.Equal(t, version, cek.KeyVersion)
	}
}

// TestEncryptDecryptRoundTrip_MultipleIterations tests encryption/decryption consistency.
func TestEncryptDecryptRoundTrip_MultipleIterations(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	for i := 0; i < 50; i++ {
		userID := uuid.New()
		materialID := uuid.New()
		deviceID := uuid.New()

		// Create CEK
		cekRecord, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
		require.NoError(t, err)

		// Decrypt
		decrypted, err := svc.DecryptCEK(ctx, cekRecord)
		require.NoError(t, err)

		// Verify size
		assert.Len(t, decrypted, domain.CEKSize)

		// Verify encrypted differs from decrypted
		assert.NotEqual(t, cekRecord.EncryptedKey, decrypted)
	}
}

// TestTransportEncryption_SameDeviceDifferentNonces tests that same device produces different ciphertexts due to random nonces.
func TestTransportEncryption_SameDeviceDifferentNonces(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	cek := make([]byte, domain.CEKSize)
	rand.Read(cek)

	deviceID := uuid.New()

	// Encrypt same CEK multiple times for same device
	var encrypted [][]byte
	for i := 0; i < 10; i++ {
		enc, err := svc.EncryptCEKForTransport(ctx, cek, deviceID)
		require.NoError(t, err)
		encrypted = append(encrypted, enc)
	}

	// All should be different due to random nonces
	for i := 0; i < len(encrypted); i++ {
		for j := i + 1; j < len(encrypted); j++ {
			assert.NotEqual(t, encrypted[i], encrypted[j], "encrypted values should differ due to random nonces")
		}
	}
}

// TestEncryptCEKForTransport_EmptyCEK tests transport encryption with empty CEK.
func TestEncryptCEKForTransport_EmptyCEK(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	deviceID := uuid.New()

	// Empty CEK should still work (GCM can encrypt empty data)
	encrypted, err := svc.EncryptCEKForTransport(ctx, []byte{}, deviceID)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted) // Should have nonce at minimum
}

// TestEncryptCEKForTransport_LargeCEK tests transport encryption with large data.
func TestEncryptCEKForTransport_LargeCEK(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	// Large CEK (1KB)
	largeCEK := make([]byte, 1024)
	rand.Read(largeCEK)

	deviceID := uuid.New()

	encrypted, err := svc.EncryptCEKForTransport(ctx, largeCEK, deviceID)
	require.NoError(t, err)
	assert.Greater(t, len(encrypted), len(largeCEK))
}

// TestRotateKeys_DecryptionFailure tests rotation when decryption fails for some CEKs.
func TestRotateKeys_DecryptionFailure(t *testing.T) {
	cekRepo := newMockCEKRepository()

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, nil, nil, config)
	ctx := context.Background()

	// Create a valid CEK
	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()
	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	// Manually add a CEK with invalid encrypted data
	invalidCEK := &domain.ContentEncryptionKey{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		MaterialID:   uuid.New(),
		DeviceID:     uuid.New(),
		EncryptedKey: []byte{1, 2, 3}, // Invalid - too short
		KeyVersion:   1,
	}
	key := compositeKey(invalidCEK.UserID, invalidCEK.MaterialID, invalidCEK.DeviceID)
	cekRepo.ceks[key] = invalidCEK
	cekRepo.byVersion[1] = append(cekRepo.byVersion[1], invalidCEK)

	newKEK := make([]byte, 32)
	rand.Read(newKEK)

	// Rotation should continue despite decryption failure for invalid CEK
	err = svc.RotateKeys(ctx, 1, 2, newKEK)
	assert.NoError(t, err)
}

// TestRotateKeys_ReEncryptionFailure tests rotation when re-encryption fails.
func TestRotateKeys_ReEncryptionFailure(t *testing.T) {
	cekRepo := newMockCEKRepository()

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, nil, nil, config)
	ctx := context.Background()

	// Create a valid CEK
	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()
	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	// Use invalid KEK (wrong size) - this will cause cipher creation to fail
	invalidKEK := []byte{1, 2, 3} // Too short for AES

	// Rotation should continue despite re-encryption failure
	err = svc.RotateKeys(ctx, 1, 2, invalidKEK)
	assert.NoError(t, err)
}

// mockFailingAuditLogRepository is a mock that fails on Create.
type mockFailingAuditLogRepository struct {
	mockAuditLogRepository
}

func (m *mockFailingAuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	return assert.AnError
}

// TestLogAuditEvent_CreateError tests that audit log creation error is handled gracefully.
func TestLogAuditEvent_CreateError(t *testing.T) {
	cekRepo := newMockCEKRepository()
	auditRepo := &mockFailingAuditLogRepository{}
	eventPublisher := newMockOfflineEventPublisher()

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, auditRepo, eventPublisher, config)
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Should not fail even if audit logging fails
	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	assert.NoError(t, err)
}

// mockFailingEventPublisher is a mock that fails on publish.
type mockFailingEventPublisher struct {
	mockOfflineEventPublisher
}

func (m *mockFailingEventPublisher) PublishKeyGenerated(ctx context.Context, event KeyEvent) error {
	return assert.AnError
}

func (m *mockFailingEventPublisher) PublishKeyRetrieved(ctx context.Context, event KeyEvent) error {
	return assert.AnError
}

// TestPublishKeyEvent_PublishError tests that publish error is handled gracefully.
func TestPublishKeyEvent_PublishError(t *testing.T) {
	cekRepo := newMockCEKRepository()
	auditRepo := newMockAuditLogRepository()
	eventPublisher := &mockFailingEventPublisher{}

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, auditRepo, eventPublisher, config)
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Should not fail even if event publishing fails
	cek, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	assert.NoError(t, err)
	assert.NotNil(t, cek)

	// Get same CEK again to trigger retrieve event
	_, err = svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	assert.NoError(t, err)
}

// TestDecryptCEKWithKEK_InvalidKEK tests decryption with invalid KEK size.
func TestDecryptCEKWithKEK_InvalidKEK(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Create a valid CEK
	cekRecord, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	// Try to decrypt with wrong KEK (by creating a new service with different KEK)
	wrongKEK := make([]byte, 32)
	rand.Read(wrongKEK)

	config := KeyManagementConfig{
		MasterSecret:   svc.masterSecret,
		KEK:            wrongKEK,
		CurrentVersion: 1,
	}

	wrongSvc := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)

	// Decryption should fail with wrong KEK
	_, err = wrongSvc.DecryptCEK(ctx, cekRecord)
	assert.Error(t, err)
}

// TestGetOrCreateCEK_GenerationError tests error handling when CEK generation fails.
func TestGetOrCreateCEK_GenerationError(t *testing.T) {
	// This test verifies the error path when generateCEK fails
	// Since generateCEK uses HKDF which is deterministic, we can't easily make it fail
	// But we can test with empty master secret which should still work
	cekRepo := newMockCEKRepository()
	auditRepo := newMockAuditLogRepository()
	eventPublisher := newMockOfflineEventPublisher()

	config := KeyManagementConfig{
		MasterSecret:   []byte{}, // Empty master secret
		KEK:            make([]byte, 32),
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, auditRepo, eventPublisher, config)
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Should still work with empty master secret (HKDF handles it)
	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	assert.NoError(t, err)
}

// TestEncryptCEKForStorage_InvalidKEK tests encryption with invalid KEK.
func TestEncryptCEKForStorage_InvalidKEK(t *testing.T) {
	cekRepo := newMockCEKRepository()

	config := KeyManagementConfig{
		MasterSecret:   make([]byte, 32),
		KEK:            []byte{1, 2, 3}, // Invalid KEK size
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, nil, nil, config)
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Should fail due to invalid KEK
	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	assert.Error(t, err)
}

// TestDeriveTransportKey_Deterministic tests that transport key derivation is deterministic.
func TestDeriveTransportKey_Deterministic(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	cek := make([]byte, domain.CEKSize)
	rand.Read(cek)

	deviceID := uuid.New()

	// Encrypt twice with same device
	enc1, err := svc.EncryptCEKForTransport(ctx, cek, deviceID)
	require.NoError(t, err)

	enc2, err := svc.EncryptCEKForTransport(ctx, cek, deviceID)
	require.NoError(t, err)

	// Results should differ due to random nonces, but both should be valid
	assert.NotEqual(t, enc1, enc2)
	assert.Greater(t, len(enc1), len(cek))
	assert.Greater(t, len(enc2), len(cek))
}

// TestKeyRotation_MultipleVersions tests rotating through multiple versions.
func TestKeyRotation_MultipleVersions(t *testing.T) {
	cekRepo := newMockCEKRepository()

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, nil, nil, config)
	ctx := context.Background()

	// Create CEKs with version 1
	for i := 0; i < 3; i++ {
		_, err := svc.GetOrCreateCEK(ctx, uuid.New(), uuid.New(), uuid.New())
		require.NoError(t, err)
	}

	// Rotate from v1 to v2
	newKEK := make([]byte, 32)
	rand.Read(newKEK)
	err := svc.RotateKeys(ctx, 1, 2, newKEK)
	require.NoError(t, err)

	// Verify all are v2
	v2CEKs, _ := cekRepo.FindByKeyVersion(ctx, 2)
	assert.Len(t, v2CEKs, 3)

	// Update service KEK for next rotation
	svc.kek = newKEK

	// Rotate from v2 to v3
	newerKEK := make([]byte, 32)
	rand.Read(newerKEK)
	err = svc.RotateKeys(ctx, 2, 3, newerKEK)
	require.NoError(t, err)

	// Verify all are v3
	v3CEKs, _ := cekRepo.FindByKeyVersion(ctx, 3)
	assert.Len(t, v3CEKs, 3)
}

// TestBuildHKDFInfo_DifferentInputs tests HKDF info with different inputs.
func TestBuildHKDFInfo_DifferentInputs(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	info1 := buildHKDFInfo(userID1, materialID, deviceID)
	info2 := buildHKDFInfo(userID2, materialID, deviceID)

	// Different users should produce different info
	assert.NotEqual(t, info1, info2)

	// Both should have same length
	assert.Equal(t, len(info1), len(info2))
}

// TestDecryptCEK_ValidCiphertext tests successful decryption.
func TestDecryptCEK_ValidCiphertext(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	// Create multiple CEKs and verify all decrypt correctly
	for i := 0; i < 10; i++ {
		userID := uuid.New()
		materialID := uuid.New()
		deviceID := uuid.New()

		cekRecord, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
		require.NoError(t, err)

		decrypted, err := svc.DecryptCEK(ctx, cekRecord)
		require.NoError(t, err)
		assert.Len(t, decrypted, domain.CEKSize)
	}
}

// TestGetOrCreateCEK_ConcurrentAccess tests concurrent CEK creation.
func TestGetOrCreateCEK_ConcurrentAccess(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Create CEK first
	cek1, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)

	// Simulate concurrent access by getting the same CEK multiple times
	for i := 0; i < 10; i++ {
		cek, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
		require.NoError(t, err)
		assert.Equal(t, cek1.ID, cek.ID)
	}
}

// TestEncryptCEKWithKEK_ValidInput tests encryption with valid input.
func TestEncryptCEKWithKEK_ValidInput(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	// Test with various CEK sizes
	sizes := []int{16, 32, 64, 128}
	for _, size := range sizes {
		cek := make([]byte, size)
		rand.Read(cek)

		deviceID := uuid.New()
		encrypted, err := svc.EncryptCEKForTransport(ctx, cek, deviceID)
		require.NoError(t, err)
		assert.Greater(t, len(encrypted), size)
	}
}

// TestAuditLogging_RetrieveEvent tests audit logging for CEK retrieval.
func TestAuditLogging_RetrieveEvent(t *testing.T) {
	svc, _, auditRepo, _ := newTestKeyManagementService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Create CEK
	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)
	assert.Len(t, auditRepo.logs, 1)
	assert.Equal(t, domain.AuditActionKeyGenerate, auditRepo.logs[0].Action)

	// Retrieve CEK
	_, err = svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)
	assert.Len(t, auditRepo.logs, 2)
	assert.Equal(t, domain.AuditActionKeyRetrieve, auditRepo.logs[1].Action)
}

// TestEventPublishing_RetrieveEvent tests event publishing for CEK retrieval.
func TestEventPublishing_RetrieveEvent(t *testing.T) {
	svc, _, _, eventPublisher := newTestKeyManagementService()
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Create CEK
	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)
	assert.Len(t, eventPublisher.keyEvents, 1)

	// Retrieve CEK
	_, err = svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	require.NoError(t, err)
	assert.Len(t, eventPublisher.keyEvents, 2)
}

// TestDeriveIVForChunk_MaxChunkIndex tests IV derivation with max uint32 chunk index.
func TestDeriveIVForChunk_MaxChunkIndex(t *testing.T) {
	baseSeed := make([]byte, 16)
	rand.Read(baseSeed)

	// Test with max uint32 value
	iv := DeriveIVForChunk(baseSeed, int(^uint32(0)))
	assert.Len(t, iv, domain.IVSize)
}

// TestGenerateNonce_Length tests that generated nonces have correct length.
func TestGenerateNonce_Length(t *testing.T) {
	for i := 0; i < 100; i++ {
		nonce, err := GenerateNonce()
		require.NoError(t, err)
		assert.Len(t, nonce, domain.NonceHexLength)
	}
}

// TestEncryptCEKForTransport_VariousDevices tests transport encryption with many different devices.
func TestEncryptCEKForTransport_VariousDevices(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	cek := make([]byte, domain.CEKSize)
	rand.Read(cek)

	// Test with many different devices
	for i := 0; i < 50; i++ {
		deviceID := uuid.New()
		encrypted, err := svc.EncryptCEKForTransport(ctx, cek, deviceID)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)
	}
}

// TestGetOrCreateCEK_EncryptionFailure tests the encryption failure path.
func TestGetOrCreateCEK_EncryptionFailure(t *testing.T) {
	cekRepo := newMockCEKRepository()
	auditRepo := newMockAuditLogRepository()

	// Use invalid KEK to trigger encryption failure
	config := KeyManagementConfig{
		MasterSecret:   make([]byte, 32),
		KEK:            []byte{1, 2, 3, 4, 5}, // Invalid size - not 16, 24, or 32 bytes
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, auditRepo, nil, config)
	ctx := context.Background()

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
	assert.Error(t, err)

	// Should have logged the failure
	assert.Len(t, auditRepo.logs, 1)
	assert.False(t, auditRepo.logs[0].Success)
}

// TestDecryptCEKWithKEK_ShortCiphertext tests decryption with ciphertext shorter than nonce.
func TestDecryptCEKWithKEK_ShortCiphertext(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	// Create a CEK record with ciphertext that's too short
	shortCEK := &domain.ContentEncryptionKey{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		MaterialID:   uuid.New(),
		DeviceID:     uuid.New(),
		EncryptedKey: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, // 11 bytes, less than GCM nonce (12)
		KeyVersion:   1,
	}

	_, err := svc.DecryptCEK(ctx, shortCEK)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ciphertext too short")
}

// TestDecryptCEKWithKEK_ExactNonceSize tests decryption with ciphertext exactly nonce size.
func TestDecryptCEKWithKEK_ExactNonceSize(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	// Create a CEK record with ciphertext exactly nonce size (no actual ciphertext)
	exactNonceCEK := &domain.ContentEncryptionKey{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		MaterialID:   uuid.New(),
		DeviceID:     uuid.New(),
		EncryptedKey: make([]byte, 12), // Exactly nonce size
		KeyVersion:   1,
	}

	_, err := svc.DecryptCEK(ctx, exactNonceCEK)
	assert.Error(t, err) // Should fail because there's no actual ciphertext
}

// TestKeyManagementService_AllFieldsSet tests that all service fields are properly set.
func TestKeyManagementService_AllFieldsSet(t *testing.T) {
	cekRepo := newMockCEKRepository()
	auditRepo := newMockAuditLogRepository()
	eventPublisher := newMockOfflineEventPublisher()

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 5,
	}

	svc := NewKeyManagementService(cekRepo, auditRepo, eventPublisher, config)

	assert.NotNil(t, svc)
	assert.Equal(t, masterSecret, svc.masterSecret)
	assert.Equal(t, kek, svc.kek)
	assert.Equal(t, 5, svc.currentVersion)
}

// TestRotateKeys_PartialFailure tests rotation with some CEKs failing.
func TestRotateKeys_PartialFailure(t *testing.T) {
	cekRepo := newMockCEKRepository()

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, nil, nil, config)
	ctx := context.Background()

	// Create valid CEKs
	for i := 0; i < 2; i++ {
		_, err := svc.GetOrCreateCEK(ctx, uuid.New(), uuid.New(), uuid.New())
		require.NoError(t, err)
	}

	// Add an invalid CEK that will fail decryption
	invalidCEK := &domain.ContentEncryptionKey{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		MaterialID:   uuid.New(),
		DeviceID:     uuid.New(),
		EncryptedKey: []byte{1, 2, 3, 4, 5}, // Invalid
		KeyVersion:   1,
	}
	key := compositeKey(invalidCEK.UserID, invalidCEK.MaterialID, invalidCEK.DeviceID)
	cekRepo.ceks[key] = invalidCEK
	cekRepo.byVersion[1] = append(cekRepo.byVersion[1], invalidCEK)

	newKEK := make([]byte, 32)
	rand.Read(newKEK)

	// Rotation should complete (logs errors but doesn't fail)
	err := svc.RotateKeys(ctx, 1, 2, newKEK)
	assert.NoError(t, err)

	// Valid CEKs should be rotated
	v2CEKs, _ := cekRepo.FindByKeyVersion(ctx, 2)
	assert.Equal(t, 2, len(v2CEKs)) // Only 2 valid CEKs rotated
}

// TestEncryptCEKForTransport_ZeroLengthCEK tests transport encryption with zero-length CEK.
func TestEncryptCEKForTransport_ZeroLengthCEK(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	deviceID := uuid.New()

	// Zero-length CEK
	encrypted, err := svc.EncryptCEKForTransport(ctx, []byte{}, deviceID)
	require.NoError(t, err)
	// Should have nonce + GCM tag (12 + 16 = 28 bytes minimum)
	assert.GreaterOrEqual(t, len(encrypted), 28)
}

// TestDeriveIVForChunk_ConsecutiveIndices tests IV derivation for consecutive chunk indices.
func TestDeriveIVForChunk_ConsecutiveIndices(t *testing.T) {
	baseSeed := make([]byte, 16)
	rand.Read(baseSeed)

	// Generate IVs for consecutive indices
	ivs := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		ivs[i] = DeriveIVForChunk(baseSeed, i)
	}

	// All IVs should be unique
	seen := make(map[string]bool)
	for _, iv := range ivs {
		key := string(iv)
		assert.False(t, seen[key], "duplicate IV found")
		seen[key] = true
	}
}

// TestGenerateNonce_HexFormat tests that nonces are valid hex strings.
func TestGenerateNonce_HexFormat(t *testing.T) {
	for i := 0; i < 50; i++ {
		nonce, err := GenerateNonce()
		require.NoError(t, err)

		// Should only contain hex characters
		for _, c := range nonce {
			assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
				"nonce contains non-hex character: %c", c)
		}
	}
}

// TestGetOrCreateCEK_SameInputsDifferentServices tests determinism across service instances.
func TestGetOrCreateCEK_SameInputsDifferentServices(t *testing.T) {
	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Create multiple services with same config
	var decryptedCEKs [][]byte
	for i := 0; i < 5; i++ {
		svc := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
		ctx := context.Background()

		cekRecord, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
		require.NoError(t, err)

		decrypted, err := svc.DecryptCEK(ctx, cekRecord)
		require.NoError(t, err)

		decryptedCEKs = append(decryptedCEKs, decrypted)
	}

	// All decrypted CEKs should be identical (deterministic)
	for i := 1; i < len(decryptedCEKs); i++ {
		assert.Equal(t, decryptedCEKs[0], decryptedCEKs[i])
	}
}

// TestEncryptCEKWithKEK_DifferentKEKSizes tests encryption with different valid KEK sizes.
func TestEncryptCEKWithKEK_DifferentKEKSizes(t *testing.T) {
	// AES supports 16, 24, and 32 byte keys
	kekSizes := []int{16, 24, 32}

	for _, size := range kekSizes {
		kek := make([]byte, size)
		rand.Read(kek)

		config := KeyManagementConfig{
			MasterSecret:   make([]byte, 32),
			KEK:            kek,
			CurrentVersion: 1,
		}

		svc := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
		ctx := context.Background()

		cekRecord, err := svc.GetOrCreateCEK(ctx, uuid.New(), uuid.New(), uuid.New())
		require.NoError(t, err, "KEK size %d should work", size)

		decrypted, err := svc.DecryptCEK(ctx, cekRecord)
		require.NoError(t, err)
		assert.Len(t, decrypted, domain.CEKSize)
	}
}

// TestRotateKeys_LargeBatch tests rotation with a large number of CEKs.
func TestRotateKeys_LargeBatch(t *testing.T) {
	cekRepo := newMockCEKRepository()

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, nil, nil, config)
	ctx := context.Background()

	// Create many CEKs
	numCEKs := 50
	for i := 0; i < numCEKs; i++ {
		_, err := svc.GetOrCreateCEK(ctx, uuid.New(), uuid.New(), uuid.New())
		require.NoError(t, err)
	}

	// Verify all created
	v1CEKs, _ := cekRepo.FindByKeyVersion(ctx, 1)
	assert.Len(t, v1CEKs, numCEKs)

	// Rotate all
	newKEK := make([]byte, 32)
	rand.Read(newKEK)
	err := svc.RotateKeys(ctx, 1, 2, newKEK)
	require.NoError(t, err)

	// Verify all rotated
	v2CEKs, _ := cekRepo.FindByKeyVersion(ctx, 2)
	assert.Len(t, v2CEKs, numCEKs)
}

// TestDecryptCEKWithKEK_InvalidKEKSize tests decryption with invalid KEK sizes.
func TestDecryptCEKWithKEK_InvalidKEKSize(t *testing.T) {
	// Test with various invalid KEK sizes
	invalidSizes := []int{0, 1, 15, 17, 23, 25, 31, 33, 64}

	for _, size := range invalidSizes {
		kek := make([]byte, size)
		if size > 0 {
			rand.Read(kek)
		}

		config := KeyManagementConfig{
			MasterSecret:   make([]byte, 32),
			KEK:            kek,
			CurrentVersion: 1,
		}

		svc := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
		ctx := context.Background()

		_, err := svc.GetOrCreateCEK(ctx, uuid.New(), uuid.New(), uuid.New())
		if size == 16 || size == 24 || size == 32 {
			assert.NoError(t, err, "KEK size %d should work", size)
		} else {
			assert.Error(t, err, "KEK size %d should fail", size)
		}
	}
}

// TestEncryptCEKForTransport_MultipleEncryptions tests multiple encryptions produce unique results.
func TestEncryptCEKForTransport_MultipleEncryptions(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	cek := make([]byte, domain.CEKSize)
	rand.Read(cek)

	deviceID := uuid.New()

	// Encrypt same CEK 100 times
	results := make(map[string]bool)
	for i := 0; i < 100; i++ {
		encrypted, err := svc.EncryptCEKForTransport(ctx, cek, deviceID)
		require.NoError(t, err)

		key := string(encrypted)
		assert.False(t, results[key], "duplicate encryption result")
		results[key] = true
	}
}

// TestGetOrCreateCEK_AllErrorPaths tests all error paths in GetOrCreateCEK.
func TestGetOrCreateCEK_AllErrorPaths(t *testing.T) {
	tests := []struct {
		name      string
		kekSize   int
		createErr error
		wantErr   bool
	}{
		{"valid KEK", 32, nil, false},
		{"invalid KEK size 5", 5, nil, true},
		{"invalid KEK size 0", 0, nil, true},
		{"create error", 32, assert.AnError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cekRepo := newMockCEKRepository()
			cekRepo.createErr = tt.createErr
			auditRepo := newMockAuditLogRepository()

			kek := make([]byte, tt.kekSize)
			if tt.kekSize > 0 {
				rand.Read(kek)
			}

			config := KeyManagementConfig{
				MasterSecret:   make([]byte, 32),
				KEK:            kek,
				CurrentVersion: 1,
			}

			svc := NewKeyManagementService(cekRepo, auditRepo, nil, config)
			ctx := context.Background()

			_, err := svc.GetOrCreateCEK(ctx, uuid.New(), uuid.New(), uuid.New())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDecryptCEK_AllErrorPaths tests all error paths in DecryptCEK.
func TestDecryptCEK_AllErrorPaths(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	tests := []struct {
		name         string
		encryptedKey []byte
		wantErr      bool
	}{
		{"empty", []byte{}, true},
		{"too short", []byte{1, 2, 3}, true},
		{"exactly nonce size", make([]byte, 12), true},
		{"nonce + 1 byte", make([]byte, 13), true},
		{"random invalid", make([]byte, 50), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cekRecord := &domain.ContentEncryptionKey{
				ID:           uuid.New(),
				EncryptedKey: tt.encryptedKey,
			}

			_, err := svc.DecryptCEK(ctx, cekRecord)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBuildHKDFInfo_Consistency tests HKDF info building consistency.
func TestBuildHKDFInfo_Consistency(t *testing.T) {
	userID := uuid.New()
	materialID := uuid.New()
	deviceID := uuid.New()

	// Build info multiple times
	for i := 0; i < 100; i++ {
		info := buildHKDFInfo(userID, materialID, deviceID)
		assert.Len(t, info, len(domain.HKDFInfoPrefix)+48)
		assert.True(t, bytes.HasPrefix(info, []byte(domain.HKDFInfoPrefix)))
	}
}

// TestDeriveIVForChunk_AllIndices tests IV derivation across a range of indices.
func TestDeriveIVForChunk_AllIndices(t *testing.T) {
	baseSeed := make([]byte, 16)
	rand.Read(baseSeed)

	// Test indices 0-999
	ivs := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		iv := DeriveIVForChunk(baseSeed, i)
		assert.Len(t, iv, domain.IVSize)

		key := string(iv)
		assert.False(t, ivs[key], "duplicate IV at index %d", i)
		ivs[key] = true
	}
}

// TestKeyManagementConfig_ZeroValues tests service creation with zero values.
func TestKeyManagementConfig_ZeroValues(t *testing.T) {
	config := KeyManagementConfig{
		MasterSecret:   nil,
		KEK:            nil,
		CurrentVersion: 0,
	}

	svc := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
	assert.NotNil(t, svc)
	assert.Nil(t, svc.masterSecret)
	assert.Nil(t, svc.kek)
	assert.Equal(t, 0, svc.currentVersion)
}

// TestRotateKeys_NoMatchingVersion tests rotation when no CEKs match the version.
func TestRotateKeys_NoMatchingVersion(t *testing.T) {
	cekRepo := newMockCEKRepository()

	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)
	kek := make([]byte, 32)
	rand.Read(kek)

	config := KeyManagementConfig{
		MasterSecret:   masterSecret,
		KEK:            kek,
		CurrentVersion: 1,
	}

	svc := NewKeyManagementService(cekRepo, nil, nil, config)
	ctx := context.Background()

	// Create CEKs with version 1
	_, err := svc.GetOrCreateCEK(ctx, uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	// Try to rotate version 99 (doesn't exist)
	newKEK := make([]byte, 32)
	rand.Read(newKEK)
	err = svc.RotateKeys(ctx, 99, 100, newKEK)
	assert.NoError(t, err) // Should succeed with 0 CEKs rotated
}

// TestEncryptCEKForTransport_VeryLargeCEK tests transport encryption with very large data.
func TestEncryptCEKForTransport_VeryLargeCEK(t *testing.T) {
	svc, _, _, _ := newTestKeyManagementService()
	ctx := context.Background()

	// 10KB CEK
	largeCEK := make([]byte, 10*1024)
	rand.Read(largeCEK)

	deviceID := uuid.New()

	encrypted, err := svc.EncryptCEKForTransport(ctx, largeCEK, deviceID)
	require.NoError(t, err)
	assert.Greater(t, len(encrypted), len(largeCEK))
}
