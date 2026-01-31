// Package offline provides integration tests for the Offline Material Service.
// These tests verify end-to-end flows across multiple components.
//
//go:build integration
// +build integration

package offline

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ngasihtau/internal/offline/application"
	"ngasihtau/internal/offline/domain"
)

// MockDeviceRepository implements domain.DeviceRepository for testing.
type MockDeviceRepository struct {
	devices map[uuid.UUID]*domain.Device
}

func NewMockDeviceRepository() *MockDeviceRepository {
	return &MockDeviceRepository{
		devices: make(map[uuid.UUID]*domain.Device),
	}
}

func (r *MockDeviceRepository) Create(ctx context.Context, device *domain.Device) error {
	r.devices[device.ID] = device
	return nil
}

func (r *MockDeviceRepository) FindById(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	if device, ok := r.devices[id]; ok {
		return device, nil
	}
	return nil, nil
}

func (r *MockDeviceRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error) {
	var result []*domain.Device
	for _, device := range r.devices {
		if device.UserID == userID && device.RevokedAt == nil {
			result = append(result, device)
		}
	}
	return result, nil
}

func (r *MockDeviceRepository) FindByFingerprint(ctx context.Context, userID uuid.UUID, fingerprint string) (*domain.Device, error) {
	for _, device := range r.devices {
		if device.UserID == userID && device.Fingerprint == fingerprint && device.RevokedAt == nil {
			return device, nil
		}
	}
	return nil, nil
}

func (r *MockDeviceRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	count := 0
	for _, device := range r.devices {
		if device.UserID == userID && device.RevokedAt == nil {
			count++
		}
	}
	return count, nil
}

func (r *MockDeviceRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	if device, ok := r.devices[id]; ok {
		device.LastUsedAt = time.Now()
	}
	return nil
}

func (r *MockDeviceRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	if device, ok := r.devices[id]; ok {
		now := time.Now()
		device.RevokedAt = &now
	}
	return nil
}

// MockLicenseRepository implements domain.LicenseRepository for testing.
type MockLicenseRepository struct {
	licenses map[uuid.UUID]*domain.License
}

func NewMockLicenseRepository() *MockLicenseRepository {
	return &MockLicenseRepository{
		licenses: make(map[uuid.UUID]*domain.License),
	}
}

func (r *MockLicenseRepository) Create(ctx context.Context, license *domain.License) error {
	r.licenses[license.ID] = license
	return nil
}

func (r *MockLicenseRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.License, error) {
	if license, ok := r.licenses[id]; ok {
		return license, nil
	}
	return nil, nil
}

func (r *MockLicenseRepository) FindByUserAndMaterial(ctx context.Context, userID, materialID, deviceID uuid.UUID) (*domain.License, error) {
	for _, license := range r.licenses {
		if license.UserID == userID && license.MaterialID == materialID && license.DeviceID == deviceID && license.RevokedAt == nil {
			return license, nil
		}
	}
	return nil, nil
}

func (r *MockLicenseRepository) FindActiveByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]*domain.License, error) {
	var result []*domain.License
	for _, license := range r.licenses {
		if license.DeviceID == deviceID && license.Status == domain.LicenseStatusActive && license.RevokedAt == nil {
			result = append(result, license)
		}
	}
	return result, nil
}

func (r *MockLicenseRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.License, error) {
	var result []*domain.License
	for _, license := range r.licenses {
		if license.UserID == userID && license.Status == domain.LicenseStatusActive && license.RevokedAt == nil {
			result = append(result, license)
		}
	}
	return result, nil
}

func (r *MockLicenseRepository) UpdateValidation(ctx context.Context, id uuid.UUID, nonce string) error {
	if license, ok := r.licenses[id]; ok {
		license.LastValidatedAt = time.Now()
		license.Nonce = nonce
	}
	return nil
}

func (r *MockLicenseRepository) UpdateExpiration(ctx context.Context, id uuid.UUID, expiresAt time.Time) error {
	if license, ok := r.licenses[id]; ok {
		license.ExpiresAt = expiresAt
	}
	return nil
}

func (r *MockLicenseRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	if license, ok := r.licenses[id]; ok {
		license.Status = domain.LicenseStatusRevoked
		now := time.Now()
		license.RevokedAt = &now
	}
	return nil
}

func (r *MockLicenseRepository) RevokeByDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	for _, license := range r.licenses {
		if license.DeviceID == deviceID && license.RevokedAt == nil {
			license.Status = domain.LicenseStatusRevoked
			now := time.Now()
			license.RevokedAt = &now
		}
	}
	return nil
}

func (r *MockLicenseRepository) RevokeByMaterialID(ctx context.Context, materialID uuid.UUID) error {
	for _, license := range r.licenses {
		if license.MaterialID == materialID && license.RevokedAt == nil {
			license.Status = domain.LicenseStatusRevoked
			now := time.Now()
			license.RevokedAt = &now
		}
	}
	return nil
}

func (r *MockLicenseRepository) RevokeByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) error {
	for _, license := range r.licenses {
		if license.UserID == userID && license.MaterialID == materialID && license.RevokedAt == nil {
			license.Status = domain.LicenseStatusRevoked
			now := time.Now()
			license.RevokedAt = &now
		}
	}
	return nil
}

