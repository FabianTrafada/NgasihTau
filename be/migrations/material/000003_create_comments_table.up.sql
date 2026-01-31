-- Migration: Create comments table

CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    edited BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT comments_content_not_empty CHECK (LENGTH(TRIM(content)) > 0)
);

-- Indexes for frequently queried columns
CREATE INDEX idx_comments_material_id ON comments(material_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);
CREATE INDEX idx_comments_parent_id ON comments(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_comments_deleted_at ON comments(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_comments_created_at ON comments(created_at);

-- Composite index for material comments listing (threaded)
CREATE INDEX idx_comments_material_id_parent_id ON comments(material_id, parent_id NULLS FIRST, created_at);

-- Trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_comments_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_comments_updated_at
    BEFORE UPDATE ON comments
    FOR EACH ROW
    EXECUTE FUNCTION update_comments_updated_at_column();

COMMENT ON TABLE comments IS 'User comments on learning materials';
COMMENT ON COLUMN comments.material_id IS 'Reference to the material being commented on';
COMMENT ON COLUMN comments.user_id IS 'User ID of the commenter (from User Service)';
COMMENT ON COLUMN comments.parent_id IS 'Reference to parent comment for threaded replies';
COMMENT ON COLUMN comments.edited IS 'Flag indicating if the comment has been edited';
