// Package domain contains recommendation system entities.
// Implements TikTok-style recommendation algorithm based on user interactions.
package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// InteractionType represents the type of user interaction with a pod.
type InteractionType string

const (
	InteractionView             InteractionType = "view"
	InteractionStar             InteractionType = "star"
	InteractionUnstar           InteractionType = "unstar"
	InteractionFollow           InteractionType = "follow"
	InteractionUnfollow         InteractionType = "unfollow"
	InteractionFork             InteractionType = "fork"
	InteractionShare            InteractionType = "share"
	InteractionTimeSpent        InteractionType = "time_spent"
	InteractionMaterialView     InteractionType = "material_view"
	InteractionMaterialBookmark InteractionType = "material_bookmark"
	InteractionSearchClick      InteractionType = "search_click"
	InteractionUpvote           InteractionType = "upvote"
	InteractionRemoveUpvote     InteractionType = "remove_upvote"
	InteractionDownvote         InteractionType = "downvote"
	InteractionRemoveDownvote   InteractionType = "remove_downvote"
)

// DefaultInteractionWeights defines the base weights for each interaction type.
var DefaultInteractionWeights = map[InteractionType]float64{
	InteractionView:             1.0,
	InteractionStar:             5.0,
	InteractionUnstar:           -3.0,
	InteractionFollow:           8.0,
	InteractionUnfollow:         -5.0,
	InteractionFork:             10.0,
	InteractionShare:            6.0,
	InteractionTimeSpent:        0.1, // Per second, capped at 300 seconds
	InteractionMaterialView:     2.0,
	InteractionMaterialBookmark: 4.0,
	InteractionSearchClick:      3.0,
	InteractionUpvote:           7.0,  // Trust indicator - higher weight than star
	InteractionRemoveUpvote:     -4.0, // Negative weight for removing upvote
	InteractionDownvote:         -7.0, // Negative trust indicator
	InteractionRemoveDownvote:   4.0,  // Positive weight for removing downvote
}

// MaxTimeSpentSeconds is the cap for time spent weighting.
const MaxTimeSpentSeconds = 300

// InteractionMetadata contains additional context for an interaction.
type InteractionMetadata struct {
	TimeSpentSeconds int       `json:"time_spent_seconds,omitempty"`
	MaterialID       uuid.UUID `json:"material_id,omitempty"`
	SearchQuery      string    `json:"search_query,omitempty"`
	Referrer         string    `json:"referrer,omitempty"`
	DeviceType       string    `json:"device_type,omitempty"`
	ScrollDepth      float64   `json:"scroll_depth,omitempty"` // 0.0 - 1.0
}

