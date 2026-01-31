"""
Synthetic Data Generator and Model Training for Learning Pulse Module

This module generates synthetic training data representing all ten Learning Personas
with realistic behavioral patterns, trains a Random Forest classifier, and saves
the trained model for production use.

Requirements:
- 8.1: Generate synthetic data representing all ten Learning Personas
- 8.2: Create at least 1000 samples per persona (10,000+ total)
- 8.3: Ensure synthetic data distributions match realistic learning behavior patterns
- 8.4: Split data into 80% training and 20% testing sets
"""

import numpy as np
from typing import Dict, List, Tuple, Optional
from dataclasses import dataclass
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
from sklearn.preprocessing import MinMaxScaler
from sklearn.metrics import classification_report, accuracy_score
import joblib
import os
from pathlib import Path

from .models import LearningPersona, FeatureVector


# Feature names in order (must match FeatureVector model)
FEATURE_NAMES = [
    # Chat features (8)
    "chat_message_ratio",
    "question_frequency",
    "avg_message_length_norm",
    "feedback_ratio",
    "feedback_engagement",
    "session_count_norm",
    "messages_per_session",
    "session_duration_norm",
    # Material features (7)
    "time_spent_norm",
    "view_count_norm",
    "material_diversity",
    "avg_time_per_view_norm",
    "bookmark_ratio",
    "scroll_depth",
    "material_engagement_score",
    # Activity features (7)
    "active_days_ratio",
    "session_frequency",
    "consistency_score",
    "late_night_ratio",
    "weekend_ratio",
    "peak_hour_norm",
    "engagement_trend_encoded",
    # Quiz features (3)
    "quiz_score_norm",
    "quiz_completion_norm",
    "quiz_attempt_frequency",
]

NUM_FEATURES = len(FEATURE_NAMES)  # 25 features


@dataclass
class FeatureDistribution:
    """
    Defines the distribution parameters for a single feature.
    
    Uses truncated normal distribution to ensure values stay within [0, 1].
    """
    mean: float
    std: float
    min_val: float = 0.0
    max_val: float = 1.0


def truncated_normal(mean: float, std: float, min_val: float, max_val: float, 
                     size: int, rng: np.random.Generator) -> np.ndarray:
    """
    Generate samples from a truncated normal distribution.
    
    Values are clipped to [min_val, max_val] range.
    """
    samples = rng.normal(mean, std, size)
    return np.clip(samples, min_val, max_val)


# ============================================================================
# Persona-Specific Feature Distributions
# ============================================================================
# Each persona has distinct behavioral patterns reflected in feature distributions.
# Features not explicitly defined use default distributions.

DEFAULT_DISTRIBUTION = FeatureDistribution(mean=0.5, std=0.15)

# Default distributions for all features
DEFAULT_FEATURE_DISTRIBUTIONS: Dict[str, FeatureDistribution] = {
    name: DEFAULT_DISTRIBUTION for name in FEATURE_NAMES
}


def get_skimmer_profile() -> Dict[str, FeatureDistribution]:
    """
    Skimmer: Quick browsing, low engagement, jumps between materials.
    
    Key indicators:
    - Low time per material
    - High material diversity (many different materials, shallow engagement)
    - Low scroll depth
    - High session frequency but short sessions
    """
    profile = DEFAULT_FEATURE_DISTRIBUTIONS.copy()
    profile.update({
        # Chat: Low engagement, few questions
        "chat_message_ratio": FeatureDistribution(0.4, 0.15),
        "question_frequency": FeatureDistribution(0.2, 0.1),
        "avg_message_length_norm": FeatureDistribution(0.3, 0.15),
        "feedback_ratio": FeatureDistribution(0.5, 0.2),
        "feedback_engagement": FeatureDistribution(0.2, 0.1),
        "session_count_norm": FeatureDistribution(0.6, 0.2),
        "messages_per_session": FeatureDistribution(0.3, 0.15),
        "session_duration_norm": FeatureDistribution(0.2, 0.1),
        # Material: High diversity, low depth
        "time_spent_norm": FeatureDistribution(0.2, 0.1),
        "view_count_norm": FeatureDistribution(0.7, 0.15),
        "material_diversity": FeatureDistribution(0.8, 0.1),
        "avg_time_per_view_norm": FeatureDistribution(0.15, 0.08),
        "bookmark_ratio": FeatureDistribution(0.1, 0.08),
        "scroll_depth": FeatureDistribution(0.25, 0.12),
        "material_engagement_score": FeatureDistribution(0.25, 0.1),
        # Activity: Frequent but inconsistent
        "active_days_ratio": FeatureDistribution(0.5, 0.2),
        "session_frequency": FeatureDistribution(0.7, 0.15),
        "consistency_score": FeatureDistribution(0.4, 0.15),
        "late_night_ratio": FeatureDistribution(0.2, 0.15),
        "weekend_ratio": FeatureDistribution(0.3, 0.15),
        # Quiz: Low engagement
        "quiz_score_norm": FeatureDistribution(0.45, 0.2),
        "quiz_completion_norm": FeatureDistribution(0.3, 0.15),
        "quiz_attempt_frequency": FeatureDistribution(0.2, 0.1),
    })
    return profile


