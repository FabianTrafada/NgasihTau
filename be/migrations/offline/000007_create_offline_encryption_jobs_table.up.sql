-- Migration: Create offline_encryption_jobs table
-- Tracks async encryption job processing

CREATE TABLE offline_encryption_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    device_id UUID NOT NULL REFERENCES offline_devices(id) ON DELETE CASCADE,
    license_id UUID NOT NULL REFERENCES offline_licenses(id) ON DELETE CASCADE,
    priority INTEGER NOT NULL DEFAULT 2,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT offline_encryption_jobs_priority_check CHECK (priority BETWEEN 1 AND 3),
    CONSTRAINT offline_encryption_jobs_status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    CONSTRAINT offline_encryption_jobs_retry_count_check CHECK (retry_count >= 0)
);

-- Indexes for job queue processing
CREATE INDEX idx_offline_encryption_jobs_status ON offline_encryption_jobs(status);
CREATE INDEX idx_offline_encryption_jobs_material ON offline_encryption_jobs(material_id);
CREATE INDEX idx_offline_encryption_jobs_created ON offline_encryption_jobs(created_at);
CREATE INDEX idx_offline_encryption_jobs_pending ON offline_encryption_jobs(priority, created_at) 
    WHERE status = 'pending';
CREATE INDEX idx_offline_encryption_jobs_user ON offline_encryption_jobs(user_id);

COMMENT ON TABLE offline_encryption_jobs IS 'Async encryption job queue for large file processing';
COMMENT ON COLUMN offline_encryption_jobs.priority IS 'Job priority: 1=high, 2=normal, 3=low';
COMMENT ON COLUMN offline_encryption_jobs.status IS 'Job status: pending, processing, completed, failed';
COMMENT ON COLUMN offline_encryption_jobs.retry_count IS 'Number of retry attempts';
