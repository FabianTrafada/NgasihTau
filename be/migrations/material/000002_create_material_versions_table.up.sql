-- Migration: Create material_versions table
-- Requirements: 5.1

CREATE TABLE material_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    uploader_id UUID NOT NULL,
    changelog TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT material_versions_version_check CHECK (version >= 1),
    CONSTRAINT material_versions_file_size_check CHECK (file_size > 0),
    CONSTRAINT material_versions_unique UNIQUE (material_id, version)
);

-- Indexes for frequently queried columns
CREATE INDEX idx_material_versions_material_id ON material_versions(material_id);
CREATE INDEX idx_material_versions_uploader_id ON material_versions(uploader_id);
CREATE INDEX idx_material_versions_created_at ON material_versions(created_at);

-- Composite index for version history listing
CREATE INDEX idx_material_versions_material_id_version ON material_versions(material_id, version DESC);

COMMENT ON TABLE material_versions IS 'Version history for learning materials';
COMMENT ON COLUMN material_versions.material_id IS 'Reference to the parent material';
COMMENT ON COLUMN material_versions.version IS 'Version number (1, 2, 3, ...)';
COMMENT ON COLUMN material_versions.uploader_id IS 'User ID who uploaded this version (from User Service)';
COMMENT ON COLUMN material_versions.changelog IS 'Description of changes in this version';
