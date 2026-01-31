package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/user/application"
	"ngasihtau/internal/user/domain"
)

// BehaviorRepository implements the BehaviorRepository interface.
// It aggregates user behavior data from multiple databases.
type BehaviorRepository struct {
	userDB *pgxpool.Pool
	aiDB   *pgxpool.Pool
	podDB  *pgxpool.Pool
}

// NewBehaviorRepository creates a new BehaviorRepository.
// aiDB and podDB can be nil if those databases are not available.
func NewBehaviorRepository(userDB, aiDB, podDB *pgxpool.Pool) application.BehaviorRepository {
	return &BehaviorRepository{
		userDB: userDB,
		aiDB:   aiDB,
		podDB:  podDB,
	}
}

// GetChatBehavior retrieves chat behavior metrics for a user.
func (r *BehaviorRepository) GetChatBehavior(ctx context.Context, userID uuid.UUID, days int) (*domain.ChatBehavior, error) {
	// If aiDB is not available, return defaults
	if r.aiDB == nil {
		return &domain.ChatBehavior{}, nil
	}

	since := time.Now().AddDate(0, 0, -days)

	query := `
		WITH session_stats AS (
			SELECT 
				COUNT(DISTINCT cs.id) as unique_sessions,
				COALESCE(SUM(EXTRACT(EPOCH FROM (cs.updated_at - cs.created_at)) / 60), 0) as total_duration_minutes
			FROM chat_sessions cs
			WHERE cs.user_id = $1 AND cs.created_at >= $2
		),
		message_stats AS (
			SELECT 
				COUNT(*) as total_messages,
				COUNT(*) FILTER (WHERE cm.role = 'user') as user_messages,
				COUNT(*) FILTER (WHERE cm.role = 'assistant') as assistant_messages,
				COUNT(*) FILTER (WHERE cm.role = 'user' AND cm.content LIKE '%?%') as question_count,
				COALESCE(AVG(LENGTH(cm.content)) FILTER (WHERE cm.role = 'user'), 0) as avg_message_length,
				COUNT(*) FILTER (WHERE cm.feedback = 'thumbs_up') as thumbs_up_count,
				COUNT(*) FILTER (WHERE cm.feedback = 'thumbs_down') as thumbs_down_count
			FROM chat_messages cm
			JOIN chat_sessions cs ON cm.session_id = cs.id
			WHERE cs.user_id = $1 AND cm.created_at >= $2
		)
		SELECT 
			COALESCE(ms.total_messages, 0)::int,
			COALESCE(ms.user_messages, 0)::int,
			COALESCE(ms.assistant_messages, 0)::int,
			COALESCE(ms.question_count, 0)::int,
			COALESCE(ms.avg_message_length, 0)::float8,
			COALESCE(ms.thumbs_up_count, 0)::int,
			COALESCE(ms.thumbs_down_count, 0)::int,
			COALESCE(ss.unique_sessions, 0)::int,
			COALESCE(ss.total_duration_minutes, 0)::float8
		FROM message_stats ms, session_stats ss
	`

	var cb domain.ChatBehavior
	err := r.aiDB.QueryRow(ctx, query, userID, since).Scan(
		&cb.TotalMessages,
		&cb.UserMessages,
		&cb.AssistantMessages,
		&cb.QuestionCount,
		&cb.AvgMessageLength,
		&cb.ThumbsUpCount,
		&cb.ThumbsDownCount,
		&cb.UniqueSessions,
		&cb.TotalSessionDurationMinutes,
	)
	if err != nil {
		return &domain.ChatBehavior{}, err
	}

	return &cb, nil
}

