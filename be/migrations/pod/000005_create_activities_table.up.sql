-- Migration: Create activities table
-- Requirements: 12

CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pod_id UUID NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT activities_action_check CHECK (action IN (
        'material_uploaded',
        'material_updated',
        'material_deleted',
        'collaborator_added',
        'collaborator_removed',
        'collaborator_verified',
        'pod_updated',
        'pod_forked'
    ))
);

-- Indexes for frequently queried columns
CREATE INDEX idx_activities_pod_id ON activities(pod_id);
CREATE INDEX idx_activities_user_id ON activities(user_id);
CREATE INDEX idx_activities_action ON activities(action);
CREATE INDEX idx_activities_created_at ON activities(created_at DESC);
CREATE INDEX idx_activities_metadata ON activities USING GIN(metadata);

-- Composite index for activity feed queries
CREATE INDEX idx_activities_pod_created ON activities(pod_id, created_at DESC);

COMMENT ON TABLE activities IS 'Activity log for Knowledge Pods';
COMMENT ON COLUMN activities.action IS 'Type of activity: material_uploaded, collaborator_added, pod_updated, etc.';
COMMENT ON COLUMN activities.metadata IS 'Additional context data for the activity (JSON)';
