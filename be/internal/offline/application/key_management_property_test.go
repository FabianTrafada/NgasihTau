// Package application contains property-based tests for the Key Management Service.
package application

import (
	"bytes"
	"context"
	"crypto/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"ngasihtau/internal/offline/domain"
)

// **Feature: offline-material-backend, Property 1: CEK Derivation Determinism**
//
// *For any* valid combination of master_secret, user_id, material_id, and device_id,
// calling the CEK derivation function multiple times SHALL always produce the same CEK value.
//
// **Validates: Requirements 1.1, 1.3**

func TestProperty1_CEKDerivationDeterminism(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 1.1: Same inputs always produce same CEK
	properties.Property("same inputs always produce same CEK", prop.ForAll(
		func(masterSecretBytes []byte, userIDBytes, materialIDBytes, deviceIDBytes [16]byte) bool {
			// Ensure master secret is at least 32 bytes
			if len(masterSecretBytes) < 32 {
				masterSecretBytes = append(masterSecretBytes, make([]byte, 32-len(masterSecretBytes))...)
			}
			masterSecret := masterSecretBytes[:32]

			kek := make([]byte, 32)
			rand.Read(kek)

			config := KeyManagementConfig{
				MasterSecret:   masterSecret,
				KEK:            kek,
				CurrentVersion: 1,
			}

			userID := uuid.UUID(userIDBytes)
			materialID := uuid.UUID(materialIDBytes)
			deviceID := uuid.UUID(deviceIDBytes)

			// Create two independent services with same config
			svc1 := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
			svc2 := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)

			ctx := context.Background()

			// Generate CEKs from both services
			cek1, err1 := svc1.GetOrCreateCEK(ctx, userID, materialID, deviceID)
			cek2, err2 := svc2.GetOrCreateCEK(ctx, userID, materialID, deviceID)

			if err1 != nil || err2 != nil {
				return false
			}

			// Decrypt both CEKs
			decrypted1, err1 := svc1.DecryptCEK(ctx, cek1)
			decrypted2, err2 := svc2.DecryptCEK(ctx, cek2)

			if err1 != nil || err2 != nil {
				return false
			}

			// The decrypted CEKs should be identical
			return bytes.Equal(decrypted1, decrypted2)
		},
		gen.SliceOfN(32, gen.UInt8()),
		genUUIDBytes(),
		genUUIDBytes(),
		genUUIDBytes(),
	))

	// Property 1.2: Multiple calls with same service produce same CEK
	properties.Property("multiple calls with same service produce same CEK", prop.ForAll(
		func(userIDBytes, materialIDBytes, deviceIDBytes [16]byte) bool {
			masterSecret := make([]byte, 32)
			rand.Read(masterSecret)
			kek := make([]byte, 32)
			rand.Read(kek)

			config := KeyManagementConfig{
				MasterSecret:   masterSecret,
				KEK:            kek,
				CurrentVersion: 1,
			}

			userID := uuid.UUID(userIDBytes)
			materialID := uuid.UUID(materialIDBytes)
			deviceID := uuid.UUID(deviceIDBytes)

			svc := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
			ctx := context.Background()

			// Call multiple times
			cek1, err1 := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
			cek2, err2 := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
			cek3, err3 := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)

			if err1 != nil || err2 != nil || err3 != nil {
				return false
			}

			// All should return the same CEK record
			return cek1.ID == cek2.ID && cek2.ID == cek3.ID
		},
		genUUIDBytes(),
		genUUIDBytes(),
		genUUIDBytes(),
	))

	// Property 1.3: Different inputs produce different CEKs
	properties.Property("different inputs produce different CEKs", prop.ForAll(
		func(userIDBytes1, userIDBytes2, materialIDBytes, deviceIDBytes [16]byte) bool {
			// Skip if user IDs are the same
			if userIDBytes1 == userIDBytes2 {
				return true
			}

			masterSecret := make([]byte, 32)
			rand.Read(masterSecret)
			kek := make([]byte, 32)
			rand.Read(kek)

			config := KeyManagementConfig{
				MasterSecret:   masterSecret,
				KEK:            kek,
				CurrentVersion: 1,
			}

			userID1 := uuid.UUID(userIDBytes1)
			userID2 := uuid.UUID(userIDBytes2)
			materialID := uuid.UUID(materialIDBytes)
			deviceID := uuid.UUID(deviceIDBytes)

			svc := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
			ctx := context.Background()

			cek1, err1 := svc.GetOrCreateCEK(ctx, userID1, materialID, deviceID)
			cek2, err2 := svc.GetOrCreateCEK(ctx, userID2, materialID, deviceID)

			if err1 != nil || err2 != nil {
				return false
			}

			// Decrypt both
			decrypted1, err1 := svc.DecryptCEK(ctx, cek1)
			decrypted2, err2 := svc.DecryptCEK(ctx, cek2)

			if err1 != nil || err2 != nil {
				return false
			}

			// Different user IDs should produce different CEKs
			return !bytes.Equal(decrypted1, decrypted2)
		},
		genUUIDBytes(),
		genUUIDBytes(),
		genUUIDBytes(),
		genUUIDBytes(),
	))

	properties.TestingRun(t)
}

