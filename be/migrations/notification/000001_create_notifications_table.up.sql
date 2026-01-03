-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE notifications (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                               user_id UUID NOT NULL,
                               type VARCHAR(50) NOT NULL,
                               title VARCHAR(255) NOT NULL,
                               message TEXT,
                               data JSONB,
                               read BOOLEAN NOT NULL DEFAULT FALSE,
                               created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

                               CONSTRAINT notifications_type_check CHECK (type IN ('pod_invite', 'new_material', 'comment_reply', 'new_follower', 'material_processed'))
);

-- Indexes for user_id and read status (as per task requirements)
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_user_id_read ON notifications(user_id, read);
CREATE INDEX idx_notifications_user_id_created_at ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

COMMENT ON TABLE notifications IS 'In-app notifications for users';
COMMENT ON COLUMN notifications.type IS 'Notification type: pod_invite, new_material, comment_reply, new_follower, material_processed';
COMMENT ON COLUMN notifications.data IS 'Additional notification data as JSON (pod_id, material_id, etc.)';
