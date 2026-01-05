-- Migration: Create chat_sessions table

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create helper function for updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE chat_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    material_id UUID,
    pod_id UUID,
    mode VARCHAR(20) NOT NULL DEFAULT 'material',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT chat_sessions_mode_check CHECK (mode IN ('material', 'pod')),
    CONSTRAINT chat_sessions_context_check CHECK (
        (mode = 'material' AND material_id IS NOT NULL) OR
        (mode = 'pod' AND pod_id IS NOT NULL)
    )
);

-- Indexes for frequently queried columns
CREATE INDEX idx_chat_sessions_user_id ON chat_sessions(user_id);
CREATE INDEX idx_chat_sessions_material_id ON chat_sessions(material_id) WHERE material_id IS NOT NULL;
CREATE INDEX idx_chat_sessions_pod_id ON chat_sessions(pod_id) WHERE pod_id IS NOT NULL;
CREATE INDEX idx_chat_sessions_user_material ON chat_sessions(user_id, material_id) WHERE material_id IS NOT NULL;
CREATE INDEX idx_chat_sessions_user_pod ON chat_sessions(user_id, pod_id) WHERE pod_id IS NOT NULL;
CREATE INDEX idx_chat_sessions_created_at ON chat_sessions(created_at);

-- Trigger to auto-update updated_at
CREATE TRIGGER update_chat_sessions_updated_at
    BEFORE UPDATE ON chat_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE chat_sessions IS 'AI chat sessions for material or pod context';
COMMENT ON COLUMN chat_sessions.mode IS 'Chat mode: material (single) or pod (pod-wide)';
COMMENT ON COLUMN chat_sessions.material_id IS 'Reference to material for single-material chat';
COMMENT ON COLUMN chat_sessions.pod_id IS 'Reference to pod for pod-wide chat';
