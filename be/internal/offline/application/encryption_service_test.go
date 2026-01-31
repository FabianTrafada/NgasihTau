package application

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ngasihtau/internal/offline/domain"
)

// TestEncryptChunk tests AES-256-GCM encryption for a single chunk.
func TestEncryptChunk(t *testing.T) {
	// Generate random CEK
	cek := make([]byte, domain.CEKSize)
	_, err := io.ReadFull(rand.Reader, cek)
	require.NoError(t, err)

	// Generate random IV
	iv := make([]byte, domain.IVSize)
	_, err = io.ReadFull(rand.Reader, iv)
	require.NoError(t, err)

	// Test data
	plaintext := []byte("Hello, World! This is a test message for encryption.")

	// Encrypt
	ciphertext, authTag, err := EncryptChunk(plaintext, cek, iv)
	require.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.Len(t, authTag, domain.AuthTagSize)

	// Verify ciphertext is different from plaintext
	assert.NotEqual(t, plaintext, ciphertext)
}

// TestEncryptChunkWithTag tests encryption with auth tag appended.
func TestEncryptChunkWithTag(t *testing.T) {
	cek := make([]byte, domain.CEKSize)
	_, err := io.ReadFull(rand.Reader, cek)
	require.NoError(t, err)

	iv := make([]byte, domain.IVSize)
	_, err = io.ReadFull(rand.Reader, iv)
	require.NoError(t, err)

	plaintext := []byte("Test message for encryption with tag.")

	ciphertextWithTag, err := EncryptChunkWithTag(plaintext, cek, iv)
	require.NoError(t, err)
	assert.NotEmpty(t, ciphertextWithTag)

	// Ciphertext with tag should be longer than plaintext by auth tag size
	assert.Greater(t, len(ciphertextWithTag), len(plaintext))
}

// TestDecryptChunkWithTag tests decryption of chunk with auth tag.
func TestDecryptChunkWithTag(t *testing.T) {
	cek := make([]byte, domain.CEKSize)
	_, err := io.ReadFull(rand.Reader, cek)
	require.NoError(t, err)

	iv := make([]byte, domain.IVSize)
	_, err = io.ReadFull(rand.Reader, iv)
	require.NoError(t, err)

	plaintext := []byte("Test message for round-trip encryption.")

	// Encrypt
	ciphertextWithTag, err := EncryptChunkWithTag(plaintext, cek, iv)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := DecryptChunkWithTag(ciphertextWithTag, cek, iv)
	require.NoError(t, err)

	// Verify round-trip
	assert.Equal(t, plaintext, decrypted)
}

// TestEncryptionRoundTrip tests that encryption followed by decryption produces original content.
func TestEncryptionRoundTrip(t *testing.T) {
	testCases := []struct {
		name string
		size int
	}{
		{"small", 100},
		{"medium", 1024},
		{"large", 10 * 1024},
		{"chunk_boundary", int(domain.DefaultChunkSize)},
		{"multi_chunk", int(domain.DefaultChunkSize)*2 + 500},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate random content
			content := make([]byte, tc.size)
			_, err := io.ReadFull(rand.Reader, content)
			require.NoError(t, err)

			// Generate CEK
			cek := make([]byte, domain.CEKSize)
			_, err = io.ReadFull(rand.Reader, cek)
			require.NoError(t, err)

			// Generate base seed
			baseSeed := make([]byte, domain.IVSize)
			_, err = io.ReadFull(rand.Reader, baseSeed)
			require.NoError(t, err)

			// Create encryption service (we'll use the internal function directly)
			svc := &EncryptionService{}

			// Encrypt
			encryptedContent, chunks, err := svc.encryptFileInChunks(content, cek, baseSeed)
			require.NoError(t, err)
			assert.NotEmpty(t, encryptedContent)
			assert.NotEmpty(t, chunks)

			// Decrypt
			decrypted, err := DecryptFile(encryptedContent, cek, chunks)
			require.NoError(t, err)

			// Verify
			assert.Equal(t, content, decrypted)
		})
	}
}