def get_struggler_profile() -> Dict[str, FeatureDistribution]:
    """
    Struggler: High AI chat usage, repeated questions, low comprehension.
    
    Key indicators:
    - High question frequency
    - Low feedback ratio (more negative feedback)
    - Low quiz scores
    - High chat engagement but poor outcomes
    """
    profile = DEFAULT_FEATURE_DISTRIBUTIONS.copy()
    profile.update({
        # Chat: High engagement, many questions, poor feedback
        "chat_message_ratio": FeatureDistribution(0.6, 0.15),
        "question_frequency": FeatureDistribution(0.75, 0.12),
        "avg_message_length_norm": FeatureDistribution(0.5, 0.2),
        "feedback_ratio": FeatureDistribution(0.3, 0.15),
        "feedback_engagement": FeatureDistribution(0.6, 0.15),
        "session_count_norm": FeatureDistribution(0.6, 0.2),
        "messages_per_session": FeatureDistribution(0.7, 0.15),
        "session_duration_norm": FeatureDistribution(0.6, 0.2),
        # Material: Moderate engagement, struggling to understand
        "time_spent_norm": FeatureDistribution(0.5, 0.2),
        "view_count_norm": FeatureDistribution(0.5, 0.2),
        "material_diversity": FeatureDistribution(0.4, 0.15),
        "avg_time_per_view_norm": FeatureDistribution(0.5, 0.2),
        "bookmark_ratio": FeatureDistribution(0.3, 0.15),
        "scroll_depth": FeatureDistribution(0.5, 0.2),
        "material_engagement_score": FeatureDistribution(0.4, 0.15),
        # Activity: Regular but struggling
        "active_days_ratio": FeatureDistribution(0.5, 0.2),
        "session_frequency": FeatureDistribution(0.5, 0.2),
        "consistency_score": FeatureDistribution(0.5, 0.2),
        # Quiz: Low scores
        "quiz_score_norm": FeatureDistribution(0.3, 0.12),
        "quiz_completion_norm": FeatureDistribution(0.5, 0.2),
        "quiz_attempt_frequency": FeatureDistribution(0.6, 0.2),
    })
    return profile


def get_anxious_profile() -> Dict[str, FeatureDistribution]:
    """
    Anxious: Erratic patterns, late-night activity, high question frequency.
    
    Key indicators:
    - High late-night ratio
    - High variance in activity
    - Excessive questions
    - Inconsistent patterns
    """
    profile = DEFAULT_FEATURE_DISTRIBUTIONS.copy()
    profile.update({
        # Chat: Excessive questions, seeking reassurance
        "chat_message_ratio": FeatureDistribution(0.65, 0.15),
        "question_frequency": FeatureDistribution(0.8, 0.1),
        "avg_message_length_norm": FeatureDistribution(0.6, 0.2),
        "feedback_ratio": FeatureDistribution(0.4, 0.2),
        "feedback_engagement": FeatureDistribution(0.5, 0.2),
        "session_count_norm": FeatureDistribution(0.7, 0.15),
        "messages_per_session": FeatureDistribution(0.6, 0.2),
        "session_duration_norm": FeatureDistribution(0.5, 0.2),
        # Material: Moderate but anxious engagement
        "time_spent_norm": FeatureDistribution(0.5, 0.25),
        "view_count_norm": FeatureDistribution(0.6, 0.2),
        "material_diversity": FeatureDistribution(0.5, 0.2),
        "avg_time_per_view_norm": FeatureDistribution(0.4, 0.2),
        "bookmark_ratio": FeatureDistribution(0.5, 0.2),
        "scroll_depth": FeatureDistribution(0.5, 0.25),
        "material_engagement_score": FeatureDistribution(0.45, 0.2),
        # Activity: High late-night, inconsistent
        "active_days_ratio": FeatureDistribution(0.6, 0.2),
        "session_frequency": FeatureDistribution(0.7, 0.15),
        "consistency_score": FeatureDistribution(0.3, 0.15),
        "late_night_ratio": FeatureDistribution(0.6, 0.15),
        "weekend_ratio": FeatureDistribution(0.5, 0.2),
        # Quiz: Moderate, anxiety affects performance
        "quiz_score_norm": FeatureDistribution(0.5, 0.2),
        "quiz_completion_norm": FeatureDistribution(0.6, 0.2),
        "quiz_attempt_frequency": FeatureDistribution(0.7, 0.15),
    })
    return profile


