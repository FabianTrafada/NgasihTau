-- Migration: Create materials table

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pod_id UUID NOT NULL,
    uploader_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    file_type VARCHAR(50) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    current_version INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'processing',
    view_count INTEGER NOT NULL DEFAULT 0,
    download_count INTEGER NOT NULL DEFAULT 0,
    average_rating DECIMAL(2,1) NOT NULL DEFAULT 0,
    rating_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT materials_file_type_check CHECK (file_type IN ('pdf', 'docx', 'pptx')),
    CONSTRAINT materials_status_check CHECK (status IN ('processing', 'ready', 'processing_failed')),
    CONSTRAINT materials_current_version_check CHECK (current_version >= 1),
    CONSTRAINT materials_view_count_check CHECK (view_count >= 0),
    CONSTRAINT materials_download_count_check CHECK (download_count >= 0),
    CONSTRAINT materials_average_rating_check CHECK (average_rating >= 0 AND average_rating <= 5),
    CONSTRAINT materials_rating_count_check CHECK (rating_count >= 0),
    CONSTRAINT materials_file_size_check CHECK (file_size > 0)
);

-- Indexes for frequently queried columns
CREATE INDEX idx_materials_pod_id ON materials(pod_id);
CREATE INDEX idx_materials_uploader_id ON materials(uploader_id);
CREATE INDEX idx_materials_status ON materials(status);
CREATE INDEX idx_materials_file_type ON materials(file_type);
CREATE INDEX idx_materials_deleted_at ON materials(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_materials_created_at ON materials(created_at);
CREATE INDEX idx_materials_average_rating ON materials(average_rating DESC);
CREATE INDEX idx_materials_view_count ON materials(view_count DESC);
CREATE INDEX idx_materials_download_count ON materials(download_count DESC);

-- Composite index for pod materials listing
CREATE INDEX idx_materials_pod_id_created_at ON materials(pod_id, created_at DESC);

-- Trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_materials_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_materials_updated_at
    BEFORE UPDATE ON materials
    FOR EACH ROW
    EXECUTE FUNCTION update_materials_updated_at_column();

COMMENT ON TABLE materials IS 'Learning materials uploaded to Knowledge Pods';
COMMENT ON COLUMN materials.pod_id IS 'Reference to the Knowledge Pod (from Pod Service)';
COMMENT ON COLUMN materials.uploader_id IS 'User ID of the uploader (from User Service)';
COMMENT ON COLUMN materials.file_type IS 'File type: pdf, docx, or pptx';
COMMENT ON COLUMN materials.status IS 'Processing status: processing, ready, or processing_failed';
COMMENT ON COLUMN materials.current_version IS 'Current version number of the material';
COMMENT ON COLUMN materials.average_rating IS 'Calculated average rating (1-5 scale)';
