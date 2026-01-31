import apiClient from "@/lib/api-client";

/**
 * Behavior Data Types matching backend learningpulse.BehaviorData
 */
export interface ChatBehavior {
  total_messages: number;
  user_messages: number;
  assistant_messages: number;
  question_count: number;
  avg_message_length: number;
  thumbs_up_count: number;
  thumbs_down_count: number;
  unique_sessions: number;
  total_session_duration_minutes: number;
}

export interface MaterialInteraction {
  total_time_spent_seconds: number;
  total_views: number;
  unique_materials_viewed: number;
  bookmark_count: number;
  avg_scroll_depth: number;
}

export interface ActivityPattern {
  active_days: number;
  total_sessions: number;
  peak_hour: number;
  late_night_sessions: number;
  weekend_sessions: number;
  total_weekday_sessions: number;
  daily_activity_variance: number;
}

export interface QuizPerformance {
  quiz_attempts: number;
  avg_score: number;
  completion_rate: number;
}

export interface BehaviorData {
  user_id: string;
  analysis_period_days: number;
  chat: ChatBehavior;
  material: MaterialInteraction;
  activity: ActivityPattern;
  quiz?: QuizPerformance;
}

export interface PredictPersonaRequest {
  user_id: string;
  behavior_data: BehaviorData;
  quiz_score?: number | null;
  previous_persona?: string | null;
}

export interface FeatureSummary {
  chat_engagement: "low" | "medium" | "high";
  material_consumption: "low" | "medium" | "high";
  activity_consistency: "low" | "medium" | "high";
  key_indicators: string[];
}

export interface Recommendation {
  id: string;
  title: string;
  description: string;
  action_type: string;
  priority: number;
}

export interface PredictPersonaResponse {
  user_id: string;
  persona: string;
  confidence: number;
  is_low_confidence: boolean;
  recommendations: Recommendation[];
  feature_summary: FeatureSummary;
  override_info?: string | null;
  flags: string[];
  processing_time_ms: number;
}

/**
 * Fetch user behavior data from AI service
 * Endpoint: GET /api/v1/ai/behavior
 */
export async function getBehaviorData(): Promise<BehaviorData> {
  const response = await apiClient.get<{ data: BehaviorData }>("/api/v1/ai/behavior");
  return response.data.data;
}

/**
 * Predict user persona using Learning Pulse service
 * Endpoint: POST /api/v1/learning-pulse/predict-persona
 */
export async function predictPersona(
  request: PredictPersonaRequest
): Promise<PredictPersonaResponse> {
  const response = await apiClient.post<PredictPersonaResponse>(
    "/api/v1/learning-pulse/predict-persona",
    request
  );
  // Learning Pulse returns data directly (not wrapped in {data: ...})
  return response.data;
}

/**
 * Fetch behavior data and predict persona in one call
 * Convenience function that combines getBehaviorData + predictPersona
 */
export async function fetchUserPersona(
  userId: string,
  previousPersona?: string | null
): Promise<PredictPersonaResponse> {
  // Step 1: Get behavior data from AI service
  const behaviorData = await getBehaviorData();

  // Step 2: Build request for Learning Pulse
  const request: PredictPersonaRequest = {
    user_id: userId,
    behavior_data: behaviorData,
    quiz_score: behaviorData.quiz?.avg_score ?? null,
    previous_persona: previousPersona ?? null,
  };

  // Step 3: Predict persona
  return predictPersona(request);
}

/**
 * Build a complete UserBehavior object for Learning Pulse
 * This matches the UserBehavior type in types/userBehavior.ts
 */
export async function buildUserBehavior(
  userId: string,
  previousPersona?: string | null
): Promise<{
  user_id: string;
  behavior_data: BehaviorData;
  quiz_score: number;
  previous_persona: string | null;
}> {
  const behaviorData = await getBehaviorData();

  return {
    user_id: userId,
    behavior_data: behaviorData,
    quiz_score: behaviorData.quiz?.avg_score ?? 0,
    previous_persona: previousPersona ?? null,
  };
}
