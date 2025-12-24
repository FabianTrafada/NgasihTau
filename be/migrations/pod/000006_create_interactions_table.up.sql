-- Migration: Create pod interactions and user preference tables
-- Purpose: Track user interactions for recommendation system (TikTok-style algorithm)

-- ===========================================
-- Interaction Types Enum
-- ===========================================
CREATE TYPE interaction_type AS ENUM (
    'view',           -- User viewed the pod
    'star',           -- User starred the pod
    'unstar',         -- User removed star
    'follow',         -- User followed the pod
    'unfollow',       -- User unfollowed the pod
    'fork',           -- User forked the pod
    'share',          -- User shared the pod
    'time_spent',     -- Track time spent viewing
    'material_view',  -- User viewed a material in pod
    'material_bookmark', -- User bookmarked a material
    'search_click'    -- User clicked from search results
);

-- ===========================================
-- Pod Interactions Table (Event Sourcing Style)
-- ===========================================
-- Stores every user interaction as an immutable event
-- This allows for flexible analysis and algorithm tuning
CREATE TABLE pod_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    pod_id UUID NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    interaction_type interaction_type NOT NULL,
    
    -- Weight for this specific interaction (allows dynamic weighting)
    weight DECIMAL(5,2) NOT NULL DEFAULT 1.0,
    
    -- Additional context (time_spent_seconds, material_id, search_query, etc.)
    metadata JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Session tracking for grouping interactions
    session_id UUID,
    
    CONSTRAINT pod_interactions_weight_positive CHECK (weight >= 0)
);

-- Indexes for efficient querying
CREATE INDEX idx_pod_interactions_user_id ON pod_interactions(user_id);
CREATE INDEX idx_pod_interactions_pod_id ON pod_interactions(pod_id);
CREATE INDEX idx_pod_interactions_user_pod ON pod_interactions(user_id, pod_id);
CREATE INDEX idx_pod_interactions_type ON pod_interactions(interaction_type);
CREATE INDEX idx_pod_interactions_created_at ON pod_interactions(created_at DESC);
CREATE INDEX idx_pod_interactions_user_recent ON pod_interactions(user_id, created_at DESC);
CREATE INDEX idx_pod_interactions_session ON pod_interactions(session_id) WHERE session_id IS NOT NULL;

-- Partial index for high-value interactions
CREATE INDEX idx_pod_interactions_high_value ON pod_interactions(user_id, pod_id, created_at DESC)
    WHERE interaction_type IN ('star', 'follow', 'fork', 'material_bookmark');

-- ===========================================
-- User Category Scores (Aggregated Preferences)
-- ===========================================
-- Pre-computed scores per category for fast recommendation queries
-- Updated periodically or on significant interactions
CREATE TABLE user_category_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    category VARCHAR(100) NOT NULL,
    
    -- Aggregated score based on interactions
    score DECIMAL(10,4) NOT NULL DEFAULT 0,
    
    -- Interaction counts for transparency
    view_count INTEGER NOT NULL DEFAULT 0,
    star_count INTEGER NOT NULL DEFAULT 0,
    follow_count INTEGER NOT NULL DEFAULT 0,
    fork_count INTEGER NOT NULL DEFAULT 0,
    total_time_spent_seconds INTEGER NOT NULL DEFAULT 0,
    
    -- Decay factor - scores decrease over time if no new interactions
    last_interaction_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT user_category_scores_unique UNIQUE(user_id, category),
    CONSTRAINT user_category_scores_score_positive CHECK (score >= 0)
);

-- Indexes
CREATE INDEX idx_user_category_scores_user_id ON user_category_scores(user_id);
CREATE INDEX idx_user_category_scores_category ON user_category_scores(category);
CREATE INDEX idx_user_category_scores_user_score ON user_category_scores(user_id, score DESC);
CREATE INDEX idx_user_category_scores_last_interaction ON user_category_scores(last_interaction_at DESC);

-- ===========================================
-- User Tag Scores (Fine-grained Preferences)
-- ===========================================
-- Similar to category scores but for specific tags
CREATE TABLE user_tag_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    tag VARCHAR(100) NOT NULL,
    
    score DECIMAL(10,4) NOT NULL DEFAULT 0,
    interaction_count INTEGER NOT NULL DEFAULT 0,
    
    last_interaction_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT user_tag_scores_unique UNIQUE(user_id, tag),
    CONSTRAINT user_tag_scores_score_positive CHECK (score >= 0)
);

