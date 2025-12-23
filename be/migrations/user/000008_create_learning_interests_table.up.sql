-- Migration: Create learning interests tables
-- Requirements: Learning interests for first-time users to get personalized recommendations

-- Predefined learning interests (provided by system)
CREATE TABLE predefined_interests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(100),
    category VARCHAR(100),
    display_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- User learning interests (both predefined and custom)
CREATE TABLE user_learning_interests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    predefined_interest_id UUID REFERENCES predefined_interests(id) ON DELETE CASCADE,
    custom_interest VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Either predefined_interest_id OR custom_interest must be set, not both
    CONSTRAINT user_interests_type_check CHECK (
        (predefined_interest_id IS NOT NULL AND custom_interest IS NULL) OR
        (predefined_interest_id IS NULL AND custom_interest IS NOT NULL)
    ),
    -- Prevent duplicate predefined interests per user
    CONSTRAINT user_predefined_interest_unique UNIQUE (user_id, predefined_interest_id),
    -- Prevent duplicate custom interests per user (case-insensitive)
    CONSTRAINT user_custom_interest_unique UNIQUE (user_id, custom_interest)
);

-- Track if user has completed onboarding (selected interests)
ALTER TABLE users ADD COLUMN IF NOT EXISTS onboarding_completed BOOLEAN NOT NULL DEFAULT FALSE;

-- Indexes
CREATE INDEX idx_predefined_interests_category ON predefined_interests(category);
CREATE INDEX idx_predefined_interests_display_order ON predefined_interests(display_order);
CREATE INDEX idx_predefined_interests_is_active ON predefined_interests(is_active) WHERE is_active = TRUE;

CREATE INDEX idx_user_learning_interests_user_id ON user_learning_interests(user_id);
CREATE INDEX idx_user_learning_interests_predefined ON user_learning_interests(predefined_interest_id) WHERE predefined_interest_id IS NOT NULL;
CREATE INDEX idx_user_learning_interests_custom ON user_learning_interests(custom_interest) WHERE custom_interest IS NOT NULL;

CREATE INDEX idx_users_onboarding_completed ON users(onboarding_completed);

-- Trigger to auto-update updated_at for predefined_interests
CREATE TRIGGER update_predefined_interests_updated_at
    BEFORE UPDATE ON predefined_interests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Seed some predefined interests
INSERT INTO predefined_interests (name, slug, description, icon, category, display_order) VALUES
-- Programming & Development
('Machine Learning', 'machine-learning', 'Artificial intelligence and machine learning algorithms', 'brain', 'Technology', 1),
('Web Development', 'web-development', 'Frontend and backend web development', 'globe', 'Technology', 2),
('Mobile Development', 'mobile-development', 'iOS and Android app development', 'smartphone', 'Technology', 3),
('Data Science', 'data-science', 'Data analysis, visualization, and statistics', 'chart-bar', 'Technology', 4),
('Cloud Computing', 'cloud-computing', 'AWS, GCP, Azure and cloud infrastructure', 'cloud', 'Technology', 5),
('Cybersecurity', 'cybersecurity', 'Information security and ethical hacking', 'shield', 'Technology', 6),
('DevOps', 'devops', 'CI/CD, containerization, and infrastructure automation', 'settings', 'Technology', 7),
('Blockchain', 'blockchain', 'Cryptocurrency and decentralized applications', 'link', 'Technology', 8),
('Game Development', 'game-development', 'Video game design and development', 'gamepad', 'Technology', 9),

-- Design
('UI/UX Design', 'ui-ux-design', 'User interface and experience design', 'palette', 'Design', 10),
('Graphic Design', 'graphic-design', 'Visual design and digital art', 'brush', 'Design', 11),
('Product Design', 'product-design', 'End-to-end product design process', 'box', 'Design', 12),

-- Business & Marketing
('Digital Marketing', 'digital-marketing', 'SEO, social media, and online marketing', 'megaphone', 'Business', 13),
('Entrepreneurship', 'entrepreneurship', 'Startup building and business strategy', 'rocket', 'Business', 14),
('Project Management', 'project-management', 'Agile, Scrum, and project planning', 'clipboard', 'Business', 15),
('Finance', 'finance', 'Personal finance and investment', 'dollar-sign', 'Business', 16),

-- Science & Math
('Mathematics', 'mathematics', 'Pure and applied mathematics', 'calculator', 'Science', 17),
('Physics', 'physics', 'Classical and modern physics', 'atom', 'Science', 18),
('Chemistry', 'chemistry', 'Organic and inorganic chemistry', 'flask', 'Science', 19),
('Biology', 'biology', 'Life sciences and biotechnology', 'leaf', 'Science', 20),

-- Languages
('English', 'english', 'English language learning', 'book', 'Language', 21),
('Japanese', 'japanese', 'Japanese language and culture', 'book', 'Language', 22),
('Mandarin', 'mandarin', 'Chinese Mandarin language', 'book', 'Language', 23),

-- Creative
('Photography', 'photography', 'Digital photography and editing', 'camera', 'Creative', 24),
('Video Production', 'video-production', 'Video editing and filmmaking', 'video', 'Creative', 25),
('Music Production', 'music-production', 'Audio production and sound design', 'music', 'Creative', 26),
('Writing', 'writing', 'Creative and technical writing', 'pen-tool', 'Creative', 27);

COMMENT ON TABLE predefined_interests IS 'System-defined learning interest options';
COMMENT ON TABLE user_learning_interests IS 'User selected learning interests (predefined or custom)';
COMMENT ON COLUMN users.onboarding_completed IS 'Whether user has completed initial onboarding (selecting interests)';
