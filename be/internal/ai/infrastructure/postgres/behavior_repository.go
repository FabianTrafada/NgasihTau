package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/ai/infrastructure/learningpulse"
)

// BehaviorDataRepository aggregates user behavior data from multiple tables.
type BehaviorDataRepository struct {
	aiDB   *pgxpool.Pool
	podDB  *pgxpool.Pool
	userDB *pgxpool.Pool
}

// NewBehaviorDataRepository creates a new behavior data repository.
func NewBehaviorDataRepository(aiDB, podDB, userDB *pgxpool.Pool) *BehaviorDataRepository {
	return &BehaviorDataRepository{
		aiDB:   aiDB,
		podDB:  podDB,
		userDB: userDB,
	}
}

// GetBehaviorData retrieves aggregated behavior data for a user over the last 30 days.
func (r *BehaviorDataRepository) GetBehaviorData(ctx context.Context, userID uuid.UUID) (*learningpulse.BehaviorData, error) {
	analysisPeriod := 30
	since := time.Now().AddDate(0, 0, -analysisPeriod)

	chatBehavior, err := r.getChatBehavior(ctx, userID, since)
	if err != nil {
		chatBehavior = learningpulse.ChatBehavior{}
	}

	materialInteraction, err := r.getMaterialInteraction(ctx, userID, since)
	if err != nil {
		materialInteraction = learningpulse.MaterialInteraction{AvgScrollDepth: 0.5}
	}

	activityPattern, err := r.getActivityPattern(ctx, userID, since)
	if err != nil {
		activityPattern = learningpulse.ActivityPattern{PeakHour: 12}
	}

	return &learningpulse.BehaviorData{
		UserID:             userID.String(),
		AnalysisPeriodDays: analysisPeriod,
		Chat:               chatBehavior,
		Material:           materialInteraction,
		Activity:           activityPattern,
	}, nil
}

func (r *BehaviorDataRepository) getChatBehavior(ctx context.Context, userID uuid.UUID, since time.Time) (learningpulse.ChatBehavior, error) {
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

	var cb learningpulse.ChatBehavior
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
		return learningpulse.ChatBehavior{}, err
	}

	return cb, nil
}


func (r *BehaviorDataRepository) getMaterialInteraction(ctx context.Context, userID uuid.UUID, since time.Time) (learningpulse.MaterialInteraction, error) {
	// If podDB is not available, return defaults
	if r.podDB == nil {
		return learningpulse.MaterialInteraction{AvgScrollDepth: 0.5}, nil
	}

	query := `
		SELECT 
			COALESCE(SUM(time_spent_seconds), 0)::int as total_time_spent,
			COUNT(*) as total_views,
			COUNT(DISTINCT material_id) as unique_materials,
			COALESCE(AVG(scroll_depth), 0.5)::float8 as avg_scroll_depth
		FROM pod_interactions
		WHERE user_id = $1 AND created_at >= $2 AND interaction_type = 'view'
	`

	var mi learningpulse.MaterialInteraction
	err := r.podDB.QueryRow(ctx, query, userID, since).Scan(
		&mi.TotalTimeSpentSeconds,
		&mi.TotalViews,
		&mi.UniqueMaterialsViewed,
		&mi.AvgScrollDepth,
	)
	if err != nil {
		return learningpulse.MaterialInteraction{AvgScrollDepth: 0.5}, err
	}

	// Get bookmark count separately
	bookmarkQuery := `
		SELECT COUNT(*) FROM bookmarks WHERE user_id = $1 AND created_at >= $2
	`
	r.podDB.QueryRow(ctx, bookmarkQuery, userID, since).Scan(&mi.BookmarkCount)

	return mi, nil
}

func (r *BehaviorDataRepository) getActivityPattern(ctx context.Context, userID uuid.UUID, since time.Time) (learningpulse.ActivityPattern, error) {
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

	var ap learningpulse.ActivityPattern
	err := r.aiDB.QueryRow(ctx, query, userID, since).Scan(
		&ap.ActiveDays,
		&ap.TotalSessions,
		&ap.PeakHour,
		&ap.LateNightSessions,
		&ap.WeekendSessions,
		&ap.TotalWeekdaySessions,
		&ap.DailyActivityVariance,
	)
	if err != nil {
		return learningpulse.ActivityPattern{PeakHour: 12}, err
	}

	return ap, nil
}
