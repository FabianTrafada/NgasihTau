-- Migration: Create teacher_verifications table

CREATE TABLE teacher_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    full_name VARCHAR(255) NOT NULL,
    id_number VARCHAR(100) NOT NULL,
    credential_type VARCHAR(50) NOT NULL,
    document_ref VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT teacher_verifications_user_unique UNIQUE(user_id),
    CONSTRAINT teacher_verifications_status_check CHECK (status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT teacher_verifications_credential_type_check CHECK (credential_type IN ('government_id', 'educator_card', 'professional_cert'))
);

-- Indexes for status and user_id
CREATE INDEX idx_teacher_verifications_status ON teacher_verifications(status);
CREATE INDEX idx_teacher_verifications_user_id ON teacher_verifications(user_id);

-- Trigger to auto-update updated_at
CREATE TRIGGER update_teacher_verifications_updated_at
    BEFORE UPDATE ON teacher_verifications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE teacher_verifications IS 'Teacher verification requests for role upgrade from student to teacher';
COMMENT ON COLUMN teacher_verifications.credential_type IS 'Type of verification document: government_id (KTP/NIK), educator_card (Kartu Pendidik), professional_cert (BNSP Certificate)';
COMMENT ON COLUMN teacher_verifications.document_ref IS 'Reference to verification document, not actual file content (for future government API integration)';
COMMENT ON COLUMN teacher_verifications.status IS 'Verification status: pending, approved, or rejected';
