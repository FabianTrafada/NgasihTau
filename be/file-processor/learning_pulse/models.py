"""
Pydantic Data Models for Learning Pulse Module

This module defines all input/output data models used throughout the Learning Pulse
predictive analytics system, including:
- Behavior data models (chat, material, activity, quiz)
- Feature vector model for ML input
- Classification and guardrail result models
- API request/response models
"""

from enum import Enum
from typing import Optional, List, Dict
from pydantic import BaseModel, Field


class EngagementTrend(str, Enum):
    """Engagement trend over the analysis period."""
    INCREASING = "increasing"
    STABLE = "stable"
    DECLINING = "declining"


class LearningPersona(str, Enum):
    """
    Learning Persona classifications.
    
    Each persona represents a distinct learning behavior pattern that helps
    teachers understand how students engage with educational materials.
    """
    SKIMMER = "skimmer"
    STRUGGLER = "struggler"
    ANXIOUS = "anxious"
    BURNOUT = "burnout"
    MASTER = "master"
    PROCRASTINATOR = "procrastinator"
    DEEP_DIVER = "deep_diver"
    SOCIAL_LEARNER = "social_learner"
    PERFECTIONIST = "perfectionist"
    LOST = "lost"


# ============================================================================
# Input Behavior Models
# ============================================================================

class ChatBehavior(BaseModel):
    """Chat interaction metrics from chat_sessions and chat_messages tables."""
    total_messages: int = Field(default=0, ge=0)
    user_messages: int = Field(default=0, ge=0)
    assistant_messages: int = Field(default=0, ge=0)
    question_count: int = Field(default=0, ge=0)  # Messages ending with "?"
    avg_message_length: float = Field(default=0.0, ge=0)
    thumbs_up_count: int = Field(default=0, ge=0)
    thumbs_down_count: int = Field(default=0, ge=0)
    unique_sessions: int = Field(default=0, ge=0)
    total_session_duration_minutes: float = Field(default=0.0, ge=0)


class MaterialInteraction(BaseModel):
    """Material consumption metrics from pod_interactions table."""
    total_time_spent_seconds: int = Field(default=0, ge=0)
    total_views: int = Field(default=0, ge=0)
    unique_materials_viewed: int = Field(default=0, ge=0)
    bookmark_count: int = Field(default=0, ge=0)
    avg_scroll_depth: float = Field(default=0.5, ge=0.0, le=1.0)


class ActivityPattern(BaseModel):
    """Temporal activity metrics derived from interaction timestamps."""
    active_days: int = Field(default=0, ge=0)
    total_sessions: int = Field(default=0, ge=0)
    peak_hour: int = Field(default=12, ge=0, le=23)
    late_night_sessions: int = Field(default=0, ge=0)  # Sessions between 23:00-05:00
    weekend_sessions: int = Field(default=0, ge=0)
    total_weekday_sessions: int = Field(default=0, ge=0)
    daily_activity_variance: float = Field(default=0.0, ge=0)  # Lower = more consistent


class QuizPerformance(BaseModel):
    """Quiz/assessment metrics (optional, may not be available)."""
    quiz_attempts: int = Field(default=0, ge=0)
    avg_score: float = Field(default=0.0, ge=0.0, le=100.0)
    completion_rate: float = Field(default=0.0, ge=0.0, le=1.0)


class BehaviorData(BaseModel):
    """Complete behavior data for a user."""
    user_id: str
    analysis_period_days: int = Field(default=30, ge=1)
    chat: ChatBehavior = Field(default_factory=ChatBehavior)
    material: MaterialInteraction = Field(default_factory=MaterialInteraction)
    activity: ActivityPattern = Field(default_factory=ActivityPattern)
    quiz: Optional[QuizPerformance] = None


# ============================================================================
# Feature Vector Model
# ============================================================================

