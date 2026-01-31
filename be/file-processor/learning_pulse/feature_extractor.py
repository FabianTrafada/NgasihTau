"""
Feature Extractor for Learning Pulse Module

Responsible for transforming raw behavior data into normalized feature vectors
suitable for ML model input. Handles:
- Chat feature extraction (Requirements 2.1-2.8)
- Material feature extraction (Requirements 3.1-3.7)
- Activity feature extraction (Requirements 4.1-4.7)
- Feature normalization (Requirements 9.1-9.5)
- Missing value imputation (Requirements 9.2-9.4)
"""

from typing import Optional
from pathlib import Path
import logging

import joblib

from .models import (
    BehaviorData,
    FeatureVector,
    EngagementTrend,
)

logger = logging.getLogger("learning_pulse.feature_extractor")


class FeatureExtractor:
    """
    Extracts and normalizes features from behavior data.
    
    Transforms raw BehaviorData into a normalized FeatureVector with all
    values in the [0.0, 1.0] range for ML model input.
    """
    
    # Normalization constants (max expected values for min-max scaling)
    MAX_MESSAGES = 1000
    MAX_MESSAGE_LENGTH = 500
    MAX_SESSIONS = 100
    MAX_SESSION_DURATION = 600  # minutes
    MAX_TIME_SPENT = 36000  # seconds (10 hours)
    MAX_VIEWS = 200
    MAX_SESSIONS_PER_DAY = 10
    MAX_QUIZ_ATTEMPTS = 50
    
    def __init__(self, scaler_path: Optional[str] = None):
        """
        Initialize the feature extractor.
        
        Args:
            scaler_path: Optional path to a saved scaler for consistent inference
        """
        self.scaler = None
        self.scaler_path = scaler_path
        if scaler_path:
            self._load_scaler(scaler_path)
    
    def _load_scaler(self, scaler_path: str) -> None:
        """
        Load a saved scaler from file using joblib.
        
        Validates: Requirement 9.5 - Serialize the scaler alongside the model
        
        Args:
            scaler_path: Path to the saved scaler file
        """
        try:
            path = Path(scaler_path)
            if path.exists():
                self.scaler = joblib.load(scaler_path)
                logger.info(f"Loaded scaler from {scaler_path}")
            else:
                logger.warning(f"Scaler file not found at {scaler_path}, using default normalization")
        except Exception as e:
            logger.error(f"Failed to load scaler from {scaler_path}: {e}")
            self.scaler = None
    
    def save_scaler(self, scaler_path: str, scaler: object) -> None:
        """
        Save a scaler to file using joblib for consistent inference.
        
        Validates: Requirement 9.5 - Serialize the scaler alongside the model
        
        Args:
            scaler_path: Path where the scaler should be saved
            scaler: The scaler object to save (e.g., sklearn MinMaxScaler)
        """
        try:
            path = Path(scaler_path)
            path.parent.mkdir(parents=True, exist_ok=True)
            joblib.dump(scaler, scaler_path)
            self.scaler = scaler
            self.scaler_path = scaler_path
            logger.info(f"Saved scaler to {scaler_path}")
        except Exception as e:
            logger.error(f"Failed to save scaler to {scaler_path}: {e}")
            raise
    
    def extract(self, data: BehaviorData) -> FeatureVector:
        """
        Extract normalized features from behavior data.
        
        Combines chat, material, activity, and quiz features into a single
        normalized feature vector. All features are scaled to [0, 1] range.
        
        Validates:
        - Requirement 9.1: Normalize all numeric features to 0-1 range using min-max scaling
        - Requirement 9.2: Handle missing values by imputing with feature-specific defaults
        - Requirement 9.3: Use 0 for count-based features when missing
        - Requirement 9.4: Use 0.5 for ratio-based features when missing
        
        Args:
            data: Complete behavior data for a user
            
        Returns:
            Normalized feature vector for ML model input with all values in [0, 1]
        """
        # Extract features from each category
        chat_features = self._extract_chat_features(data)
        material_features = self._extract_material_features(data)
        activity_features = self._extract_activity_features(data)
        quiz_features = self._extract_quiz_features(data)
        
        # Combine all features into a FeatureVector
        # All features are already normalized to [0, 1] by the extraction methods
        feature_vector = FeatureVector(
            # Chat features (8)
            chat_message_ratio=chat_features["chat_message_ratio"],
            question_frequency=chat_features["question_frequency"],
            avg_message_length_norm=chat_features["avg_message_length_norm"],
            feedback_ratio=chat_features["feedback_ratio"],
            feedback_engagement=chat_features["feedback_engagement"],
            session_count_norm=chat_features["session_count_norm"],
            messages_per_session=chat_features["messages_per_session"],
            session_duration_norm=chat_features["session_duration_norm"],
            
            # Material features (7)
            time_spent_norm=material_features["time_spent_norm"],
            view_count_norm=material_features["view_count_norm"],
            material_diversity=material_features["material_diversity"],
            avg_time_per_view_norm=material_features["avg_time_per_view_norm"],
            bookmark_ratio=material_features["bookmark_ratio"],
            scroll_depth=material_features["scroll_depth"],
            material_engagement_score=material_features["material_engagement_score"],
            
            # Activity features (7)
            active_days_ratio=activity_features["active_days_ratio"],
            session_frequency=activity_features["session_frequency"],
            consistency_score=activity_features["consistency_score"],
            late_night_ratio=activity_features["late_night_ratio"],
            weekend_ratio=activity_features["weekend_ratio"],
            peak_hour_norm=activity_features["peak_hour_norm"],
            engagement_trend_encoded=activity_features["engagement_trend_encoded"],
            
            # Quiz features (3)
            quiz_score_norm=quiz_features["quiz_score_norm"],
            quiz_completion_norm=quiz_features["quiz_completion_norm"],
            quiz_attempt_frequency=quiz_features["quiz_attempt_frequency"],
        )
        
        logger.debug(f"Extracted feature vector for user {data.user_id}")
        return feature_vector
    
    def _extract_chat_features(self, data: BehaviorData) -> dict:
        """
        Extract chat-related features.
        
        Calculates:
        - chat_message_ratio: user_messages / total_messages
        - question_frequency: question_count / user_messages
        - avg_message_length_norm: normalized average message length
        - feedback_ratio: thumbs_up / (thumbs_up + thumbs_down)
        - feedback_engagement: total_feedback / total_messages
        - session_count_norm: normalized session count
        - messages_per_session: total_messages / unique_sessions
        - session_duration_norm: normalized session duration
        
        Validates: Requirements 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8
        """
        chat = data.chat
        
        # Requirement 2.2: Calculate ratio of user messages to total messages
        # Default to 0.5 (neutral) if no messages exist
        chat_message_ratio = self._safe_divide(
            chat.user_messages, 
            chat.total_messages, 
            default=0.5
        )
        
        # Requirement 2.3: Calculate question frequency (questions / user messages)
        # Default to 0 if no user messages exist
        question_frequency = self._safe_divide(
            chat.question_count, 
            chat.user_messages, 
            default=0.0
        )
        # Normalize to [0, 1] - cap at 1.0 (100% questions)
        question_frequency = min(1.0, question_frequency)
        
        # Requirement 2.4: Normalize average message length
        avg_message_length_norm = self._normalize(
            chat.avg_message_length, 
            self.MAX_MESSAGE_LENGTH
        )
        
        # Requirements 2.5, 2.6: Calculate feedback ratio
        # thumbs_up / (thumbs_up + thumbs_down), default to 0.5 if no feedback
        total_feedback = chat.thumbs_up_count + chat.thumbs_down_count
        feedback_ratio = self._safe_divide(
            chat.thumbs_up_count, 
            total_feedback, 
            default=0.5
        )
        
        # Calculate feedback engagement: total_feedback / total_messages
        # Default to 0 if no messages exist
        feedback_engagement = self._safe_divide(
            total_feedback, 
            chat.total_messages, 
            default=0.0
        )
        # Normalize to [0, 1] - cap at 1.0
        feedback_engagement = min(1.0, feedback_engagement)
        
        # Requirement 2.7: Normalize session count
        session_count_norm = self._normalize(
            chat.unique_sessions, 
            self.MAX_SESSIONS
        )
        
        # Requirement 2.8: Calculate messages per session
        # Default to 0 if no sessions exist
        raw_messages_per_session = self._safe_divide(
            chat.total_messages, 
            chat.unique_sessions, 
            default=0.0
        )
        # Normalize: assume max ~20 messages per session is high engagement
        messages_per_session = self._normalize(raw_messages_per_session, 20.0)
        
        # Normalize session duration
        session_duration_norm = self._normalize(
            chat.total_session_duration_minutes, 
            self.MAX_SESSION_DURATION
        )
        
        return {
            "chat_message_ratio": chat_message_ratio,
            "question_frequency": question_frequency,
            "avg_message_length_norm": avg_message_length_norm,
            "feedback_ratio": feedback_ratio,
            "feedback_engagement": feedback_engagement,
            "session_count_norm": session_count_norm,
            "messages_per_session": messages_per_session,
            "session_duration_norm": session_duration_norm,
        }
    
    def _extract_material_features(self, data: BehaviorData) -> dict:
        """
        Extract material interaction features.
        
        Calculates:
        - time_spent_norm: normalized total time spent
        - view_count_norm: normalized view count
        - material_diversity: unique_materials / total_views
        - avg_time_per_view_norm: normalized time per view
        - bookmark_ratio: bookmarks / views
        - scroll_depth: average scroll depth
        - material_engagement_score: composite engagement score
        
        Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7
        """
        material = data.material
        
        # Requirement 3.1: Calculate total time spent on materials (normalized)
        time_spent_norm = self._normalize(
            material.total_time_spent_seconds,
            self.MAX_TIME_SPENT
        )
        
        # Requirement 3.2: Count total material views (normalized)
        view_count_norm = self._normalize(
            material.total_views,
            self.MAX_VIEWS
        )
        
        # Requirement 3.3: Calculate material diversity (unique / total views)
        # Default to 0 if no views exist
        material_diversity = self._safe_divide(
            material.unique_materials_viewed,
            material.total_views,
            default=0.0
        )
        # Cap at 1.0 (unique can't exceed total views in valid data)
        material_diversity = min(1.0, material_diversity)
        
        # Requirement 3.4: Calculate average time per material view
        # Default to 0 if no views exist
        avg_time_per_view = self._safe_divide(
            material.total_time_spent_seconds,
            material.total_views,
            default=0.0
        )
        # Normalize: assume max ~600 seconds (10 minutes) per view is high engagement
        avg_time_per_view_norm = self._normalize(avg_time_per_view, 600.0)
        
        # Requirements 3.5, 3.6: Calculate bookmark ratio (bookmarks / views)
        # Default to 0 if no views exist
        bookmark_ratio = self._safe_divide(
            material.bookmark_count,
            material.total_views,
            default=0.0
        )
        # Cap at 1.0 (can't have more bookmarks than views in typical usage)
        bookmark_ratio = min(1.0, bookmark_ratio)
        
        # Requirement 3.7: Scroll depth average (already in [0, 1] range)
        scroll_depth = material.avg_scroll_depth
        # Ensure it's within bounds
        scroll_depth = max(0.0, min(1.0, scroll_depth))
        
        # Calculate material_engagement_score as a composite metric
        # Weighted average of time_spent_norm, scroll_depth, and bookmark_ratio
        # Weights: time_spent (40%), scroll_depth (40%), bookmark_ratio (20%)
        material_engagement_score = (
            0.4 * time_spent_norm +
            0.4 * scroll_depth +
            0.2 * bookmark_ratio
        )
        # Ensure it's within [0, 1] bounds
        material_engagement_score = max(0.0, min(1.0, material_engagement_score))
        
        return {
            "time_spent_norm": time_spent_norm,
            "view_count_norm": view_count_norm,
            "material_diversity": material_diversity,
            "avg_time_per_view_norm": avg_time_per_view_norm,
            "bookmark_ratio": bookmark_ratio,
            "scroll_depth": scroll_depth,
            "material_engagement_score": material_engagement_score,
        }
    
    def _extract_activity_features(self, data: BehaviorData) -> dict:
        """
        Extract activity pattern features.
        
        Calculates:
        - active_days_ratio: active_days / analysis_period
        - session_frequency: sessions / active_days
        - consistency_score: 1 - normalized_variance
        - late_night_ratio: late_night_sessions / total_sessions
        - weekend_ratio: weekend_sessions / total_sessions
        - peak_hour_norm: peak_hour / 24
        - engagement_trend_encoded: -1, 0, 1 encoded as 0, 0.5, 1
        
        Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7
        """
        activity = data.activity
        
        # Requirement 4.1: Calculate active_days_ratio
        # active_days / analysis_period_days
        # Default to 0 if analysis_period_days is 0 (shouldn't happen due to validation)
        active_days_ratio = self._safe_divide(
            activity.active_days,
            data.analysis_period_days,
            default=0.0
        )
        # Cap at 1.0 (active_days can't exceed analysis_period in valid data)
        active_days_ratio = min(1.0, active_days_ratio)
        
        # Requirement 4.2: Calculate session_frequency
        # total_sessions / active_days
        # Default to 0 if no active days
        raw_session_frequency = self._safe_divide(
            activity.total_sessions,
            activity.active_days,
            default=0.0
        )
        # Normalize: assume max ~10 sessions per day is high frequency
        session_frequency = self._normalize(raw_session_frequency, self.MAX_SESSIONS_PER_DAY)
        
        # Requirement 4.3: Calculate consistency_score
        # consistency_score = 1 - normalized_variance
        # Higher variance = less consistent, so we invert it
        # Normalize variance: assume max variance of 10.0 is very inconsistent
        MAX_VARIANCE = 10.0
        normalized_variance = self._normalize(activity.daily_activity_variance, MAX_VARIANCE)
        consistency_score = 1.0 - normalized_variance
        # Ensure it's within [0, 1] bounds
        consistency_score = max(0.0, min(1.0, consistency_score))
        
        # Requirement 4.5: Calculate late_night_ratio
        # late_night_sessions / total_sessions
        # Default to 0 if no sessions
        late_night_ratio = self._safe_divide(
            activity.late_night_sessions,
            activity.total_sessions,
            default=0.0
        )
        # Cap at 1.0 (late_night_sessions can't exceed total_sessions)
        late_night_ratio = min(1.0, late_night_ratio)
        
        # Requirement 4.6: Calculate weekend_ratio
        # weekend_sessions / total_sessions
        # Default to 0 if no sessions
        weekend_ratio = self._safe_divide(
            activity.weekend_sessions,
            activity.total_sessions,
            default=0.0
        )
        # Cap at 1.0 (weekend_sessions can't exceed total_sessions)
        weekend_ratio = min(1.0, weekend_ratio)
        
        # Requirement 4.4: Normalize peak_hour
        # peak_hour is 0-23, normalize to [0, 1]
        peak_hour_norm = activity.peak_hour / 23.0 if activity.peak_hour <= 23 else 1.0
        # Ensure it's within [0, 1] bounds
        peak_hour_norm = max(0.0, min(1.0, peak_hour_norm))
        
        # Requirement 4.7: Calculate engagement_trend_encoded
        # Determine engagement trend and encode as:
        # declining = 0, stable = 0.5, increasing = 1
        engagement_trend = self._calculate_engagement_trend(data)
        if engagement_trend == EngagementTrend.DECLINING:
            engagement_trend_encoded = 0.0
        elif engagement_trend == EngagementTrend.STABLE:
            engagement_trend_encoded = 0.5
        else:  # INCREASING
            engagement_trend_encoded = 1.0
        
        return {
            "active_days_ratio": active_days_ratio,
            "session_frequency": session_frequency,
            "consistency_score": consistency_score,
            "late_night_ratio": late_night_ratio,
            "weekend_ratio": weekend_ratio,
            "peak_hour_norm": peak_hour_norm,
            "engagement_trend_encoded": engagement_trend_encoded,
        }
    
    def _extract_quiz_features(self, data: BehaviorData) -> dict:
        """
        Extract quiz performance features.
        
        Uses default values when quiz data is missing:
        - Requirement 9.3: Use 0 for count-based features (quiz_attempt_frequency)
        - Requirement 9.4: Use 0.5 for ratio-based features (quiz_score_norm, quiz_completion_norm)
        
        Calculates:
        - quiz_score_norm: avg_score / 100 (normalized to [0, 1])
        - quiz_completion_norm: completion_rate (already in [0, 1])
        - quiz_attempt_frequency: normalized quiz attempts
        
        Args:
            data: Complete behavior data for a user
            
        Returns:
            Dictionary with quiz features, using defaults if quiz data is missing
        """
        # Check if quiz data is available
        if data.quiz is None:
            # Requirement 9.2, 9.3, 9.4: Impute missing values with defaults
            # - 0.5 for ratio-based features (quiz_score_norm, quiz_completion_norm)
            # - 0.0 for count-based features (quiz_attempt_frequency)
            logger.debug(f"No quiz data for user {data.user_id}, using default values")
            return {
                "quiz_score_norm": 0.5,  # Ratio-based default
                "quiz_completion_norm": 0.5,  # Ratio-based default
                "quiz_attempt_frequency": 0.0,  # Count-based default
            }
        
        quiz = data.quiz
        
        # Normalize quiz score: avg_score is 0-100, normalize to [0, 1]
        # If avg_score is 0 and no attempts, use default 0.5
        if quiz.quiz_attempts == 0:
            quiz_score_norm = 0.5  # Ratio-based default for no data
        else:
            quiz_score_norm = quiz.avg_score / 100.0
            # Ensure it's within [0, 1] bounds
            quiz_score_norm = max(0.0, min(1.0, quiz_score_norm))
        
        # Quiz completion rate is already in [0, 1] range
        # If no attempts, use default 0.5
        if quiz.quiz_attempts == 0:
            quiz_completion_norm = 0.5  # Ratio-based default for no data
        else:
            quiz_completion_norm = quiz.completion_rate
            # Ensure it's within [0, 1] bounds
            quiz_completion_norm = max(0.0, min(1.0, quiz_completion_norm))
        
        # Normalize quiz attempt frequency
        # Use 0 as default for count-based feature
        quiz_attempt_frequency = self._normalize(
            quiz.quiz_attempts,
            self.MAX_QUIZ_ATTEMPTS
        )
        
        return {
            "quiz_score_norm": quiz_score_norm,
            "quiz_completion_norm": quiz_completion_norm,
            "quiz_attempt_frequency": quiz_attempt_frequency,
        }
    
    def _calculate_engagement_trend(self, data: BehaviorData) -> EngagementTrend:
        """
        Determine if engagement is increasing, stable, or declining.
        
        Based on activity variance and session patterns over the analysis period.
        Uses heuristics based on:
        - Daily activity variance (high variance may indicate declining engagement)
        - Session frequency relative to active days
        - Late night activity patterns (may indicate stress/declining)
        
        Validates: Requirement 4.7
        """
        activity = data.activity
        
        # If no sessions at all, consider it stable (neutral)
        if activity.total_sessions == 0:
            return EngagementTrend.STABLE
        
        # Calculate session frequency (sessions per active day)
        session_frequency = self._safe_divide(
            activity.total_sessions,
            activity.active_days,
            default=0.0
        )
        
        # Calculate active days ratio
        active_days_ratio = self._safe_divide(
            activity.active_days,
            data.analysis_period_days,
            default=0.0
        )
        
        # Calculate late night ratio
        late_night_ratio = self._safe_divide(
            activity.late_night_sessions,
            activity.total_sessions,
            default=0.0
        )
        
        # Heuristics for engagement trend:
        # 
        # DECLINING indicators:
        # - High variance (inconsistent activity) > 5.0
        # - Low active days ratio (< 0.3) with some sessions
        # - High late night ratio (> 0.4) may indicate stress
        #
        # INCREASING indicators:
        # - Low variance (consistent activity) < 2.0
        # - High active days ratio (> 0.6)
        # - High session frequency (> 2 sessions per active day)
        #
        # STABLE: everything else
        
        declining_score = 0
        increasing_score = 0
        
        # Variance-based scoring
        if activity.daily_activity_variance > 5.0:
            declining_score += 1
        elif activity.daily_activity_variance < 2.0:
            increasing_score += 1
        
        # Active days ratio scoring
        if active_days_ratio < 0.3 and activity.total_sessions > 0:
            declining_score += 1
        elif active_days_ratio > 0.6:
            increasing_score += 1
        
        # Session frequency scoring
        if session_frequency > 2.0:
            increasing_score += 1
        elif session_frequency < 0.5 and activity.active_days > 0:
            declining_score += 1
        
        # Late night activity scoring (high late night may indicate stress)
        if late_night_ratio > 0.4:
            declining_score += 1
        
        # Determine trend based on scores
        if declining_score >= 2:
            return EngagementTrend.DECLINING
        elif increasing_score >= 2:
            return EngagementTrend.INCREASING
        else:
            return EngagementTrend.STABLE
    
    def _safe_divide(self, numerator: float, denominator: float, default: float = 0.0) -> float:
        """Safely divide two numbers, returning default if denominator is zero."""
        if denominator == 0:
            return default
        return numerator / denominator
    
    def _normalize(self, value: float, max_value: float) -> float:
        """Normalize a value to [0, 1] range using min-max scaling."""
        if max_value <= 0:
            return 0.0
        return min(1.0, max(0.0, value / max_value))
