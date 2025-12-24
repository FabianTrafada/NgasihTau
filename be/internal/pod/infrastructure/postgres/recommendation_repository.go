// Package postgres provides PostgreSQL implementation of recommendation repository.
package postgres

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// PodPopularityRepository implements domain.PodPopularityRepository.
type PodPopularityRepository struct {
	db *pgxpool.Pool
}

// NewPodPopularityRepository creates a new PodPopularityRepository.
func NewPodPopularityRepository(db *pgxpool.Pool) *PodPopularityRepository {
	return &PodPopularityRepository{db: db}
}

var _ domain.PodPopularityRepository = (*PodPopularityRepository)(nil)

// Upsert creates or updates popularity score.
func (r *PodPopularityRepository) Upsert(ctx context.Context, score *domain.PodPopularityScore) error {
	query := `
		INSERT INTO pod_popularity_scores (
			pod_id, total_views, total_stars, total_follows, total_forks,
			trending_score, engagement_rate, avg_time_spent_seconds, return_visitor_rate,
			calculated_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (pod_id) DO UPDATE SET
			total_views = EXCLUDED.total_views,
			total_stars = EXCLUDED.total_stars,
			total_follows = EXCLUDED.total_follows,
			total_forks = EXCLUDED.total_forks,
			trending_score = EXCLUDED.trending_score,
			engagement_rate = EXCLUDED.engagement_rate,
			avg_time_spent_seconds = EXCLUDED.avg_time_spent_seconds,
			return_visitor_rate = EXCLUDED.return_visitor_rate,
			calculated_at = EXCLUDED.calculated_at,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query,
		score.PodID,
		score.TotalViews,
		score.TotalStars,
		score.TotalFollows,
		score.TotalForks,
		score.TrendingScore,
		score.EngagementRate,
		score.AvgTimeSpentSeconds,
		score.ReturnVisitorRate,
		score.CalculatedAt,
		score.CreatedAt,
		score.UpdatedAt,
	)
	if err != nil {
		return errors.Internal("failed to upsert popularity score", err)
	}

	return nil
}

// FindByPodID returns popularity score for a pod.
func (r *PodPopularityRepository) FindByPodID(ctx context.Context, podID uuid.UUID) (*domain.PodPopularityScore, error) {
	query := `
		SELECT pod_id, total_views, total_stars, total_follows, total_forks,
			   trending_score, engagement_rate, avg_time_spent_seconds, return_visitor_rate,
			   calculated_at, created_at, updated_at
		FROM pod_popularity_scores
		WHERE pod_id = $1
	`

	var score domain.PodPopularityScore
	err := r.db.QueryRow(ctx, query, podID).Scan(
		&score.PodID,
		&score.TotalViews,
		&score.TotalStars,
		&score.TotalFollows,
		&score.TotalForks,
		&score.TrendingScore,
		&score.EngagementRate,
		&score.AvgTimeSpentSeconds,
		&score.ReturnVisitorRate,
		&score.CalculatedAt,
		&score.CreatedAt,
		&score.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Internal("failed to find popularity score", err)
	}

	return &score, nil
}

// GetTrendingPods returns pods ordered by trending score.
func (r *PodPopularityRepository) GetTrendingPods(ctx context.Context, limit, offset int) ([]*domain.PodPopularityScore, error) {
	query := `
		SELECT pod_id, total_views, total_stars, total_follows, total_forks,
			   trending_score, engagement_rate, avg_time_spent_seconds, return_visitor_rate,
			   calculated_at, created_at, updated_at
		FROM pod_popularity_scores
		ORDER BY trending_score DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, errors.Internal("failed to get trending pods", err)
	}
	defer rows.Close()

	var scores []*domain.PodPopularityScore
	for rows.Next() {
		var score domain.PodPopularityScore
		err := rows.Scan(
			&score.PodID,
			&score.TotalViews,
			&score.TotalStars,
			&score.TotalFollows,
			&score.TotalForks,
			&score.TrendingScore,
			&score.EngagementRate,
			&score.AvgTimeSpentSeconds,
			&score.ReturnVisitorRate,
			&score.CalculatedAt,
			&score.CreatedAt,
			&score.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan popularity score", err)
		}
		scores = append(scores, &score)
	}

	return scores, nil
}

