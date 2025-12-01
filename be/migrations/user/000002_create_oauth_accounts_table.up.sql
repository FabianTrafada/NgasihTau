-- Migration: Create oauth_accounts table
-- Requirements: 1 (Google OAuth login)

CREATE TABLE oauth_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT oauth_accounts_provider_check CHECK (provider IN ('google')),
    CONSTRAINT oauth_accounts_provider_user_unique UNIQUE (provider, provider_user_id)
);

-- Indexes for frequently queried columns
CREATE INDEX idx_oauth_accounts_user_id ON oauth_accounts(user_id);
CREATE INDEX idx_oauth_accounts_provider ON oauth_accounts(provider);
CREATE INDEX idx_oauth_accounts_provider_user_id ON oauth_accounts(provider, provider_user_id);

COMMENT ON TABLE oauth_accounts IS 'OAuth provider accounts linked to users';
COMMENT ON COLUMN oauth_accounts.provider IS 'OAuth provider name (e.g., google)';
COMMENT ON COLUMN oauth_accounts.provider_user_id IS 'User ID from the OAuth provider';
