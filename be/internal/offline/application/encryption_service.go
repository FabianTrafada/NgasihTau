// Package application contains the business logic and use cases for the Offline Material Service.
package application

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/offline/domain"
)

// MinIOStorageClient defines the interface for MinIO storage operations.
type MinIOStorageClient interface {
	// GetObject retrieves an object from storage.
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error)
	// PutObject stores an object in storage.
	PutObject(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error
	// DeleteObject deletes an object from storage.
	DeleteObject(ctx context.Context, objectKey string) error
	// DeleteObjects deletes multiple objects from storage.
	DeleteObjects(ctx context.Context, objectKeys []string) error
	// GetObjectInfo returns metadata about an object.
	GetObjectInfo(ctx context.Context, objectKey string) (*ObjectInfo, error)
	// GeneratePresignedGetURL generates a presigned URL for downloading.
	GeneratePresignedGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
}

// ObjectInfo represents metadata about a stored object.
type ObjectInfo struct {
	Size        int64
	ContentType string
	ETag        string
}

// MaterialAccessChecker defines the interface for checking material access.
type MaterialAccessChecker interface {
	// GetMaterialFileKey returns the storage key for a material's file.
	GetMaterialFileKey(ctx context.Context, materialID uuid.UUID) (string, string, error) // returns objectKey, fileType, error
}

// EncryptionService handles material encryption operations.
// Implements Requirement 2: Encryption Service.
type EncryptionService struct {
	storage               MinIOStorageClient
	materialChecker       MaterialAccessChecker
	encryptedMaterialRepo domain.EncryptedMaterialRepository
	eventPublisher        OfflineEventPublisher
	encryptedBucket       string
}

// EncryptionServiceConfig holds configuration for the Encryption Service.
type EncryptionServiceConfig struct {
	EncryptedBucket string
}

// NewEncryptionService creates a new Encryption Service.
func NewEncryptionService(
	storage MinIOStorageClient,
	materialChecker MaterialAccessChecker,
	encryptedMaterialRepo domain.EncryptedMaterialRepository,
	eventPublisher OfflineEventPublisher,
	config EncryptionServiceConfig,
) *EncryptionService {
	return &EncryptionService{
		storage:               storage,
		materialChecker:       materialChecker,
		encryptedMaterialRepo: encryptedMaterialRepo,
		eventPublisher:        eventPublisher,
		encryptedBucket:       config.EncryptedBucket,
	}
}

// EncryptMaterialInput contains input for encrypting a material.
type EncryptMaterialInput struct {
	MaterialID uuid.UUID
	LicenseID  uuid.UUID
	CEKID      uuid.UUID
	CEK        []byte // Decrypted CEK (32 bytes for AES-256)
	UserID     uuid.UUID
	DeviceID   uuid.UUID
}

// EncryptMaterialOutput contains the result of encrypting a material.
type EncryptMaterialOutput struct {
	Manifest         *domain.DownloadManifest
	EncryptedFileURL string
}

