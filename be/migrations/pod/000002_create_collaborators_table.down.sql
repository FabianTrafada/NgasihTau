-- Migration: Drop collaborators table

DROP TRIGGER IF EXISTS update_collaborators_updated_at ON collaborators;
DROP FUNCTION IF EXISTS update_collaborators_updated_at_column();
DROP TABLE IF EXISTS collaborators;
