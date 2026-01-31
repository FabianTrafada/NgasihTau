-- Rollback: Drop comments table

DROP TRIGGER IF EXISTS update_comments_updated_at ON comments;
DROP FUNCTION IF EXISTS update_comments_updated_at_column();
DROP TABLE IF EXISTS comments;