// EncryptMaterial encrypts a material file and returns the manifest.
// Implements Requirement 2.1-2.7: Material encryption with AES-256-GCM.
func (s *EncryptionService) EncryptMaterial(ctx context.Context, input EncryptMaterialInput) (*EncryptMaterialOutput, error) {
	// Validate CEK size
	if len(input.CEK) != domain.CEKSize {
		return nil, domain.NewOfflineError(domain.ErrCodeInvalidKey,
			fmt.Sprintf("invalid CEK size: expected %d, got %d", domain.CEKSize, len(input.CEK)))
	}

	// Get material file info
	objectKey, fileType, err := s.materialChecker.GetMaterialFileKey(ctx, input.MaterialID)
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeMaterialNotFound, "failed to get material file key", err)
	}

	// Validate file type
	if !domain.IsSupportedFileType(fileType) {
		return nil, domain.NewOfflineError(domain.ErrCodeUnsupportedFileType,
			fmt.Sprintf("file type '%s' is not supported for offline encryption", fileType))
	}

	// Get file info for size
	fileInfo, err := s.storage.GetObjectInfo(ctx, objectKey)
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to get file info", err)
	}

	// Download original file
	reader, err := s.storage.GetObject(ctx, objectKey)
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to download original file", err)
	}
	defer reader.Close()

	// Read entire file into memory (for chunking)
	originalContent, err := io.ReadAll(reader)
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to read file content", err)
	}

	// Calculate original hash
	originalHash := calculateSHA256(originalContent)

	// Generate base seed for IV derivation
	baseSeed, err := generateBaseSeed()
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeEncryptionFailed, "failed to generate base seed", err)
	}

	// Encrypt file in chunks
	encryptedContent, chunks, err := s.encryptFileInChunks(originalContent, input.CEK, baseSeed)
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeEncryptionFailed, "failed to encrypt file", err)
	}

	// Calculate encrypted hash
	encryptedHash := calculateSHA256(encryptedContent)

	// Generate encrypted file key
	encryptedFileKey := generateEncryptedFileKey(input.MaterialID, input.CEKID)

	// Store encrypted file
	err = s.storage.PutObject(ctx, encryptedFileKey, bytes.NewReader(encryptedContent), int64(len(encryptedContent)), "application/octet-stream")
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to store encrypted file", err)
	}

	// Create manifest
	manifest := domain.NewDownloadManifest(
		input.MaterialID,
		input.LicenseID,
		fileInfo.Size,
		originalHash,
		encryptedHash,
		fileType,
		chunks,
	)

	// Store encrypted material record
	encryptedMaterial := domain.NewEncryptedMaterial(input.MaterialID, input.CEKID, *manifest, encryptedFileKey)
	if err := s.encryptedMaterialRepo.Create(ctx, encryptedMaterial); err != nil {
		// Cleanup: delete the encrypted file we just uploaded
		if delErr := s.storage.DeleteObject(ctx, encryptedFileKey); delErr != nil {
			log.Error().Err(delErr).Str("key", encryptedFileKey).Msg("failed to cleanup encrypted file after db error")
		}
		return nil, domain.WrapOfflineError(domain.ErrCodeDatabaseError, "failed to store encrypted material record", err)
	}

	// Generate presigned URL for download
	presignedURL, err := s.storage.GeneratePresignedGetURL(ctx, encryptedFileKey, 1*time.Hour)
	if err != nil {
		log.Warn().Err(err).Msg("failed to generate presigned URL, returning empty")
		presignedURL = ""
	}

	log.Info().
		Str("material_id", input.MaterialID.String()).
		Str("license_id", input.LicenseID.String()).
		Int("total_chunks", len(chunks)).
		Int64("original_size", fileInfo.Size).
		Int("encrypted_size", len(encryptedContent)).
		Msg("material encrypted successfully")

	return &EncryptMaterialOutput{
		Manifest:         manifest,
		EncryptedFileURL: presignedURL,
	}, nil
}

// encryptFileInChunks encrypts file content in 1MB chunks using AES-256-GCM.
// Implements Requirement 2.1, 2.2, 2.3, 2.4: Chunk-based encryption with unique IVs.
func (s *EncryptionService) encryptFileInChunks(content, cek, baseSeed []byte) ([]byte, []domain.EncryptedChunk, error) {
	chunkSize := int(domain.DefaultChunkSize)
	totalChunks := CalculateChunkCount(int64(len(content)))

	var encryptedContent bytes.Buffer
	chunks := make([]domain.EncryptedChunk, 0, totalChunks)

	// Create AES cipher
	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	var offset int64 = 0
	for i := 0; i < totalChunks; i++ {
		// Calculate chunk boundaries
		start := i * chunkSize
		end := start + chunkSize
		if end > len(content) {
			end = len(content)
		}
		chunkData := content[start:end]

		// Derive unique IV for this chunk
		iv := DeriveIVForChunk(baseSeed, i)

		// Encrypt chunk with AES-256-GCM
		// GCM Seal appends the auth tag to the ciphertext
		ciphertext := gcm.Seal(nil, iv, chunkData, nil)

		// Extract auth tag (last 16 bytes of ciphertext)
		authTagStart := len(ciphertext) - gcm.Overhead()
		authTag := ciphertext[authTagStart:]
		ciphertextOnly := ciphertext[:authTagStart]

		// Write to output: IV + Ciphertext + AuthTag
		encryptedChunkData := make([]byte, 0, len(iv)+len(ciphertext))
		encryptedChunkData = append(encryptedChunkData, iv...)
		encryptedChunkData = append(encryptedChunkData, ciphertextOnly...)
		encryptedChunkData = append(encryptedChunkData, authTag...)

		encryptedContent.Write(encryptedChunkData)

		// Record chunk metadata
		chunk := domain.EncryptedChunk{
			Index:   i,
			Offset:  offset,
			Size:    int64(len(encryptedChunkData)),
			IV:      iv,
			AuthTag: authTag,
		}
		chunks = append(chunks, chunk)

		offset += int64(len(encryptedChunkData))
	}

	return encryptedContent.Bytes(), chunks, nil
}

