-- Rollback: Drop ratings table

DROP TRIGGER IF EXISTS update_ratings_updated_at ON ratings;
DROP FUNCTION IF EXISTS update_ratings_updated_at_column();
DROP TABLE IF EXISTS ratings;