-- Indexes
CREATE INDEX idx_user_tag_scores_user_id ON user_tag_scores(user_id);
CREATE INDEX idx_user_tag_scores_tag ON user_tag_scores(tag);
CREATE INDEX idx_user_tag_scores_user_score ON user_tag_scores(user_id, score DESC);

-- ===========================================
-- Pod Popularity Scores (Cached Metrics)
-- ===========================================
-- Pre-computed popularity metrics updated periodically
CREATE TABLE pod_popularity_scores (
    pod_id UUID PRIMARY KEY REFERENCES pods(id) ON DELETE CASCADE,
    
    -- Raw counts (denormalized for performance)
    total_views INTEGER NOT NULL DEFAULT 0,
    total_stars INTEGER NOT NULL DEFAULT 0,
    total_follows INTEGER NOT NULL DEFAULT 0,
    total_forks INTEGER NOT NULL DEFAULT 0,
    
    -- Time-weighted scores (recent activity counts more)
    trending_score DECIMAL(10,4) NOT NULL DEFAULT 0,
    
    -- Engagement rate = (stars + follows + forks) / views
    engagement_rate DECIMAL(5,4) NOT NULL DEFAULT 0,
    
    -- Quality signals
    avg_time_spent_seconds DECIMAL(10,2) NOT NULL DEFAULT 0,
    return_visitor_rate DECIMAL(5,4) NOT NULL DEFAULT 0,
    
    -- Timestamps
    calculated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_pod_popularity_trending ON pod_popularity_scores(trending_score DESC);
CREATE INDEX idx_pod_popularity_engagement ON pod_popularity_scores(engagement_rate DESC);

-- ===========================================
-- Interaction Weights Configuration
-- ===========================================
-- Configurable weights for different interaction types
-- Allows tuning the algorithm without code changes
CREATE TABLE interaction_weights (
    interaction_type interaction_type PRIMARY KEY,
    base_weight DECIMAL(5,2) NOT NULL DEFAULT 1.0,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Insert default weights
INSERT INTO interaction_weights (interaction_type, base_weight, description) VALUES
    ('view', 1.0, 'Basic view of a pod'),
    ('star', 5.0, 'User starred the pod - strong positive signal'),
    ('unstar', -3.0, 'User removed star - negative signal'),
    ('follow', 8.0, 'User followed the pod - very strong signal'),
    ('unfollow', -5.0, 'User unfollowed - negative signal'),
    ('fork', 10.0, 'User forked the pod - highest positive signal'),
    ('share', 6.0, 'User shared the pod - strong endorsement'),
    ('time_spent', 0.1, 'Per second of time spent (capped)'),
    ('material_view', 2.0, 'Viewed material inside pod'),
    ('material_bookmark', 4.0, 'Bookmarked a material'),
    ('search_click', 3.0, 'Clicked from search results');

-- ===========================================
-- Trigger to update updated_at
-- ===========================================
CREATE OR REPLACE FUNCTION update_recommendation_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_user_category_scores_updated_at
    BEFORE UPDATE ON user_category_scores
    FOR EACH ROW
    EXECUTE FUNCTION update_recommendation_updated_at();

CREATE TRIGGER update_user_tag_scores_updated_at
    BEFORE UPDATE ON user_tag_scores
    FOR EACH ROW
    EXECUTE FUNCTION update_recommendation_updated_at();

CREATE TRIGGER update_pod_popularity_scores_updated_at
    BEFORE UPDATE ON pod_popularity_scores
    FOR EACH ROW
    EXECUTE FUNCTION update_recommendation_updated_at();

-- ===========================================
-- Comments
-- ===========================================
COMMENT ON TABLE pod_interactions IS 'Immutable log of all user interactions with pods for recommendation algorithm';
COMMENT ON TABLE user_category_scores IS 'Aggregated user preference scores per category for fast recommendations';
COMMENT ON TABLE user_tag_scores IS 'Aggregated user preference scores per tag for fine-grained recommendations';
COMMENT ON TABLE pod_popularity_scores IS 'Pre-computed popularity metrics for pods';
COMMENT ON TABLE interaction_weights IS 'Configurable weights for different interaction types';

COMMENT ON COLUMN pod_interactions.weight IS 'Actual weight applied (may differ from base_weight due to context)';
COMMENT ON COLUMN pod_interactions.metadata IS 'JSON with context: time_spent_seconds, material_id, referrer, etc.';
COMMENT ON COLUMN user_category_scores.score IS 'Weighted sum of interactions, with time decay applied';
COMMENT ON COLUMN pod_popularity_scores.trending_score IS 'Time-weighted popularity score for trending recommendations';