// DecryptChunk decrypts a single encrypted chunk.
// This is primarily used for testing the encryption round-trip.
func DecryptChunk(encryptedChunk []byte, cek []byte, iv []byte) ([]byte, error) {
	if len(cek) != domain.CEKSize {
		return nil, fmt.Errorf("invalid CEK size: expected %d, got %d", domain.CEKSize, len(cek))
	}

	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// The encrypted chunk format is: IV (12) + Ciphertext + AuthTag (16)
	// But we receive the ciphertext+authtag portion (IV is passed separately)
	// For GCM.Open, we need ciphertext with auth tag appended

	plaintext, err := gcm.Open(nil, iv, encryptedChunk, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// DecryptFile decrypts an entire encrypted file using the manifest.
func DecryptFile(encryptedContent []byte, cek []byte, chunks []domain.EncryptedChunk) ([]byte, error) {
	if len(cek) != domain.CEKSize {
		return nil, fmt.Errorf("invalid CEK size: expected %d, got %d", domain.CEKSize, len(cek))
	}

	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	var decryptedContent bytes.Buffer

	for _, chunk := range chunks {
		// Extract chunk data from encrypted content
		chunkStart := chunk.Offset
		chunkEnd := chunkStart + chunk.Size
		if chunkEnd > int64(len(encryptedContent)) {
			return nil, fmt.Errorf("chunk %d extends beyond encrypted content", chunk.Index)
		}

		encryptedChunkData := encryptedContent[chunkStart:chunkEnd]

		// Parse chunk: IV (12) + Ciphertext + AuthTag (16)
		if len(encryptedChunkData) < domain.IVSize+domain.AuthTagSize {
			return nil, fmt.Errorf("chunk %d is too small", chunk.Index)
		}

		iv := encryptedChunkData[:domain.IVSize]
		ciphertextWithTag := encryptedChunkData[domain.IVSize:]

		// Decrypt
		plaintext, err := gcm.Open(nil, iv, ciphertextWithTag, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt chunk %d: %w", chunk.Index, err)
		}

		decryptedContent.Write(plaintext)
	}

	return decryptedContent.Bytes(), nil
}

// GetEncryptedMaterial retrieves an existing encrypted material.
func (s *EncryptionService) GetEncryptedMaterial(ctx context.Context, materialID, cekID uuid.UUID) (*domain.EncryptedMaterial, error) {
	material, err := s.encryptedMaterialRepo.FindByMaterialAndCEK(ctx, materialID, cekID)
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeMaterialNotFound, "encrypted material not found", err)
	}
	return material, nil
}

// GetEncryptedFile returns a reader for the encrypted file.
func (s *EncryptionService) GetEncryptedFile(ctx context.Context, materialID uuid.UUID) (io.ReadCloser, error) {
	materials, err := s.encryptedMaterialRepo.FindByMaterialID(ctx, materialID)
	if err != nil || len(materials) == 0 {
		return nil, domain.NewOfflineError(domain.ErrCodeMaterialNotFound, "no encrypted material found")
	}

	// Use the first available encrypted material
	material := materials[0]

	reader, err := s.storage.GetObject(ctx, material.EncryptedFileURL)
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to get encrypted file", err)
	}

	return reader, nil
}

// GetEncryptedChunk returns a specific encrypted chunk.
func (s *EncryptionService) GetEncryptedChunk(ctx context.Context, materialID uuid.UUID, chunkIndex int) ([]byte, error) {
	materials, err := s.encryptedMaterialRepo.FindByMaterialID(ctx, materialID)
	if err != nil || len(materials) == 0 {
		return nil, domain.NewOfflineError(domain.ErrCodeMaterialNotFound, "no encrypted material found")
	}

	material := materials[0]
	manifest := material.Manifest

	if chunkIndex < 0 || chunkIndex >= len(manifest.Chunks) {
		return nil, domain.NewOfflineError(domain.ErrCodeInvalidRequest,
			fmt.Sprintf("invalid chunk index: %d (total chunks: %d)", chunkIndex, len(manifest.Chunks)))
	}

	chunk := manifest.Chunks[chunkIndex]

	// Get the encrypted file
	reader, err := s.storage.GetObject(ctx, material.EncryptedFileURL)
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to get encrypted file", err)
	}
	defer reader.Close()

	// Read entire file (in production, we'd use range requests)
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to read encrypted file", err)
	}

	// Extract chunk
	if chunk.Offset+chunk.Size > int64(len(content)) {
		return nil, domain.NewOfflineError(domain.ErrCodeInternalError, "chunk extends beyond file")
	}

	return content[chunk.Offset : chunk.Offset+chunk.Size], nil
}