def get_burnout_profile() -> Dict[str, FeatureDistribution]:
    """
    Burnout: Declining engagement over time, fatigue indicators.
    
    Key indicators:
    - Declining engagement trend
    - Was previously engaged (moderate baseline)
    - Reduced session frequency
    - Lower recent activity
    """
    profile = DEFAULT_FEATURE_DISTRIBUTIONS.copy()
    profile.update({
        # Chat: Declining engagement
        "chat_message_ratio": FeatureDistribution(0.4, 0.15),
        "question_frequency": FeatureDistribution(0.3, 0.15),
        "avg_message_length_norm": FeatureDistribution(0.35, 0.15),
        "feedback_ratio": FeatureDistribution(0.4, 0.2),
        "feedback_engagement": FeatureDistribution(0.25, 0.15),
        "session_count_norm": FeatureDistribution(0.35, 0.15),
        "messages_per_session": FeatureDistribution(0.3, 0.15),
        "session_duration_norm": FeatureDistribution(0.3, 0.15),
        # Material: Reduced engagement
        "time_spent_norm": FeatureDistribution(0.3, 0.15),
        "view_count_norm": FeatureDistribution(0.35, 0.15),
        "material_diversity": FeatureDistribution(0.4, 0.2),
        "avg_time_per_view_norm": FeatureDistribution(0.35, 0.15),
        "bookmark_ratio": FeatureDistribution(0.2, 0.1),
        "scroll_depth": FeatureDistribution(0.4, 0.2),
        "material_engagement_score": FeatureDistribution(0.3, 0.15),
        # Activity: Declining trend is key
        "active_days_ratio": FeatureDistribution(0.35, 0.15),
        "session_frequency": FeatureDistribution(0.3, 0.15),
        "consistency_score": FeatureDistribution(0.4, 0.2),
        "late_night_ratio": FeatureDistribution(0.3, 0.2),
        "engagement_trend_encoded": FeatureDistribution(0.15, 0.1),  # Declining = 0
        # Quiz: Declining performance
        "quiz_score_norm": FeatureDistribution(0.45, 0.2),
        "quiz_completion_norm": FeatureDistribution(0.35, 0.15),
        "quiz_attempt_frequency": FeatureDistribution(0.25, 0.15),
    })
    return profile


def get_master_profile() -> Dict[str, FeatureDistribution]:
    """
    Master: High comprehension, good quiz scores, efficient learning.
    
    Key indicators:
    - High quiz scores
    - Good feedback ratio
    - Efficient time usage
    - Consistent engagement
    """
    profile = DEFAULT_FEATURE_DISTRIBUTIONS.copy()
    profile.update({
        # Chat: Efficient, good feedback
        "chat_message_ratio": FeatureDistribution(0.55, 0.15),
        "question_frequency": FeatureDistribution(0.4, 0.15),
        "avg_message_length_norm": FeatureDistribution(0.6, 0.15),
        "feedback_ratio": FeatureDistribution(0.8, 0.1),
        "feedback_engagement": FeatureDistribution(0.6, 0.15),
        "session_count_norm": FeatureDistribution(0.6, 0.15),
        "messages_per_session": FeatureDistribution(0.5, 0.15),
        "session_duration_norm": FeatureDistribution(0.55, 0.15),
        # Material: Efficient engagement
        "time_spent_norm": FeatureDistribution(0.6, 0.15),
        "view_count_norm": FeatureDistribution(0.55, 0.15),
        "material_diversity": FeatureDistribution(0.6, 0.15),
        "avg_time_per_view_norm": FeatureDistribution(0.55, 0.15),
        "bookmark_ratio": FeatureDistribution(0.5, 0.15),
        "scroll_depth": FeatureDistribution(0.75, 0.12),
        "material_engagement_score": FeatureDistribution(0.7, 0.12),
        # Activity: Consistent, healthy patterns
        "active_days_ratio": FeatureDistribution(0.7, 0.15),
        "session_frequency": FeatureDistribution(0.6, 0.15),
        "consistency_score": FeatureDistribution(0.75, 0.12),
        "late_night_ratio": FeatureDistribution(0.15, 0.1),
        "engagement_trend_encoded": FeatureDistribution(0.75, 0.15),  # Stable/Increasing
        # Quiz: High scores
        "quiz_score_norm": FeatureDistribution(0.8, 0.1),
        "quiz_completion_norm": FeatureDistribution(0.85, 0.1),
        "quiz_attempt_frequency": FeatureDistribution(0.6, 0.15),
    })
    return profile


