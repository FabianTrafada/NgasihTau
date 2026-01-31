package http

import (
	"testing"
)

func TestParseRangeHeader(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		wantStart   int64
		wantEnd     int64
		wantErr     bool
	}{
		{
			name:      "valid range with start and end",
			header:    "bytes=0-999",
			wantStart: 0,
			wantEnd:   999,
			wantErr:   false,
		},
		{
			name:      "valid range with start only",
			header:    "bytes=100-",
			wantStart: 100,
			wantEnd:   -1,
			wantErr:   false,
		},
		{
			name:      "valid range from beginning",
			header:    "bytes=0-",
			wantStart: 0,
			wantEnd:   -1,
			wantErr:   false,
		},
		{
			name:      "large range",
			header:    "bytes=1048576-2097151",
			wantStart: 1048576,
			wantEnd:   2097151,
			wantErr:   false,
		},
		{
			name:    "invalid format - no bytes prefix",
			header:  "0-999",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong prefix",
			header:  "chars=0-999",
			wantErr: true,
		},
		{
			name:    "invalid format - no start",
			header:  "bytes=-999",
			wantErr: true,
		},
		{
			name:    "invalid format - start greater than end",
			header:  "bytes=1000-500",
			wantErr: true,
		},
		{
			name:    "invalid format - non-numeric start",
			header:  "bytes=abc-999",
			wantErr: true,
		},
		{
			name:    "invalid format - non-numeric end",
			header:  "bytes=0-xyz",
			wantErr: true,
		},
		{
			name:    "invalid format - empty",
			header:  "",
			wantErr: true,
		},
		{
			name:    "invalid format - just bytes=",
			header:  "bytes=",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := parseRangeHeader(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRangeHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if start != tt.wantStart {
					t.Errorf("parseRangeHeader() start = %v, want %v", start, tt.wantStart)
				}
				if end != tt.wantEnd {
					t.Errorf("parseRangeHeader() end = %v, want %v", end, tt.wantEnd)
				}
			}
		})
	}
}

func TestToManifestResponse(t *testing.T) {
	// Test nil input
	result := ToManifestResponse(nil)
	if result != nil {
		t.Error("ToManifestResponse(nil) should return nil")
	}
}
