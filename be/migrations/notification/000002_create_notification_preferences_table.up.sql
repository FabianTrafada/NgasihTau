CREATE TABLE notification_preferences (
                                          user_id UUID PRIMARY KEY,
                                          email_pod_invite BOOLEAN NOT NULL DEFAULT TRUE,
                                          email_new_material BOOLEAN NOT NULL DEFAULT TRUE,
                                          email_comment_reply BOOLEAN NOT NULL DEFAULT TRUE,
                                          inapp_pod_invite BOOLEAN NOT NULL DEFAULT TRUE,
                                          inapp_new_material BOOLEAN NOT NULL DEFAULT TRUE,
                                          inapp_comment_reply BOOLEAN NOT NULL DEFAULT TRUE,
                                          updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_notification_preferences_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER update_notification_preferences_updated_at
    BEFORE UPDATE ON notification_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_notification_preferences_updated_at();

COMMENT ON TABLE notification_preferences IS 'User preferences for email and in-app notifications';
COMMENT ON COLUMN notification_preferences.email_pod_invite IS 'Receive email for pod invitations';
COMMENT ON COLUMN notification_preferences.inapp_pod_invite IS 'Receive in-app notification for pod invitations';