def get_procrastinator_profile() -> Dict[str, FeatureDistribution]:
    """
    Procrastinator: Last-minute cramming, irregular patterns.
    
    Key indicators:
    - Low consistency
    - Burst activity patterns
    - Deadline-driven behavior
    - Irregular session frequency
    """
    profile = DEFAULT_FEATURE_DISTRIBUTIONS.copy()
    profile.update({
        # Chat: Burst patterns
        "chat_message_ratio": FeatureDistribution(0.5, 0.2),
        "question_frequency": FeatureDistribution(0.5, 0.2),
        "avg_message_length_norm": FeatureDistribution(0.45, 0.2),
        "feedback_ratio": FeatureDistribution(0.5, 0.2),
        "feedback_engagement": FeatureDistribution(0.4, 0.2),
        "session_count_norm": FeatureDistribution(0.4, 0.2),
        "messages_per_session": FeatureDistribution(0.6, 0.2),
        "session_duration_norm": FeatureDistribution(0.5, 0.25),
        # Material: Cramming behavior
        "time_spent_norm": FeatureDistribution(0.45, 0.25),
        "view_count_norm": FeatureDistribution(0.5, 0.25),
        "material_diversity": FeatureDistribution(0.5, 0.2),
        "avg_time_per_view_norm": FeatureDistribution(0.4, 0.2),
        "bookmark_ratio": FeatureDistribution(0.25, 0.15),
        "scroll_depth": FeatureDistribution(0.45, 0.2),
        "material_engagement_score": FeatureDistribution(0.4, 0.2),
        # Activity: Very inconsistent, burst patterns
        "active_days_ratio": FeatureDistribution(0.3, 0.15),
        "session_frequency": FeatureDistribution(0.4, 0.25),
        "consistency_score": FeatureDistribution(0.2, 0.1),
        "late_night_ratio": FeatureDistribution(0.45, 0.2),
        "weekend_ratio": FeatureDistribution(0.5, 0.2),
        # Quiz: Variable performance
        "quiz_score_norm": FeatureDistribution(0.5, 0.2),
        "quiz_completion_norm": FeatureDistribution(0.45, 0.2),
        "quiz_attempt_frequency": FeatureDistribution(0.35, 0.2),
    })
    return profile


def get_deep_diver_profile() -> Dict[str, FeatureDistribution]:
    """
    Deep Diver: Long sessions, thorough material consumption.
    
    Key indicators:
    - High time per material
    - High scroll depth
    - Many bookmarks
    - Focused on fewer materials but deeply
    """
    profile = DEFAULT_FEATURE_DISTRIBUTIONS.copy()
    profile.update({
        # Chat: Thoughtful, detailed questions
        "chat_message_ratio": FeatureDistribution(0.55, 0.15),
        "question_frequency": FeatureDistribution(0.5, 0.15),
        "avg_message_length_norm": FeatureDistribution(0.7, 0.12),
        "feedback_ratio": FeatureDistribution(0.7, 0.15),
        "feedback_engagement": FeatureDistribution(0.55, 0.15),
        "session_count_norm": FeatureDistribution(0.5, 0.15),
        "messages_per_session": FeatureDistribution(0.55, 0.15),
        "session_duration_norm": FeatureDistribution(0.75, 0.12),
        # Material: Deep engagement
        "time_spent_norm": FeatureDistribution(0.8, 0.1),
        "view_count_norm": FeatureDistribution(0.5, 0.15),
        "material_diversity": FeatureDistribution(0.35, 0.15),
        "avg_time_per_view_norm": FeatureDistribution(0.8, 0.1),
        "bookmark_ratio": FeatureDistribution(0.7, 0.12),
        "scroll_depth": FeatureDistribution(0.85, 0.08),
        "material_engagement_score": FeatureDistribution(0.8, 0.1),
        # Activity: Focused, long sessions
        "active_days_ratio": FeatureDistribution(0.55, 0.15),
        "session_frequency": FeatureDistribution(0.45, 0.15),
        "consistency_score": FeatureDistribution(0.65, 0.15),
        "late_night_ratio": FeatureDistribution(0.25, 0.15),
        # Quiz: Good performance
        "quiz_score_norm": FeatureDistribution(0.7, 0.15),
        "quiz_completion_norm": FeatureDistribution(0.75, 0.12),
        "quiz_attempt_frequency": FeatureDistribution(0.5, 0.15),
    })
    return profile


