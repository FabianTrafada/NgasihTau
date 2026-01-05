-- Migration: Create upload_requests table
-- Purpose: Teacher-to-teacher collaboration for uploading materials to other teachers' pods

CREATE TABLE upload_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id UUID NOT NULL,
    pod_id UUID NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    pod_owner_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    message TEXT,
    rejection_reason TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT upload_requests_unique_requester_pod UNIQUE(requester_id, pod_id),
    CONSTRAINT upload_requests_status_check CHECK (status IN ('pending', 'approved', 'rejected', 'revoked'))
);

-- Indexes for pod_owner and requester
CREATE INDEX idx_upload_requests_pod_owner ON upload_requests(pod_owner_id, status);
CREATE INDEX idx_upload_requests_requester ON upload_requests(requester_id);
CREATE INDEX idx_upload_requests_pod_id ON upload_requests(pod_id);
CREATE INDEX idx_upload_requests_status ON upload_requests(status);
CREATE INDEX idx_upload_requests_created_at ON upload_requests(created_at);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_upload_requests_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_upload_requests_updated_at
    BEFORE UPDATE ON upload_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_upload_requests_updated_at();

-- Comments for documentation
COMMENT ON TABLE upload_requests IS 'Upload requests from teachers to upload materials to other teachers pods';
COMMENT ON COLUMN upload_requests.requester_id IS 'Teacher ID requesting upload permission';
COMMENT ON COLUMN upload_requests.pod_id IS 'Target pod ID for upload permission';
COMMENT ON COLUMN upload_requests.pod_owner_id IS 'Pod owner ID for quick lookup and filtering';
COMMENT ON COLUMN upload_requests.status IS 'Request status: pending, approved, rejected, revoked';
COMMENT ON COLUMN upload_requests.message IS 'Optional message from requester explaining the request';
COMMENT ON COLUMN upload_requests.rejection_reason IS 'Optional reason provided when request is rejected';
COMMENT ON COLUMN upload_requests.expires_at IS 'Optional expiration time for approved upload permission';
