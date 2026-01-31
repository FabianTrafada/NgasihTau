-- Migration: Create offline_audit_logs table
-- Stores audit logs for security-sensitive operations

CREATE TABLE offline_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    device_id UUID,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id UUID NOT NULL,
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL DEFAULT true,
    error_code VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for audit log queries
CREATE INDEX idx_offline_audit_logs_user_id ON offline_audit_logs(user_id);
CREATE INDEX idx_offline_audit_logs_device_id ON offline_audit_logs(device_id);
CREATE INDEX idx_offline_audit_logs_created_at ON offline_audit_logs(created_at);
CREATE INDEX idx_offline_audit_logs_action ON offline_audit_logs(action);
CREATE INDEX idx_offline_audit_logs_resource ON offline_audit_logs(resource, resource_id);
CREATE INDEX idx_offline_audit_logs_success ON offline_audit_logs(success) WHERE success = false;

-- Partition hint: Consider partitioning by created_at for large deployments
COMMENT ON TABLE offline_audit_logs IS 'Audit trail for security-sensitive offline operations';
COMMENT ON COLUMN offline_audit_logs.action IS 'Action performed: key_generate, license_issue, license_validate, etc.';
COMMENT ON COLUMN offline_audit_logs.resource IS 'Resource type: license, device, cek, material';
COMMENT ON COLUMN offline_audit_logs.error_code IS 'Error code if operation failed';