def get_social_learner_profile() -> Dict[str, FeatureDistribution]:
    """
    Social Learner: High collaboration, peer interactions.
    
    Key indicators:
    - High session count
    - Collaborative pod activity
    - Regular engagement
    - Good feedback engagement
    """
    profile = DEFAULT_FEATURE_DISTRIBUTIONS.copy()
    profile.update({
        # Chat: High engagement, collaborative
        "chat_message_ratio": FeatureDistribution(0.6, 0.15),
        "question_frequency": FeatureDistribution(0.55, 0.15),
        "avg_message_length_norm": FeatureDistribution(0.55, 0.15),
        "feedback_ratio": FeatureDistribution(0.65, 0.15),
        "feedback_engagement": FeatureDistribution(0.7, 0.12),
        "session_count_norm": FeatureDistribution(0.8, 0.1),
        "messages_per_session": FeatureDistribution(0.6, 0.15),
        "session_duration_norm": FeatureDistribution(0.55, 0.15),
        # Material: Moderate, shared learning
        "time_spent_norm": FeatureDistribution(0.55, 0.15),
        "view_count_norm": FeatureDistribution(0.6, 0.15),
        "material_diversity": FeatureDistribution(0.6, 0.15),
        "avg_time_per_view_norm": FeatureDistribution(0.5, 0.15),
        "bookmark_ratio": FeatureDistribution(0.5, 0.15),
        "scroll_depth": FeatureDistribution(0.6, 0.15),
        "material_engagement_score": FeatureDistribution(0.6, 0.12),
        # Activity: Very active, social patterns
        "active_days_ratio": FeatureDistribution(0.7, 0.12),
        "session_frequency": FeatureDistribution(0.75, 0.12),
        "consistency_score": FeatureDistribution(0.65, 0.15),
        "late_night_ratio": FeatureDistribution(0.2, 0.12),
        "weekend_ratio": FeatureDistribution(0.4, 0.15),
        # Quiz: Good collaborative performance
        "quiz_score_norm": FeatureDistribution(0.65, 0.15),
        "quiz_completion_norm": FeatureDistribution(0.7, 0.12),
        "quiz_attempt_frequency": FeatureDistribution(0.6, 0.15),
    })
    return profile


def get_perfectionist_profile() -> Dict[str, FeatureDistribution]:
    """
    Perfectionist: Excessive review, high self-correction.
    
    Key indicators:
    - Multiple views of same material
    - High bookmark ratio
    - Low material diversity (revisiting same content)
    - High time spent
    """
    profile = DEFAULT_FEATURE_DISTRIBUTIONS.copy()
    profile.update({
        # Chat: Detailed, seeking confirmation
        "chat_message_ratio": FeatureDistribution(0.55, 0.15),
        "question_frequency": FeatureDistribution(0.6, 0.15),
        "avg_message_length_norm": FeatureDistribution(0.65, 0.15),
        "feedback_ratio": FeatureDistribution(0.6, 0.15),
        "feedback_engagement": FeatureDistribution(0.6, 0.15),
        "session_count_norm": FeatureDistribution(0.65, 0.15),
        "messages_per_session": FeatureDistribution(0.55, 0.15),
        "session_duration_norm": FeatureDistribution(0.65, 0.15),
        # Material: High revisits, low diversity
        "time_spent_norm": FeatureDistribution(0.75, 0.12),
        "view_count_norm": FeatureDistribution(0.75, 0.12),
        "material_diversity": FeatureDistribution(0.25, 0.12),
        "avg_time_per_view_norm": FeatureDistribution(0.6, 0.15),
        "bookmark_ratio": FeatureDistribution(0.75, 0.1),
        "scroll_depth": FeatureDistribution(0.8, 0.1),
        "material_engagement_score": FeatureDistribution(0.7, 0.12),
        # Activity: Regular, methodical
        "active_days_ratio": FeatureDistribution(0.65, 0.15),
        "session_frequency": FeatureDistribution(0.6, 0.15),
        "consistency_score": FeatureDistribution(0.7, 0.12),
        "late_night_ratio": FeatureDistribution(0.3, 0.15),
        # Quiz: Good but anxious about scores
        "quiz_score_norm": FeatureDistribution(0.7, 0.12),
        "quiz_completion_norm": FeatureDistribution(0.8, 0.1),
        "quiz_attempt_frequency": FeatureDistribution(0.7, 0.12),
    })
    return profile


def get_lost_profile() -> Dict[str, FeatureDistribution]:
    """
    Lost: Random navigation, no clear learning path.
    
    Key indicators:
    - Low material diversity (stuck on few materials)
    - Low engagement overall
    - No clear pattern
    - Low quiz engagement
    """
    profile = DEFAULT_FEATURE_DISTRIBUTIONS.copy()
    profile.update({
        # Chat: Low engagement
        "chat_message_ratio": FeatureDistribution(0.4, 0.2),
        "question_frequency": FeatureDistribution(0.25, 0.15),
        "avg_message_length_norm": FeatureDistribution(0.3, 0.15),
        "feedback_ratio": FeatureDistribution(0.5, 0.25),
        "feedback_engagement": FeatureDistribution(0.15, 0.1),
        "session_count_norm": FeatureDistribution(0.25, 0.15),
        "messages_per_session": FeatureDistribution(0.25, 0.15),
        "session_duration_norm": FeatureDistribution(0.2, 0.12),
        # Material: Low engagement, stuck
        "time_spent_norm": FeatureDistribution(0.2, 0.12),
        "view_count_norm": FeatureDistribution(0.2, 0.12),
        "material_diversity": FeatureDistribution(0.2, 0.12),
        "avg_time_per_view_norm": FeatureDistribution(0.25, 0.15),
        "bookmark_ratio": FeatureDistribution(0.1, 0.08),
        "scroll_depth": FeatureDistribution(0.25, 0.15),
        "material_engagement_score": FeatureDistribution(0.15, 0.1),
        # Activity: Low, no pattern
        "active_days_ratio": FeatureDistribution(0.2, 0.12),
        "session_frequency": FeatureDistribution(0.2, 0.12),
        "consistency_score": FeatureDistribution(0.3, 0.2),
        "late_night_ratio": FeatureDistribution(0.3, 0.2),
        # Quiz: Very low engagement
        "quiz_score_norm": FeatureDistribution(0.3, 0.2),
        "quiz_completion_norm": FeatureDistribution(0.2, 0.12),
        "quiz_attempt_frequency": FeatureDistribution(0.1, 0.08),
    })
    return profile