// CleanupEncryptedMaterial removes encrypted material and its file.
// Implements Requirement 2.7: Cleanup on failure.
func (s *EncryptionService) CleanupEncryptedMaterial(ctx context.Context, materialID, cekID uuid.UUID) error {
	material, err := s.encryptedMaterialRepo.FindByMaterialAndCEK(ctx, materialID, cekID)
	if err != nil {
		// Material doesn't exist, nothing to clean up
		return nil
	}

	// Delete the encrypted file from storage
	if err := s.storage.DeleteObject(ctx, material.EncryptedFileURL); err != nil {
		log.Warn().Err(err).Str("key", material.EncryptedFileURL).Msg("failed to delete encrypted file during cleanup")
	}

	// Delete the database record
	if err := s.encryptedMaterialRepo.Delete(ctx, material.ID); err != nil {
		return domain.WrapOfflineError(domain.ErrCodeDatabaseError, "failed to delete encrypted material record", err)
	}

	log.Info().
		Str("material_id", materialID.String()).
		Str("cek_id", cekID.String()).
		Msg("encrypted material cleaned up")

	return nil
}

// CleanupPartialEncryption cleans up any partial encryption results.
// Called when encryption fails partway through.
func (s *EncryptionService) CleanupPartialEncryption(ctx context.Context, encryptedFileKey string) error {
	if encryptedFileKey == "" {
		return nil
	}

	if err := s.storage.DeleteObject(ctx, encryptedFileKey); err != nil {
		log.Warn().Err(err).Str("key", encryptedFileKey).Msg("failed to cleanup partial encryption")
		return err
	}

	log.Debug().Str("key", encryptedFileKey).Msg("cleaned up partial encryption")
	return nil
}

// CalculateChunkCount calculates the number of chunks for a file of given size.
// Implements Requirement 2.2: File chunking (1MB chunks).
func CalculateChunkCount(fileSize int64) int {
	if fileSize <= 0 {
		return 0
	}
	chunkSize := domain.DefaultChunkSize
	count := fileSize / chunkSize
	if fileSize%chunkSize != 0 {
		count++
	}
	return int(count)
}

// generateBaseSeed generates a random base seed for IV derivation.
func generateBaseSeed() ([]byte, error) {
	seed := make([]byte, domain.IVSize)
	if _, err := io.ReadFull(rand.Reader, seed); err != nil {
		return nil, fmt.Errorf("failed to generate random seed: %w", err)
	}
	return seed, nil
}

// calculateSHA256 calculates the SHA-256 hash of content and returns it as a hex string.
func calculateSHA256(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}

// generateEncryptedFileKey generates the storage key for an encrypted file.
func generateEncryptedFileKey(materialID, cekID uuid.UUID) string {
	return fmt.Sprintf("encrypted/%s/%s.enc", materialID.String(), cekID.String())
}

// GetFileTypeFromPath extracts the file type from a file path.
func GetFileTypeFromPath(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return ""
	}
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}

// GetFileTypeFromContentType extracts the file type from a content type.
func GetFileTypeFromContentType(contentType string) string {
	switch contentType {
	case "application/pdf":
		return "pdf"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "docx"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return "pptx"
	case "application/msword":
		return "doc"
	case "application/vnd.ms-powerpoint":
		return "ppt"
	default:
		return ""
	}
}

// EncryptChunk encrypts a single chunk of data using AES-256-GCM.
// This is a standalone function for testing and direct chunk encryption.
// Implements Requirement 2.1: AES-256-GCM encryption.
func EncryptChunk(plaintext, cek, iv []byte) ([]byte, []byte, error) {
	if len(cek) != domain.CEKSize {
		return nil, nil, fmt.Errorf("invalid CEK size: expected %d, got %d", domain.CEKSize, len(cek))
	}
	if len(iv) != domain.IVSize {
		return nil, nil, fmt.Errorf("invalid IV size: expected %d, got %d", domain.IVSize, len(iv))
	}

	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// GCM Seal appends the auth tag to the ciphertext
	ciphertext := gcm.Seal(nil, iv, plaintext, nil)

	// Extract auth tag (last 16 bytes)
	authTagStart := len(ciphertext) - gcm.Overhead()
	authTag := make([]byte, gcm.Overhead())
	copy(authTag, ciphertext[authTagStart:])

	// Return ciphertext without auth tag, and auth tag separately
	ciphertextOnly := ciphertext[:authTagStart]

	return ciphertextOnly, authTag, nil
}

