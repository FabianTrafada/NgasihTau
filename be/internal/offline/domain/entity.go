package domain

import (
	"time"

	"github.com/google/uuid"
)

type DeviceStatus string

const (
	DeviceStatusActive  DeviceStatus = "active"
	DeviceStatusRevoked DeviceStatus = "revoked"
)

// LicenseStatus represents the status of a license
type LicenseStatus string

const (
	LicenseStatusActive  LicenseStatus = "active"
	LicenseStatusExpired LicenseStatus = "expired"
	LicenseStatusRevoked LicenseStatus = "revoked"
)

// JobStatus represents the status of an encryption job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// Platform represents the device platform
type Platform string

const (
	PlatformIOS     Platform = "ios"
	PlatformAndroid Platform = "android"
	PlatformDesktop Platform = "desktop"
)

type Device struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Fingerprint string     `json:"fingerprint"`
	Name        string     `json:"name"`
	Platform    Platform   `json:"platform"`
	LastUsedAt  time.Time  `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}

func (d *Device) IsActive() bool {
	return d.RevokedAt == nil
}

func (d *Device) IsRevoked() bool {
	return d.RevokedAt != nil
}

func (d *Device) Status() DeviceStatus {
	if d.IsRevoked() {
		return DeviceStatusRevoked
	}

	return DeviceStatusActive
}

func NewDevice(userID uuid.UUID, fingerprint, name string, platform Platform) *Device {
	now := time.Now()
	return &Device{
		ID:          uuid.New(),
		UserID:      userID,
		Fingerprint: fingerprint,
		Name:        name,
		Platform:    platform,
		LastUsedAt:  now,
		CreatedAt:   now,
	}
}

type License struct {
	ID                 uuid.UUID     `json:"id"`
	UserID             uuid.UUID     `json:"user_id"`
	MaterialID         uuid.UUID     `json:"material_id"`
	DeviceID           uuid.UUID     `json:"device_id"`
	EncryptedCEK       []byte        `json:"-"`
	Status             LicenseStatus `json:"status"`
	ExpiresAt          time.Time     `json:"expires_at"`
	OfflineGracePeriod time.Duration `json:"offline_grace_period"`
	LastValidatedAt    time.Time     `json:"last_validated_at"`
	Nonce              string        `json:"nonce"`
	CreatedAt          time.Time     `json:"created_at"`
	RevokedAt          *time.Time    `json:"revoked_at,omitempty"`
}

const DefaultLicenseExpiration = 30 * 24 * time.Hour

const DefaultOfflineGracePeriod = 72 * time.Hour

func (l *License) IsActive() bool {
	return l.Status == LicenseStatusActive && l.RevokedAt == nil
}

func (l *License) IsExpired() bool {
	return time.Now().After(l.ExpiresAt)
}

func (l *License) IsRevoked() bool {
	return l.RevokedAt != nil || l.Status == LicenseStatusRevoked
}

func (l *License) IsOfflineGraceExpired() bool {
	return time.Since(l.LastValidatedAt) > l.OfflineGracePeriod
}

func (l *License) CanAccess() bool {
	return l.IsActive() && !l.IsExpired() && !l.IsOfflineGraceExpired()
}

func NewLicense(userID, materialID, deviceID uuid.UUID, encryptedCEK []byte, nonce string) *License {
	now := time.Now()
	return &License{
		ID:                 uuid.New(),
		UserID:             userID,
		MaterialID:         materialID,
		DeviceID:           deviceID,
		EncryptedCEK:       encryptedCEK,
		Status:             LicenseStatusActive,
		ExpiresAt:          now.Add(DefaultLicenseExpiration),
		OfflineGracePeriod: DefaultOfflineGracePeriod,
		LastValidatedAt:    now,
		Nonce:              nonce,
		CreatedAt:          now,
	}
}

type ContentEncryptionKey struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	MaterialID   uuid.UUID `json:"material_id"`
	DeviceID     uuid.UUID `json:"device_id"`
	EncryptedKey []byte    `json:"-"` // Encrypted with KEK, never expose
	KeyVersion   int       `json:"key_version"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewContentEncryptionKey(userID, materialID, deviceID uuid.UUID, encryptedKey []byte, keyVersion int) *ContentEncryptionKey {
	return &ContentEncryptionKey{
		ID:           uuid.New(),
		UserID:       userID,
		MaterialID:   materialID,
		DeviceID:     deviceID,
		EncryptedKey: encryptedKey,
		KeyVersion:   keyVersion,
		CreatedAt:    time.Now(),
	}
}

type EncryptedChunk struct {
	Index   int    `json:"index"`
	Offset  int64  `json:"offset"`
	Size    int64  `json:"size"`
	IV      []byte `json:"iv"`
	AuthTag []byte `json:"auth_tag"`
}

type DownloadManifest struct {
	MaterialID    uuid.UUID        `json:"material_id"`
	LicenseID     uuid.UUID        `json:"license_id"`
	TotalChunks   int              `json:"total_chunks"`
	TotalSize     int64            `json:"total_size"`
	OriginalHash  string           `json:"original_hash"`
	EncryptedHash string           `json:"encrypted_hash"`
	ChunkSize     int64            `json:"chunk_size"`
	Chunks        []EncryptedChunk `json:"chunks"`
	FileType      string           `json:"file_type"`
	CreatedAt     time.Time        `json:"created_at"`
}

const DefaultChunkSize int64 = 1024 * 1024 // 1MB

