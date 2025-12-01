-- Migration: Create pod_stars table
-- Requirements: 3.2

CREATE TABLE pod_stars (
    user_id UUID NOT NULL,
    pod_id UUID NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (user_id, pod_id)
);

-- Indexes for frequently queried columns
CREATE INDEX idx_pod_stars_pod_id ON pod_stars(pod_id);
CREATE INDEX idx_pod_stars_user_id ON pod_stars(user_id);
CREATE INDEX idx_pod_stars_created_at ON pod_stars(created_at);

COMMENT ON TABLE pod_stars IS 'User stars/bookmarks for Knowledge Pods';
COMMENT ON COLUMN pod_stars.user_id IS 'User ID who starred the pod';
COMMENT ON COLUMN pod_stars.pod_id IS 'Pod ID that was starred';