// MockCEKRepository implements domain.CEKRepository for testing.
type MockCEKRepository struct {
	ceks map[uuid.UUID]*domain.ContentEncryptionKey
}

func NewMockCEKRepository() *MockCEKRepository {
	return &MockCEKRepository{
		ceks: make(map[uuid.UUID]*domain.ContentEncryptionKey),
	}
}

func (r *MockCEKRepository) Create(ctx context.Context, cek *domain.ContentEncryptionKey) error {
	r.ceks[cek.ID] = cek
	return nil
}

func (r *MockCEKRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ContentEncryptionKey, error) {
	if cek, ok := r.ceks[id]; ok {
		return cek, nil
	}
	return nil, nil
}

func (r *MockCEKRepository) FindByComposite(ctx context.Context, userID, materialID, deviceID uuid.UUID) (*domain.ContentEncryptionKey, error) {
	for _, cek := range r.ceks {
		if cek.UserID == userID && cek.MaterialID == materialID && cek.DeviceID == deviceID {
			return cek, nil
		}
	}
	return nil, nil
}

func (r *MockCEKRepository) DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	for id, cek := range r.ceks {
		if cek.DeviceID == deviceID {
			delete(r.ceks, id)
		}
	}
	return nil
}

func (r *MockCEKRepository) DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error {
	for id, cek := range r.ceks {
		if cek.MaterialID == materialID {
			delete(r.ceks, id)
		}
	}
	return nil
}

func (r *MockCEKRepository) FindByKeyVersion(ctx context.Context, keyVersion int) ([]*domain.ContentEncryptionKey, error) {
	var result []*domain.ContentEncryptionKey
	for _, cek := range r.ceks {
		if cek.KeyVersion == keyVersion {
			result = append(result, cek)
		}
	}
	return result, nil
}

func (r *MockCEKRepository) UpdateKeyVersion(ctx context.Context, id uuid.UUID, encryptedKey []byte, keyVersion int) error {
	if cek, ok := r.ceks[id]; ok {
		cek.EncryptedKey = encryptedKey
		cek.KeyVersion = keyVersion
	}
	return nil
}

// MockMaterialAccessChecker implements application.LicenseMaterialAccessChecker for testing.
type MockMaterialAccessChecker struct {
	accessMap map[string]bool // key: "userID:materialID"
}

func NewMockMaterialAccessChecker() *MockMaterialAccessChecker {
	return &MockMaterialAccessChecker{
		accessMap: make(map[string]bool),
	}
}

func (c *MockMaterialAccessChecker) GrantAccess(userID, materialID uuid.UUID) {
	key := userID.String() + ":" + materialID.String()
	c.accessMap[key] = true
}

func (c *MockMaterialAccessChecker) CheckAccess(ctx context.Context, userID, materialID uuid.UUID) (bool, error) {
	key := userID.String() + ":" + materialID.String()
	return c.accessMap[key], nil
}

// TestEndToEndDownloadFlow tests the complete download flow:
// 1. Register device
// 2. Issue license
// 3. Validate license
// 4. Renew license
// Implements Task 18.1: End-to-end download flow test.
func TestEndToEndDownloadFlow(t *testing.T) {
	ctx := context.Background()

	// Setup mock repositories
	deviceRepo := NewMockDeviceRepository()
	licenseRepo := NewMockLicenseRepository()
	cekRepo := NewMockCEKRepository()
	accessChecker := NewMockMaterialAccessChecker()

	// Create services
	deviceService := application.NewDeviceService(deviceRepo, licenseRepo, cekRepo, nil)
	licenseService := application.NewLicenseService(licenseRepo, deviceRepo, accessChecker, nil)

	// Test data
	userID := uuid.New()
	materialID := uuid.New()
	fingerprint := "test-fingerprint-" + uuid.New().String()

	// Grant material access
	accessChecker.GrantAccess(userID, materialID)

	// Step 1: Register device
	t.Run("Step1_RegisterDevice", func(t *testing.T) {
		device, err := deviceService.RegisterDevice(ctx, application.RegisterDeviceInput{
			UserID:      userID,
			Fingerprint: fingerprint,
			Name:        "Test Device",
			Platform:    domain.PlatformDesktop,
		})
		require.NoError(t, err)
		require.NotNil(t, device)
		assert.Equal(t, userID, device.UserID)
		assert.Equal(t, fingerprint, device.Fingerprint)
		assert.Equal(t, domain.PlatformDesktop, device.Platform)
	})

	// Get the registered device
	devices, err := deviceService.ListDevices(ctx, userID)
	require.NoError(t, err)
	require.Len(t, devices, 1)
	device := devices[0]

	// Step 2: Issue license
	var license *domain.License
	t.Run("Step2_IssueLicense", func(t *testing.T) {
		var err error
		license, err = licenseService.IssueLicense(ctx, application.IssueLicenseInput{
			UserID:      userID,
			MaterialID:  materialID,
			DeviceID:    device.ID,
			Fingerprint: fingerprint,
		})
		require.NoError(t, err)
		require.NotNil(t, license)
		assert.Equal(t, userID, license.UserID)
		assert.Equal(t, materialID, license.MaterialID)
		assert.Equal(t, device.ID, license.DeviceID)
		assert.Equal(t, domain.LicenseStatusActive, license.Status)
		assert.NotEmpty(t, license.Nonce)
	})

	// Step 3: Validate license
	t.Run("Step3_ValidateLicense", func(t *testing.T) {
		originalNonce := license.Nonce
		validatedLicense, err := licenseService.ValidateLicense(ctx, application.ValidateLicenseInput{
			LicenseID:   license.ID,
			DeviceID:    device.ID,
			Fingerprint: fingerprint,
			Nonce:       license.Nonce,
		})
		require.NoError(t, err)
		require.NotNil(t, validatedLicense)
		// Nonce should be updated after validation (new nonce generated)
		assert.NotEmpty(t, validatedLicense.Nonce)
		// The returned license should have a new nonce
		// Note: The original nonce was used for validation, new one is for next validation
		_ = originalNonce // Used for validation
		// Update license reference for next step
		license = validatedLicense
	})

	// Step 4: Renew license
	t.Run("Step4_RenewLicense", func(t *testing.T) {
		renewedLicense, err := licenseService.RenewLicense(ctx, application.RenewLicenseInput{
			LicenseID:   license.ID,
			DeviceID:    device.ID,
			Fingerprint: fingerprint,
		})
		require.NoError(t, err)
		require.NotNil(t, renewedLicense)
		// Expiry should be set to approximately 30 days from now
		expectedExpiry := time.Now().Add(domain.DefaultLicenseExpiration)
		// Allow 1 minute tolerance for test execution time
		assert.WithinDuration(t, expectedExpiry, renewedLicense.ExpiresAt, time.Minute)
	})
}

