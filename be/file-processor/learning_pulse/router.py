"""
FastAPI Router for Learning Pulse Module

Exposes REST API endpoints for persona prediction and health checks.
Integrates Feature Extractor, Persona Classifier, Logic Guardrails,
and Recommendation Engine into a cohesive prediction pipeline.
Handles Requirements 7.1-7.6, 10.1-10.4.
"""

import time
import logging
from pathlib import Path
from typing import Optional

from fastapi import APIRouter, HTTPException

from .models import (
    PredictRequest,
    PredictResponse,
    HealthResponse,
    FeatureSummary,
    LearningPersona,
    FeatureVector,
)
from .feature_extractor import FeatureExtractor
from .persona_classifier import PersonaClassifier
from .guardrails import LogicGuardrails
from .recommendations import RecommendationEngine

logger = logging.getLogger("learning_pulse.router")

# Create router with prefix and tags
router = APIRouter(prefix="/api/v1/learning-pulse", tags=["Learning Pulse"])

# Module components (initialized on startup)
_feature_extractor: Optional[FeatureExtractor] = None
_persona_classifier: Optional[PersonaClassifier] = None
_logic_guardrails: Optional[LogicGuardrails] = None
_recommendation_engine: Optional[RecommendationEngine] = None
_initialized: bool = False


def initialize_components(
    model_path: Optional[str] = None,
    scaler_path: Optional[str] = None,
    metadata_path: Optional[str] = None,
) -> bool:
    """
    Initialize Learning Pulse components.
    
    Called during application startup to load models and initialize components.
    Uses default paths if not provided.
    
    Returns:
        True if initialization successful, False otherwise
    """
    global _feature_extractor, _persona_classifier, _logic_guardrails, _recommendation_engine, _initialized
    
    # Default paths relative to this module
    module_dir = Path(__file__).parent
    models_dir = module_dir / "models"
    
    if model_path is None:
        model_path = str(models_dir / "persona_classifier.joblib")
    if scaler_path is None:
        scaler_path = str(models_dir / "feature_scaler.joblib")
    if metadata_path is None:
        metadata_path = str(models_dir / "model_metadata.json")
    
    try:
        _feature_extractor = FeatureExtractor(scaler_path=scaler_path)
        _persona_classifier = PersonaClassifier(
            model_path=model_path,
            scaler_path=scaler_path,
            metadata_path=metadata_path,
        )
        _logic_guardrails = LogicGuardrails()
        _recommendation_engine = RecommendationEngine()
        _initialized = True
        
        logger.info("Learning Pulse components initialized successfully")
        return True
    except Exception as e:
        logger.error(f"Failed to initialize Learning Pulse components: {e}")
        _initialized = False
        return False


def _generate_feature_summary(features: FeatureVector) -> FeatureSummary:
    """
    Generate human-readable feature summary for transparency.
    
    Categorizes engagement levels and identifies key indicators.
    Validates: Requirement 7.3, Task 12.1
    """
    # Categorize chat engagement based on multiple chat features
    chat_score = (
        features.chat_message_ratio * 0.3 +
        features.messages_per_session * 0.3 +
        features.session_count_norm * 0.2 +
        features.feedback_engagement * 0.2
    )
    if chat_score < 0.33:
        chat_engagement = "low"
    elif chat_score < 0.66:
        chat_engagement = "medium"
    else:
        chat_engagement = "high"
    
    # Categorize material consumption
    material_score = (
        features.time_spent_norm * 0.3 +
        features.view_count_norm * 0.2 +
        features.scroll_depth * 0.3 +
        features.material_engagement_score * 0.2
    )
    if material_score < 0.33:
        material_consumption = "low"
    elif material_score < 0.66:
        material_consumption = "medium"
    else:
        material_consumption = "high"
    
    # Categorize activity consistency
    if features.consistency_score < 0.33:
        activity_consistency = "low"
    elif features.consistency_score < 0.66:
        activity_consistency = "medium"
    else:
        activity_consistency = "high"
    
    # Identify key indicators
    key_indicators = []
    
    if features.question_frequency > 0.7:
        key_indicators.append("High question frequency in chat")
    if features.late_night_ratio > 0.5:
        key_indicators.append("Significant late-night activity")
    if features.scroll_depth > 0.8:
        key_indicators.append("Thorough material reading")
    if features.scroll_depth < 0.3:
        key_indicators.append("Quick material scanning")
    if features.engagement_trend_encoded < 0.25:
        key_indicators.append("Declining engagement trend")
    elif features.engagement_trend_encoded > 0.75:
        key_indicators.append("Increasing engagement trend")
    if features.bookmark_ratio > 0.5:
        key_indicators.append("Active bookmarking behavior")
    if features.quiz_score_norm > 0.8:
        key_indicators.append("Strong quiz performance")
    elif features.quiz_score_norm < 0.4 and features.quiz_attempt_frequency > 0.1:
        key_indicators.append("Struggling with quizzes")
    if features.active_days_ratio > 0.7:
        key_indicators.append("Consistent daily activity")
    elif features.active_days_ratio < 0.2:
        key_indicators.append("Sporadic activity pattern")
    
    return FeatureSummary(
        chat_engagement=chat_engagement,
        material_consumption=material_consumption,
        activity_consistency=activity_consistency,
        key_indicators=key_indicators,
    )


