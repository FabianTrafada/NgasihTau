package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewMetrics(t *testing.T) {
	// Reset default registry for clean test
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg

	m := NewMetrics()

	if m == nil {
		t.Fatal("NewMetrics returned nil")
	}

	// Verify counter metrics are initialized
	if m.DeviceRegistrations == nil {
		t.Error("DeviceRegistrations counter is nil")
	}
	if m.LicensesIssued == nil {
		t.Error("LicensesIssued counter is nil")
	}
	if m.DownloadsTotal == nil {
		t.Error("DownloadsTotal counter is nil")
	}

	// Verify histogram metrics are initialized
	if m.DownloadLatency == nil {
		t.Error("DownloadLatency histogram is nil")
	}
	if m.RequestDuration == nil {
		t.Error("RequestDuration histogram is nil")
	}

	// Verify gauge metrics are initialized
	if m.ActiveDevices == nil {
		t.Error("ActiveDevices gauge is nil")
	}
	if m.ActiveLicenses == nil {
		t.Error("ActiveLicenses gauge is nil")
	}
}

func TestMetrics_RecordDownload(t *testing.T) {
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg

	m := NewMetrics()

	// Record successful download
	m.RecordDownload("material-123", true, 0.5)

	// Verify counter incremented
	count := testutil.ToFloat64(m.DownloadsTotal)
	if count != 1 {
		t.Errorf("Expected DownloadsTotal=1, got %f", count)
	}

	// Record failed download
	m.RecordDownload("material-456", false, 1.0)

	failedCount := testutil.ToFloat64(m.DownloadsFailed)
	if failedCount != 1 {
		t.Errorf("Expected DownloadsFailed=1, got %f", failedCount)
	}
}

func TestMetrics_RecordLicenseOperation(t *testing.T) {
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg

	m := NewMetrics()

	tests := []struct {
		operation string
		counter   func() float64
	}{
		{"issue", func() float64 { return testutil.ToFloat64(m.LicensesIssued) }},
		{"validate", func() float64 { return testutil.ToFloat64(m.LicensesValidated) }},
		{"renew", func() float64 { return testutil.ToFloat64(m.LicensesRenewed) }},
		{"revoke", func() float64 { return testutil.ToFloat64(m.LicensesRevoked) }},
	}

	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			before := tt.counter()
			m.RecordLicenseOperation(tt.operation, 0.01)
			after := tt.counter()

			if after != before+1 {
				t.Errorf("Expected counter to increment by 1, got %f -> %f", before, after)
			}
		})
	}
}

func TestMetrics_RecordEncryptionJob(t *testing.T) {
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg

	m := NewMetrics()

	// Queue a job
	m.RecordEncryptionJob("queued", 0)
	if testutil.ToFloat64(m.EncryptionJobsQueued) != 1 {
		t.Error("EncryptionJobsQueued should be 1")
	}
	if testutil.ToFloat64(m.PendingEncryptionJobs) != 1 {
		t.Error("PendingEncryptionJobs should be 1")
	}

	// Complete job successfully
	m.RecordEncryptionJob("success", 5.0)
	if testutil.ToFloat64(m.EncryptionJobsSuccess) != 1 {
		t.Error("EncryptionJobsSuccess should be 1")
	}
	if testutil.ToFloat64(m.PendingEncryptionJobs) != 0 {
		t.Error("PendingEncryptionJobs should be 0 after success")
	}

	// Queue and fail a job
	m.RecordEncryptionJob("queued", 0)
	m.RecordEncryptionJob("failed", 2.0)
	if testutil.ToFloat64(m.EncryptionJobsFailed) != 1 {
		t.Error("EncryptionJobsFailed should be 1")
	}
}

func TestMetrics_RecordSecurityEvent(t *testing.T) {
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg

	m := NewMetrics()

	// Record replay attack
	m.RecordSecurityEvent("replay_attack")
	if testutil.ToFloat64(m.SecurityViolations) != 1 {
		t.Error("SecurityViolations should be 1")
	}
	if testutil.ToFloat64(m.ReplayAttacksBlocked) != 1 {
		t.Error("ReplayAttacksBlocked should be 1")
	}

	// Record device blocked
	m.RecordSecurityEvent("device_blocked")
	if testutil.ToFloat64(m.DevicesBlocked) != 1 {
		t.Error("DevicesBlocked should be 1")
	}
	if testutil.ToFloat64(m.BlockedDevices) != 1 {
		t.Error("BlockedDevices gauge should be 1")
	}

	// Record rate limit
	m.RecordSecurityEvent("rate_limit")
	if testutil.ToFloat64(m.RateLimitExceeded) != 1 {
		t.Error("RateLimitExceeded should be 1")
	}
}

func TestMetrics_UpdateGauges(t *testing.T) {
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg

	m := NewMetrics()

	// Update device count
	m.UpdateDeviceCount(10)
	if testutil.ToFloat64(m.ActiveDevices) != 10 {
		t.Error("ActiveDevices should be 10")
	}

	// Update license count
	m.UpdateLicenseCount(25)
	if testutil.ToFloat64(m.ActiveLicenses) != 25 {
		t.Error("ActiveLicenses should be 25")
	}

	// Update cache hit ratio
	m.UpdateCacheHitRatio(0.85)
	if testutil.ToFloat64(m.CacheHitRatio) != 0.85 {
		t.Error("CacheHitRatio should be 0.85")
	}
}

func TestMetrics_DeviceUnblocked(t *testing.T) {
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg

	m := NewMetrics()

	// Block a device
	m.RecordSecurityEvent("device_blocked")
	if testutil.ToFloat64(m.BlockedDevices) != 1 {
		t.Error("BlockedDevices should be 1")
	}

	// Unblock device
	m.DeviceUnblocked()
	if testutil.ToFloat64(m.BlockedDevices) != 0 {
		t.Error("BlockedDevices should be 0 after unblock")
	}
}

func TestMetrics_RecordRequest(t *testing.T) {
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg

	m := NewMetrics()

	// Record various status codes
	m.RecordRequest("GET", "/api/v1/offline/devices", 200, 0.05)
	m.RecordRequest("POST", "/api/v1/offline/licenses", 201, 0.1)
	m.RecordRequest("GET", "/api/v1/offline/download", 404, 0.02)
	m.RecordRequest("POST", "/api/v1/offline/sync", 500, 0.5)

	// Verify histogram has observations (can't easily check labels without more setup)
	// Just verify no panic occurred
}

func TestMetrics_RecordAuditEvent(t *testing.T) {
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg

	m := NewMetrics()

	m.RecordAuditEvent("device_register", true)
	m.RecordAuditEvent("license_issue", true)
	m.RecordAuditEvent("download_failed", false)

	// Verify no panic - counter vec with labels
}

func TestMetrics_RecordError(t *testing.T) {
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg

	m := NewMetrics()

	m.RecordError("device_not_found")
	m.RecordError("license_expired")
	m.RecordError("rate_limit_exceeded")

	// Verify no panic - counter vec with labels
}
