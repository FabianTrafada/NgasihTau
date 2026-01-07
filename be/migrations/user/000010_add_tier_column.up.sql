-- Migration: Add tier column to users table

ALTER TABLE users ADD COLUMN tier VARCHAR(20) NOT NULL DEFAULT 'free';

-- Add constraint to ensure valid tier values
ALTER TABLE users ADD CONSTRAINT users_tier_check CHECK (tier IN ('free', 'premium', 'pro'));

-- Index for tier column for efficient filtering
CREATE INDEX idx_users_tier ON users(tier);

COMMENT ON COLUMN users.tier IS 'User subscription tier: free, premium, or pro';