# ============================================================================
# Persona Profile Registry
# ============================================================================

PERSONA_PROFILES: Dict[LearningPersona, callable] = {
    LearningPersona.SKIMMER: get_skimmer_profile,
    LearningPersona.STRUGGLER: get_struggler_profile,
    LearningPersona.ANXIOUS: get_anxious_profile,
    LearningPersona.BURNOUT: get_burnout_profile,
    LearningPersona.MASTER: get_master_profile,
    LearningPersona.PROCRASTINATOR: get_procrastinator_profile,
    LearningPersona.DEEP_DIVER: get_deep_diver_profile,
    LearningPersona.SOCIAL_LEARNER: get_social_learner_profile,
    LearningPersona.PERFECTIONIST: get_perfectionist_profile,
    LearningPersona.LOST: get_lost_profile,
}


# ============================================================================
# Synthetic Data Generation
# ============================================================================

class SyntheticDataGenerator:
    """
    Generates synthetic training data for Learning Persona classification.
    
    Each persona has distinct feature distributions that reflect realistic
    learning behavior patterns.
    
    Requirements:
    - 8.1: Generate synthetic data representing all ten Learning Personas
    - 8.2: Create at least 1000 samples per persona (10,000+ total)
    - 8.3: Ensure synthetic data distributions match realistic learning behavior patterns
    - 8.4: Split data into 80% training and 20% testing sets
    """
    
    def __init__(self, seed: Optional[int] = 42):
        """
        Initialize the generator with a random seed for reproducibility.
        
        Args:
            seed: Random seed for reproducibility. Default is 42.
        """
        self.rng = np.random.default_rng(seed)
        self.personas = list(LearningPersona)
    
    def generate_samples_for_persona(
        self, 
        persona: LearningPersona, 
        n_samples: int
    ) -> np.ndarray:
        """
        Generate synthetic feature vectors for a specific persona.
        
        Args:
            persona: The Learning Persona to generate samples for
            n_samples: Number of samples to generate
            
        Returns:
            numpy array of shape (n_samples, NUM_FEATURES)
        """
        profile_fn = PERSONA_PROFILES.get(persona)
        if profile_fn is None:
            raise ValueError(f"Unknown persona: {persona}")
        
        profile = profile_fn()
        samples = np.zeros((n_samples, NUM_FEATURES))
        
        for i, feature_name in enumerate(FEATURE_NAMES):
            dist = profile.get(feature_name, DEFAULT_DISTRIBUTION)
            samples[:, i] = truncated_normal(
                mean=dist.mean,
                std=dist.std,
                min_val=dist.min_val,
                max_val=dist.max_val,
                size=n_samples,
                rng=self.rng
            )
        
        return samples

    
    def generate_dataset(
        self, 
        samples_per_persona: int = 1000
    ) -> Tuple[np.ndarray, np.ndarray, List[str]]:
        """
        Generate complete synthetic dataset for all personas.
        
        Args:
            samples_per_persona: Number of samples per persona (default: 1000)
            
        Returns:
            Tuple of (features, labels, persona_names)
            - features: numpy array of shape (n_samples * 10, NUM_FEATURES)
            - labels: numpy array of integer labels
            - persona_names: list of persona names for label mapping
        """
        all_features = []
        all_labels = []
        persona_names = [p.value for p in self.personas]
        
        for persona_idx, persona in enumerate(self.personas):
            samples = self.generate_samples_for_persona(persona, samples_per_persona)
            labels = np.full(samples_per_persona, persona_idx)
            
            all_features.append(samples)
            all_labels.append(labels)
        
        features = np.vstack(all_features)
        labels = np.concatenate(all_labels)
        
        return features, labels, persona_names
    
    def generate_train_test_split(
        self,
        samples_per_persona: int = 1000,
        test_size: float = 0.2,
        stratify: bool = True
    ) -> Tuple[np.ndarray, np.ndarray, np.ndarray, np.ndarray, List[str]]:
        """
        Generate dataset and split into training and testing sets.
        
        Implements Requirement 8.4: Split data into 80% training and 20% testing sets.
        
        Args:
            samples_per_persona: Number of samples per persona (default: 1000)
            test_size: Fraction of data for testing (default: 0.2 = 20%)
            stratify: Whether to stratify split by persona (default: True)
            
        Returns:
            Tuple of (X_train, X_test, y_train, y_test, persona_names)
        """
        features, labels, persona_names = self.generate_dataset(samples_per_persona)
        
        stratify_labels = labels if stratify else None
        
        X_train, X_test, y_train, y_test = train_test_split(
            features,
            labels,
            test_size=test_size,
            random_state=self.rng.integers(0, 2**31),
            stratify=stratify_labels
        )
        
        return X_train, X_test, y_train, y_test, persona_names


    def get_dataset_statistics(
        self, 
        features: np.ndarray, 
        labels: np.ndarray,
        persona_names: List[str]
    ) -> Dict:
        """
        Calculate statistics about the generated dataset.
        
        Args:
            features: Feature array
            labels: Label array
            persona_names: List of persona names
            
        Returns:
            Dictionary with dataset statistics
        """
        unique, counts = np.unique(labels, return_counts=True)
        
        stats = {
            "total_samples": len(labels),
            "num_features": features.shape[1],
            "num_personas": len(persona_names),
            "samples_per_persona": dict(zip(
                [persona_names[i] for i in unique],
                counts.tolist()
            )),
            "feature_stats": {
                name: {
                    "mean": float(features[:, i].mean()),
                    "std": float(features[:, i].std()),
                    "min": float(features[:, i].min()),
                    "max": float(features[:, i].max()),
                }
                for i, name in enumerate(FEATURE_NAMES)
            }
        }
        
        return stats