class FeatureVector(BaseModel):
    """
    Normalized feature vector for ML model input.
    
    All features are normalized to [0.0, 1.0] range.
    Total features: 25
    """
    # Chat features (8)
    chat_message_ratio: float = Field(ge=0.0, le=1.0)
    question_frequency: float = Field(ge=0.0, le=1.0)
    avg_message_length_norm: float = Field(ge=0.0, le=1.0)
    feedback_ratio: float = Field(ge=0.0, le=1.0)
    feedback_engagement: float = Field(ge=0.0, le=1.0)
    session_count_norm: float = Field(ge=0.0, le=1.0)
    messages_per_session: float = Field(ge=0.0, le=1.0)
    session_duration_norm: float = Field(ge=0.0, le=1.0)
    
    # Material features (7)
    time_spent_norm: float = Field(ge=0.0, le=1.0)
    view_count_norm: float = Field(ge=0.0, le=1.0)
    material_diversity: float = Field(ge=0.0, le=1.0)
    avg_time_per_view_norm: float = Field(ge=0.0, le=1.0)
    bookmark_ratio: float = Field(ge=0.0, le=1.0)
    scroll_depth: float = Field(ge=0.0, le=1.0)
    material_engagement_score: float = Field(ge=0.0, le=1.0)
    
    # Activity features (7)
    active_days_ratio: float = Field(ge=0.0, le=1.0)
    session_frequency: float = Field(ge=0.0, le=1.0)
    consistency_score: float = Field(ge=0.0, le=1.0)
    late_night_ratio: float = Field(ge=0.0, le=1.0)
    weekend_ratio: float = Field(ge=0.0, le=1.0)
    peak_hour_norm: float = Field(ge=0.0, le=1.0)
    engagement_trend_encoded: float = Field(ge=0.0, le=1.0)
    
    # Quiz features (3) - optional, defaults for missing data
    quiz_score_norm: float = Field(default=0.5, ge=0.0, le=1.0)
    quiz_completion_norm: float = Field(default=0.5, ge=0.0, le=1.0)
    quiz_attempt_frequency: float = Field(default=0.0, ge=0.0, le=1.0)


# ============================================================================
# Classification Result Models
# ============================================================================

class ClassificationResult(BaseModel):
    """Raw classification result from ML model."""
    persona: LearningPersona
    confidence: float = Field(ge=0.0, le=1.0)
    probabilities: Dict[str, float]  # All persona probabilities
    is_low_confidence: bool  # True if confidence < 0.5


class GuardrailOverride(BaseModel):
    """Record of a guardrail override."""
    original_persona: LearningPersona
    final_persona: LearningPersona
    rule_triggered: str
    reason: str


class GuardrailFlags(BaseModel):
    """Flags raised by guardrails without overriding."""
    potential_anxious: bool = False
    potential_burnout: bool = False
    needs_attention: bool = False
    flag_reasons: List[str] = Field(default_factory=list)


class GuardrailResult(BaseModel):
    """Result after applying logic guardrails."""
    persona: LearningPersona
    confidence: float = Field(ge=0.0, le=1.0)
    was_overridden: bool
    override: Optional[GuardrailOverride] = None
    flags: GuardrailFlags = Field(default_factory=GuardrailFlags)


# ============================================================================
# Recommendation Models
# ============================================================================

class Recommendation(BaseModel):
    """Single actionable recommendation."""
    id: str
    title: str
    description: str
    action_type: str  # "content", "ui", "notification", "feature"
    priority: int = Field(ge=1)  # 1 = highest


class PersonaRecommendations(BaseModel):
    """Recommendations for a specific persona."""
    persona: LearningPersona
    summary: str
    recommendations: List[Recommendation]
    ui_hints: Dict[str, str] = Field(default_factory=dict)


# ============================================================================
# API Request/Response Models
# ============================================================================

class PredictRequest(BaseModel):
    """Request body for persona prediction."""
    user_id: str
    behavior_data: BehaviorData
    quiz_score: Optional[float] = Field(default=None, ge=0.0, le=100.0)
    previous_persona: Optional[str] = None


class FeatureSummary(BaseModel):
    """Summary of extracted features for transparency."""
    chat_engagement: str  # "low", "medium", "high"
    material_consumption: str  # "low", "medium", "high"
    activity_consistency: str  # "low", "medium", "high"
    key_indicators: List[str] = Field(default_factory=list)


class PredictResponse(BaseModel):
    """Response for persona prediction."""
    user_id: str
    persona: str
    confidence: float = Field(ge=0.0, le=1.0)
    is_low_confidence: bool
    recommendations: List[Recommendation]
    feature_summary: FeatureSummary
    override_info: Optional[GuardrailOverride] = None
    flags: List[str] = Field(default_factory=list)
    processing_time_ms: float = Field(ge=0)


class HealthResponse(BaseModel):
    """Health check response for Learning Pulse."""
    status: str  # "healthy" or "unhealthy"
    model_loaded: bool
    model_version: str
    last_training_date: str
    error_details: Optional[str] = None
