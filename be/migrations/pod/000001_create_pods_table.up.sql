-- Migration: Create pods table

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE pods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    visibility VARCHAR(20) NOT NULL DEFAULT 'public',
    categories TEXT[],
    tags TEXT[],
    star_count INTEGER NOT NULL DEFAULT 0,
    fork_count INTEGER NOT NULL DEFAULT 0,
    view_count INTEGER NOT NULL DEFAULT 0,
    forked_from_id UUID REFERENCES pods(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT pods_visibility_check CHECK (visibility IN ('public', 'private')),
    CONSTRAINT pods_star_count_check CHECK (star_count >= 0),
    CONSTRAINT pods_fork_count_check CHECK (fork_count >= 0),
    CONSTRAINT pods_view_count_check CHECK (view_count >= 0)
);

-- Indexes for frequently queried columns
CREATE INDEX idx_pods_owner_id ON pods(owner_id);
CREATE INDEX idx_pods_slug ON pods(slug);
CREATE INDEX idx_pods_visibility ON pods(visibility);
CREATE INDEX idx_pods_forked_from_id ON pods(forked_from_id) WHERE forked_from_id IS NOT NULL;
CREATE INDEX idx_pods_deleted_at ON pods(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_pods_created_at ON pods(created_at);
CREATE INDEX idx_pods_star_count ON pods(star_count DESC);
CREATE INDEX idx_pods_categories ON pods USING GIN(categories);
CREATE INDEX idx_pods_tags ON pods USING GIN(tags);

-- Trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_pods_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_pods_updated_at
    BEFORE UPDATE ON pods
    FOR EACH ROW
    EXECUTE FUNCTION update_pods_updated_at_column();

COMMENT ON TABLE pods IS 'Knowledge Pods - collaborative learning units containing materials';
COMMENT ON COLUMN pods.owner_id IS 'User ID of the pod owner (from User Service)';
COMMENT ON COLUMN pods.slug IS 'URL-friendly unique identifier for the pod';
COMMENT ON COLUMN pods.visibility IS 'Pod visibility: public or private';
COMMENT ON COLUMN pods.forked_from_id IS 'Reference to original pod if this is a fork';
