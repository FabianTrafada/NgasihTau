-- Migration: Create pod_downvotes table for negative trust indicators

CREATE TABLE pod_downvotes (
    user_id UUID NOT NULL,
    pod_id UUID NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, pod_id)
);

-- Index for querying downvotes by pod
CREATE INDEX idx_pod_downvotes_pod_id ON pod_downvotes(pod_id);

-- Index for querying user's downvoted pods
CREATE INDEX idx_pod_downvotes_user_id ON pod_downvotes(user_id);

COMMENT ON TABLE pod_downvotes IS 'Tracks user downvotes on pods as negative trust indicators';
COMMENT ON COLUMN pod_downvotes.user_id IS 'User ID from User Service (no FK due to cross-database)';