// RecalculateForPod recalculates popularity metrics for a specific pod.
func (r *PodPopularityRepository) RecalculateForPod(ctx context.Context, podID uuid.UUID) error {
	// Calculate metrics from interactions in the last 7 days for trending
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	query := `
		WITH recent_interactions AS (
			SELECT 
				COUNT(*) FILTER (WHERE interaction_type = 'view') as view_count,
				COUNT(*) FILTER (WHERE interaction_type = 'star') as star_count,
				COUNT(*) FILTER (WHERE interaction_type = 'follow') as follow_count,
				COUNT(*) FILTER (WHERE interaction_type = 'fork') as fork_count,
				COALESCE(AVG((metadata->>'time_spent_seconds')::int) FILTER (WHERE interaction_type = 'time_spent'), 0) as avg_time,
				COUNT(DISTINCT user_id) as unique_users,
				COUNT(*) FILTER (WHERE interaction_type = 'view') as total_views_recent
			FROM pod_interactions
			WHERE pod_id = $1 AND created_at >= $2
		),
		all_time AS (
			SELECT 
				COUNT(*) FILTER (WHERE interaction_type = 'view') as total_views,
				COUNT(*) FILTER (WHERE interaction_type = 'star') - COUNT(*) FILTER (WHERE interaction_type = 'unstar') as net_stars,
				COUNT(*) FILTER (WHERE interaction_type = 'follow') - COUNT(*) FILTER (WHERE interaction_type = 'unfollow') as net_follows,
				COUNT(*) FILTER (WHERE interaction_type = 'fork') as total_forks,
				COUNT(DISTINCT user_id) as total_unique_users
			FROM pod_interactions
			WHERE pod_id = $1
		),
		return_visitors AS (
			SELECT COUNT(*) as returning FROM (
				SELECT user_id
				FROM pod_interactions
				WHERE pod_id = $1 AND interaction_type = 'view'
				GROUP BY user_id
				HAVING COUNT(*) > 1
			) sub
		)
		INSERT INTO pod_popularity_scores (
			pod_id, total_views, total_stars, total_follows, total_forks,
			trending_score, engagement_rate, avg_time_spent_seconds, return_visitor_rate,
			calculated_at, created_at, updated_at
		)
		SELECT 
			$1,
			COALESCE(a.total_views, 0),
			GREATEST(COALESCE(a.net_stars, 0), 0),
			GREATEST(COALESCE(a.net_follows, 0), 0),
			COALESCE(a.total_forks, 0),
			-- Trending score: weighted sum of recent activity
			(COALESCE(r.view_count, 0) * 1.0 + COALESCE(r.star_count, 0) * 5.0 + 
			 COALESCE(r.follow_count, 0) * 8.0 + COALESCE(r.fork_count, 0) * 10.0),
			-- Engagement rate
			CASE WHEN a.total_views > 0 
				THEN (a.net_stars + a.net_follows + a.total_forks)::decimal / a.total_views 
				ELSE 0 END,
			COALESCE(r.avg_time, 0),
			-- Return visitor rate
			CASE WHEN a.total_unique_users > 0 
				THEN rv.returning::decimal / a.total_unique_users 
				ELSE 0 END,
			NOW(),
			NOW(),
			NOW()
		FROM recent_interactions r, all_time a, return_visitors rv
		ON CONFLICT (pod_id) DO UPDATE SET
			total_views = EXCLUDED.total_views,
			total_stars = EXCLUDED.total_stars,
			total_follows = EXCLUDED.total_follows,
			total_forks = EXCLUDED.total_forks,
			trending_score = EXCLUDED.trending_score,
			engagement_rate = EXCLUDED.engagement_rate,
			avg_time_spent_seconds = EXCLUDED.avg_time_spent_seconds,
			return_visitor_rate = EXCLUDED.return_visitor_rate,
			calculated_at = NOW(),
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, podID, sevenDaysAgo)
	if err != nil {
		return errors.Internal("failed to recalculate popularity", err)
	}

	return nil
}

// RecalculateAll recalculates popularity metrics for all pods.
func (r *PodPopularityRepository) RecalculateAll(ctx context.Context) error {
	// Get all pod IDs that have interactions
	query := `SELECT DISTINCT pod_id FROM pod_interactions`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return errors.Internal("failed to get pod IDs", err)
	}
	defer rows.Close()

	var podIDs []uuid.UUID
	for rows.Next() {
		var podID uuid.UUID
		if err := rows.Scan(&podID); err != nil {
			return errors.Internal("failed to scan pod ID", err)
		}
		podIDs = append(podIDs, podID)
	}

	// Recalculate each pod
	for _, podID := range podIDs {
		if err := r.RecalculateForPod(ctx, podID); err != nil {
			// Log error but continue with other pods
			continue
		}
	}

	return nil
}

// ===========================================
// Recommendation Repository
// ===========================================

// RecommendationRepository implements domain.RecommendationRepository.
type RecommendationRepository struct {
	db                *pgxpool.Pool
	podRepo           domain.PodRepository
	categoryScoreRepo domain.UserCategoryScoreRepository
	tagScoreRepo      domain.UserTagScoreRepository
	popularityRepo    domain.PodPopularityRepository
	interactionRepo   domain.InteractionRepository
}

// NewRecommendationRepository creates a new RecommendationRepository.
func NewRecommendationRepository(
	db *pgxpool.Pool,
	podRepo domain.PodRepository,
	categoryScoreRepo domain.UserCategoryScoreRepository,
	tagScoreRepo domain.UserTagScoreRepository,
	popularityRepo domain.PodPopularityRepository,
	interactionRepo domain.InteractionRepository,
) *RecommendationRepository {
	return &RecommendationRepository{
		db:                db,
		podRepo:           podRepo,
		categoryScoreRepo: categoryScoreRepo,
		tagScoreRepo:      tagScoreRepo,
		popularityRepo:    popularityRepo,
		interactionRepo:   interactionRepo,
	}
}

var _ domain.RecommendationRepository = (*RecommendationRepository)(nil)

// GetPersonalizedFeed returns personalized pod recommendations for a user.
func (r *RecommendationRepository) GetPersonalizedFeed(ctx context.Context, userID uuid.UUID, config *domain.RecommendationConfig, limit, offset int) ([]*domain.RecommendedPod, error) {
	if config == nil {
		config = domain.DefaultRecommendationConfig()
	}

	// Get user preferences
	profile, err := r.GetUserPreferenceProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	// If user doesn't have enough data, return trending feed
	if !profile.HasEnoughData {
		pods, err := r.GetTrendingFeed(ctx, limit, offset)
		if err != nil {
			return nil, err
		}

		// Convert to RecommendedPod
		result := make([]*domain.RecommendedPod, len(pods))
		for i, pod := range pods {
			result[i] = &domain.RecommendedPod{
				Pod:                 pod,
				RecommendationScore: 0,
				PopularityScore:     1.0,
			}
		}
		return result, nil
	}

	// Build the recommendation query with scoring
	query := r.buildRecommendationQuery(profile, config, limit, offset)

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, errors.Internal("failed to get recommendations", err)
	}
	defer rows.Close()

	return r.scanRecommendedPods(ctx, rows, profile, config)
}

// buildRecommendationQuery constructs the SQL for personalized recommendations.
func (r *RecommendationRepository) buildRecommendationQuery(profile *domain.UserPreferenceProfile, config *domain.RecommendationConfig, limit, offset int) string {
	// Extract top categories for matching
	categoryList := make([]string, 0, len(profile.TopCategories))
	for _, cat := range profile.TopCategories {
		categoryList = append(categoryList, cat.Category)
	}

	// Extract top tags for matching
	tagList := make([]string, 0, len(profile.TopTags))
	for _, tag := range profile.TopTags {
		tagList = append(tagList, tag.Tag)
	}

	query := `
		WITH user_prefs AS (
			SELECT 
				category, score as cat_score
			FROM user_category_scores
			WHERE user_id = $1
		),
		user_tags AS (
			SELECT tag, score as tag_score
			FROM user_tag_scores
			WHERE user_id = $1
		),
		scored_pods AS (
			SELECT 
				p.id,
				p.owner_id,
				p.name,
				p.slug,
				p.description,
				p.visibility,
				p.categories,
				p.tags,
				p.star_count,
				p.fork_count,
				p.view_count,
				p.forked_from_id,
				p.created_at,
				p.updated_at,
				-- Category match score
				COALESCE((
					SELECT SUM(up.cat_score)
					FROM user_prefs up
					WHERE up.category = ANY(p.categories)
				), 0) as category_score,
				-- Tag match score  
				COALESCE((
					SELECT SUM(ut.tag_score)
					FROM user_tags ut
					WHERE ut.tag = ANY(p.tags)
				), 0) as tag_score,
				-- Popularity score (normalized)
				COALESCE(pop.trending_score, 0) as popularity_score,
				-- Recency score (exponential decay)
				EXP(-EXTRACT(EPOCH FROM (NOW() - p.created_at)) / (86400 * 30)) as recency_score
			FROM pods p
			LEFT JOIN pod_popularity_scores pop ON p.id = pop.pod_id
			WHERE p.visibility = 'public' 
				AND p.deleted_at IS NULL
				-- Exclude pods user has already interacted with heavily
				AND p.id NOT IN (
					SELECT DISTINCT pod_id FROM pod_interactions 
					WHERE user_id = $1 
					AND interaction_type IN ('star', 'follow', 'fork')
				)
		)
		SELECT *,
			-- Final recommendation score
			(category_score * 0.35 + tag_score * 0.15 + popularity_score * 0.0001 + recency_score * 100 * 0.15) as recommendation_score
		FROM scored_pods
		ORDER BY recommendation_score DESC
		LIMIT $2 OFFSET $3
	`

	return query
}

// scanRecommendedPods scans rows into RecommendedPod.
func (r *RecommendationRepository) scanRecommendedPods(ctx context.Context, rows pgx.Rows, profile *domain.UserPreferenceProfile, config *domain.RecommendationConfig) ([]*domain.RecommendedPod, error) {
	var recommendations []*domain.RecommendedPod

	for rows.Next() {
		var pod domain.Pod
		var categoryScore, tagScore, popularityScore, recencyScore, recommendationScore float64

		err := rows.Scan(
			&pod.ID,
			&pod.OwnerID,
			&pod.Name,
			&pod.Slug,
			&pod.Description,
			&pod.Visibility,
			&pod.Categories,
			&pod.Tags,
			&pod.StarCount,
			&pod.ForkCount,
			&pod.ViewCount,
			&pod.ForkedFromID,
			&pod.CreatedAt,
			&pod.UpdatedAt,
			&categoryScore,
			&tagScore,
			&popularityScore,
			&recencyScore,
			&recommendationScore,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan recommended pod", err)
		}

		// Find matched categories and tags
		matchedCategories := findMatchedCategories(pod.Categories, profile.TopCategories)
		matchedTags := findMatchedTags(pod.Tags, profile.TopTags)

		recommendations = append(recommendations, &domain.RecommendedPod{
			Pod:                 &pod,
			RecommendationScore: recommendationScore,
			CategoryMatchScore:  categoryScore,
			TagMatchScore:       tagScore,
			PopularityScore:     popularityScore,
			RecencyScore:        recencyScore,
			MatchedCategories:   matchedCategories,
			MatchedTags:         matchedTags,
		})
	}

	return recommendations, nil
}

// GetTrendingFeed returns trending pods for cold start or anonymous users.
func (r *RecommendationRepository) GetTrendingFeed(ctx context.Context, limit, offset int) ([]*domain.Pod, error) {
	query := `
		SELECT p.id, p.owner_id, p.name, p.slug, p.description, p.visibility,
			   p.categories, p.tags, p.star_count, p.fork_count, p.view_count,
			   p.forked_from_id, p.created_at, p.updated_at
		FROM pods p
		LEFT JOIN pod_popularity_scores pop ON p.id = pop.pod_id
		WHERE p.visibility = 'public' AND p.deleted_at IS NULL
		ORDER BY COALESCE(pop.trending_score, 0) DESC, p.star_count DESC, p.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, errors.Internal("failed to get trending feed", err)
	}
	defer rows.Close()

	var pods []*domain.Pod
	for rows.Next() {
		var pod domain.Pod
		err := rows.Scan(
			&pod.ID,
			&pod.OwnerID,
			&pod.Name,
			&pod.Slug,
			&pod.Description,
			&pod.Visibility,
			&pod.Categories,
			&pod.Tags,
			&pod.StarCount,
			&pod.ForkCount,
			&pod.ViewCount,
			&pod.ForkedFromID,
			&pod.CreatedAt,
			&pod.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan pod", err)
		}
		pods = append(pods, &pod)
	}

	return pods, nil
}

