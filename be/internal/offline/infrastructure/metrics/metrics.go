// Package metrics provides Prometheus metrics for the Offline Material Service.
// Implements Phase 6: Monitoring & Observability.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "ngasihtau"
	subsystem = "offline"
)

// Metrics contains all Prometheus metrics for the Offline Material Service.
type Metrics struct {
	// Counter metrics
	DeviceRegistrations   prometheus.Counter
	DeviceDeregistrations prometheus.Counter
	LicensesIssued        prometheus.Counter
	LicensesValidated     prometheus.Counter
	LicensesRenewed       prometheus.Counter
	LicensesRevoked       prometheus.Counter
	DownloadsTotal        prometheus.Counter
	DownloadsFailed       prometheus.Counter
	EncryptionJobsQueued  prometheus.Counter
	EncryptionJobsSuccess prometheus.Counter
	EncryptionJobsFailed  prometheus.Counter
	RateLimitExceeded     prometheus.Counter
	SecurityViolations    prometheus.Counter
	ReplayAttacksBlocked  prometheus.Counter
	DevicesBlocked        prometheus.Counter

	// Counter vectors (with labels)
	DownloadsByMaterial *prometheus.CounterVec
	ErrorsByType        *prometheus.CounterVec
	AuditEvents         *prometheus.CounterVec

	// Histogram metrics
	DownloadLatency       prometheus.Histogram
	LicenseValidationTime prometheus.Histogram
	EncryptionDuration    prometheus.Histogram
	DecryptionDuration    prometheus.Histogram
	KeyDerivationTime     prometheus.Histogram
	RequestDuration       *prometheus.HistogramVec

	// Gauge metrics
	ActiveDevices         prometheus.Gauge
	ActiveLicenses        prometheus.Gauge
	PendingEncryptionJobs prometheus.Gauge
	BlockedDevices        prometheus.Gauge
	CacheHitRatio         prometheus.Gauge
}

// NewMetrics creates and registers all Prometheus metrics.
func NewMetrics() *Metrics {
	m := &Metrics{
		// Counter metrics
		DeviceRegistrations: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "device_registrations_total",
			Help:      "Total number of device registrations",
		}),
		DeviceDeregistrations: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "device_deregistrations_total",
			Help:      "Total number of device deregistrations",
		}),
		LicensesIssued: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "licenses_issued_total",
			Help:      "Total number of licenses issued",
		}),
		LicensesValidated: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "licenses_validated_total",
			Help:      "Total number of license validations",
		}),
		LicensesRenewed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "licenses_renewed_total",
			Help:      "Total number of license renewals",
		}),
		LicensesRevoked: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "licenses_revoked_total",
			Help:      "Total number of licenses revoked",
		}),
		DownloadsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "downloads_total",
			Help:      "Total number of successful downloads",
		}),
		DownloadsFailed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "downloads_failed_total",
			Help:      "Total number of failed downloads",
		}),
		EncryptionJobsQueued: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "encryption_jobs_queued_total",
			Help:      "Total number of encryption jobs queued",
		}),
		EncryptionJobsSuccess: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "encryption_jobs_success_total",
			Help:      "Total number of successful encryption jobs",
		}),
		EncryptionJobsFailed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "encryption_jobs_failed_total",
			Help:      "Total number of failed encryption jobs",
		}),
		RateLimitExceeded: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "rate_limit_exceeded_total",
			Help:      "Total number of rate limit exceeded events",
		}),
		SecurityViolations: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "security_violations_total",
			Help:      "Total number of security violations detected",
		}),
		ReplayAttacksBlocked: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "replay_attacks_blocked_total",
			Help:      "Total number of replay attacks blocked",
		}),
		DevicesBlocked: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "devices_blocked_total",
			Help:      "Total number of devices blocked due to validation failures",
		}),

		// Counter vectors
		DownloadsByMaterial: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "downloads_by_material_total",
			Help:      "Total downloads per material",
		}, []string{"material_id"}),
		ErrorsByType: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "errors_by_type_total",
			Help:      "Total errors by error type",
		}, []string{"error_type"}),
		AuditEvents: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "audit_events_total",
			Help:      "Total audit events by action",
		}, []string{"action", "success"}),
	}

	// Initialize histogram metrics separately to avoid line length issues
	m.initHistograms()
	m.initGauges()

	return m
}


