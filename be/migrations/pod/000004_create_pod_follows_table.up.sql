-- Migration: Create pod_follows table
-- Requirements: 12

CREATE TABLE pod_follows (
    user_id UUID NOT NULL,
    pod_id UUID NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (user_id, pod_id)
);

-- Indexes for frequently queried columns
CREATE INDEX idx_pod_follows_pod_id ON pod_follows(pod_id);
CREATE INDEX idx_pod_follows_user_id ON pod_follows(user_id);
CREATE INDEX idx_pod_follows_created_at ON pod_follows(created_at);

COMMENT ON TABLE pod_follows IS 'User follows for Knowledge Pods to receive activity updates';
COMMENT ON COLUMN pod_follows.user_id IS 'User ID who follows the pod';
COMMENT ON COLUMN pod_follows.pod_id IS 'Pod ID that is being followed';
