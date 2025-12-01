-- Migration: Create backup_codes table
-- Requirements: 1.1 (Two-Factor Authentication)

CREATE TABLE backup_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for frequently queried columns
CREATE INDEX idx_backup_codes_user_id ON backup_codes(user_id);
CREATE INDEX idx_backup_codes_user_unused ON backup_codes(user_id, used) WHERE used = FALSE;

COMMENT ON TABLE backup_codes IS 'Backup codes for 2FA account recovery';
COMMENT ON COLUMN backup_codes.code_hash IS 'Bcrypt hash of the backup code';
COMMENT ON COLUMN backup_codes.used IS 'Whether the backup code has been used';
