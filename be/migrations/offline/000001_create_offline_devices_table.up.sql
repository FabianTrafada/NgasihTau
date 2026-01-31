-- Migration: Create offline_devices table
-- Tracks registered user devices for offline access

CREATE TABLE offline_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    fingerprint VARCHAR(512) NOT NULL,
    name VARCHAR(255) NOT NULL,
    platform VARCHAR(50) NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT offline_devices_platform_check CHECK (platform IN ('ios', 'android', 'desktop')),
    CONSTRAINT offline_devices_fingerprint_length CHECK (LENGTH(fingerprint) >= 32),
    CONSTRAINT unique_user_fingerprint UNIQUE (user_id, fingerprint)
);

-- Indexes for frequently queried columns
CREATE INDEX idx_offline_devices_user_id ON offline_devices(user_id);
CREATE INDEX idx_offline_devices_fingerprint ON offline_devices(fingerprint);
CREATE INDEX idx_offline_devices_user_active ON offline_devices(user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_offline_devices_last_used ON offline_devices(last_used_at DESC);

COMMENT ON TABLE offline_devices IS 'Registered user devices for offline material access';
COMMENT ON COLUMN offline_devices.fingerprint IS 'Unique device fingerprint for binding licenses';
COMMENT ON COLUMN offline_devices.platform IS 'Device platform: ios, android, or desktop';
COMMENT ON COLUMN offline_devices.revoked_at IS 'Timestamp when device was deregistered';
