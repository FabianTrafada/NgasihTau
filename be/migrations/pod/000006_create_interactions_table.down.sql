-- Rollback: Drop recommendation system tables

DROP TRIGGER IF EXISTS update_pod_popularity_scores_updated_at ON pod_popularity_scores;
DROP TRIGGER IF EXISTS update_user_tag_scores_updated_at ON user_tag_scores;
DROP TRIGGER IF EXISTS update_user_category_scores_updated_at ON user_category_scores;
DROP FUNCTION IF EXISTS update_recommendation_updated_at();

DROP TABLE IF EXISTS interaction_weights;
DROP TABLE IF EXISTS pod_popularity_scores;
DROP TABLE IF EXISTS user_tag_scores;
DROP TABLE IF EXISTS user_category_scores;
DROP TABLE IF EXISTS pod_interactions;

DROP TYPE IF EXISTS interaction_type;
