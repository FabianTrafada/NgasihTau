-- Migration: Create offline_sync_state table
-- Tracks device sync state for offline/online transitions

CREATE TABLE offline_sync_state (
    device_id UUID PRIMARY KEY REFERENCES offline_devices(id) ON DELETE CASCADE,
    last_sync_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    sync_version INTEGER NOT NULL DEFAULT 1,
    pending_changes JSONB,
    
    CONSTRAINT offline_sync_state_version_check CHECK (sync_version >= 1)
);

-- Index for sync queries
CREATE INDEX idx_offline_sync_state_last_sync ON offline_sync_state(last_sync_at);

COMMENT ON TABLE offline_sync_state IS 'Device sync state for offline/online conflict resolution';
COMMENT ON COLUMN offline_sync_state.sync_version IS 'Incremental version for sync tracking';
COMMENT ON COLUMN offline_sync_state.pending_changes IS 'JSON of pending changes to sync';