// TestCalculateChunkCount tests chunk count calculation.
func TestCalculateChunkCount(t *testing.T) {
	testCases := []struct {
		name     string
		fileSize int64
		expected int
	}{
		{"zero", 0, 0},
		{"one_byte", 1, 1},
		{"exactly_one_chunk", domain.DefaultChunkSize, 1},
		{"one_byte_over", domain.DefaultChunkSize + 1, 2},
		{"two_chunks", domain.DefaultChunkSize * 2, 2},
		{"two_and_half", domain.DefaultChunkSize*2 + domain.DefaultChunkSize/2, 3},
		{"ten_chunks", domain.DefaultChunkSize * 10, 10},
		{"ten_plus_one", domain.DefaultChunkSize*10 + 1, 11},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateChunkCount(tc.fileSize)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestDeriveIVForChunk_Encryption tests IV derivation for chunks.
func TestDeriveIVForChunk_Encryption(t *testing.T) {
	baseSeed := make([]byte, domain.IVSize)
	_, err := io.ReadFull(rand.Reader, baseSeed)
	require.NoError(t, err)

	// Generate IVs for multiple chunks
	ivs := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		ivs[i] = DeriveIVForChunk(baseSeed, i)
		assert.Len(t, ivs[i], domain.IVSize)
	}

	// Verify all IVs are unique
	assert.True(t, ValidateIVUniqueness(ivs))

	// Verify determinism - same seed and index should produce same IV
	iv0Again := DeriveIVForChunk(baseSeed, 0)
	assert.Equal(t, ivs[0], iv0Again)
}

// TestIVUniqueness tests that IVs are unique across chunks.
func TestIVUniqueness(t *testing.T) {
	baseSeed := make([]byte, domain.IVSize)
	_, err := io.ReadFull(rand.Reader, baseSeed)
	require.NoError(t, err)

	// Generate many IVs
	ivs := GenerateIVsForChunks(baseSeed, 1000)
	assert.Len(t, ivs, 1000)

	// All should be unique
	assert.True(t, ValidateIVUniqueness(ivs))
}

// TestAuthenticationTagIntegrity tests that modifying ciphertext causes decryption to fail.
func TestAuthenticationTagIntegrity(t *testing.T) {
	cek := make([]byte, domain.CEKSize)
	_, err := io.ReadFull(rand.Reader, cek)
	require.NoError(t, err)

	iv := make([]byte, domain.IVSize)
	_, err = io.ReadFull(rand.Reader, iv)
	require.NoError(t, err)

	plaintext := []byte("Test message for integrity check.")

	// Encrypt
	ciphertextWithTag, err := EncryptChunkWithTag(plaintext, cek, iv)
	require.NoError(t, err)

	// Modify a byte in the ciphertext
	modified := ModifyByte(ciphertextWithTag, 5)

	// Decryption should fail
	_, err = DecryptChunkWithTag(modified, cek, iv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
}

// TestInvalidCEKSize tests that invalid CEK sizes are rejected.
func TestInvalidCEKSize(t *testing.T) {
	iv := make([]byte, domain.IVSize)
	plaintext := []byte("test")

	// Too short
	shortCEK := make([]byte, 16)
	_, _, err := EncryptChunk(plaintext, shortCEK, iv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CEK size")

	// Too long
	longCEK := make([]byte, 64)
	_, _, err = EncryptChunk(plaintext, longCEK, iv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CEK size")
}

// TestInvalidIVSize tests that invalid IV sizes are rejected.
func TestInvalidIVSize(t *testing.T) {
	cek := make([]byte, domain.CEKSize)
	plaintext := []byte("test")

	// Too short
	shortIV := make([]byte, 8)
	_, _, err := EncryptChunk(plaintext, cek, shortIV)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid IV size")

	// Too long
	longIV := make([]byte, 16)
	_, _, err = EncryptChunk(plaintext, cek, longIV)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid IV size")
}

// TestValidateManifestCompleteness tests manifest validation.
func TestValidateManifestCompleteness(t *testing.T) {
	// Valid manifest
	validManifest := &domain.DownloadManifest{
		MaterialID:    uuid.New(),
		LicenseID:     uuid.New(),
		TotalChunks:   5,
		TotalSize:     5 * 1024 * 1024,
		OriginalHash:  "abc123",
		EncryptedHash: "def456",
		ChunkSize:     domain.DefaultChunkSize,
		Chunks:        []domain.EncryptedChunk{{Index: 0}},
		FileType:      "pdf",
	}
	assert.True(t, ValidateManifestCompleteness(validManifest))

	// Nil manifest
	assert.False(t, ValidateManifestCompleteness(nil))

	// Missing MaterialID
	invalidManifest := *validManifest
	invalidManifest.MaterialID = uuid.Nil
	assert.False(t, ValidateManifestCompleteness(&invalidManifest))

	// Missing LicenseID
	invalidManifest = *validManifest
	invalidManifest.LicenseID = uuid.Nil
	assert.False(t, ValidateManifestCompleteness(&invalidManifest))

	// Zero TotalChunks
	invalidManifest = *validManifest
	invalidManifest.TotalChunks = 0
	assert.False(t, ValidateManifestCompleteness(&invalidManifest))

	// Zero TotalSize
	invalidManifest = *validManifest
	invalidManifest.TotalSize = 0
	assert.False(t, ValidateManifestCompleteness(&invalidManifest))

	// Empty OriginalHash
	invalidManifest = *validManifest
	invalidManifest.OriginalHash = ""
	assert.False(t, ValidateManifestCompleteness(&invalidManifest))

	// Empty EncryptedHash
	invalidManifest = *validManifest
	invalidManifest.EncryptedHash = ""
	assert.False(t, ValidateManifestCompleteness(&invalidManifest))

	// Zero ChunkSize
	invalidManifest = *validManifest
	invalidManifest.ChunkSize = 0
	assert.False(t, ValidateManifestCompleteness(&invalidManifest))

	// Empty Chunks
	invalidManifest = *validManifest
	invalidManifest.Chunks = nil
	assert.False(t, ValidateManifestCompleteness(&invalidManifest))

	// Empty FileType
	invalidManifest = *validManifest
	invalidManifest.FileType = ""
	assert.False(t, ValidateManifestCompleteness(&invalidManifest))
}

// TestGetFileTypeFromPath tests file type extraction from path.
func TestGetFileTypeFromPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"document.pdf", "pdf"},
		{"document.PDF", "pdf"},
		{"report.docx", "docx"},
		{"presentation.pptx", "pptx"},
		{"file.txt", "txt"},
		{"noextension", ""},
		{"/path/to/file.pdf", "pdf"},
		{"file.tar.gz", "gz"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			result := GetFileTypeFromPath(tc.path)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestGetFileTypeFromContentType tests file type extraction from content type.
func TestGetFileTypeFromContentType(t *testing.T) {
	testCases := []struct {
		contentType string
		expected    string
	}{
		{"application/pdf", "pdf"},
		{"application/vnd.openxmlformats-officedocument.wordprocessingml.document", "docx"},
		{"application/vnd.openxmlformats-officedocument.presentationml.presentation", "pptx"},
		{"application/msword", "doc"},
		{"application/vnd.ms-powerpoint", "ppt"},
		{"text/plain", ""},
		{"application/octet-stream", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.contentType, func(t *testing.T) {
			result := GetFileTypeFromContentType(tc.contentType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestSupportedFileTypes tests file type validation.
func TestSupportedFileTypes(t *testing.T) {
	// Supported types
	assert.True(t, domain.IsSupportedFileType("pdf"))
	assert.True(t, domain.IsSupportedFileType("docx"))
	assert.True(t, domain.IsSupportedFileType("pptx"))

	// Unsupported types
	assert.False(t, domain.IsSupportedFileType("txt"))
	assert.False(t, domain.IsSupportedFileType("jpg"))
	assert.False(t, domain.IsSupportedFileType("mp4"))
	assert.False(t, domain.IsSupportedFileType(""))
}

// TestCalculateSHA256 tests hash calculation.
func TestCalculateSHA256(t *testing.T) {
	content := []byte("Hello, World!")
	hash := calculateSHA256(content)

	// SHA-256 produces 64 hex characters
	assert.Len(t, hash, 64)

	// Same content should produce same hash
	hash2 := calculateSHA256(content)
	assert.Equal(t, hash, hash2)

	// Different content should produce different hash
	hash3 := calculateSHA256([]byte("Different content"))
	assert.NotEqual(t, hash, hash3)
}

// TestGenerateEncryptedFileKey tests encrypted file key generation.
func TestGenerateEncryptedFileKey(t *testing.T) {
	materialID := uuid.New()
	cekID := uuid.New()

	key := generateEncryptedFileKey(materialID, cekID)

	assert.Contains(t, key, "encrypted/")
	assert.Contains(t, key, materialID.String())
	assert.Contains(t, key, cekID.String())
	assert.Contains(t, key, ".enc")
}

// TestEmptyContent tests encryption of empty content.
func TestEmptyContent(t *testing.T) {
	cek := make([]byte, domain.CEKSize)
	_, err := io.ReadFull(rand.Reader, cek)
	require.NoError(t, err)

	baseSeed := make([]byte, domain.IVSize)
	_, err = io.ReadFull(rand.Reader, baseSeed)
	require.NoError(t, err)

	svc := &EncryptionService{}

	// Empty content should produce no chunks
	encryptedContent, chunks, err := svc.encryptFileInChunks([]byte{}, cek, baseSeed)
	require.NoError(t, err)
	assert.Empty(t, encryptedContent)
	assert.Empty(t, chunks)
}

// TestLargeFile tests encryption of a large file (multiple chunks).
func TestLargeFile(t *testing.T) {
	// Create 5MB of random content
	content := make([]byte, 5*1024*1024)
	_, err := io.ReadFull(rand.Reader, content)
	require.NoError(t, err)

	cek := make([]byte, domain.CEKSize)
	_, err = io.ReadFull(rand.Reader, cek)
	require.NoError(t, err)

	baseSeed := make([]byte, domain.IVSize)
	_, err = io.ReadFull(rand.Reader, baseSeed)
	require.NoError(t, err)

	svc := &EncryptionService{}

	// Encrypt
	encryptedContent, chunks, err := svc.encryptFileInChunks(content, cek, baseSeed)
	require.NoError(t, err)

	// Should have 5 chunks
	assert.Len(t, chunks, 5)

	// Decrypt and verify
	decrypted, err := DecryptFile(encryptedContent, cek, chunks)
	require.NoError(t, err)
	assert.Equal(t, content, decrypted)
}

// Mock implementations for testing

type mockMinIOStorageClient struct {
	objects map[string][]byte
}

func newMockMinIOStorageClient() *mockMinIOStorageClient {
	return &mockMinIOStorageClient{
		objects: make(map[string][]byte),
	}
}

func (m *mockMinIOStorageClient) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	data, ok := m.objects[objectKey]
	if !ok {
		return nil, domain.NewOfflineError(domain.ErrCodeMaterialNotFound, "object not found")
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *mockMinIOStorageClient) PutObject(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	m.objects[objectKey] = data
	return nil
}

func (m *mockMinIOStorageClient) DeleteObject(ctx context.Context, objectKey string) error {
	delete(m.objects, objectKey)
	return nil
}

func (m *mockMinIOStorageClient) DeleteObjects(ctx context.Context, objectKeys []string) error {
	for _, key := range objectKeys {
		delete(m.objects, key)
	}
	return nil
}

func (m *mockMinIOStorageClient) GetObjectInfo(ctx context.Context, objectKey string) (*ObjectInfo, error) {
	data, ok := m.objects[objectKey]
	if !ok {
		return nil, domain.NewOfflineError(domain.ErrCodeMaterialNotFound, "object not found")
	}
	return &ObjectInfo{
		Size:        int64(len(data)),
		ContentType: "application/octet-stream",
	}, nil
}

func (m *mockMinIOStorageClient) GeneratePresignedGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	return "https://example.com/presigned/" + objectKey, nil
}

type mockMaterialAccessChecker struct {
	materials map[uuid.UUID]struct {
		objectKey string
		fileType  string
	}
}

func newMockMaterialAccessChecker() *mockMaterialAccessChecker {
	return &mockMaterialAccessChecker{
		materials: make(map[uuid.UUID]struct {
			objectKey string
			fileType  string
		}),
	}
}

func (m *mockMaterialAccessChecker) GetMaterialFileKey(ctx context.Context, materialID uuid.UUID) (string, string, error) {
	mat, ok := m.materials[materialID]
	if !ok {
		return "", "", domain.NewOfflineError(domain.ErrCodeMaterialNotFound, "material not found")
	}
	return mat.objectKey, mat.fileType, nil
}

func (m *mockMaterialAccessChecker) AddMaterial(materialID uuid.UUID, objectKey, fileType string) {
	m.materials[materialID] = struct {
		objectKey string
		fileType  string
	}{objectKey, fileType}
}

type mockEncryptedMaterialRepository struct {
	materials map[uuid.UUID]*domain.EncryptedMaterial
}

func newMockEncryptedMaterialRepository() *mockEncryptedMaterialRepository {
	return &mockEncryptedMaterialRepository{
		materials: make(map[uuid.UUID]*domain.EncryptedMaterial),
	}
}

func (m *mockEncryptedMaterialRepository) Create(ctx context.Context, material *domain.EncryptedMaterial) error {
	m.materials[material.ID] = material
	return nil
}

func (m *mockEncryptedMaterialRepository) FindById(ctx context.Context, id uuid.UUID) (*domain.EncryptedMaterial, error) {
	mat, ok := m.materials[id]
	if !ok {
		return nil, domain.NewOfflineError(domain.ErrCodeMaterialNotFound, "not found")
	}
	return mat, nil
}

func (m *mockEncryptedMaterialRepository) FindByMaterialAndCEK(ctx context.Context, materialID, cekID uuid.UUID) (*domain.EncryptedMaterial, error) {
	for _, mat := range m.materials {
		if mat.MaterialID == materialID && mat.CEKID == cekID {
			return mat, nil
		}
	}
	return nil, domain.NewOfflineError(domain.ErrCodeMaterialNotFound, "not found")
}

func (m *mockEncryptedMaterialRepository) FindByMaterialID(ctx context.Context, materialID uuid.UUID) ([]*domain.EncryptedMaterial, error) {
	var result []*domain.EncryptedMaterial
	for _, mat := range m.materials {
		if mat.MaterialID == materialID {
			result = append(result, mat)
		}
	}
	return result, nil
}

func (m *mockEncryptedMaterialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.materials, id)
	return nil
}

func (m *mockEncryptedMaterialRepository) DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error {
	for id, mat := range m.materials {
		if mat.MaterialID == materialID {
			delete(m.materials, id)
		}
	}
	return nil
}

// TestEncryptMaterialIntegration tests the full encryption flow with mocks.
func TestEncryptMaterialIntegration(t *testing.T) {
	// Setup mocks
	storage := newMockMinIOStorageClient()
	materialChecker := newMockMaterialAccessChecker()
	encryptedMaterialRepo := newMockEncryptedMaterialRepository()

	// Create test material
	materialID := uuid.New()
	objectKey := "materials/" + materialID.String() + "/file.pdf"
	content := make([]byte, 2*1024*1024+500) // 2.5 MB
	_, err := io.ReadFull(rand.Reader, content)
	require.NoError(t, err)

	storage.objects[objectKey] = content
	materialChecker.AddMaterial(materialID, objectKey, "pdf")

	// Create service
	svc := NewEncryptionService(
		storage,
		materialChecker,
		encryptedMaterialRepo,
		nil, // no event publisher for this test
		EncryptionServiceConfig{EncryptedBucket: "encrypted"},
	)

	// Generate CEK
	cek := make([]byte, domain.CEKSize)
	_, err = io.ReadFull(rand.Reader, cek)
	require.NoError(t, err)

	// Encrypt
	input := EncryptMaterialInput{
		MaterialID: materialID,
		LicenseID:  uuid.New(),
		CEKID:      uuid.New(),
		CEK:        cek,
		UserID:     uuid.New(),
		DeviceID:   uuid.New(),
	}

	output, err := svc.EncryptMaterial(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, output)
	require.NotNil(t, output.Manifest)

	// Verify manifest
	assert.True(t, ValidateManifestCompleteness(output.Manifest))
	assert.Equal(t, materialID, output.Manifest.MaterialID)
	assert.Equal(t, input.LicenseID, output.Manifest.LicenseID)
	assert.Equal(t, 3, output.Manifest.TotalChunks) // 2.5 MB = 3 chunks
	assert.Equal(t, "pdf", output.Manifest.FileType)
	assert.NotEmpty(t, output.Manifest.OriginalHash)
	assert.NotEmpty(t, output.Manifest.EncryptedHash)
	assert.NotEmpty(t, output.EncryptedFileURL)

	// Verify encrypted file was stored
	encryptedFileKey := generateEncryptedFileKey(materialID, input.CEKID)
	encryptedContent, ok := storage.objects[encryptedFileKey]
	assert.True(t, ok)
	assert.NotEmpty(t, encryptedContent)

	// Verify we can decrypt
	decrypted, err := DecryptFile(encryptedContent, cek, output.Manifest.Chunks)
	require.NoError(t, err)
	assert.Equal(t, content, decrypted)
}

// TestEncryptMaterialUnsupportedFileType tests rejection of unsupported file types.
func TestEncryptMaterialUnsupportedFileType(t *testing.T) {
	storage := newMockMinIOStorageClient()
	materialChecker := newMockMaterialAccessChecker()
	encryptedMaterialRepo := newMockEncryptedMaterialRepository()

	materialID := uuid.New()
	objectKey := "materials/" + materialID.String() + "/file.txt"
	storage.objects[objectKey] = []byte("test content")
	materialChecker.AddMaterial(materialID, objectKey, "txt") // Unsupported type

	svc := NewEncryptionService(
		storage,
		materialChecker,
		encryptedMaterialRepo,
		nil,
		EncryptionServiceConfig{EncryptedBucket: "encrypted"},
	)

	cek := make([]byte, domain.CEKSize)
	_, _ = io.ReadFull(rand.Reader, cek)

	input := EncryptMaterialInput{
		MaterialID: materialID,
		LicenseID:  uuid.New(),
		CEKID:      uuid.New(),
		CEK:        cek,
		UserID:     uuid.New(),
		DeviceID:   uuid.New(),
	}

	_, err := svc.EncryptMaterial(context.Background(), input)
	require.Error(t, err)

	offlineErr, ok := domain.GetOfflineError(err)
	require.True(t, ok)
	assert.Equal(t, domain.ErrCodeUnsupportedFileType, offlineErr.Code)
}

// TestEncryptMaterialInvalidCEK tests rejection of invalid CEK.
func TestEncryptMaterialInvalidCEK(t *testing.T) {
	storage := newMockMinIOStorageClient()
	materialChecker := newMockMaterialAccessChecker()
	encryptedMaterialRepo := newMockEncryptedMaterialRepository()

	svc := NewEncryptionService(
		storage,
		materialChecker,
		encryptedMaterialRepo,
		nil,
		EncryptionServiceConfig{EncryptedBucket: "encrypted"},
	)

	input := EncryptMaterialInput{
		MaterialID: uuid.New(),
		LicenseID:  uuid.New(),
		CEKID:      uuid.New(),
		CEK:        []byte("too short"), // Invalid CEK
		UserID:     uuid.New(),
		DeviceID:   uuid.New(),
	}

	_, err := svc.EncryptMaterial(context.Background(), input)
	require.Error(t, err)

	offlineErr, ok := domain.GetOfflineError(err)
	require.True(t, ok)
	assert.Equal(t, domain.ErrCodeInvalidKey, offlineErr.Code)
}