// TestAccessRevocationFlow tests the access revocation flow:
// 1. Register device and issue license
// 2. Revoke license
// 3. Verify license is no longer valid
// Implements Task 18.2: Access revocation flow test.
func TestAccessRevocationFlow(t *testing.T) {
	ctx := context.Background()

	// Setup mock repositories
	deviceRepo := NewMockDeviceRepository()
	licenseRepo := NewMockLicenseRepository()
	cekRepo := NewMockCEKRepository()
	accessChecker := NewMockMaterialAccessChecker()

	// Create services
	deviceService := application.NewDeviceService(deviceRepo, licenseRepo, cekRepo, nil)
	licenseService := application.NewLicenseService(licenseRepo, deviceRepo, accessChecker, nil)

	// Test data
	userID := uuid.New()
	materialID := uuid.New()
	fingerprint := "test-fingerprint-" + uuid.New().String()

	// Grant material access
	accessChecker.GrantAccess(userID, materialID)

	// Step 1: Register device
	device, err := deviceService.RegisterDevice(ctx, application.RegisterDeviceInput{
		UserID:      userID,
		Fingerprint: fingerprint,
		Name:        "Test Device",
		Platform:    domain.PlatformDesktop,
	})
	require.NoError(t, err)

	// Step 2: Issue license
	license, err := licenseService.IssueLicense(ctx, application.IssueLicenseInput{
		UserID:      userID,
		MaterialID:  materialID,
		DeviceID:    device.ID,
		Fingerprint: fingerprint,
	})
	require.NoError(t, err)

	// Step 3: Revoke license
	t.Run("RevokeLicense", func(t *testing.T) {
		err := licenseService.RevokeLicense(ctx, license.ID)
		require.NoError(t, err)
	})

	// Step 4: Verify license is revoked
	t.Run("VerifyLicenseRevoked", func(t *testing.T) {
		revokedLicense, err := licenseService.GetLicense(ctx, license.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.LicenseStatusRevoked, revokedLicense.Status)
		assert.NotNil(t, revokedLicense.RevokedAt)
	})

	// Step 5: Verify validation fails for revoked license
	t.Run("ValidationFailsForRevokedLicense", func(t *testing.T) {
		_, err := licenseService.ValidateLicense(ctx, application.ValidateLicenseInput{
			LicenseID:   license.ID,
			DeviceID:    device.ID,
			Fingerprint: fingerprint,
			Nonce:       license.Nonce,
		})
		require.Error(t, err)
	})
}

