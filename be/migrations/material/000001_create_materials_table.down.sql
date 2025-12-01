-- Rollback: Drop materials table
-- Requirements: 5, 5.1, 5.2, 5.3, 5.4

DROP TRIGGER IF EXISTS update_materials_updated_at ON materials;
DROP FUNCTION IF EXISTS update_materials_updated_at_column();
DROP TABLE IF EXISTS materials;