// ToJSON converts metadata to JSON bytes.
func (m *InteractionMetadata) ToJSON() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// PodInteraction represents a single user interaction event.
type PodInteraction struct {
	ID              uuid.UUID            `json:"id"`
	UserID          uuid.UUID            `json:"user_id"`
	PodID           uuid.UUID            `json:"pod_id"`
	InteractionType InteractionType      `json:"interaction_type"`
	Weight          float64              `json:"weight"`
	Metadata        *InteractionMetadata `json:"metadata,omitempty"`
	SessionID       *uuid.UUID           `json:"session_id,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
}

// NewPodInteraction creates a new interaction with calculated weight.
func NewPodInteraction(userID, podID uuid.UUID, interactionType InteractionType, metadata *InteractionMetadata) *PodInteraction {
	weight := calculateWeight(interactionType, metadata)

	return &PodInteraction{
		ID:              uuid.New(),
		UserID:          userID,
		PodID:           podID,
		InteractionType: interactionType,
		Weight:          weight,
		Metadata:        metadata,
		CreatedAt:       time.Now(),
	}
}

// calculateWeight computes the weight for an interaction.
func calculateWeight(interactionType InteractionType, metadata *InteractionMetadata) float64 {
	baseWeight, ok := DefaultInteractionWeights[interactionType]
	if !ok {
		baseWeight = 1.0
	}

	// Special handling for time_spent - multiply by seconds (capped)
	if interactionType == InteractionTimeSpent && metadata != nil {
		seconds := metadata.TimeSpentSeconds
		if seconds > MaxTimeSpentSeconds {
			seconds = MaxTimeSpentSeconds
		}
		return baseWeight * float64(seconds)
	}

	return baseWeight
}

// UserCategoryScore represents aggregated preference for a category.
type UserCategoryScore struct {
	ID                    uuid.UUID `json:"id"`
	UserID                uuid.UUID `json:"user_id"`
	Category              string    `json:"category"`
	Score                 float64   `json:"score"`
	ViewCount             int       `json:"view_count"`
	StarCount             int       `json:"star_count"`
	FollowCount           int       `json:"follow_count"`
	ForkCount             int       `json:"fork_count"`
	TotalTimeSpentSeconds int       `json:"total_time_spent_seconds"`
	LastInteractionAt     time.Time `json:"last_interaction_at"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// UserTagScore represents aggregated preference for a specific tag.
type UserTagScore struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	Tag               string    `json:"tag"`
	Score             float64   `json:"score"`
	InteractionCount  int       `json:"interaction_count"`
	LastInteractionAt time.Time `json:"last_interaction_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// PodPopularityScore represents pre-computed popularity metrics.
type PodPopularityScore struct {
	PodID               uuid.UUID `json:"pod_id"`
	TotalViews          int       `json:"total_views"`
	TotalStars          int       `json:"total_stars"`
	TotalFollows        int       `json:"total_follows"`
	TotalForks          int       `json:"total_forks"`
	TrendingScore       float64   `json:"trending_score"`
	EngagementRate      float64   `json:"engagement_rate"`
	AvgTimeSpentSeconds float64   `json:"avg_time_spent_seconds"`
	ReturnVisitorRate   float64   `json:"return_visitor_rate"`
	CalculatedAt        time.Time `json:"calculated_at"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// InteractionWeight represents configurable weight for an interaction type.
type InteractionWeight struct {
	InteractionType InteractionType `json:"interaction_type"`
	BaseWeight      float64         `json:"base_weight"`
	Description     string          `json:"description"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// RecommendedPod represents a pod with its recommendation score.
type RecommendedPod struct {
	Pod                 *Pod    `json:"pod"`
	RecommendationScore float64 `json:"recommendation_score"`

	// Score breakdown for transparency/debugging
	CategoryMatchScore float64 `json:"category_match_score"`
	TagMatchScore      float64 `json:"tag_match_score"`
	PopularityScore    float64 `json:"popularity_score"`
	RecencyScore       float64 `json:"recency_score"`
	InterestMatchScore float64 `json:"interest_match_score"`

	// Reason for recommendation
	MatchedCategories []string `json:"matched_categories,omitempty"`
	MatchedTags       []string `json:"matched_tags,omitempty"`
}

// RecommendationConfig holds algorithm configuration.
type RecommendationConfig struct {
	// Weight factors (must sum to 1.0)
	CategoryWeight   float64 `json:"category_weight"`   // Default: 0.35
	TagWeight        float64 `json:"tag_weight"`        // Default: 0.15
	PopularityWeight float64 `json:"popularity_weight"` // Default: 0.20
	RecencyWeight    float64 `json:"recency_weight"`    // Default: 0.15
	InterestWeight   float64 `json:"interest_weight"`   // Default: 0.15

	// Time decay settings
	DecayHalfLifeDays int `json:"decay_half_life_days"` // Default: 7

	// Diversity settings
	MaxSameCategoryPercent float64 `json:"max_same_category_percent"` // Default: 0.4

	// Cold start settings
	MinInteractionsForPersonalization int `json:"min_interactions_for_personalization"` // Default: 5
}

// DefaultRecommendationConfig returns the default configuration.
func DefaultRecommendationConfig() *RecommendationConfig {
	return &RecommendationConfig{
		CategoryWeight:                    0.35,
		TagWeight:                         0.15,
		PopularityWeight:                  0.20,
		RecencyWeight:                     0.15,
		InterestWeight:                    0.15,
		DecayHalfLifeDays:                 7,
		MaxSameCategoryPercent:            0.4,
		MinInteractionsForPersonalization: 5,
	}
}

// UserPreferenceProfile represents a user's aggregated preferences.
type UserPreferenceProfile struct {
	UserID            uuid.UUID            `json:"user_id"`
	TopCategories     []CategoryPreference `json:"top_categories"`
	TopTags           []TagPreference      `json:"top_tags"`
	TotalInteractions int                  `json:"total_interactions"`
	LastInteractionAt time.Time            `json:"last_interaction_at"`
	HasEnoughData     bool                 `json:"has_enough_data"` // For cold start handling
}

// CategoryPreference represents preference for a category.
type CategoryPreference struct {
	Category string  `json:"category"`
	Score    float64 `json:"score"`
	Rank     int     `json:"rank"`
}

// TagPreference represents preference for a tag.
type TagPreference struct {
	Tag   string  `json:"tag"`
	Score float64 `json:"score"`
	Rank  int     `json:"rank"`
}
