-- Migration: Create pod_upvotes table
-- Requirements: 5.1, 5.2, 5.3

CREATE TABLE pod_upvotes (
    user_id UUID NOT NULL,
    pod_id UUID NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (user_id, pod_id)
);

-- Indexes for frequently queried columns
CREATE INDEX idx_pod_upvotes_pod_id ON pod_upvotes(pod_id);
CREATE INDEX idx_pod_upvotes_user_id ON pod_upvotes(user_id);
CREATE INDEX idx_pod_upvotes_created_at ON pod_upvotes(created_at);

COMMENT ON TABLE pod_upvotes IS 'User upvotes for Knowledge Pods - trust indicator system';
COMMENT ON COLUMN pod_upvotes.user_id IS 'User ID who upvoted the pod';
COMMENT ON COLUMN pod_upvotes.pod_id IS 'Pod ID that was upvoted';
