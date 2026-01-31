package application

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"ngasihtau/internal/offline/domain"
)

// TestProperty6_EncryptionRoundTrip tests that encryption followed by decryption
// produces content identical to the original.
// Feature: offline-material-backend, Property 6: Encryption Round-Trip
// **Validates: Requirements 2.1**
func TestProperty6_EncryptionRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("encryption round-trip preserves content", prop.ForAll(
		func(contentSize int) bool {
			// Generate random content
			if contentSize <= 0 {
				return true // Skip empty content
			}
			content := make([]byte, contentSize)
			if _, err := io.ReadFull(rand.Reader, content); err != nil {
				return false
			}

			// Generate random CEK
			cek := make([]byte, domain.CEKSize)
			if _, err := io.ReadFull(rand.Reader, cek); err != nil {
				return false
			}

			// Generate base seed for IV derivation
			baseSeed := make([]byte, domain.IVSize)
			if _, err := io.ReadFull(rand.Reader, baseSeed); err != nil {
				return false
			}

			// Encrypt
			svc := &EncryptionService{}
			encryptedContent, chunks, err := svc.encryptFileInChunks(content, cek, baseSeed)
			if err != nil {
				return false
			}

			// Decrypt
			decrypted, err := DecryptFile(encryptedContent, cek, chunks)
			if err != nil {
				return false
			}

			// Verify content is identical
			return bytes.Equal(content, decrypted)
		},
		// Generate content sizes from 1 byte to 5MB
		gen.IntRange(1, 5*1024*1024),
	))

	properties.TestingRun(t)
}

// TestProperty7_ChunkSizeCalculation tests that chunk count equals ceil(S / 1048576).
// Feature: offline-material-backend, Property 7: Chunk Size Calculation
// **Validates: Requirements 2.2**
func TestProperty7_ChunkSizeCalculation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("chunk count equals ceil(fileSize / chunkSize)", prop.ForAll(
		func(fileSize int64) bool {
			if fileSize < 0 {
				return true // Skip negative sizes
			}

			chunkCount := CalculateChunkCount(fileSize)

			// Calculate expected: ceil(fileSize / 1048576)
			chunkSize := domain.DefaultChunkSize
			var expected int
			if fileSize == 0 {
				expected = 0
			} else {
				expected = int(fileSize / chunkSize)
				if fileSize%chunkSize != 0 {
					expected++
				}
			}

			return chunkCount == expected
		},
		gen.Int64Range(0, 100*1024*1024), // 0 to 100MB
	))

	properties.TestingRun(t)
}

// TestProperty8_IVUniquenessPerChunk tests that all IVs for chunks are unique.
// Feature: offline-material-backend, Property 8: IV Uniqueness Per Chunk
// **Validates: Requirements 2.3**
func TestProperty8_IVUniquenessPerChunk(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("all IVs for chunks are unique", prop.ForAll(
		func(chunkCount int) bool {
			if chunkCount <= 0 {
				return true // Skip invalid chunk counts
			}

			// Generate random base seed
			baseSeed := make([]byte, domain.IVSize)
			if _, err := io.ReadFull(rand.Reader, baseSeed); err != nil {
				return false
			}

			// Generate IVs for all chunks
			ivs := GenerateIVsForChunks(baseSeed, chunkCount)

			// Verify all IVs are unique
			return ValidateIVUniqueness(ivs)
		},
		gen.IntRange(1, 1000), // 1 to 1000 chunks
	))

	properties.TestingRun(t)
}

