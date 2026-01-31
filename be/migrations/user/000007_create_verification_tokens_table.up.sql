-- Create verification_tokens table for email verification and password reset
-- Implements requirement 1.2: Email Verification & Password Reset
CREATE TABLE IF NOT EXISTS verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    token_type VARCHAR(50) NOT NULL, -- 'email_verification' or 'password_reset'
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    
    -- Ensure unique token hash
    CONSTRAINT unique_token_hash UNIQUE (token_hash)
);

-- Index for finding tokens by user and type
CREATE INDEX idx_verification_tokens_user_type ON verification_tokens(user_id, token_type);

-- Index for cleanup of expired tokens
CREATE INDEX idx_verification_tokens_expires_at ON verification_tokens(expires_at);

-- Index for finding unused tokens
CREATE INDEX idx_verification_tokens_used_at ON verification_tokens(used_at) WHERE used_at IS NULL;