// EncryptChunkWithTag encrypts a chunk and returns ciphertext with auth tag appended.
func EncryptChunkWithTag(plaintext, cek, iv []byte) ([]byte, error) {
	if len(cek) != domain.CEKSize {
		return nil, fmt.Errorf("invalid CEK size: expected %d, got %d", domain.CEKSize, len(cek))
	}
	if len(iv) != domain.IVSize {
		return nil, fmt.Errorf("invalid IV size: expected %d, got %d", domain.IVSize, len(iv))
	}

	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// GCM Seal appends the auth tag to the ciphertext
	return gcm.Seal(nil, iv, plaintext, nil), nil
}

// DecryptChunkWithTag decrypts a chunk that has the auth tag appended.
func DecryptChunkWithTag(ciphertextWithTag, cek, iv []byte) ([]byte, error) {
	if len(cek) != domain.CEKSize {
		return nil, fmt.Errorf("invalid CEK size: expected %d, got %d", domain.CEKSize, len(cek))
	}
	if len(iv) != domain.IVSize {
		return nil, fmt.Errorf("invalid IV size: expected %d, got %d", domain.IVSize, len(iv))
	}

	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, iv, ciphertextWithTag, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (authentication failed): %w", err)
	}

	return plaintext, nil
}

// GenerateIVsForChunks generates unique IVs for all chunks of a file.
// Implements Requirement 2.3: Unique IV per chunk.
func GenerateIVsForChunks(baseSeed []byte, chunkCount int) [][]byte {
	ivs := make([][]byte, chunkCount)
	for i := 0; i < chunkCount; i++ {
		ivs[i] = DeriveIVForChunk(baseSeed, i)
	}
	return ivs
}

// ValidateIVUniqueness checks that all IVs in a slice are unique.
func ValidateIVUniqueness(ivs [][]byte) bool {
	seen := make(map[string]bool)
	for _, iv := range ivs {
		key := string(iv)
		if seen[key] {
			return false
		}
		seen[key] = true
	}
	return true
}

// ValidateManifestCompleteness checks that a manifest has all required fields.
// Implements Property 10: Manifest Completeness.
func ValidateManifestCompleteness(manifest *domain.DownloadManifest) bool {
	if manifest == nil {
		return false
	}
	if manifest.MaterialID == uuid.Nil {
		return false
	}
	if manifest.LicenseID == uuid.Nil {
		return false
	}
	if manifest.TotalChunks <= 0 {
		return false
	}
	if manifest.TotalSize <= 0 {
		return false
	}
	if manifest.OriginalHash == "" {
		return false
	}
	if manifest.EncryptedHash == "" {
		return false
	}
	if manifest.ChunkSize <= 0 {
		return false
	}
	if len(manifest.Chunks) == 0 {
		return false
	}
	if manifest.FileType == "" {
		return false
	}
	return true
}

// DeriveIVForChunkExported is an exported wrapper for DeriveIVForChunk for testing.
// It derives a unique IV for a specific chunk index from a base seed.
func DeriveIVForChunkExported(baseSeed []byte, chunkIndex int) []byte {
	return DeriveIVForChunk(baseSeed, chunkIndex)
}

// ModifyByte modifies a single byte in a slice (for testing authentication tag integrity).
func ModifyByte(data []byte, index int) []byte {
	if index < 0 || index >= len(data) {
		return data
	}
	modified := make([]byte, len(data))
	copy(modified, data)
	modified[index] ^= 0xFF // Flip all bits
	return modified
}

// GenerateRandomCEK generates a random 32-byte CEK for testing.
func GenerateRandomCEK() ([]byte, error) {
	cek := make([]byte, domain.CEKSize)
	if _, err := io.ReadFull(rand.Reader, cek); err != nil {
		return nil, err
	}
	return cek, nil
}

// GenerateRandomIV generates a random 12-byte IV for testing.
func GenerateRandomIV() ([]byte, error) {
	iv := make([]byte, domain.IVSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	return iv, nil
}

// DeriveIVForChunk derives a unique IV for a specific chunk index.
// This function is defined in key_management_service.go but we need it here too.
// The IV is derived by combining a base seed with the chunk index.
func deriveIVForChunkInternal(baseSeed []byte, chunkIndex int) []byte {
	iv := make([]byte, domain.IVSize)

	// Use first 8 bytes from seed
	if len(baseSeed) >= 8 {
		copy(iv, baseSeed[:8])
	}

	// Use last 4 bytes for chunk index (big-endian)
	binary.BigEndian.PutUint32(iv[8:], uint32(chunkIndex))

	return iv
}