// TestProperty9_AuthenticationTagIntegrity tests that modifying ciphertext
// causes authentication verification to fail.
// Feature: offline-material-backend, Property 9: Authentication Tag Integrity
// **Validates: Requirements 2.4**
func TestProperty9_AuthenticationTagIntegrity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("modifying ciphertext causes authentication to fail", prop.ForAll(
		func(contentSize int, modifyIndex int) bool {
			if contentSize <= 0 {
				return true // Skip empty content
			}

			// Generate random content
			content := make([]byte, contentSize)
			if _, err := io.ReadFull(rand.Reader, content); err != nil {
				return false
			}

			// Generate random CEK
			cek := make([]byte, domain.CEKSize)
			if _, err := io.ReadFull(rand.Reader, cek); err != nil {
				return false
			}

			// Generate random IV
			iv := make([]byte, domain.IVSize)
			if _, err := io.ReadFull(rand.Reader, iv); err != nil {
				return false
			}

			// Encrypt
			ciphertextWithTag, err := EncryptChunkWithTag(content, cek, iv)
			if err != nil {
				return false
			}

			// Ensure modifyIndex is within bounds
			if len(ciphertextWithTag) == 0 {
				return true
			}
			actualModifyIndex := modifyIndex % len(ciphertextWithTag)
			if actualModifyIndex < 0 {
				actualModifyIndex = -actualModifyIndex
			}

			// Modify a byte in the ciphertext
			modified := ModifyByte(ciphertextWithTag, actualModifyIndex)

			// Decryption should fail
			_, err = DecryptChunkWithTag(modified, cek, iv)
			return err != nil // Should return an error
		},
		gen.IntRange(1, 10*1024), // Content size 1 byte to 10KB
		gen.Int(),                // Random modify index
	))

	properties.TestingRun(t)
}

// TestProperty10_ManifestCompleteness tests that successful encryption returns
// a manifest with all required fields.
// Feature: offline-material-backend, Property 10: Manifest Completeness
// **Validates: Requirements 2.6, 4.8**
func TestProperty10_ManifestCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("successful encryption returns complete manifest", prop.ForAll(
		func(contentSize int) bool {
			if contentSize <= 0 {
				return true // Skip empty content
			}

			// Generate random content
			content := make([]byte, contentSize)
			if _, err := io.ReadFull(rand.Reader, content); err != nil {
				return false
			}

			// Generate random CEK
			cek := make([]byte, domain.CEKSize)
			if _, err := io.ReadFull(rand.Reader, cek); err != nil {
				return false
			}

			// Generate base seed
			baseSeed := make([]byte, domain.IVSize)
			if _, err := io.ReadFull(rand.Reader, baseSeed); err != nil {
				return false
			}

			// Encrypt
			svc := &EncryptionService{}
			encryptedContent, chunks, err := svc.encryptFileInChunks(content, cek, baseSeed)
			if err != nil {
				return false
			}

			// Create manifest
			materialID := uuid.New()
			licenseID := uuid.New()
			originalHash := calculateSHA256(content)
			encryptedHash := calculateSHA256(encryptedContent)

			manifest := domain.NewDownloadManifest(
				materialID,
				licenseID,
				int64(contentSize),
				originalHash,
				encryptedHash,
				"pdf",
				chunks,
			)

			// Validate completeness
			return ValidateManifestCompleteness(manifest)
		},
		gen.IntRange(1, 5*1024*1024), // 1 byte to 5MB
	))

	properties.TestingRun(t)
}

// TestProperty_EncryptionDeterminism tests that encrypting the same content
// with the same CEK and IV produces the same ciphertext.
func TestProperty_EncryptionDeterminism(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("same inputs produce same ciphertext", prop.ForAll(
		func(contentSize int) bool {
			if contentSize <= 0 {
				return true
			}

			// Generate random content
			content := make([]byte, contentSize)
			if _, err := io.ReadFull(rand.Reader, content); err != nil {
				return false
			}

			// Generate random CEK
			cek := make([]byte, domain.CEKSize)
			if _, err := io.ReadFull(rand.Reader, cek); err != nil {
				return false
			}

			// Generate random IV
			iv := make([]byte, domain.IVSize)
			if _, err := io.ReadFull(rand.Reader, iv); err != nil {
				return false
			}

			// Encrypt twice
			ciphertext1, err := EncryptChunkWithTag(content, cek, iv)
			if err != nil {
				return false
			}

			ciphertext2, err := EncryptChunkWithTag(content, cek, iv)
			if err != nil {
				return false
			}

			// Should be identical
			return bytes.Equal(ciphertext1, ciphertext2)
		},
		gen.IntRange(1, 10*1024),
	))

	properties.TestingRun(t)
}