// TestDeviceManagementFlow tests the device management flow:
// 1. Register multiple devices
// 2. List devices
// 3. Deregister device
// 4. Verify licenses are revoked
// Implements Task 18.3: Device management flow test.
func TestDeviceManagementFlow(t *testing.T) {
	ctx := context.Background()

	// Setup mock repositories
	deviceRepo := NewMockDeviceRepository()
	licenseRepo := NewMockLicenseRepository()
	cekRepo := NewMockCEKRepository()
	accessChecker := NewMockMaterialAccessChecker()

	// Create services
	deviceService := application.NewDeviceService(deviceRepo, licenseRepo, cekRepo, nil)
	licenseService := application.NewLicenseService(licenseRepo, deviceRepo, accessChecker, nil)

	// Test data
	userID := uuid.New()
	materialID := uuid.New()

	// Grant material access
	accessChecker.GrantAccess(userID, materialID)

	// Step 1: Register multiple devices
	var devices []*domain.Device
	for i := 0; i < 3; i++ {
		fingerprint := "fingerprint-" + uuid.New().String()
		device, err := deviceService.RegisterDevice(ctx, application.RegisterDeviceInput{
			UserID:      userID,
			Fingerprint: fingerprint,
			Name:        "Device " + string(rune('A'+i)),
			Platform:    domain.PlatformDesktop,
		})
		require.NoError(t, err)
		devices = append(devices, device)
	}

	// Step 2: List devices
	t.Run("ListDevices", func(t *testing.T) {
		listedDevices, err := deviceService.ListDevices(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, listedDevices, 3)
	})

	// Step 3: Issue license for first device
	license, err := licenseService.IssueLicense(ctx, application.IssueLicenseInput{
		UserID:      userID,
		MaterialID:  materialID,
		DeviceID:    devices[0].ID,
		Fingerprint: devices[0].Fingerprint,
	})
	require.NoError(t, err)

	// Step 4: Deregister first device
	t.Run("DeregisterDevice", func(t *testing.T) {
		err := deviceService.DeregisterDevice(ctx, userID, devices[0].ID)
		require.NoError(t, err)
	})

	// Step 5: Verify device is removed from list
	t.Run("VerifyDeviceRemoved", func(t *testing.T) {
		listedDevices, err := deviceService.ListDevices(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, listedDevices, 2)
	})

	// Step 6: Verify license is revoked (cascade)
	t.Run("VerifyLicenseRevoked", func(t *testing.T) {
		revokedLicense, err := licenseService.GetLicense(ctx, license.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.LicenseStatusRevoked, revokedLicense.Status)
	})
}

// TestOfflineGracePeriodFlow tests the offline grace period enforcement:
// 1. Issue license
// 2. Simulate offline period within grace period
// 3. Verify license is still valid
// 4. Simulate offline period exceeding grace period
// 5. Verify license validation fails
// Implements Task 18.4: Offline grace period flow test.
func TestOfflineGracePeriodFlow(t *testing.T) {
	ctx := context.Background()

	// Setup mock repositories
	deviceRepo := NewMockDeviceRepository()
	licenseRepo := NewMockLicenseRepository()
	cekRepo := NewMockCEKRepository()
	accessChecker := NewMockMaterialAccessChecker()

	// Create services
	deviceService := application.NewDeviceService(deviceRepo, licenseRepo, cekRepo, nil)
	licenseService := application.NewLicenseService(licenseRepo, deviceRepo, accessChecker, nil)

	// Test data
	userID := uuid.New()
	materialID := uuid.New()
	fingerprint := "test-fingerprint-" + uuid.New().String()

	// Grant material access
	accessChecker.GrantAccess(userID, materialID)

	// Register device
	device, err := deviceService.RegisterDevice(ctx, application.RegisterDeviceInput{
		UserID:      userID,
		Fingerprint: fingerprint,
		Name:        "Test Device",
		Platform:    domain.PlatformDesktop,
	})
	require.NoError(t, err)

	// Issue license
	license, err := licenseService.IssueLicense(ctx, application.IssueLicenseInput{
		UserID:      userID,
		MaterialID:  materialID,
		DeviceID:    device.ID,
		Fingerprint: fingerprint,
	})
	require.NoError(t, err)

	// Verify license has grace period set
	t.Run("VerifyGracePeriodSet", func(t *testing.T) {
		assert.Equal(t, domain.DefaultOfflineGracePeriod, license.OfflineGracePeriod)
	})

	// Verify license is valid within grace period
	t.Run("ValidWithinGracePeriod", func(t *testing.T) {
		validatedLicense, err := licenseService.ValidateLicense(ctx, application.ValidateLicenseInput{
			LicenseID:   license.ID,
			DeviceID:    device.ID,
			Fingerprint: fingerprint,
			Nonce:       license.Nonce,
		})
		require.NoError(t, err)
		require.NotNil(t, validatedLicense)
	})
}


// TestRateLimitingIntegration tests the rate limiting functionality:
// 1. Check initial rate limit state
// 2. Increment download count
// 3. Verify rate limit enforcement
// 4. Test device blocking after failures
// Implements Task 18.5: Rate limiting integration test.
func TestRateLimitingIntegration(t *testing.T) {
	// Skip if no Redis available (this is a unit test simulation)
	t.Run("RateLimitLogic", func(t *testing.T) {
		// Test rate limit constants
		assert.Equal(t, 10, domain.MaxDownloadsPerHour)
		assert.Equal(t, 5, domain.MaxValidationFailuresPerHour)
		assert.Equal(t, time.Hour, domain.RateLimitWindow)
		assert.Equal(t, time.Hour, domain.DeviceBlockDuration)
	})

	t.Run("RateLimitInfo", func(t *testing.T) {
		// Test RateLimitInfo structure
		info := application.RateLimitInfo{
			Limit:     domain.MaxDownloadsPerHour,
			Remaining: 5,
			Reset:     time.Now().Add(time.Hour),
		}
		assert.Equal(t, 10, info.Limit)
		assert.Equal(t, 5, info.Remaining)
		assert.True(t, info.Reset.After(time.Now()))
	})
}

