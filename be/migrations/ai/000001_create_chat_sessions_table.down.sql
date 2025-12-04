-- Migration: Drop chat_sessions table

DROP TRIGGER IF EXISTS update_chat_sessions_updated_at ON chat_sessions;
DROP TABLE IF EXISTS chat_sessions;