func NewDownloadManifest(materialID, licenseID uuid.UUID, totalSize int64, originalHash, encryptedHash, fileType string, chunks []EncryptedChunk) *DownloadManifest {
	return &DownloadManifest{
		MaterialID:    materialID,
		LicenseID:     licenseID,
		TotalChunks:   len(chunks),
		TotalSize:     totalSize,
		OriginalHash:  originalHash,
		EncryptedHash: encryptedHash,
		ChunkSize:     DefaultChunkSize,
		Chunks:        chunks,
		FileType:      fileType,
		CreatedAt:     time.Now(),
	}
}

type AuditLog struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	DeviceID   *uuid.UUID `json:"device_id,omitempty"`
	Action     string     `json:"action"`
	Resource   string     `json:"resource"`
	ResourceID uuid.UUID  `json:"resource_id"`
	IPAddress  string     `json:"ip_address"`
	UserAgent  string     `json:"user_agent"`
	Success    bool       `json:"success"`
	ErrorCode  *string    `json:"error_code,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

const (
	AuditActionKeyGenerate      = "key.generate"
	AuditActionKeyRetrieve      = "key.retrieve"
	AuditActionLicenseIssue     = "license.issue"
	AuditActionLicenseValidate  = "license.validate"
	AuditActionLicenseRenew     = "license.renew"
	AuditActionLicenseRevoke    = "license.revoke"
	AuditActionDeviceRegister   = "device.register"
	AuditActionDeviceDeregister = "device.deregister"
	AuditActionMaterialDownload = "material.download"
	AuditActionMaterialEncrypt  = "material.encrypt"
)

const (
	AuditResourceCEK      = "cek"
	AuditResourceLicense  = "license"
	AuditResourceDevice   = "device"
	AuditResourceMaterial = "material"
)

func NewAuditLog(userID uuid.UUID, deviceID *uuid.UUID, action, resource string, resourceID uuid.UUID, ipAddress, userAgent string, success bool, errorCode *string) *AuditLog {
	return &AuditLog{
		ID:         uuid.New(),
		UserID:     userID,
		DeviceID:   deviceID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Success:    success,
		ErrorCode:  errorCode,
		CreatedAt:  time.Now(),
	}
}

type EncryptedMaterial struct {
	ID               uuid.UUID        `json:"id"`
	MaterialID       uuid.UUID        `json:"material_id"`
	CEKID            uuid.UUID        `json:"cek_id"`
	Manifest         DownloadManifest `json:"manifest"`
	EncryptedFileURL string           `json:"encrypted_file_url"`
	CreatedAt        time.Time        `json:"created_at"`
}

func NewEncryptedMaterial(materialID, cekID uuid.UUID, manifest DownloadManifest, encryptedFileURL string) *EncryptedMaterial {
	return &EncryptedMaterial{
		ID:               uuid.New(),
		MaterialID:       materialID,
		CEKID:            cekID,
		Manifest:         manifest,
		EncryptedFileURL: encryptedFileURL,
		CreatedAt:        time.Now(),
	}
}

type DeviceRateLimit struct {
	DeviceID       uuid.UUID  `json:"device_id"`
	FailedAttempts int        `json:"failed_attempts"`
	BlockedUntil   *time.Time `json:"blocked_until,omitempty"`
	LastAttemptAt  time.Time  `json:"last_attempt_at"`
}

const MaxFailedAttempts = 5

const DeviceBlockDuration = 1 * time.Hour

func (d *DeviceRateLimit) IsBlocked() bool {
	if d.BlockedUntil == nil {
		return false
	}
	return time.Now().Before(*d.BlockedUntil)
}

func (d *DeviceRateLimit) ShouldBlock() bool {
	return d.FailedAttempts >= MaxFailedAttempts
}

type EncryptionJob struct {
	ID          uuid.UUID  `json:"id"`
	MaterialID  uuid.UUID  `json:"material_id"`
	UserID      uuid.UUID  `json:"user_id"`
	DeviceID    uuid.UUID  `json:"device_id"`
	LicenseID   uuid.UUID  `json:"license_id"`
	Priority    int        `json:"priority"` // 1=high, 2=normal, 3=low
	Status      JobStatus  `json:"status"`
	Error       *string    `json:"error,omitempty"`
	RetryCount  int        `json:"retry_count"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

const (
	JobPriorityHigh   = 1
	JobPriorityNormal = 2
	JobPriorityLow    = 3
)

const MaxJobRetries = 3

func (j *EncryptionJob) IsPending() bool {
	return j.Status == JobStatusPending
}

func (j *EncryptionJob) IsProcessing() bool {
	return j.Status == JobStatusProcessing
}

func (j *EncryptionJob) IsCompleted() bool {
	return j.Status == JobStatusCompleted
}

func (j *EncryptionJob) IsFailed() bool {
	return j.Status == JobStatusFailed
}

func (j *EncryptionJob) CanRetry() bool {
	return j.RetryCount < MaxJobRetries
}

func NewEncryptionJob(materialID, userID, deviceID, licenseID uuid.UUID, priority int) *EncryptionJob {
	return &EncryptionJob{
		ID:         uuid.New(),
		MaterialID: materialID,
		UserID:     userID,
		DeviceID:   deviceID,
		LicenseID:  licenseID,
		Priority:   priority,
		Status:     JobStatusPending,
		RetryCount: 0,
		CreatedAt:  time.Now(),
	}
}

type SyncState struct {
	DeviceID       uuid.UUID `json:"device_id"`
	LastSyncAt     time.Time `json:"last_sync_at"`
	SyncVersion    int       `json:"sync_version"`
	PendingChanges []byte    `json:"pending_changes,omitempty"` // JSONB
}

func NewSyncState(deviceID uuid.UUID) *SyncState {
	return &SyncState{
		DeviceID:    deviceID,
		LastSyncAt:  time.Now(),
		SyncVersion: 1,
	}
}