// initHistograms initializes histogram metrics.
func (m *Metrics) initHistograms() {
	m.DownloadLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "download_latency_seconds",
		Help:      "Download preparation latency in seconds",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	})

	m.LicenseValidationTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "license_validation_seconds",
		Help:      "License validation duration in seconds",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
	})

	m.EncryptionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "encryption_duration_seconds",
		Help:      "Material encryption duration in seconds",
		Buckets:   []float64{0.1, 0.5, 1, 2.5, 5, 10, 30, 60, 120},
	})

	m.DecryptionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "decryption_duration_seconds",
		Help:      "Material decryption duration in seconds",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	})

	m.KeyDerivationTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "key_derivation_seconds",
		Help:      "Key derivation duration in seconds",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
	})

	m.RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_duration_seconds",
		Help:      "HTTP request duration in seconds",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"method", "endpoint", "status"})
}

// initGauges initializes gauge metrics.
func (m *Metrics) initGauges() {
	m.ActiveDevices = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "active_devices",
		Help:      "Current number of active registered devices",
	})

	m.ActiveLicenses = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "active_licenses",
		Help:      "Current number of active (non-expired) licenses",
	})

	m.PendingEncryptionJobs = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pending_encryption_jobs",
		Help:      "Current number of pending encryption jobs",
	})

	m.BlockedDevices = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "blocked_devices",
		Help:      "Current number of blocked devices",
	})

	m.CacheHitRatio = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "cache_hit_ratio",
		Help:      "Cache hit ratio (0-1)",
	})
}

// RecordDownload records a download event.
func (m *Metrics) RecordDownload(materialID string, success bool, latencySeconds float64) {
	if success {
		m.DownloadsTotal.Inc()
		m.DownloadsByMaterial.WithLabelValues(materialID).Inc()
	} else {
		m.DownloadsFailed.Inc()
	}
	m.DownloadLatency.Observe(latencySeconds)
}

// RecordLicenseOperation records a license operation.
func (m *Metrics) RecordLicenseOperation(operation string, durationSeconds float64) {
	switch operation {
	case "issue":
		m.LicensesIssued.Inc()
	case "validate":
		m.LicensesValidated.Inc()
		m.LicenseValidationTime.Observe(durationSeconds)
	case "renew":
		m.LicensesRenewed.Inc()
	case "revoke":
		m.LicensesRevoked.Inc()
	}
}

// RecordEncryptionJob records an encryption job event.
func (m *Metrics) RecordEncryptionJob(status string, durationSeconds float64) {
	switch status {
	case "queued":
		m.EncryptionJobsQueued.Inc()
		m.PendingEncryptionJobs.Inc()
	case "success":
		m.EncryptionJobsSuccess.Inc()
		m.PendingEncryptionJobs.Dec()
		m.EncryptionDuration.Observe(durationSeconds)
	case "failed":
		m.EncryptionJobsFailed.Inc()
		m.PendingEncryptionJobs.Dec()
	}
}

// RecordSecurityEvent records a security-related event.
func (m *Metrics) RecordSecurityEvent(eventType string) {
	m.SecurityViolations.Inc()
	switch eventType {
	case "replay_attack":
		m.ReplayAttacksBlocked.Inc()
	case "device_blocked":
		m.DevicesBlocked.Inc()
		m.BlockedDevices.Inc()
	case "rate_limit":
		m.RateLimitExceeded.Inc()
	}
}

// RecordAuditEvent records an audit event.
func (m *Metrics) RecordAuditEvent(action string, success bool) {
	successStr := "true"
	if !success {
		successStr = "false"
	}
	m.AuditEvents.WithLabelValues(action, successStr).Inc()
}

// RecordError records an error by type.
func (m *Metrics) RecordError(errorType string) {
	m.ErrorsByType.WithLabelValues(errorType).Inc()
}

// RecordRequest records an HTTP request.
func (m *Metrics) RecordRequest(method, endpoint string, status int, durationSeconds float64) {
	statusStr := "2xx"
	switch {
	case status >= 500:
		statusStr = "5xx"
	case status >= 400:
		statusStr = "4xx"
	case status >= 300:
		statusStr = "3xx"
	}
	m.RequestDuration.WithLabelValues(method, endpoint, statusStr).Observe(durationSeconds)
}

// UpdateDeviceCount updates the active device gauge.
func (m *Metrics) UpdateDeviceCount(count int) {
	m.ActiveDevices.Set(float64(count))
}

// UpdateLicenseCount updates the active license gauge.
func (m *Metrics) UpdateLicenseCount(count int) {
	m.ActiveLicenses.Set(float64(count))
}

// UpdateCacheHitRatio updates the cache hit ratio gauge.
func (m *Metrics) UpdateCacheHitRatio(ratio float64) {
	m.CacheHitRatio.Set(ratio)
}

// DeviceUnblocked decrements the blocked devices gauge.
func (m *Metrics) DeviceUnblocked() {
	m.BlockedDevices.Dec()
}