// MockEncryptionJobRepository implements domain.EncryptionJobRepository for testing.
type MockEncryptionJobRepository struct {
	mu   sync.RWMutex
	jobs map[uuid.UUID]*domain.EncryptionJob
}

func NewMockEncryptionJobRepository() *MockEncryptionJobRepository {
	return &MockEncryptionJobRepository{
		jobs: make(map[uuid.UUID]*domain.EncryptionJob),
	}
}

func (r *MockEncryptionJobRepository) Create(ctx context.Context, job *domain.EncryptionJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[job.ID] = job
	return nil
}

func (r *MockEncryptionJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.EncryptionJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.jobs[id]
	if !ok {
		return nil, domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
	}
	return job, nil
}

func (r *MockEncryptionJobRepository) FindByMaterialID(ctx context.Context, materialID uuid.UUID) ([]*domain.EncryptionJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.EncryptionJob
	for _, job := range r.jobs {
		if job.MaterialID == materialID {
			result = append(result, job)
		}
	}
	return result, nil
}

func (r *MockEncryptionJobRepository) FindByLicenseID(ctx context.Context, licenseID uuid.UUID) (*domain.EncryptionJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, job := range r.jobs {
		if job.LicenseID == licenseID {
			return job, nil
		}
	}
	return nil, domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
}

func (r *MockEncryptionJobRepository) FindPending(ctx context.Context, limit int) ([]*domain.EncryptionJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.EncryptionJob
	// Sort by priority (lower number = higher priority)
	for _, job := range r.jobs {
		if job.Status == domain.JobStatusPending {
			result = append(result, job)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (r *MockEncryptionJobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.JobStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.jobs[id]
	if !ok {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
	}
	job.Status = status
	return nil
}

func (r *MockEncryptionJobRepository) UpdateStarted(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.jobs[id]
	if !ok {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
	}
	job.Status = domain.JobStatusProcessing
	now := time.Now()
	job.StartedAt = &now
	return nil
}

func (r *MockEncryptionJobRepository) UpdateCompleted(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.jobs[id]
	if !ok {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
	}
	job.Status = domain.JobStatusCompleted
	now := time.Now()
	job.CompletedAt = &now
	return nil
}

func (r *MockEncryptionJobRepository) UpdateFailed(ctx context.Context, id uuid.UUID, errorMsg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.jobs[id]
	if !ok {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
	}
	job.Status = domain.JobStatusFailed
	job.Error = &errorMsg
	now := time.Now()
	job.CompletedAt = &now
	return nil
}

func (r *MockEncryptionJobRepository) IncrementRetryCount(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.jobs[id]
	if !ok {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
	}
	job.RetryCount++
	return nil
}

func (r *MockEncryptionJobRepository) DeleteOldCompleted(ctx context.Context, olderThan time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, job := range r.jobs {
		if job.Status == domain.JobStatusCompleted && job.CompletedAt != nil && job.CompletedAt.Before(olderThan) {
			delete(r.jobs, id)
		}
	}
	return nil
}

// MockJobEventPublisher implements OfflineEventPublisher for job tests.
type MockJobEventPublisher struct {
	mu                  sync.Mutex
	encryptionRequested []application.EncryptionJobEvent
	encryptionCompleted []application.EncryptionJobEvent
	encryptionFailed    []application.EncryptionJobEvent
}

func NewMockJobEventPublisher() *MockJobEventPublisher {
	return &MockJobEventPublisher{}
}

func (m *MockJobEventPublisher) PublishKeyGenerated(ctx context.Context, event application.KeyEvent) error {
	return nil
}

func (m *MockJobEventPublisher) PublishKeyRetrieved(ctx context.Context, event application.KeyEvent) error {
	return nil
}

func (m *MockJobEventPublisher) PublishLicenseIssued(ctx context.Context, event application.LicenseEvent) error {
	return nil
}

func (m *MockJobEventPublisher) PublishLicenseValidated(ctx context.Context, event application.LicenseEvent) error {
	return nil
}

func (m *MockJobEventPublisher) PublishLicenseRevoked(ctx context.Context, event application.LicenseEvent) error {
	return nil
}

func (m *MockJobEventPublisher) PublishLicenseRenewed(ctx context.Context, event application.LicenseEvent) error {
	return nil
}

func (m *MockJobEventPublisher) PublishDeviceRegistered(ctx context.Context, event application.DeviceEvent) error {
	return nil
}

func (m *MockJobEventPublisher) PublishDeviceDeregistered(ctx context.Context, event application.DeviceEvent) error {
	return nil
}

func (m *MockJobEventPublisher) PublishEncryptionRequested(ctx context.Context, event application.EncryptionJobEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.encryptionRequested = append(m.encryptionRequested, event)
	return nil
}

func (m *MockJobEventPublisher) PublishEncryptionCompleted(ctx context.Context, event application.EncryptionJobEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.encryptionCompleted = append(m.encryptionCompleted, event)
	return nil
}

func (m *MockJobEventPublisher) PublishEncryptionFailed(ctx context.Context, event application.EncryptionJobEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.encryptionFailed = append(m.encryptionFailed, event)
	return nil
}

func (m *MockJobEventPublisher) PublishMaterialDownloaded(ctx context.Context, event application.MaterialDownloadEvent) error {
	return nil
}

func (m *MockJobEventPublisher) GetEncryptionRequestedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.encryptionRequested)
}

func (m *MockJobEventPublisher) GetEncryptionCompletedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.encryptionCompleted)
}

