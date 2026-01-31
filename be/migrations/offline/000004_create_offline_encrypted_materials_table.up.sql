-- Migration: Create offline_encrypted_materials table
-- Stores encrypted file metadata and manifest

CREATE TABLE offline_encrypted_materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    cek_id UUID NOT NULL REFERENCES offline_ceks(id) ON DELETE CASCADE,
    manifest JSONB NOT NULL,
    encrypted_file_url VARCHAR(1024) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_encrypted_material UNIQUE (material_id, cek_id)
);

-- Indexes for encrypted material lookups
CREATE INDEX idx_offline_encrypted_materials_material_id ON offline_encrypted_materials(material_id);
CREATE INDEX idx_offline_encrypted_materials_cek_id ON offline_encrypted_materials(cek_id);
CREATE INDEX idx_offline_encrypted_materials_created_at ON offline_encrypted_materials(created_at);

COMMENT ON TABLE offline_encrypted_materials IS 'Cached encrypted material metadata';
COMMENT ON COLUMN offline_encrypted_materials.manifest IS 'JSON manifest with chunk info, hashes, and encryption metadata';
COMMENT ON COLUMN offline_encrypted_materials.encrypted_file_url IS 'MinIO URL for the encrypted file';