// GetSimilarPods returns pods similar to a given pod.
func (r *RecommendationRepository) GetSimilarPods(ctx context.Context, podID uuid.UUID, limit int) ([]*domain.Pod, error) {
	// Find pods with overlapping categories and tags
	query := `
		WITH target_pod AS (
			SELECT categories, tags FROM pods WHERE id = $1
		)
		SELECT p.id, p.owner_id, p.name, p.slug, p.description, p.visibility,
			   p.categories, p.tags, p.star_count, p.fork_count, p.view_count,
			   p.forked_from_id, p.created_at, p.updated_at,
			   -- Similarity score based on overlapping categories and tags
			   (
				   COALESCE(array_length(p.categories & (SELECT categories FROM target_pod), 1), 0) * 2 +
				   COALESCE(array_length(p.tags & (SELECT tags FROM target_pod), 1), 0)
			   ) as similarity_score
		FROM pods p
		WHERE p.id != $1 
			AND p.visibility = 'public' 
			AND p.deleted_at IS NULL
			AND (
				p.categories && (SELECT categories FROM target_pod)
				OR p.tags && (SELECT tags FROM target_pod)
			)
		ORDER BY similarity_score DESC, p.star_count DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, podID, limit)
	if err != nil {
		return nil, errors.Internal("failed to get similar pods", err)
	}
	defer rows.Close()

	var pods []*domain.Pod
	for rows.Next() {
		var pod domain.Pod
		var similarityScore int
		err := rows.Scan(
			&pod.ID,
			&pod.OwnerID,
			&pod.Name,
			&pod.Slug,
			&pod.Description,
			&pod.Visibility,
			&pod.Categories,
			&pod.Tags,
			&pod.StarCount,
			&pod.ForkCount,
			&pod.ViewCount,
			&pod.ForkedFromID,
			&pod.CreatedAt,
			&pod.UpdatedAt,
			&similarityScore,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan pod", err)
		}
		pods = append(pods, &pod)
	}

	return pods, nil
}

// GetUserPreferenceProfile returns aggregated user preferences.
func (r *RecommendationRepository) GetUserPreferenceProfile(ctx context.Context, userID uuid.UUID) (*domain.UserPreferenceProfile, error) {
	// Get interaction count
	interactionCount, err := r.interactionRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	config := domain.DefaultRecommendationConfig()
	hasEnoughData := interactionCount >= config.MinInteractionsForPersonalization

	// Get top categories
	categoryScores, err := r.categoryScoreRepo.FindByUserID(ctx, userID, 10)
	if err != nil {
		return nil, err
	}

	topCategories := make([]domain.CategoryPreference, len(categoryScores))
	for i, cs := range categoryScores {
		topCategories[i] = domain.CategoryPreference{
			Category: cs.Category,
			Score:    cs.Score,
			Rank:     i + 1,
		}
	}

	// Get top tags
	tagScores, err := r.tagScoreRepo.FindByUserID(ctx, userID, 20)
	if err != nil {
		return nil, err
	}

	topTags := make([]domain.TagPreference, len(tagScores))
	for i, ts := range tagScores {
		topTags[i] = domain.TagPreference{
			Tag:   ts.Tag,
			Score: ts.Score,
			Rank:  i + 1,
		}
	}

	// Get last interaction time
	var lastInteractionAt time.Time
	if len(categoryScores) > 0 {
		lastInteractionAt = categoryScores[0].LastInteractionAt
	}

	return &domain.UserPreferenceProfile{
		UserID:            userID,
		TopCategories:     topCategories,
		TopTags:           topTags,
		TotalInteractions: interactionCount,
		LastInteractionAt: lastInteractionAt,
		HasEnoughData:     hasEnoughData,
	}, nil
}

// GetPersonalizedFeedExcluding returns recommendations excluding specific pods.
func (r *RecommendationRepository) GetPersonalizedFeedExcluding(ctx context.Context, userID uuid.UUID, excludePodIDs []uuid.UUID, config *domain.RecommendationConfig, limit int) ([]*domain.RecommendedPod, error) {
	// For now, use the regular feed and filter
	// TODO: Optimize with SQL exclusion
	recommendations, err := r.GetPersonalizedFeed(ctx, userID, config, limit*2, 0)
	if err != nil {
		return nil, err
	}

	// Filter out excluded pods
	excludeMap := make(map[uuid.UUID]bool)
	for _, id := range excludePodIDs {
		excludeMap[id] = true
	}

	var filtered []*domain.RecommendedPod
	for _, rec := range recommendations {
		if !excludeMap[rec.Pod.ID] {
			filtered = append(filtered, rec)
			if len(filtered) >= limit {
				break
			}
		}
	}

	return filtered, nil
}

// Helper functions

func findMatchedCategories(podCategories []string, userPrefs []domain.CategoryPreference) []string {
	prefMap := make(map[string]bool)
	for _, p := range userPrefs {
		prefMap[p.Category] = true
	}

	var matched []string
	for _, cat := range podCategories {
		if prefMap[cat] {
			matched = append(matched, cat)
		}
	}
	return matched
}

func findMatchedTags(podTags []string, userPrefs []domain.TagPreference) []string {
	prefMap := make(map[string]bool)
	for _, p := range userPrefs {
		prefMap[p.Tag] = true
	}

	var matched []string
	for _, tag := range podTags {
		if prefMap[tag] {
			matched = append(matched, tag)
		}
	}
	return matched
}

// Utility function for time decay calculation
func calculateTimeDecay(createdAt time.Time, halfLifeDays int) float64 {
	daysSinceCreation := time.Since(createdAt).Hours() / 24
	halfLife := float64(halfLifeDays)
	return math.Pow(0.5, daysSinceCreation/halfLife)
}
