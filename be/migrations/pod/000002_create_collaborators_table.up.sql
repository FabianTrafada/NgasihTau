-- Migration: Create collaborators table
-- Requirements: 4

CREATE TABLE collaborators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pod_id UUID NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'viewer',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    invited_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT collaborators_pod_user_unique UNIQUE(pod_id, user_id),
    CONSTRAINT collaborators_role_check CHECK (role IN ('viewer', 'contributor', 'admin')),
    CONSTRAINT collaborators_status_check CHECK (status IN ('pending', 'pending_verification', 'verified'))
);

-- Indexes for frequently queried columns
CREATE INDEX idx_collaborators_pod_id ON collaborators(pod_id);
CREATE INDEX idx_collaborators_user_id ON collaborators(user_id);
CREATE INDEX idx_collaborators_status ON collaborators(status);
CREATE INDEX idx_collaborators_invited_by ON collaborators(invited_by);

-- Trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_collaborators_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_collaborators_updated_at
    BEFORE UPDATE ON collaborators
    FOR EACH ROW
    EXECUTE FUNCTION update_collaborators_updated_at_column();

COMMENT ON TABLE collaborators IS 'Pod collaborators with roles and verification status';
COMMENT ON COLUMN collaborators.role IS 'Collaborator role: viewer, contributor, or admin';
COMMENT ON COLUMN collaborators.status IS 'Verification status: pending, pending_verification, or verified';
COMMENT ON COLUMN collaborators.invited_by IS 'User ID who invited this collaborator';