# ============================================================================
# Utility Functions
# ============================================================================

def feature_vector_to_array(fv: FeatureVector) -> np.ndarray:
    """
    Convert a FeatureVector Pydantic model to a numpy array.
    
    Args:
        fv: FeatureVector instance
        
    Returns:
        numpy array of shape (NUM_FEATURES,)
    """
    return np.array([getattr(fv, name) for name in FEATURE_NAMES])


def array_to_feature_vector(arr: np.ndarray) -> FeatureVector:
    """
    Convert a numpy array to a FeatureVector Pydantic model.
    
    Args:
        arr: numpy array of shape (NUM_FEATURES,)
        
    Returns:
        FeatureVector instance
    """
    if len(arr) != NUM_FEATURES:
        raise ValueError(f"Expected {NUM_FEATURES} features, got {len(arr)}")
    
    return FeatureVector(**dict(zip(FEATURE_NAMES, arr.tolist())))


def get_model_dir() -> Path:
    """Get the directory for storing trained models."""
    model_dir = Path(__file__).parent / "models"
    model_dir.mkdir(exist_ok=True)
    return model_dir


def get_default_model_path() -> Path:
    """Get the default path for the trained model."""
    return get_model_dir() / "persona_classifier.joblib"


def get_default_scaler_path() -> Path:
    """Get the default path for the feature scaler."""
    return get_model_dir() / "feature_scaler.joblib"


# ============================================================================
# Main Entry Point for Training
# ============================================================================

