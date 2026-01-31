-- Migration: Create offline_ceks table
-- Stores Content Encryption Keys for offline material access

CREATE TABLE offline_ceks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES offline_devices(id) ON DELETE CASCADE,
    encrypted_key BYTEA NOT NULL,
    key_version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT offline_ceks_key_version_check CHECK (key_version >= 1),
    CONSTRAINT unique_cek_composite UNIQUE (user_id, material_id, device_id)
);

-- Composite index for CEK lookups
CREATE INDEX idx_offline_ceks_composite ON offline_ceks(user_id, material_id, device_id);
CREATE INDEX idx_offline_ceks_device_id ON offline_ceks(device_id);
CREATE INDEX idx_offline_ceks_material_id ON offline_ceks(material_id);
CREATE INDEX idx_offline_ceks_key_version ON offline_ceks(key_version);

COMMENT ON TABLE offline_ceks IS 'Content Encryption Keys for offline material decryption';
COMMENT ON COLUMN offline_ceks.encrypted_key IS 'CEK encrypted with Key Encryption Key (KEK)';
COMMENT ON COLUMN offline_ceks.key_version IS 'Key version for rotation support';
