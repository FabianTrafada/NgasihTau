-- Rollback: Remove is_verified and upvote_count columns from pods table
-- Requirements: 1.4, 2.4, 6.1, 6.2

-- Drop indexes first
DROP INDEX IF EXISTS idx_pods_trust_score;
DROP INDEX IF EXISTS idx_pods_upvote_count;
DROP INDEX IF EXISTS idx_pods_is_verified;

-- Drop constraint
ALTER TABLE pods DROP CONSTRAINT IF EXISTS pods_upvote_count_check;

-- Drop columns
ALTER TABLE pods DROP COLUMN IF EXISTS upvote_count;
ALTER TABLE pods DROP COLUMN IF EXISTS is_verified;
