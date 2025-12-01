-- Migration: Drop pods table
-- Requirements: 3, 3.1, 3.2

DROP TRIGGER IF EXISTS update_pods_updated_at ON pods;
DROP FUNCTION IF EXISTS update_pods_updated_at_column();
DROP TABLE IF EXISTS pods;
