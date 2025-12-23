-- Rollback: Drop learning interests tables

DROP TABLE IF EXISTS user_learning_interests;
DROP TABLE IF EXISTS predefined_interests;

ALTER TABLE users DROP COLUMN IF EXISTS onboarding_completed;