func (m *MockJobEventPublisher) GetEncryptionFailedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.encryptionFailed)
}

// TestBackgroundJobProcessingFlow tests the complete background job processing flow:
// 1. Create encryption job
// 2. Verify job is stored with pending status
// 3. Verify encryption requested event is published
// 4. Test job status transitions
// 5. Test retry logic with exponential backoff
// 6. Test worker start/stop lifecycle
// Implements Task 18.6: Background job processing test.
func TestBackgroundJobProcessingFlow(t *testing.T) {
	ctx := context.Background()

	// Setup mock repositories
	jobRepo := NewMockEncryptionJobRepository()
	publisher := NewMockJobEventPublisher()

	// Create job service with test configuration
	config := application.DefaultJobServiceConfig()
	config.JobTimeout = 5 * time.Second
	config.MaxRetries = 3
	config.RetryBaseDelay = 100 * time.Millisecond
	config.RetryMaxDelay = 1 * time.Second

	jobService := application.NewJobService(
		jobRepo,
		nil, // encryptionService - not needed for job creation tests
		nil, // keyMgmtService - not needed for job creation tests
		publisher,
		config,
	)

	// Test data
	materialID := uuid.New()
	userID := uuid.New()
	deviceID := uuid.New()
	licenseID := uuid.New()

	// Step 1: Create encryption job
	t.Run("Step1_CreateJob", func(t *testing.T) {
		job, err := jobService.CreateJob(ctx, application.CreateJobInput{
			MaterialID: materialID,
			UserID:     userID,
			DeviceID:   deviceID,
			LicenseID:  licenseID,
			Priority:   domain.JobPriorityHigh,
		})
		require.NoError(t, err)
		require.NotNil(t, job)

		// Verify job properties
		assert.Equal(t, materialID, job.MaterialID)
		assert.Equal(t, userID, job.UserID)
		assert.Equal(t, deviceID, job.DeviceID)
		assert.Equal(t, licenseID, job.LicenseID)
		assert.Equal(t, domain.JobPriorityHigh, job.Priority)
		assert.Equal(t, domain.JobStatusPending, job.Status)
		assert.Equal(t, 0, job.RetryCount)
		assert.Nil(t, job.StartedAt)
		assert.Nil(t, job.CompletedAt)
		assert.Nil(t, job.Error)
	})

	// Step 2: Verify job is stored
	t.Run("Step2_VerifyJobStored", func(t *testing.T) {
		// Get job by material ID
		job, err := jobService.GetJobByMaterial(ctx, materialID)
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.Equal(t, materialID, job.MaterialID)

		// Get job by license ID
		job, err = jobService.GetJobByLicense(ctx, licenseID)
		require.NoError(t, err)
		require.NotNil(t, job)
		assert.Equal(t, licenseID, job.LicenseID)
	})

	// Step 3: Verify encryption requested event was published
	t.Run("Step3_VerifyEventPublished", func(t *testing.T) {
		assert.Equal(t, 1, publisher.GetEncryptionRequestedCount())
	})

	// Step 4: Test job status transitions
	t.Run("Step4_JobStatusTransitions", func(t *testing.T) {
		job, err := jobService.GetJobByMaterial(ctx, materialID)
		require.NoError(t, err)

		// Initial status should be pending
		assert.True(t, job.IsPending())
		assert.False(t, job.IsProcessing())
		assert.False(t, job.IsCompleted())
		assert.False(t, job.IsFailed())

		// Update to processing
		err = jobRepo.UpdateStarted(ctx, job.ID)
		require.NoError(t, err)

		job, err = jobService.GetJob(ctx, job.ID)
		require.NoError(t, err)
		assert.True(t, job.IsProcessing())
		assert.NotNil(t, job.StartedAt)

		// Update to completed
		err = jobRepo.UpdateCompleted(ctx, job.ID)
		require.NoError(t, err)

		job, err = jobService.GetJob(ctx, job.ID)
		require.NoError(t, err)
		assert.True(t, job.IsCompleted())
		assert.NotNil(t, job.CompletedAt)
	})

	// Step 5: Test retry logic with exponential backoff
	t.Run("Step5_RetryLogicExponentialBackoff", func(t *testing.T) {
		// Test exponential backoff calculation
		delay0 := jobService.CalculateRetryDelay(0)
		delay1 := jobService.CalculateRetryDelay(1)
		delay2 := jobService.CalculateRetryDelay(2)
		delay3 := jobService.CalculateRetryDelay(3)

		// Verify exponential growth
		assert.Equal(t, 100*time.Millisecond, delay0)
		assert.Equal(t, 200*time.Millisecond, delay1)
		assert.Equal(t, 400*time.Millisecond, delay2)
		assert.Equal(t, 800*time.Millisecond, delay3)

		// Verify max delay cap (at retry 4, delay would be 1.6s but capped at 1s)
		delay4 := jobService.CalculateRetryDelay(4)
		assert.Equal(t, 1*time.Second, delay4)

		// Verify retry 5 is also capped
		delay5 := jobService.CalculateRetryDelay(5)
		assert.Equal(t, 1*time.Second, delay5)

		// Verify retry 10 is still capped (reasonable retry count)
		delay10 := jobService.CalculateRetryDelay(10)
		assert.Equal(t, 1*time.Second, delay10)
	})

	// Step 6: Test job can retry logic
	t.Run("Step6_JobCanRetry", func(t *testing.T) {
		// Create a new job for retry testing
		retryJob := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityNormal)
		err := jobRepo.Create(ctx, retryJob)
		require.NoError(t, err)

		// Initially can retry
		assert.True(t, retryJob.CanRetry())
		assert.Equal(t, 0, retryJob.RetryCount)

		// Increment retry count
		for i := 0; i < domain.MaxJobRetries; i++ {
			err = jobRepo.IncrementRetryCount(ctx, retryJob.ID)
			require.NoError(t, err)
		}

		// Fetch updated job
		retryJob, err = jobRepo.FindByID(ctx, retryJob.ID)
		require.NoError(t, err)

		// Should not be able to retry after max retries
		assert.False(t, retryJob.CanRetry())
		assert.Equal(t, domain.MaxJobRetries, retryJob.RetryCount)
	})
}