def train_model(
    samples_per_persona: int = 1000,
    n_estimators: int = 100,
    test_size: float = 0.2,
    seed: int = 42,
    save_model: bool = True,
    model_path: Optional[Path] = None,
    scaler_path: Optional[Path] = None,
    verbose: bool = True
) -> Dict:
    """
    Train the Learning Persona classifier with synthetic data.
    
    This function:
    1. Generates synthetic data for all 10 personas
    2. Splits data into 80% training and 20% testing
    3. Trains a Random Forest classifier
    4. Evaluates and reports accuracy metrics
    5. Saves the trained model and scaler
    
    Args:
        samples_per_persona: Number of samples per persona (default: 1000)
        n_estimators: Number of trees in Random Forest (default: 100)
        test_size: Fraction for test set (default: 0.2)
        seed: Random seed for reproducibility (default: 42)
        save_model: Whether to save the trained model (default: True)
        model_path: Path to save model (default: models/persona_classifier.joblib)
        scaler_path: Path to save scaler (default: models/feature_scaler.joblib)
        verbose: Whether to print progress (default: True)
        
    Returns:
        Dictionary with training results including accuracy and classification report
    """
    if model_path is None:
        model_path = get_default_model_path()
    if scaler_path is None:
        scaler_path = get_default_scaler_path()
    
    if verbose:
        print(f"Generating synthetic data ({samples_per_persona} samples per persona)...")
    
    # Generate data
    generator = SyntheticDataGenerator(seed=seed)
    X_train, X_test, y_train, y_test, persona_names = generator.generate_train_test_split(
        samples_per_persona=samples_per_persona,
        test_size=test_size
    )
    
    if verbose:
        print(f"  Training samples: {len(X_train)}")
        print(f"  Testing samples: {len(X_test)}")
        print(f"  Features: {X_train.shape[1]}")

    
    # Fit scaler on training data
    if verbose:
        print("Fitting feature scaler...")
    
    scaler = MinMaxScaler()
    X_train_scaled = scaler.fit_transform(X_train)
    X_test_scaled = scaler.transform(X_test)
    
    # Train Random Forest classifier
    if verbose:
        print(f"Training Random Forest classifier ({n_estimators} estimators)...")
    
    classifier = RandomForestClassifier(
        n_estimators=n_estimators,
        random_state=seed,
        n_jobs=-1,  # Use all CPU cores
        class_weight='balanced'  # Handle any class imbalance
    )
    classifier.fit(X_train_scaled, y_train)
    
    # Evaluate on test set
    if verbose:
        print("Evaluating model...")
    
    y_pred = classifier.predict(X_test_scaled)
    accuracy = accuracy_score(y_test, y_pred)
    report = classification_report(
        y_test, 
        y_pred, 
        target_names=persona_names,
        output_dict=True
    )
    report_text = classification_report(
        y_test, 
        y_pred, 
        target_names=persona_names
    )
    
    if verbose:
        print(f"\nTest Accuracy: {accuracy:.4f} ({accuracy * 100:.2f}%)")
        print("\nClassification Report:")
        print(report_text)
    
    # Save model and scaler
    from datetime import datetime
    import json
    
    training_date = datetime.now().isoformat()
    metadata_path = model_path.parent / "model_metadata.json"
    
    if save_model:
        if verbose:
            print(f"\nSaving model to {model_path}...")
        joblib.dump(classifier, model_path)
        
        if verbose:
            print(f"Saving scaler to {scaler_path}...")
        joblib.dump(scaler, scaler_path)
        
        # Save metadata to JSON file
        metadata = {
            "model_version": "1.0.0",
            "training_date": training_date,
            "accuracy": accuracy,
            "n_estimators": n_estimators,
            "n_features": NUM_FEATURES,
            "feature_names": FEATURE_NAMES,
            "persona_names": persona_names,
            "training_samples": len(X_train),
            "testing_samples": len(X_test),
            "samples_per_persona": samples_per_persona,
            "test_size": test_size,
            "seed": seed,
            "model_path": str(model_path),
            "scaler_path": str(scaler_path),
            "classification_report": report,
        }
        
        if verbose:
            print(f"Saving metadata to {metadata_path}...")
        with open(metadata_path, 'w') as f:
            json.dump(metadata, f, indent=2)
    
    # Prepare results
    results = {
        "accuracy": accuracy,
        "classification_report": report,
        "classification_report_text": report_text,
        "training_samples": len(X_train),
        "testing_samples": len(X_test),
        "samples_per_persona": samples_per_persona,
        "n_estimators": n_estimators,
        "n_features": NUM_FEATURES,
        "feature_names": FEATURE_NAMES,
        "persona_names": persona_names,
        "model_path": str(model_path) if save_model else None,
        "scaler_path": str(scaler_path) if save_model else None,
        "metadata_path": str(metadata_path) if save_model else None,
        "training_date": training_date,
        "seed": seed,
    }
    
    return results


# ============================================================================
# CLI Entry Point
# ============================================================================

if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(
        description="Train Learning Persona classifier with synthetic data"
    )
    parser.add_argument(
        "--samples", 
        type=int, 
        default=1000,
        help="Number of samples per persona (default: 1000)"
    )
    parser.add_argument(
        "--estimators", 
        type=int, 
        default=100,
        help="Number of Random Forest estimators (default: 100)"
    )
    parser.add_argument(
        "--test-size", 
        type=float, 
        default=0.2,
        help="Test set fraction (default: 0.2)"
    )
    parser.add_argument(
        "--seed", 
        type=int, 
        default=42,
        help="Random seed (default: 42)"
    )
    parser.add_argument(
        "--no-save", 
        action="store_true",
        help="Don't save the trained model"
    )
    parser.add_argument(
        "--quiet", 
        action="store_true",
        help="Suppress output"
    )
    
    args = parser.parse_args()
    
    results = train_model(
        samples_per_persona=args.samples,
        n_estimators=args.estimators,
        test_size=args.test_size,
        seed=args.seed,
        save_model=not args.no_save,
        verbose=not args.quiet
    )
    
    if not args.quiet:
        print(f"\n{'='*60}")
        print("Training Complete!")
        print(f"{'='*60}")
        print(f"Final Accuracy: {results['accuracy']:.4f} ({results['accuracy'] * 100:.2f}%)")
        if results['model_path']:
            print(f"Model saved to: {results['model_path']}")
            print(f"Scaler saved to: {results['scaler_path']}")
            print(f"Metadata saved to: {results['metadata_path']}")
