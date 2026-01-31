"""
Logic Guardrails for Learning Pulse Module

Rule-based override system ensuring pedagogically sound classifications.
Applies domain knowledge to correct ML predictions when they conflict
with educational best practices. Handles:
- Override rules (Requirements 5.1-5.3)
- Flag rules (Requirements 5.4-5.5)
- Override reason tracking (Requirement 5.6)
"""

from typing import Optional, List
import logging

from .models import (
    LearningPersona,
    FeatureVector,
    ClassificationResult,
    GuardrailResult,
    GuardrailOverride,
    GuardrailFlags,
)

logger = logging.getLogger("learning_pulse.guardrails")


class LogicGuardrails:
    """
    Pedagogical rules for classification override.
    
    Applies domain-specific rules to ensure ML predictions are
    educationally sound. Can override predictions or raise flags
    for attention without changing the classification.
    """
    
    # Override thresholds
    MASTER_MIN_QUIZ_SCORE = 50.0  # Requirement 5.1
    SKIMMER_MIN_TIME_PER_MATERIAL = 300  # seconds, Requirement 5.2
    DEEP_DIVER_MIN_MATERIALS = 3  # Requirement 5.3
    
    # Flag thresholds
    ANXIOUS_LATE_NIGHT_THRESHOLD = 0.5  # Requirement 5.4
    HIGH_SESSION_FREQUENCY_THRESHOLD = 3.0  # sessions per day
    
    def apply(
        self,
        prediction: ClassificationResult,
        features: FeatureVector,
        quiz_score: Optional[float] = None,
        previous_persona: Optional[LearningPersona] = None,
        total_materials_viewed: int = 0,
        avg_time_per_material: float = 0.0,
    ) -> GuardrailResult:
        """
        Apply pedagogical rules to classification.
        
        Args:
            prediction: Raw ML classification result
            features: Extracted feature vector
            quiz_score: Optional quiz score (0-100)
            previous_persona: Previous classification for trend detection
            total_materials_viewed: Total materials viewed (for Deep Diver check)
            avg_time_per_material: Average time per material in seconds
            
        Returns:
            GuardrailResult with potentially overridden persona and flags
        """
        override: Optional[GuardrailOverride] = None
        final_persona = prediction.persona
        was_overridden = False
        
        # Check override rules in priority order
        # Rule 1: Master → Struggler when quiz_score < 50% (Requirement 5.1)
        if override is None:
            override = self._check_master_override(prediction, quiz_score)
        
        # Rule 2: Deep_Diver → Lost when materials_viewed < 3 (Requirement 5.3)
        if override is None:
            override = self._check_deep_diver_override(prediction, total_materials_viewed)
        
        # Rule 3: Skimmer reconsideration when time_per_material > 300s (Requirement 5.2)
        if override is None:
            override = self._check_skimmer_reconsideration(prediction, avg_time_per_material)
        
        # Apply override if one was triggered
        if override is not None:
            final_persona = override.final_persona
            was_overridden = True
            logger.warning(
                f"Guardrail override: {override.original_persona.value} -> {override.final_persona.value} "
                f"(rule: {override.rule_triggered})"
            )
        
        # Check flag rules (these don't override, just raise flags)
        flag_reasons: List[str] = []
        
        # Check for Anxious indicators (Requirement 5.4)
        anxious_flags = self._check_anxious_flags(features)
        flag_reasons.extend(anxious_flags)
        
        # Check for Burnout indicators (Requirement 5.5)
        burnout_flags = self._check_burnout_flags(features, previous_persona)
        flag_reasons.extend(burnout_flags)
        
        # Build flags object
        flags = GuardrailFlags(
            potential_anxious=len(anxious_flags) > 0,
            potential_burnout=len(burnout_flags) > 0,
            needs_attention=len(flag_reasons) > 0 or was_overridden,
            flag_reasons=flag_reasons,
        )
        
        return GuardrailResult(
            persona=final_persona,
            confidence=prediction.confidence,
            was_overridden=was_overridden,
            override=override,
            flags=flags,
        )
    
    def _check_master_override(
        self,
        prediction: ClassificationResult,
        quiz_score: Optional[float],
    ) -> Optional[GuardrailOverride]:
        """
        Check if Master prediction should be overridden.
        
        Rule: IF persona is Master AND quiz_score < 50%, THEN override to Struggler
        (Requirement 5.1)
        """
        if prediction.persona != LearningPersona.MASTER:
            return None
        
        if quiz_score is not None and quiz_score < self.MASTER_MIN_QUIZ_SCORE:
            return GuardrailOverride(
                original_persona=LearningPersona.MASTER,
                final_persona=LearningPersona.STRUGGLER,
                rule_triggered="master_low_quiz_score",
                reason=f"Quiz score ({quiz_score:.1f}%) is below {self.MASTER_MIN_QUIZ_SCORE}% threshold for Master classification",
            )
        
        return None
    
    def _check_skimmer_reconsideration(
        self,
        prediction: ClassificationResult,
        avg_time_per_material: float,
    ) -> Optional[GuardrailOverride]:
        """
        Check if Skimmer prediction needs reconsideration.
        
        Rule: IF persona is Skimmer AND avg_time_per_material > 300s, THEN reconsider
        (Requirement 5.2)
        """
        if prediction.persona != LearningPersona.SKIMMER:
            return None
        
        if avg_time_per_material > self.SKIMMER_MIN_TIME_PER_MATERIAL:
            # Reconsider - might be Deep Diver instead
            return GuardrailOverride(
                original_persona=LearningPersona.SKIMMER,
                final_persona=LearningPersona.DEEP_DIVER,
                rule_triggered="skimmer_high_time_per_material",
                reason=f"Average time per material ({avg_time_per_material:.0f}s) exceeds {self.SKIMMER_MIN_TIME_PER_MATERIAL}s, reconsidering classification",
            )
        
        return None
    
    def _check_deep_diver_override(
        self,
        prediction: ClassificationResult,
        total_materials_viewed: int,
    ) -> Optional[GuardrailOverride]:
        """
        Check if Deep Diver prediction should be overridden.
        
        Rule: IF persona is Deep_Diver AND total_materials_viewed < 3, THEN override to Lost
        (Requirement 5.3)
        """
        if prediction.persona != LearningPersona.DEEP_DIVER:
            return None
        
        if total_materials_viewed < self.DEEP_DIVER_MIN_MATERIALS:
            return GuardrailOverride(
                original_persona=LearningPersona.DEEP_DIVER,
                final_persona=LearningPersona.LOST,
                rule_triggered="deep_diver_low_material_count",
                reason=f"Total materials viewed ({total_materials_viewed}) is below {self.DEEP_DIVER_MIN_MATERIALS} minimum for Deep Diver classification",
            )
        
        return None
    
    def _check_anxious_flags(
        self,
        features: FeatureVector,
    ) -> List[str]:
        """
        Check for Anxious behavior indicators.
        
        Rule: IF late_night_ratio > 0.5 AND session_frequency is high, THEN flag potential Anxious
        (Requirement 5.4)
        """
        flags = []
        
        # Check late night + high frequency combination
        if features.late_night_ratio > self.ANXIOUS_LATE_NIGHT_THRESHOLD:
            # session_frequency is normalized, so we check against a threshold
            if features.session_frequency > 0.3:  # Indicates high frequency
                flags.append(
                    f"High late-night activity ({features.late_night_ratio:.0%}) combined with frequent sessions indicates potential anxiety"
                )
        
        return flags
    
    def _check_burnout_flags(
        self,
        features: FeatureVector,
        previous_persona: Optional[LearningPersona],
    ) -> List[str]:
        """
        Check for Burnout indicators.
        
        Rule: IF engagement_trend is declining AND previous_persona was Master, THEN flag potential Burnout
        (Requirement 5.5)
        """
        flags = []
        
        # engagement_trend_encoded: 0=declining, 0.5=stable, 1=increasing
        is_declining = features.engagement_trend_encoded < 0.25
        
        if is_declining and previous_persona == LearningPersona.MASTER:
            flags.append(
                "Declining engagement trend from previous Master classification indicates potential burnout"
            )
        
        return flags
