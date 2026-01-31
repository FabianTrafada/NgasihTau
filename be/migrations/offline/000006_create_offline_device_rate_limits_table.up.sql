-- Migration: Create offline_device_rate_limits table
-- Tracks device validation failures and blocking status

CREATE TABLE offline_device_rate_limits (
    device_id UUID PRIMARY KEY REFERENCES offline_devices(id) ON DELETE CASCADE,
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    blocked_until TIMESTAMP WITH TIME ZONE,
    last_attempt_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT offline_device_rate_limits_failed_attempts_check CHECK (failed_attempts >= 0)
);

-- Index for checking blocked devices
CREATE INDEX idx_offline_device_rate_limits_blocked ON offline_device_rate_limits(blocked_until) 
    WHERE blocked_until IS NOT NULL;
CREATE INDEX idx_offline_device_rate_limits_last_attempt ON offline_device_rate_limits(last_attempt_at);

COMMENT ON TABLE offline_device_rate_limits IS 'Rate limiting and blocking for device validation failures';
COMMENT ON COLUMN offline_device_rate_limits.failed_attempts IS 'Count of failed validation attempts';
COMMENT ON COLUMN offline_device_rate_limits.blocked_until IS 'Device blocked until this timestamp (null if not blocked)';
