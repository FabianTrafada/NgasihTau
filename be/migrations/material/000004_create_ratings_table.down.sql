-- Rollback: Drop ratings table
-- Requirements: 5.3

DROP TRIGGER IF EXISTS update_ratings_updated_at ON ratings;
DROP FUNCTION IF EXISTS update_ratings_updated_at_column();
DROP TABLE IF EXISTS ratings;