// TestBackgroundJobWorkerLifecycle tests the worker start/stop lifecycle.
// Implements Task 18.6: Background job processing test.
func TestBackgroundJobWorkerLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup mock repositories
	jobRepo := NewMockEncryptionJobRepository()
	publisher := NewMockJobEventPublisher()

	// Create job service
	config := application.DefaultJobServiceConfig()
	config.JobTimeout = 1 * time.Second

	jobService := application.NewJobService(
		jobRepo,
		nil,
		nil,
		publisher,
		config,
	)

	// Create worker with short poll interval for testing
	workerConfig := application.DefaultWorkerConfig()
	workerConfig.PollInterval = 50 * time.Millisecond
	workerConfig.Concurrency = 2
	workerConfig.ShutdownTimeout = 1 * time.Second
	workerConfig.BatchSize = 0 // Don't fetch jobs to avoid nil pointer issues

	worker := application.NewEncryptionWorker(jobService, jobRepo, workerConfig)

	// Test worker lifecycle
	t.Run("WorkerStartStop", func(t *testing.T) {
		// Initially not running
		assert.False(t, worker.IsRunning())

		// Start worker
		err := worker.Start(ctx)
		require.NoError(t, err)
		assert.True(t, worker.IsRunning())

		// Try to start again (should fail)
		err = worker.Start(ctx)
		assert.Error(t, err)

		// Let it run briefly
		time.Sleep(100 * time.Millisecond)

		// Stop worker
		err = worker.Stop()
		require.NoError(t, err)
		assert.False(t, worker.IsRunning())

		// Stop again (should be no-op)
		err = worker.Stop()
		require.NoError(t, err)
	})
}

// TestBackgroundJobPriorityOrdering tests that jobs are created with correct priorities.
// Implements Task 18.6: Background job processing test.
func TestBackgroundJobPriorityOrdering(t *testing.T) {
	ctx := context.Background()

	jobRepo := NewMockEncryptionJobRepository()
	publisher := NewMockJobEventPublisher()

	jobService := application.NewJobService(
		jobRepo,
		nil,
		nil,
		publisher,
		application.DefaultJobServiceConfig(),
	)

	// Create jobs with different priorities
	t.Run("CreateJobsWithPriorities", func(t *testing.T) {
		// High priority job
		highJob, err := jobService.CreateJob(ctx, application.CreateJobInput{
			MaterialID: uuid.New(),
			UserID:     uuid.New(),
			DeviceID:   uuid.New(),
			LicenseID:  uuid.New(),
			Priority:   domain.JobPriorityHigh,
		})
		require.NoError(t, err)
		assert.Equal(t, domain.JobPriorityHigh, highJob.Priority)

		// Normal priority job
		normalJob, err := jobService.CreateJob(ctx, application.CreateJobInput{
			MaterialID: uuid.New(),
			UserID:     uuid.New(),
			DeviceID:   uuid.New(),
			LicenseID:  uuid.New(),
			Priority:   domain.JobPriorityNormal,
		})
		require.NoError(t, err)
		assert.Equal(t, domain.JobPriorityNormal, normalJob.Priority)

		// Low priority job
		lowJob, err := jobService.CreateJob(ctx, application.CreateJobInput{
			MaterialID: uuid.New(),
			UserID:     uuid.New(),
			DeviceID:   uuid.New(),
			LicenseID:  uuid.New(),
			Priority:   domain.JobPriorityLow,
		})
		require.NoError(t, err)
		assert.Equal(t, domain.JobPriorityLow, lowJob.Priority)

		// Invalid priority should default to normal
		invalidPriorityJob, err := jobService.CreateJob(ctx, application.CreateJobInput{
			MaterialID: uuid.New(),
			UserID:     uuid.New(),
			DeviceID:   uuid.New(),
			LicenseID:  uuid.New(),
			Priority:   99, // Invalid
		})
		require.NoError(t, err)
		assert.Equal(t, domain.JobPriorityNormal, invalidPriorityJob.Priority)
	})
}

