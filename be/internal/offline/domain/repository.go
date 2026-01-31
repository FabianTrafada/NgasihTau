package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DeviceRepository interface {
	Create(ctx context.Context, device *Device) error
	FindById(ctx context.Context, id uuid.UUID) (*Device, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*Device, error)
	FindByFingerprint(ctx context.Context, userID uuid.UUID, fingerprint string) (*Device, error)
	CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int, error)
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
}

type LicenseRepository interface {
	Create(ctx context.Context, license *License) error
	FindByID(ctx context.Context, id uuid.UUID) (*License, error)
	FindByUserAndMaterial(ctx context.Context, userID, materialID, deviceID uuid.UUID) (*License, error)
	FindActiveByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]*License, error)
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*License, error)
	UpdateValidation(ctx context.Context, id uuid.UUID, nonce string) error
	UpdateExpiration(ctx context.Context, id uuid.UUID, expiresAt time.Time) error
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeByDeviceID(ctx context.Context, deviceID uuid.UUID) error
	RevokeByMaterialID(ctx context.Context, materialID uuid.UUID) error
	RevokeByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) error
}

type CEKRepository interface {
	Create(ctx context.Context, cek *ContentEncryptionKey) error
	FindByID(ctx context.Context, id uuid.UUID) (*ContentEncryptionKey, error)
	FindByComposite(ctx context.Context, userID, materialID, deviceID uuid.UUID) (*ContentEncryptionKey, error)
	DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error
	DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error
	FindByKeyVersion(ctx context.Context, keyVersion int) ([]*ContentEncryptionKey, error)
	UpdateKeyVersion(ctx context.Context, id uuid.UUID, encryptedKey []byte, keyVersion int) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AuditLog, int, error)
	FindByDeviceID(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*AuditLog, int, error)
	FindByAction(ctx context.Context, action string, limit, offset int) ([]*AuditLog, int, error)
	CountFailedByDevice(ctx context.Context, deviceID uuid.UUID, since time.Time) (int, error)
	CountFailedByDeviceAndAction(ctx context.Context, deviceID uuid.UUID, action string, since time.Time) (int, error)
}

type EncryptedMaterialRepository interface {
	Create(ctx context.Context, material *EncryptedMaterial) error
	FindById(ctx context.Context, id uuid.UUID) (*EncryptedMaterial, error)
	FindByMaterialAndCEK(ctx context.Context, materialID, cekID uuid.UUID) (*EncryptedMaterial, error)
	FindByMaterialID(ctx context.Context, materialID uuid.UUID) ([]*EncryptedMaterial, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error
}

type DeviceRateLimitRepository interface {
	FindByDeviceID(ctx context.Context, deviceID uuid.UUID) (*DeviceRateLimit, error)
	Upsert(ctx context.Context, rateLimit *DeviceRateLimit) error
	IncrementFailedAttempts(ctx context.Context, deviceID uuid.UUID) error
	ResetFailedAttempts(ctx context.Context, deviceID uuid.UUID) error
	SetBlocked(ctx context.Context, deviceID uuid.UUID, blockedUntil time.Time) error
	ClearBlock(ctx context.Context, deviceID uuid.UUID) error
	Delete(ctx context.Context, deviceID uuid.UUID) error
}
