"""
Learning Pulse Module

A predictive analytics system that diagnoses student learning profiles by analyzing
behavioral data from chat interactions, material consumption, and activity patterns.
The system classifies students into Learning Personas and provides actionable,
pedagogically-aware recommendations to improve learning outcomes.

This module demonstrates that NgasihTau understands "WHY" a student struggles,
not just "WHEN" they struggle.
"""

from .models import (
    LearningPersona,
    ChatBehavior,
    MaterialInteraction,
    ActivityPattern,
    QuizPerformance,
    BehaviorData,
    FeatureVector,
    ClassificationResult,
    GuardrailResult,
    Recommendation,
    PredictRequest,
    PredictResponse,
    HealthResponse,
)
from .feature_extractor import FeatureExtractor
from .persona_classifier import PersonaClassifier
from .guardrails import LogicGuardrails
from .recommendations import RecommendationEngine
from .router import router as learning_pulse_router

__all__ = [
    # Enums
    "LearningPersona",
    # Input Models
    "ChatBehavior",
    "MaterialInteraction",
    "ActivityPattern",
    "QuizPerformance",
    "BehaviorData",
    # Feature Models
    "FeatureVector",
    # Classification Models
    "ClassificationResult",
    "GuardrailResult",
    "Recommendation",
    # API Models
    "PredictRequest",
    "PredictResponse",
    "HealthResponse",
    # Components
    "FeatureExtractor",
    "PersonaClassifier",
    "LogicGuardrails",
    "RecommendationEngine",
    # Router
    "learning_pulse_router",
]
