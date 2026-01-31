-- Migration: Create offline_licenses table
-- Stores licenses for offline content access authorization

CREATE TABLE offline_licenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES offline_devices(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    offline_grace_period INTERVAL NOT NULL DEFAULT '72 hours',
    last_validated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    nonce VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT offline_licenses_status_check CHECK (status IN ('active', 'expired', 'revoked')),
    CONSTRAINT offline_licenses_nonce_length CHECK (LENGTH(nonce) >= 32),
    CONSTRAINT unique_license_composite UNIQUE (user_id, material_id, device_id)
);

-- Indexes for license queries
CREATE INDEX idx_offline_licenses_user_id ON offline_licenses(user_id);
CREATE INDEX idx_offline_licenses_device_id ON offline_licenses(device_id);
CREATE INDEX idx_offline_licenses_material_id ON offline_licenses(material_id);
CREATE INDEX idx_offline_licenses_status ON offline_licenses(status);
CREATE INDEX idx_offline_licenses_expires_at ON offline_licenses(expires_at);
CREATE INDEX idx_offline_licenses_active ON offline_licenses(user_id, status) WHERE status = 'active';

COMMENT ON TABLE offline_licenses IS 'Licenses authorizing offline content access';
COMMENT ON COLUMN offline_licenses.offline_grace_period IS 'Maximum duration without online validation (default 72 hours)';
COMMENT ON COLUMN offline_licenses.nonce IS 'Unique nonce to prevent license cloning';
COMMENT ON COLUMN offline_licenses.last_validated_at IS 'Last online validation timestamp';
