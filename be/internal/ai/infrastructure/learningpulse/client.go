package learningpulse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Persona represents a learning persona type.
type Persona string

const (
	PersonaSkimmer       Persona = "skimmer"
	PersonaStruggler     Persona = "struggler"
	PersonaAnxious       Persona = "anxious"
	PersonaBurnout       Persona = "burnout"
	PersonaMaster        Persona = "master"
	PersonaProcrastinator Persona = "procrastinator"
	PersonaDeepDiver     Persona = "deep_diver"
	PersonaSocialLearner Persona = "social_learner"
	PersonaPerfectionist Persona = "perfectionist"
	PersonaLost          Persona = "lost"
	PersonaUnknown       Persona = "unknown"
)

// PredictRequest represents the request to Learning Pulse predict endpoint.
type PredictRequest struct {
	UserID       string        `json:"user_id"`
	BehaviorData *BehaviorData `json:"behavior_data"`
	QuizScore    *float64      `json:"quiz_score,omitempty"`
}

// BehaviorData contains user behavior metrics.
type BehaviorData struct {
	UserID             string              `json:"user_id"`
	AnalysisPeriodDays int                 `json:"analysis_period_days"`
	Chat               ChatBehavior        `json:"chat"`
	Material           MaterialInteraction `json:"material"`
	Activity           ActivityPattern     `json:"activity"`
	Quiz               *QuizPerformance    `json:"quiz,omitempty"`
}

// ChatBehavior contains chat interaction metrics.
type ChatBehavior struct {
	TotalMessages               int     `json:"total_messages"`
	UserMessages                int     `json:"user_messages"`
	AssistantMessages           int     `json:"assistant_messages"`
	QuestionCount               int     `json:"question_count"`
	AvgMessageLength            float64 `json:"avg_message_length"`
	ThumbsUpCount               int     `json:"thumbs_up_count"`
	ThumbsDownCount             int     `json:"thumbs_down_count"`
	UniqueSessions              int     `json:"unique_sessions"`
	TotalSessionDurationMinutes float64 `json:"total_session_duration_minutes"`
}

// MaterialInteraction contains material consumption metrics.
type MaterialInteraction struct {
	TotalTimeSpentSeconds  int     `json:"total_time_spent_seconds"`
	TotalViews             int     `json:"total_views"`
	UniqueMaterialsViewed  int     `json:"unique_materials_viewed"`
	BookmarkCount          int     `json:"bookmark_count"`
	AvgScrollDepth         float64 `json:"avg_scroll_depth"`
}

// ActivityPattern contains temporal activity metrics.
type ActivityPattern struct {
	ActiveDays            int     `json:"active_days"`
	TotalSessions         int     `json:"total_sessions"`
	PeakHour              int     `json:"peak_hour"`
	LateNightSessions     int     `json:"late_night_sessions"`
	WeekendSessions       int     `json:"weekend_sessions"`
	TotalWeekdaySessions  int     `json:"total_weekday_sessions"`
	DailyActivityVariance float64 `json:"daily_activity_variance"`
}

// QuizPerformance contains quiz metrics.
type QuizPerformance struct {
	QuizAttempts   int     `json:"quiz_attempts"`
	AvgScore       float64 `json:"avg_score"`
	CompletionRate float64 `json:"completion_rate"`
}

// PredictResponse represents the response from Learning Pulse.
type PredictResponse struct {
	UserID           string  `json:"user_id"`
	Persona          string  `json:"persona"`
	Confidence       float64 `json:"confidence"`
	IsLowConfidence  bool    `json:"is_low_confidence"`
	ProcessingTimeMs float64 `json:"processing_time_ms"`
}

// Config holds configuration for the Learning Pulse client.
type Config struct {
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
}

// Client communicates with the Learning Pulse service.
type Client struct {
	httpClient *http.Client
	baseURL    string
	maxRetries int
}

// NewClient creates a new Learning Pulse client.
func NewClient(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = 2
	}

	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    cfg.BaseURL,
		maxRetries: maxRetries,
	}
}


// GetPersona fetches the user's learning persona from the Learning Pulse service.
// If behavior data is not provided, it returns PersonaUnknown.
func (c *Client) GetPersona(ctx context.Context, userID string, behaviorData *BehaviorData) (Persona, error) {
	if behaviorData == nil {
		return PersonaUnknown, nil
	}

	reqBody := PredictRequest{
		UserID:       userID,
		BehaviorData: behaviorData,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return PersonaUnknown, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/learning-pulse/predict-persona", c.baseURL)

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
		if err != nil {
			return PersonaUnknown, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusServiceUnavailable {
			lastErr = fmt.Errorf("learning pulse service unavailable")
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return PersonaUnknown, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		var predictResp PredictResponse
		if err := json.NewDecoder(resp.Body).Decode(&predictResp); err != nil {
			return PersonaUnknown, fmt.Errorf("failed to decode response: %w", err)
		}

		return Persona(predictResp.Persona), nil
	}

	if lastErr != nil {
		return PersonaUnknown, lastErr
	}

	return PersonaUnknown, nil
}