@router.post("/predict-persona", response_model=PredictResponse)
async def predict_persona(request: PredictRequest) -> PredictResponse:
    """
    Predict learning persona for a user.
    
    Accepts behavior data and returns:
    - Classified persona with confidence score
    - Actionable recommendations
    - Feature summary for transparency
    - Any guardrail overrides or flags
    
    Requirements: 7.1, 7.2, 7.3, 7.6
    """
    start_time = time.perf_counter()
    
    # Check if components are initialized
    if not _initialized or not _persona_classifier or not _persona_classifier.is_loaded():
        raise HTTPException(
            status_code=503,
            detail="Model not available. Learning Pulse module not initialized.",
        )
    
    try:
        # Step 1: Extract features from behavior_data
        features = _feature_extractor.extract(request.behavior_data)
        
        # Step 2: Classify persona using ML model
        classification = _persona_classifier.predict(features)
        
        # Step 3: Apply logic guardrails
        # Calculate additional metrics needed for guardrails
        material = request.behavior_data.material
        avg_time_per_material = (
            material.total_time_spent_seconds / material.total_views
            if material.total_views > 0 else 0.0
        )
        
        # Parse previous persona if provided
        previous_persona = None
        if request.previous_persona:
            try:
                previous_persona = LearningPersona(request.previous_persona)
            except ValueError:
                logger.warning(f"Invalid previous_persona: {request.previous_persona}")
        
        guardrail_result = _logic_guardrails.apply(
            prediction=classification,
            features=features,
            quiz_score=request.quiz_score,
            previous_persona=previous_persona,
            total_materials_viewed=material.unique_materials_viewed,
            avg_time_per_material=avg_time_per_material,
        )
        
        # Step 4: Get recommendations for final persona
        persona_recommendations = _recommendation_engine.get_recommendations(
            guardrail_result.persona
        )
        
        # Step 5: Generate feature summary
        feature_summary = _generate_feature_summary(features)
        
        # Step 6: Calculate processing time
        processing_time_ms = (time.perf_counter() - start_time) * 1000
        
        # Build response
        return PredictResponse(
            user_id=request.user_id,
            persona=guardrail_result.persona.value,
            confidence=guardrail_result.confidence,
            is_low_confidence=classification.is_low_confidence,
            recommendations=persona_recommendations.recommendations,
            feature_summary=feature_summary,
            override_info=guardrail_result.override,
            flags=guardrail_result.flags.flag_reasons,
            processing_time_ms=processing_time_ms,
        )
        
    except Exception as e:
        logger.error(f"Prediction failed for user {request.user_id}: {e}")
        raise HTTPException(
            status_code=500,
            detail=f"Prediction failed: {str(e)}",
        )


@router.get("/health", response_model=HealthResponse)
async def health_check() -> HealthResponse:
    """
    Health check for Learning Pulse module.
    
    Returns model status, version, and last training date.
    Requirements: 10.1, 10.2, 10.3, 10.4
    """
    if _initialized and _persona_classifier and _persona_classifier.is_loaded():
        return HealthResponse(
            status="healthy",
            model_loaded=True,
            model_version=_persona_classifier.model_version,
            last_training_date=_persona_classifier.last_training_date,
        )
    else:
        return HealthResponse(
            status="unhealthy",
            model_loaded=False,
            model_version="unknown",
            last_training_date="unknown",
            error_details="Model not loaded or unavailable",
        )
