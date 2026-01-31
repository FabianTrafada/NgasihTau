package http

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"ngasihtau/internal/common/validator"
)

// TestProperty17_RequestValidationChain tests that request validation
// properly validates all fields in the correct order and returns appropriate errors.
// Property 17: Request Validation Chain
func TestProperty17_RequestValidationChain(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	v := validator.Get()

	// Property: IssueLicenseRequest validation - valid fingerprint length
	properties.Property("IssueLicenseRequest accepts valid fingerprints (32-512 chars)", prop.ForAll(
		func(fingerprintLen int) bool {
			fingerprint := generateString(fingerprintLen)
			deviceID := uuid.New().String()

			req := IssueLicenseRequest{
				DeviceID:    deviceID,
				Fingerprint: fingerprint,
			}

			err := v.Struct(&req)
			// Should pass if fingerprint is between 32 and 512 chars
			if fingerprintLen >= 32 && fingerprintLen <= 512 {
				return err == nil
			}
			// Should fail otherwise
			return err != nil
		},
		gen.IntRange(1, 600),
	))

	// Property: IssueLicenseRequest validation - invalid device ID
	properties.Property("IssueLicenseRequest rejects invalid device IDs", prop.ForAll(
		func(invalidID string) bool {
			req := IssueLicenseRequest{
				DeviceID:    invalidID,
				Fingerprint: generateString(64), // Valid fingerprint
			}

			err := v.Struct(&req)
			// Should fail for non-UUID device IDs
			_, parseErr := uuid.Parse(invalidID)
			if parseErr != nil {
				return err != nil
			}
			return true
		},
		gen.AnyString().SuchThat(func(s string) bool {
			_, err := uuid.Parse(s)
			return err != nil && len(s) > 0
		}),
	))

	// Property: ValidateLicenseRequest validation - nonce length
	properties.Property("ValidateLicenseRequest validates nonce length (16-64 chars)", prop.ForAll(
		func(nonceLen int) bool {
			nonce := generateString(nonceLen)
			fingerprint := generateString(64) // Valid fingerprint

			req := ValidateLicenseRequest{
				Fingerprint: fingerprint,
				Nonce:       nonce,
			}

			err := v.Struct(&req)
			// Should pass if nonce is between 16 and 64 chars
			if nonceLen >= 16 && nonceLen <= 64 {
				return err == nil
			}
			// Should fail otherwise
			return err != nil
		},
		gen.IntRange(1, 100),
	))

	// Property: Empty required fields always fail validation
	properties.Property("Empty required fields always fail validation", prop.ForAll(
		func(emptyField int) bool {
			switch emptyField {
			case 0:
				// Empty DeviceID
				req := IssueLicenseRequest{
					DeviceID:    "",
					Fingerprint: generateString(64),
				}
				return v.Struct(&req) != nil
			case 1:
				// Empty Fingerprint in IssueLicenseRequest
				req := IssueLicenseRequest{
					DeviceID:    uuid.New().String(),
					Fingerprint: "",
				}
				return v.Struct(&req) != nil
			case 2:
				// Empty Fingerprint in ValidateLicenseRequest
				req := ValidateLicenseRequest{
					Fingerprint: "",
					Nonce:       generateString(32),
				}
				return v.Struct(&req) != nil
			case 3:
				// Empty Nonce in ValidateLicenseRequest
				req := ValidateLicenseRequest{
					Fingerprint: generateString(64),
					Nonce:       "",
				}
				return v.Struct(&req) != nil
			}
			return true
		},
		gen.IntRange(0, 3),
	))

	// Property: Valid requests always pass validation
	properties.Property("Valid requests always pass validation", prop.ForAll(
		func(seed int64) bool {
			// Generate valid IssueLicenseRequest
			issueReq := IssueLicenseRequest{
				DeviceID:    uuid.New().String(),
				Fingerprint: generateString(64),
			}
			if v.Struct(&issueReq) != nil {
				return false
			}

			// Generate valid ValidateLicenseRequest
			validateReq := ValidateLicenseRequest{
				Fingerprint: generateString(64),
				Nonce:       generateString(32),
			}
			if v.Struct(&validateReq) != nil {
				return false
			}

			return true
		},
		gen.Int64(),
	))

	properties.TestingRun(t)
}

// TestProperty18_RangeRequestSupport tests that range request parsing
// correctly handles valid and invalid range headers.
// Property 18: Range Request Support
// **Validates: Requirements 4.5**
func TestProperty18_RangeRequestSupport(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Valid range headers with start and end are parsed correctly
	properties.Property("Valid range headers with start and end are parsed correctly", prop.ForAll(
		func(start, end int64) bool {
			if start > end {
				return true // Skip invalid ranges
			}
			header := "bytes=" + intToString(start) + "-" + intToString(end)
			parsedStart, parsedEnd, err := parseRangeHeader(header)
			if err != nil {
				return false
			}
			return parsedStart == start && parsedEnd == end
		},
		gen.Int64Range(0, 1000000),
		gen.Int64Range(0, 1000000),
	))

	// Property: Valid range headers with start only are parsed correctly
	properties.Property("Valid range headers with start only are parsed correctly", prop.ForAll(
		func(start int64) bool {
			header := "bytes=" + intToString(start) + "-"
			parsedStart, parsedEnd, err := parseRangeHeader(header)
			if err != nil {
				return false
			}
			return parsedStart == start && parsedEnd == -1
		},
		gen.Int64Range(0, 1000000),
	))

	// Property: Range start must not exceed end
	properties.Property("Range start must not exceed end", prop.ForAll(
		func(start, end int64) bool {
			if start <= end {
				return true // Skip valid ranges
			}
			header := "bytes=" + intToString(start) + "-" + intToString(end)
			_, _, err := parseRangeHeader(header)
			return err != nil
		},
		gen.Int64Range(1, 1000000),
		gen.Int64Range(0, 999999),
	))

	// Property: Invalid prefixes are rejected
	properties.Property("Invalid prefixes are rejected", prop.ForAll(
		func(prefixIdx int) bool {
			invalidPrefixes := []string{"chars", "bits", "octets", "range", "data", "content", "file", "block", "segment", "part"}
			prefix := invalidPrefixes[prefixIdx%len(invalidPrefixes)]
			header := prefix + "=0-100"
			_, _, err := parseRangeHeader(header)
			return err != nil
		},
		gen.IntRange(0, 9),
	))

	// Property: Empty range headers are rejected
	properties.Property("Empty range headers are rejected", prop.ForAll(
		func(_ int) bool {
			_, _, err := parseRangeHeader("")
			return err != nil
		},
		gen.Int(),
	))

	// Property: Range headers without bytes= prefix are rejected
	properties.Property("Range headers without bytes= prefix are rejected", prop.ForAll(
		func(start, end int64) bool {
			header := intToString(start) + "-" + intToString(end)
			_, _, err := parseRangeHeader(header)
			return err != nil
		},
		gen.Int64Range(0, 1000),
		gen.Int64Range(0, 1000),
	))

	properties.TestingRun(t)
}

// generateString generates a string of the specified length.
func generateString(length int) string {
	if length <= 0 {
		return ""
	}
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[i%len(chars)]
	}
	return string(result)
}

// intToString converts an int64 to string.
func intToString(n int64) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + intToString(-n)
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