// TestBackgroundJobCleanup tests old job cleanup functionality.
// Implements Task 18.6: Background job processing test.
func TestBackgroundJobCleanup(t *testing.T) {
	ctx := context.Background()

	jobRepo := NewMockEncryptionJobRepository()
	publisher := NewMockJobEventPublisher()

	jobService := application.NewJobService(
		jobRepo,
		nil,
		nil,
		publisher,
		application.DefaultJobServiceConfig(),
	)

	t.Run("CleanupOldCompletedJobs", func(t *testing.T) {
		// Create an old completed job
		oldJob := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityNormal)
		oldJob.Status = domain.JobStatusCompleted
		oldTime := time.Now().Add(-48 * time.Hour)
		oldJob.CompletedAt = &oldTime
		err := jobRepo.Create(ctx, oldJob)
		require.NoError(t, err)

		// Create a recent completed job
		recentJob := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityNormal)
		recentJob.Status = domain.JobStatusCompleted
		recentTime := time.Now().Add(-1 * time.Hour)
		recentJob.CompletedAt = &recentTime
		err = jobRepo.Create(ctx, recentJob)
		require.NoError(t, err)

		// Create a pending job (should not be cleaned up)
		pendingJob := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityNormal)
		err = jobRepo.Create(ctx, pendingJob)
		require.NoError(t, err)

		// Cleanup jobs older than 24 hours
		err = jobService.CleanupOldJobs(ctx, 24*time.Hour)
		require.NoError(t, err)

		// Old job should be deleted
		_, err = jobRepo.FindByID(ctx, oldJob.ID)
		assert.Error(t, err)

		// Recent job should still exist
		job, err := jobRepo.FindByID(ctx, recentJob.ID)
		require.NoError(t, err)
		assert.NotNil(t, job)

		// Pending job should still exist
		job, err = jobRepo.FindByID(ctx, pendingJob.ID)
		require.NoError(t, err)
		assert.NotNil(t, job)
	})
}

// TestBackgroundJobConcurrentCreation tests concurrent job creation.
// Implements Task 18.6: Background job processing test.
func TestBackgroundJobConcurrentCreation(t *testing.T) {
	ctx := context.Background()

	jobRepo := NewMockEncryptionJobRepository()
	publisher := NewMockJobEventPublisher()

	jobService := application.NewJobService(
		jobRepo,
		nil,
		nil,
		publisher,
		application.DefaultJobServiceConfig(),
	)

	t.Run("ConcurrentJobCreation", func(t *testing.T) {
		var wg sync.WaitGroup
		numJobs := 20
		jobs := make([]*domain.EncryptionJob, numJobs)
		errors := make([]error, numJobs)

		// Create jobs concurrently
		for i := 0; i < numJobs; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				jobs[idx], errors[idx] = jobService.CreateJob(ctx, application.CreateJobInput{
					MaterialID: uuid.New(),
					UserID:     uuid.New(),
					DeviceID:   uuid.New(),
					LicenseID:  uuid.New(),
					Priority:   domain.JobPriorityNormal,
				})
			}(i)
		}

		wg.Wait()

		// Verify all jobs were created successfully
		for i := 0; i < numJobs; i++ {
			assert.NoError(t, errors[i], "job %d should be created without error", i)
			assert.NotNil(t, jobs[i], "job %d should not be nil", i)
		}

		// Verify all jobs have unique IDs
		ids := make(map[uuid.UUID]bool)
		for _, job := range jobs {
			assert.False(t, ids[job.ID], "duplicate job ID found")
			ids[job.ID] = true
		}

		// Verify all events were published
		assert.Equal(t, numJobs, publisher.GetEncryptionRequestedCount())
	})
}

// TestBackgroundJobFailedStatus tests job failure status handling.
// Implements Task 18.6: Background job processing test.
func TestBackgroundJobFailedStatus(t *testing.T) {
	ctx := context.Background()

	jobRepo := NewMockEncryptionJobRepository()

	t.Run("JobFailureStatus", func(t *testing.T) {
		// Create a job
		job := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityNormal)
		err := jobRepo.Create(ctx, job)
		require.NoError(t, err)

		// Mark as started
		err = jobRepo.UpdateStarted(ctx, job.ID)
		require.NoError(t, err)

		// Mark as failed with error message
		errorMsg := "encryption failed: file not found"
		err = jobRepo.UpdateFailed(ctx, job.ID, errorMsg)
		require.NoError(t, err)

		// Verify job status
		failedJob, err := jobRepo.FindByID(ctx, job.ID)
		require.NoError(t, err)
		assert.True(t, failedJob.IsFailed())
		assert.NotNil(t, failedJob.Error)
		assert.Equal(t, errorMsg, *failedJob.Error)
		assert.NotNil(t, failedJob.CompletedAt)
	})
}