// **Feature: offline-material-backend, Property 2: CEK Storage Encryption Round-Trip**
//
// *For any* CEK value, encrypting it with the KEK and then decrypting SHALL produce
// the original CEK value, and the encrypted form SHALL differ from the plaintext.
//
// **Validates: Requirements 1.2**

func TestProperty2_CEKStorageEncryptionRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 2.1: Encryption round-trip preserves CEK value
	properties.Property("encryption round-trip preserves CEK value", prop.ForAll(
		func(userIDBytes, materialIDBytes, deviceIDBytes [16]byte) bool {
			masterSecret := make([]byte, 32)
			rand.Read(masterSecret)
			kek := make([]byte, 32)
			rand.Read(kek)

			config := KeyManagementConfig{
				MasterSecret:   masterSecret,
				KEK:            kek,
				CurrentVersion: 1,
			}

			userID := uuid.UUID(userIDBytes)
			materialID := uuid.UUID(materialIDBytes)
			deviceID := uuid.UUID(deviceIDBytes)

			svc := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
			ctx := context.Background()

			// Create CEK (which encrypts it for storage)
			cekRecord, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
			if err != nil {
				return false
			}

			// Decrypt the stored CEK
			decryptedCEK, err := svc.DecryptCEK(ctx, cekRecord)
			if err != nil {
				return false
			}

			// CEK should be exactly 32 bytes
			if len(decryptedCEK) != domain.CEKSize {
				return false
			}

			// Re-derive the CEK to verify it matches
			svc2 := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
			cekRecord2, err := svc2.GetOrCreateCEK(ctx, userID, materialID, deviceID)
			if err != nil {
				return false
			}

			decryptedCEK2, err := svc2.DecryptCEK(ctx, cekRecord2)
			if err != nil {
				return false
			}

			// Both decrypted CEKs should be identical
			return bytes.Equal(decryptedCEK, decryptedCEK2)
		},
		genUUIDBytes(),
		genUUIDBytes(),
		genUUIDBytes(),
	))

	// Property 2.2: Encrypted form differs from plaintext
	properties.Property("encrypted form differs from plaintext", prop.ForAll(
		func(userIDBytes, materialIDBytes, deviceIDBytes [16]byte) bool {
			masterSecret := make([]byte, 32)
			rand.Read(masterSecret)
			kek := make([]byte, 32)
			rand.Read(kek)

			config := KeyManagementConfig{
				MasterSecret:   masterSecret,
				KEK:            kek,
				CurrentVersion: 1,
			}

			userID := uuid.UUID(userIDBytes)
			materialID := uuid.UUID(materialIDBytes)
			deviceID := uuid.UUID(deviceIDBytes)

			svc := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
			ctx := context.Background()

			cekRecord, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
			if err != nil {
				return false
			}

			decryptedCEK, err := svc.DecryptCEK(ctx, cekRecord)
			if err != nil {
				return false
			}

			// Encrypted key should differ from decrypted key
			return !bytes.Equal(cekRecord.EncryptedKey, decryptedCEK)
		},
		genUUIDBytes(),
		genUUIDBytes(),
		genUUIDBytes(),
	))

	// Property 2.3: Encrypted form is longer than plaintext (includes nonce)
	properties.Property("encrypted form is longer than plaintext", prop.ForAll(
		func(userIDBytes, materialIDBytes, deviceIDBytes [16]byte) bool {
			masterSecret := make([]byte, 32)
			rand.Read(masterSecret)
			kek := make([]byte, 32)
			rand.Read(kek)

			config := KeyManagementConfig{
				MasterSecret:   masterSecret,
				KEK:            kek,
				CurrentVersion: 1,
			}

			userID := uuid.UUID(userIDBytes)
			materialID := uuid.UUID(materialIDBytes)
			deviceID := uuid.UUID(deviceIDBytes)

			svc := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
			ctx := context.Background()

			cekRecord, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
			if err != nil {
				return false
			}

			// Encrypted key should be longer than CEK size (includes nonce + auth tag)
			return len(cekRecord.EncryptedKey) > domain.CEKSize
		},
		genUUIDBytes(),
		genUUIDBytes(),
		genUUIDBytes(),
	))

	properties.TestingRun(t)
}

