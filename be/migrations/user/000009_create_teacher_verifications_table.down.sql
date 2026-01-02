-- Rollback: Drop teacher_verifications table

DROP TRIGGER IF EXISTS update_teacher_verifications_updated_at ON teacher_verifications;
DROP TABLE IF EXISTS teacher_verifications;