// TestProperty_DifferentIVsProduceDifferentCiphertext tests that different IVs
// produce different ciphertext for the same content.
func TestProperty_DifferentIVsProduceDifferentCiphertext(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("different IVs produce different ciphertext", prop.ForAll(
		func(contentSize int) bool {
			if contentSize <= 0 {
				return true
			}

			// Generate random content
			content := make([]byte, contentSize)
			if _, err := io.ReadFull(rand.Reader, content); err != nil {
				return false
			}

			// Generate random CEK
			cek := make([]byte, domain.CEKSize)
			if _, err := io.ReadFull(rand.Reader, cek); err != nil {
				return false
			}

			// Generate two different IVs
			iv1 := make([]byte, domain.IVSize)
			if _, err := io.ReadFull(rand.Reader, iv1); err != nil {
				return false
			}

			iv2 := make([]byte, domain.IVSize)
			if _, err := io.ReadFull(rand.Reader, iv2); err != nil {
				return false
			}

			// Encrypt with different IVs
			ciphertext1, err := EncryptChunkWithTag(content, cek, iv1)
			if err != nil {
				return false
			}

			ciphertext2, err := EncryptChunkWithTag(content, cek, iv2)
			if err != nil {
				return false
			}

			// Should be different (with very high probability)
			return !bytes.Equal(ciphertext1, ciphertext2)
		},
		gen.IntRange(1, 10*1024),
	))

	properties.TestingRun(t)
}

// TestProperty_ChunkBoundaryCorrectness tests that chunk boundaries are calculated correctly.
func TestProperty_ChunkBoundaryCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("chunk boundaries cover entire file", prop.ForAll(
		func(contentSize int) bool {
			if contentSize <= 0 {
				return true
			}

			// Generate random content
			content := make([]byte, contentSize)
			if _, err := io.ReadFull(rand.Reader, content); err != nil {
				return false
			}

			// Generate random CEK
			cek := make([]byte, domain.CEKSize)
			if _, err := io.ReadFull(rand.Reader, cek); err != nil {
				return false
			}

			// Generate base seed
			baseSeed := make([]byte, domain.IVSize)
			if _, err := io.ReadFull(rand.Reader, baseSeed); err != nil {
				return false
			}

			// Encrypt
			svc := &EncryptionService{}
			encryptedContent, chunks, err := svc.encryptFileInChunks(content, cek, baseSeed)
			if err != nil {
				return false
			}

			// Verify chunk count matches expected
			expectedChunks := CalculateChunkCount(int64(contentSize))
			if len(chunks) != expectedChunks {
				return false
			}

			// Verify chunks cover entire encrypted content
			var totalSize int64
			for _, chunk := range chunks {
				totalSize += chunk.Size
			}

			return totalSize == int64(len(encryptedContent))
		},
		gen.IntRange(1, 5*1024*1024),
	))

	properties.TestingRun(t)
}

// TestProperty_IVDerivationDeterminism tests that IV derivation is deterministic.
func TestProperty_IVDerivationDeterminism(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("same seed and index produce same IV", prop.ForAll(
		func(chunkIndex int) bool {
			if chunkIndex < 0 {
				chunkIndex = -chunkIndex
			}

			// Generate random base seed
			baseSeed := make([]byte, domain.IVSize)
			if _, err := io.ReadFull(rand.Reader, baseSeed); err != nil {
				return false
			}

			// Derive IV twice
			iv1 := DeriveIVForChunk(baseSeed, chunkIndex)
			iv2 := DeriveIVForChunk(baseSeed, chunkIndex)

			// Should be identical
			return bytes.Equal(iv1, iv2)
		},
		gen.IntRange(0, 10000),
	))

	properties.TestingRun(t)
}

// TestProperty_EncryptedSizeLargerThanOriginal tests that encrypted content
// is larger than original due to IV and auth tag overhead.
func TestProperty_EncryptedSizeLargerThanOriginal(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("encrypted content is larger than original", prop.ForAll(
		func(contentSize int) bool {
			if contentSize <= 0 {
				return true
			}

			// Generate random content
			content := make([]byte, contentSize)
			if _, err := io.ReadFull(rand.Reader, content); err != nil {
				return false
			}

			// Generate random CEK
			cek := make([]byte, domain.CEKSize)
			if _, err := io.ReadFull(rand.Reader, cek); err != nil {
				return false
			}

			// Generate base seed
			baseSeed := make([]byte, domain.IVSize)
			if _, err := io.ReadFull(rand.Reader, baseSeed); err != nil {
				return false
			}

			// Encrypt
			svc := &EncryptionService{}
			encryptedContent, _, err := svc.encryptFileInChunks(content, cek, baseSeed)
			if err != nil {
				return false
			}

			// Encrypted content should be larger due to IV + auth tag per chunk
			return len(encryptedContent) > contentSize
		},
		gen.IntRange(1, 1024*1024),
	))

	properties.TestingRun(t)
}
