-- Migration: Create ratings table

CREATE TABLE ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    score INTEGER NOT NULL,
    review TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT ratings_score_check CHECK (score >= 1 AND score <= 5),
    CONSTRAINT ratings_unique_user_material UNIQUE (material_id, user_id)
);

-- Indexes for frequently queried columns
CREATE INDEX idx_ratings_material_id ON ratings(material_id);
CREATE INDEX idx_ratings_user_id ON ratings(user_id);
CREATE INDEX idx_ratings_score ON ratings(score);
CREATE INDEX idx_ratings_created_at ON ratings(created_at);

-- Composite index for material ratings listing
CREATE INDEX idx_ratings_material_id_created_at ON ratings(material_id, created_at DESC);

-- Trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_ratings_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_ratings_updated_at
    BEFORE UPDATE ON ratings
    FOR EACH ROW
    EXECUTE FUNCTION update_ratings_updated_at_column();

COMMENT ON TABLE ratings IS 'User ratings and reviews for learning materials';
COMMENT ON COLUMN ratings.material_id IS 'Reference to the material being rated';
COMMENT ON COLUMN ratings.user_id IS 'User ID of the rater (from User Service)';
COMMENT ON COLUMN ratings.score IS 'Rating score from 1 to 5 stars';
COMMENT ON COLUMN ratings.review IS 'Optional text review accompanying the rating';
