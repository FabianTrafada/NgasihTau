-- Migration: Create bookmarks table
-- Requirements: 5.4

CREATE TABLE bookmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    folder VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT bookmarks_unique_user_material UNIQUE (user_id, material_id)
);

-- Indexes for frequently queried columns
CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_material_id ON bookmarks(material_id);
CREATE INDEX idx_bookmarks_folder ON bookmarks(folder) WHERE folder IS NOT NULL;
CREATE INDEX idx_bookmarks_created_at ON bookmarks(created_at);

-- Composite index for user bookmarks listing
CREATE INDEX idx_bookmarks_user_id_created_at ON bookmarks(user_id, created_at DESC);

-- Composite index for user bookmarks by folder
CREATE INDEX idx_bookmarks_user_id_folder ON bookmarks(user_id, folder);

COMMENT ON TABLE bookmarks IS 'User bookmarks for saving learning materials';
COMMENT ON COLUMN bookmarks.user_id IS 'User ID who created the bookmark (from User Service)';
COMMENT ON COLUMN bookmarks.material_id IS 'Reference to the bookmarked material';
COMMENT ON COLUMN bookmarks.folder IS 'Optional folder/collection name for organizing bookmarks';
