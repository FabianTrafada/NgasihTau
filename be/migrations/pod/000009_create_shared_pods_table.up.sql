-- Migration: Create shared_pods table
-- Requirements: 7.2
-- Purpose: Teacher shares pod with student for guided learning

CREATE TABLE shared_pods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pod_id UUID NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    teacher_id UUID NOT NULL,
    student_id UUID NOT NULL,
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT shared_pods_unique_pod_student UNIQUE(pod_id, student_id)
);

-- Indexes for student and teacher as specified in task requirements
CREATE INDEX idx_shared_pods_student ON shared_pods(student_id);
CREATE INDEX idx_shared_pods_teacher ON shared_pods(teacher_id);
CREATE INDEX idx_shared_pods_pod_id ON shared_pods(pod_id);
CREATE INDEX idx_shared_pods_created_at ON shared_pods(created_at);

-- Comments for documentation
COMMENT ON TABLE shared_pods IS 'Pods shared by teachers with students for guided learning';
COMMENT ON COLUMN shared_pods.pod_id IS 'Pod ID that is being shared';
COMMENT ON COLUMN shared_pods.teacher_id IS 'Teacher ID who shared the pod';
COMMENT ON COLUMN shared_pods.student_id IS 'Student ID receiving the shared pod';
COMMENT ON COLUMN shared_pods.message IS 'Optional message from teacher to student about the shared pod';
