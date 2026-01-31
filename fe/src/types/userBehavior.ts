export interface UserBehavior
{
  user_id: string;
  behavior_data: {
    user_id: string;
    analysis_period_days: number;
    chat: {
      total_messages: number;
      user_messages: number;
      assistant_messages: number;
      question_count: number;
      avg_message_length: number;
      thumbs_up_count: number;
      thumbs_down_count: number;
      unique_sessions: number;
      total_session_duration_minutes: number;
    },
    material: {
      total_time_spent_seconds: number;
      total_views: number;
      unique_materials_viewed: number;
      bookmark_count: number;
      avg_scroll_depth: number;
    }
    activity: {
      active_days: number;
      total_sessions: number;
      peak_hour: number;
      late_night_sessions: number;
      weekend_sessions: number;
      total_weekday_sessions: number;
      daily_activity_variance: number;
    };
    quiz: {
      quiz_attempts: number;
      avg_score: number;
      completion_rate: number;
    }
  };
  quiz_score: number;
  previous_persona: string;
}