// GetMaterialBehavior retrieves material interaction metrics for a user.
func (r *BehaviorRepository) GetMaterialBehavior(ctx context.Context, userID uuid.UUID, days int) (*domain.MaterialBehavior, error) {
	// If podDB is not available, return defaults
	if r.podDB == nil {
		return &domain.MaterialBehavior{AvgScrollDepth: 0.5}, nil
	}

	since := time.Now().AddDate(0, 0, -days)

	// Query pod_interactions for material view data
	query := `
		SELECT 
			COALESCE(SUM((metadata->>'time_spent_seconds')::int), 0)::int as total_time_spent,
			COUNT(*) as total_views,
			COUNT(DISTINCT (metadata->>'material_id')::uuid) as unique_materials,
			COALESCE(AVG((metadata->>'scroll_depth')::float), 0.5)::float8 as avg_scroll_depth
		FROM pod_interactions
		WHERE user_id = $1 AND created_at >= $2 AND interaction_type IN ('view', 'material_view')
	`

	var mb domain.MaterialBehavior
	err := r.podDB.QueryRow(ctx, query, userID, since).Scan(
		&mb.TotalTimeSpentSeconds,
		&mb.TotalViews,
		&mb.UniqueMaterialsViewed,
		&mb.AvgScrollDepth,
	)
	if err != nil {
		// Return defaults on error
		return &domain.MaterialBehavior{AvgScrollDepth: 0.5}, nil
	}

	// Get bookmark count separately (bookmarks might be in material DB or pod DB)
	bookmarkQuery := `
		SELECT COUNT(*) FROM pod_interactions 
		WHERE user_id = $1 AND created_at >= $2 AND interaction_type = 'material_bookmark'
	`
	r.podDB.QueryRow(ctx, bookmarkQuery, userID, since).Scan(&mb.BookmarkCount)

	return &mb, nil
}

// GetActivityBehavior retrieves activity pattern metrics for a user.
func (r *BehaviorRepository) GetActivityBehavior(ctx context.Context, userID uuid.UUID, days int) (*domain.ActivityBehavior, error) {
	// If aiDB is not available, return defaults
	if r.aiDB == nil {
		return &domain.ActivityBehavior{PeakHour: 12}, nil
	}

	since := time.Now().AddDate(0, 0, -days)

	query := `
		WITH daily_activity AS (
			SELECT 
				DATE(created_at) as activity_date,
				COUNT(*) as session_count,
				EXTRACT(HOUR FROM created_at) as hour,
				EXTRACT(DOW FROM created_at) as dow
			FROM chat_sessions
			WHERE user_id = $1 AND created_at >= $2
			GROUP BY DATE(created_at), EXTRACT(HOUR FROM created_at), EXTRACT(DOW FROM created_at)
		),
		aggregated AS (
			SELECT
				COUNT(DISTINCT activity_date) as active_days,
				SUM(session_count) as total_sessions,
				MODE() WITHIN GROUP (ORDER BY hour) as peak_hour,
				SUM(session_count) FILTER (WHERE hour >= 23 OR hour < 5) as late_night_sessions,
				SUM(session_count) FILTER (WHERE dow IN (0, 6)) as weekend_sessions,
				SUM(session_count) FILTER (WHERE dow NOT IN (0, 6)) as weekday_sessions,
				COALESCE(STDDEV(session_count), 0) as daily_variance
			FROM daily_activity
		)
		SELECT 
			COALESCE(active_days, 0)::int,
			COALESCE(total_sessions, 0)::int,
			COALESCE(peak_hour, 12)::int,
			COALESCE(late_night_sessions, 0)::int,
			COALESCE(weekend_sessions, 0)::int,
			COALESCE(weekday_sessions, 0)::int,
			COALESCE(daily_variance, 0)::float8
		FROM aggregated
	`

	var ab domain.ActivityBehavior
	err := r.aiDB.QueryRow(ctx, query, userID, since).Scan(
		&ab.ActiveDays,
		&ab.TotalSessions,
		&ab.PeakHour,
		&ab.LateNightSessions,
		&ab.WeekendSessions,
		&ab.TotalWeekdaySessions,
		&ab.DailyActivityVariance,
	)
	if err != nil {
		return &domain.ActivityBehavior{PeakHour: 12}, err
	}

	return &ab, nil
}

// GetQuizBehavior retrieves quiz performance metrics for a user.
// NOTE: Quiz tracking is not yet implemented, so this returns defaults.
func (r *BehaviorRepository) GetQuizBehavior(ctx context.Context, userID uuid.UUID, days int) (*domain.QuizBehavior, error) {
	// TODO: Implement when quiz tracking tables are created
	// For now, return empty defaults
	return &domain.QuizBehavior{
		QuizAttempts:   0,
		AvgScore:       0,
		CompletionRate: 0,
	}, nil
}
