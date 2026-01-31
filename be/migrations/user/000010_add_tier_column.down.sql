-- Migration: Remove tier column from users table

DROP INDEX IF EXISTS idx_users_tier;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_tier_check;
ALTER TABLE users DROP COLUMN IF EXISTS tier;