// **Feature: offline-material-backend, Property 3: Key Version Isolation**
//
// *For any* key rotation operation, CEKs created with version N SHALL remain decryptable
// with version N's KEK, while new CEKs SHALL use version N+1.
//
// **Validates: Requirements 1.4**

func TestProperty3_KeyVersionIsolation(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 3.1: CEKs created with version N use version N
	properties.Property("CEKs created with version N use version N", prop.ForAll(
		func(version int, userIDBytes, materialIDBytes, deviceIDBytes [16]byte) bool {
			// Ensure version is positive
			if version <= 0 {
				version = 1
			}
			if version > 100 {
				version = 100
			}

			masterSecret := make([]byte, 32)
			rand.Read(masterSecret)
			kek := make([]byte, 32)
			rand.Read(kek)

			config := KeyManagementConfig{
				MasterSecret:   masterSecret,
				KEK:            kek,
				CurrentVersion: version,
			}

			userID := uuid.UUID(userIDBytes)
			materialID := uuid.UUID(materialIDBytes)
			deviceID := uuid.UUID(deviceIDBytes)

			svc := NewKeyManagementService(newMockCEKRepository(), nil, nil, config)
			ctx := context.Background()

			cekRecord, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
			if err != nil {
				return false
			}

			// CEK should have the configured version
			return cekRecord.KeyVersion == version
		},
		gen.IntRange(1, 100),
		genUUIDBytes(),
		genUUIDBytes(),
		genUUIDBytes(),
	))

	// Property 3.2: Key rotation updates version correctly
	properties.Property("key rotation updates version correctly", prop.ForAll(
		func(fromVersion int, userIDBytes, materialIDBytes, deviceIDBytes [16]byte) bool {
			// Ensure version is positive
			if fromVersion <= 0 {
				fromVersion = 1
			}
			if fromVersion > 50 {
				fromVersion = 50
			}
			toVersion := fromVersion + 1

			masterSecret := make([]byte, 32)
			rand.Read(masterSecret)
			kek := make([]byte, 32)
			rand.Read(kek)
			newKEK := make([]byte, 32)
			rand.Read(newKEK)

			config := KeyManagementConfig{
				MasterSecret:   masterSecret,
				KEK:            kek,
				CurrentVersion: fromVersion,
			}

			userID := uuid.UUID(userIDBytes)
			materialID := uuid.UUID(materialIDBytes)
			deviceID := uuid.UUID(deviceIDBytes)

			cekRepo := newMockCEKRepository()
			svc := NewKeyManagementService(cekRepo, nil, nil, config)
			ctx := context.Background()

			// Create CEK with fromVersion
			cekRecord, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
			if err != nil {
				return false
			}

			// Verify initial version
			if cekRecord.KeyVersion != fromVersion {
				return false
			}

			// Rotate keys
			err = svc.RotateKeys(ctx, fromVersion, toVersion, newKEK)
			if err != nil {
				return false
			}

			// Verify CEK was updated to new version
			rotatedCEKs, err := cekRepo.FindByKeyVersion(ctx, toVersion)
			if err != nil {
				return false
			}

			// Should have exactly one CEK with new version
			return len(rotatedCEKs) == 1 && rotatedCEKs[0].KeyVersion == toVersion
		},
		gen.IntRange(1, 50),
		genUUIDBytes(),
		genUUIDBytes(),
		genUUIDBytes(),
	))

	// Property 3.3: Rotated CEKs remain decryptable with new KEK
	properties.Property("rotated CEKs remain decryptable with new KEK", prop.ForAll(
		func(userIDBytes, materialIDBytes, deviceIDBytes [16]byte) bool {
			masterSecret := make([]byte, 32)
			rand.Read(masterSecret)
			kek := make([]byte, 32)
			rand.Read(kek)
			newKEK := make([]byte, 32)
			rand.Read(newKEK)

			config := KeyManagementConfig{
				MasterSecret:   masterSecret,
				KEK:            kek,
				CurrentVersion: 1,
			}

			userID := uuid.UUID(userIDBytes)
			materialID := uuid.UUID(materialIDBytes)
			deviceID := uuid.UUID(deviceIDBytes)

			cekRepo := newMockCEKRepository()
			svc := NewKeyManagementService(cekRepo, nil, nil, config)
			ctx := context.Background()

			// Create CEK with version 1
			cekRecord, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
			if err != nil {
				return false
			}

			// Decrypt original CEK
			originalDecrypted, err := svc.DecryptCEK(ctx, cekRecord)
			if err != nil {
				return false
			}

			// Rotate keys
			err = svc.RotateKeys(ctx, 1, 2, newKEK)
			if err != nil {
				return false
			}

			// Create new service with new KEK
			newConfig := KeyManagementConfig{
				MasterSecret:   masterSecret,
				KEK:            newKEK,
				CurrentVersion: 2,
			}
			newSvc := NewKeyManagementService(cekRepo, nil, nil, newConfig)

			// Get the rotated CEK
			rotatedCEKs, err := cekRepo.FindByKeyVersion(ctx, 2)
			if err != nil || len(rotatedCEKs) == 0 {
				return false
			}

			// Decrypt with new KEK
			rotatedDecrypted, err := newSvc.DecryptCEK(ctx, rotatedCEKs[0])
			if err != nil {
				return false
			}

			// The decrypted CEK should be the same as original
			return bytes.Equal(originalDecrypted, rotatedDecrypted)
		},
		genUUIDBytes(),
		genUUIDBytes(),
		genUUIDBytes(),
	))

	// Property 3.4: Old version CEKs are removed after rotation
	properties.Property("old version CEKs are removed after rotation", prop.ForAll(
		func(numCEKs int) bool {
			// Limit number of CEKs
			if numCEKs <= 0 {
				numCEKs = 1
			}
			if numCEKs > 10 {
				numCEKs = 10
			}

			masterSecret := make([]byte, 32)
			rand.Read(masterSecret)
			kek := make([]byte, 32)
			rand.Read(kek)
			newKEK := make([]byte, 32)
			rand.Read(newKEK)

			config := KeyManagementConfig{
				MasterSecret:   masterSecret,
				KEK:            kek,
				CurrentVersion: 1,
			}

			cekRepo := newMockCEKRepository()
			svc := NewKeyManagementService(cekRepo, nil, nil, config)
			ctx := context.Background()

			// Create multiple CEKs with version 1
			userID := uuid.New()
			for i := 0; i < numCEKs; i++ {
				materialID := uuid.New()
				deviceID := uuid.New()
				_, err := svc.GetOrCreateCEK(ctx, userID, materialID, deviceID)
				if err != nil {
					return false
				}
			}

			// Verify we have numCEKs with version 1
			v1CEKs, err := cekRepo.FindByKeyVersion(ctx, 1)
			if err != nil || len(v1CEKs) != numCEKs {
				return false
			}

			// Rotate keys
			err = svc.RotateKeys(ctx, 1, 2, newKEK)
			if err != nil {
				return false
			}

			// Version 1 should have no CEKs
			v1CEKsAfter, err := cekRepo.FindByKeyVersion(ctx, 1)
			if err != nil {
				return false
			}

			// Version 2 should have all CEKs
			v2CEKs, err := cekRepo.FindByKeyVersion(ctx, 2)
			if err != nil {
				return false
			}

			return len(v1CEKsAfter) == 0 && len(v2CEKs) == numCEKs
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// genUUIDBytes generates random UUID bytes for property testing.
func genUUIDBytes() gopter.Gen {
	return gen.SliceOfN(16, gen.UInt8()).Map(func(b []byte) [16]byte {
		var arr [16]byte
		copy(arr[:], b)
		return arr
	})
}
