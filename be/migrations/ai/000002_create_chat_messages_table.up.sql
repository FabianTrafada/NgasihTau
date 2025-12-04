-- Migration: Create chat_messages table
-- Requirements: 9, 9.3

CREATE TABLE chat_messages (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                               session_id UUID NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
                               role VARCHAR(20) NOT NULL,
                               content TEXT NOT NULL,
                               sources JSONB,
                               feedback VARCHAR(20),
                               feedback_text TEXT,
                               created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

                               CONSTRAINT chat_messages_role_check CHECK (role IN ('user', 'assistant')),
                               CONSTRAINT chat_messages_feedback_check CHECK (feedback IS NULL OR feedback IN ('thumbs_up', 'thumbs_down'))
);

-- Indexes for frequently queried columns
CREATE INDEX idx_chat_messages_session_id ON chat_messages(session_id);
CREATE INDEX idx_chat_messages_session_created ON chat_messages(session_id, created_at);
CREATE INDEX idx_chat_messages_feedback ON chat_messages(feedback) WHERE feedback IS NOT NULL;
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at);

COMMENT ON TABLE chat_messages IS 'Individual messages within AI chat sessions';
COMMENT ON COLUMN chat_messages.role IS 'Message role: user or assistant';
COMMENT ON COLUMN chat_messages.sources IS 'JSON array of chunk sources used for the response';
COMMENT ON COLUMN chat_messages.feedback IS 'User feedback: thumbs_up or thumbs_down';
