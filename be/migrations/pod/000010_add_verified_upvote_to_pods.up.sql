-- Migration: Add is_verified and upvote_count columns to pods table
-- Purpose: Support trust indicators for knowledge pods (verified badge and upvote count)

-- Add is_verified column (true if created by teacher)
ALTER TABLE pods ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT FALSE;

-- Add upvote_count column (trust indicator separate from star_count)
ALTER TABLE pods ADD COLUMN upvote_count INTEGER NOT NULL DEFAULT 0;

-- Add constraint to ensure upvote_count is non-negative
ALTER TABLE pods ADD CONSTRAINT pods_upvote_count_check CHECK (upvote_count >= 0);

-- Index for filtering by verified status
CREATE INDEX idx_pods_is_verified ON pods(is_verified);

-- Index for sorting by upvote count (descending for popularity)
CREATE INDEX idx_pods_upvote_count ON pods(upvote_count DESC);

-- Composite index for trust score sorting (verified first, then by upvotes)
-- This supports queries that prioritize verified pods with high upvote counts
CREATE INDEX idx_pods_trust_score ON pods(is_verified DESC, upvote_count DESC);

-- Comments for documentation
COMMENT ON COLUMN pods.is_verified IS 'True if pod was created by a verified teacher';
COMMENT ON COLUMN pods.upvote_count IS 'Number of upvotes as trust indicator (separate from star_count which is for bookmarks)';
