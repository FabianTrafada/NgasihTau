"""
Unit Tests for Logic Guardrails

Tests the override rules and flag rules for the Learning Pulse module.
Validates Requirements 5.1, 5.2, 5.3, 5.4, 5.5, 5.6
"""

import pytest
from learning_pulse.models import (
    LearningPersona,
    FeatureVector,
    ClassificationResult,
    GuardrailResult,
    GuardrailOverride,
)
from learning_pulse.guardrails import LogicGuardrails


@pytest.fixture
def guardrails() -> LogicGuardrails:
    """Create a LogicGuardrails instance."""
    return LogicGuardrails()


@pytest.fixture
def sample_feature_vector() -> FeatureVector:
    """Create a sample feature vector with default values."""
    return FeatureVector(
        # Chat features
        chat_message_ratio=0.5,
        question_frequency=0.3,
        avg_message_length_norm=0.4,
        feedback_ratio=0.5,
        feedback_engagement=0.2,
        session_count_norm=0.3,
        messages_per_session=0.4,
        session_duration_norm=0.3,
        # Material features
        time_spent_norm=0.5,
        view_count_norm=0.4,
        material_diversity=0.6,
        avg_time_per_view_norm=0.5,
        bookmark_ratio=0.3,
        scroll_depth=0.7,
        material_engagement_score=0.5,
        # Activity features
        active_days_ratio=0.6,
        session_frequency=0.4,
        consistency_score=0.7,
        late_night_ratio=0.1,
        weekend_ratio=0.3,
        peak_hour_norm=0.5,
        engagement_trend_encoded=0.5,  # stable
        # Quiz features
        quiz_score_norm=0.7,
        quiz_completion_norm=0.8,
        quiz_attempt_frequency=0.3,
    )


def create_classification_result(
    persona: LearningPersona, confidence: float = 0.8
) -> ClassificationResult:
    """Helper to create a ClassificationResult."""
    probabilities = {p.value: 0.02 for p in LearningPersona}
    probabilities[persona.value] = confidence
    return ClassificationResult(
        persona=persona,
        confidence=confidence,
        probabilities=probabilities,
        is_low_confidence=confidence < 0.5,
    )


class TestMasterOverride:
    """Tests for Master → Struggler override rule (Requirement 5.1)."""

    def test_master_with_low_quiz_score_overridden_to_struggler(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Master with quiz_score < 50% should be overridden to Struggler."""
        prediction = create_classification_result(LearningPersona.MASTER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            quiz_score=40.0,  # Below 50% threshold
        )
        
        assert result.persona == LearningPersona.STRUGGLER
        assert result.was_overridden is True
        assert result.override is not None
        assert result.override.original_persona == LearningPersona.MASTER
        assert result.override.final_persona == LearningPersona.STRUGGLER
        assert result.override.rule_triggered == "master_low_quiz_score"

    def test_master_with_high_quiz_score_not_overridden(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Master with quiz_score >= 50% should not be overridden."""
        prediction = create_classification_result(LearningPersona.MASTER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            quiz_score=75.0,  # Above 50% threshold
        )
        
        assert result.persona == LearningPersona.MASTER
        assert result.was_overridden is False
        assert result.override is None

    def test_master_with_exactly_50_quiz_score_not_overridden(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Master with quiz_score == 50% should not be overridden (boundary)."""
        prediction = create_classification_result(LearningPersona.MASTER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            quiz_score=50.0,  # Exactly at threshold
        )
        
        assert result.persona == LearningPersona.MASTER
        assert result.was_overridden is False

    def test_master_with_no_quiz_score_not_overridden(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Master with no quiz_score should not be overridden."""
        prediction = create_classification_result(LearningPersona.MASTER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            quiz_score=None,  # No quiz data
        )
        
        assert result.persona == LearningPersona.MASTER
        assert result.was_overridden is False

    def test_non_master_with_low_quiz_score_not_affected(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Non-Master personas with low quiz scores should not be affected."""
        prediction = create_classification_result(LearningPersona.STRUGGLER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            quiz_score=30.0,  # Low score but not Master
        )
        
        assert result.persona == LearningPersona.STRUGGLER
        assert result.was_overridden is False


class TestDeepDiverOverride:
    """Tests for Deep_Diver → Lost override rule (Requirement 5.3)."""

    def test_deep_diver_with_low_materials_overridden_to_lost(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Deep_Diver with materials_viewed < 3 should be overridden to Lost."""
        prediction = create_classification_result(LearningPersona.DEEP_DIVER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            total_materials_viewed=2,  # Below 3 threshold
        )
        
        assert result.persona == LearningPersona.LOST
        assert result.was_overridden is True
        assert result.override is not None
        assert result.override.original_persona == LearningPersona.DEEP_DIVER
        assert result.override.final_persona == LearningPersona.LOST
        assert result.override.rule_triggered == "deep_diver_low_material_count"

    def test_deep_diver_with_sufficient_materials_not_overridden(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Deep_Diver with materials_viewed >= 3 should not be overridden."""
        prediction = create_classification_result(LearningPersona.DEEP_DIVER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            total_materials_viewed=5,  # Above threshold
        )
        
        assert result.persona == LearningPersona.DEEP_DIVER
        assert result.was_overridden is False

    def test_deep_diver_with_exactly_3_materials_not_overridden(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Deep_Diver with exactly 3 materials should not be overridden (boundary)."""
        prediction = create_classification_result(LearningPersona.DEEP_DIVER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            total_materials_viewed=3,  # Exactly at threshold
        )
        
        assert result.persona == LearningPersona.DEEP_DIVER
        assert result.was_overridden is False

    def test_deep_diver_with_zero_materials_overridden(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Deep_Diver with 0 materials should be overridden to Lost."""
        prediction = create_classification_result(LearningPersona.DEEP_DIVER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            total_materials_viewed=0,
        )
        
        assert result.persona == LearningPersona.LOST
        assert result.was_overridden is True


class TestSkimmerReconsideration:
    """Tests for Skimmer reconsideration rule (Requirement 5.2)."""

    def test_skimmer_with_high_time_reconsidered_to_deep_diver(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Skimmer with avg_time_per_material > 300s should be reconsidered to Deep_Diver."""
        prediction = create_classification_result(LearningPersona.SKIMMER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            avg_time_per_material=350.0,  # Above 300s threshold
        )
        
        assert result.persona == LearningPersona.DEEP_DIVER
        assert result.was_overridden is True
        assert result.override is not None
        assert result.override.original_persona == LearningPersona.SKIMMER
        assert result.override.final_persona == LearningPersona.DEEP_DIVER
        assert result.override.rule_triggered == "skimmer_high_time_per_material"

    def test_skimmer_with_low_time_not_reconsidered(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Skimmer with avg_time_per_material <= 300s should not be reconsidered."""
        prediction = create_classification_result(LearningPersona.SKIMMER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            avg_time_per_material=200.0,  # Below threshold
        )
        
        assert result.persona == LearningPersona.SKIMMER
        assert result.was_overridden is False

    def test_skimmer_with_exactly_300s_not_reconsidered(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Skimmer with exactly 300s should not be reconsidered (boundary)."""
        prediction = create_classification_result(LearningPersona.SKIMMER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            avg_time_per_material=300.0,  # Exactly at threshold
        )
        
        assert result.persona == LearningPersona.SKIMMER
        assert result.was_overridden is False


class TestOverrideReasonTracking:
    """Tests for override reason tracking (Requirement 5.6)."""

    def test_override_includes_original_persona(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Override should include original_persona."""
        prediction = create_classification_result(LearningPersona.MASTER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            quiz_score=30.0,
        )
        
        assert result.override is not None
        assert result.override.original_persona == LearningPersona.MASTER

    def test_override_includes_final_persona(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Override should include final_persona."""
        prediction = create_classification_result(LearningPersona.MASTER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            quiz_score=30.0,
        )
        
        assert result.override is not None
        assert result.override.final_persona == LearningPersona.STRUGGLER

    def test_override_includes_rule_triggered(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Override should include rule_triggered identifier."""
        prediction = create_classification_result(LearningPersona.MASTER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            quiz_score=30.0,
        )
        
        assert result.override is not None
        assert result.override.rule_triggered == "master_low_quiz_score"

    def test_override_includes_human_readable_reason(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Override should include human-readable reason."""
        prediction = create_classification_result(LearningPersona.MASTER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            quiz_score=30.0,
        )
        
        assert result.override is not None
        assert "30.0%" in result.override.reason
        assert "50.0%" in result.override.reason


class TestNoOverrideScenarios:
    """Tests for scenarios where no override should occur."""

    def test_no_override_when_no_rules_triggered(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """No override should occur when no rules are triggered."""
        prediction = create_classification_result(LearningPersona.SOCIAL_LEARNER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            quiz_score=80.0,
            total_materials_viewed=10,
            avg_time_per_material=150.0,
        )
        
        assert result.persona == LearningPersona.SOCIAL_LEARNER
        assert result.was_overridden is False
        assert result.override is None

    def test_confidence_preserved_after_override(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """Confidence should be preserved even after override."""
        prediction = create_classification_result(LearningPersona.MASTER, confidence=0.85)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
            quiz_score=30.0,
        )
        
        assert result.confidence == 0.85


class TestGuardrailThresholds:
    """Tests for guardrail threshold constants."""

    def test_master_min_quiz_score_threshold(self):
        """Master minimum quiz score threshold should be 50.0."""
        assert LogicGuardrails.MASTER_MIN_QUIZ_SCORE == 50.0

    def test_skimmer_min_time_per_material_threshold(self):
        """Skimmer minimum time per material threshold should be 300 seconds."""
        assert LogicGuardrails.SKIMMER_MIN_TIME_PER_MATERIAL == 300

    def test_deep_diver_min_materials_threshold(self):
        """Deep Diver minimum materials threshold should be 3."""
        assert LogicGuardrails.DEEP_DIVER_MIN_MATERIALS == 3


class TestAnxiousFlagRule:
    """Tests for potential_anxious flag rule (Requirement 5.4)."""

    def test_anxious_flag_set_when_high_late_night_and_high_frequency(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """
        Potential anxious flag should be set when late_night_ratio > 0.5 
        AND session_frequency is high.
        
        **Validates: Requirements 5.4**
        """
        # Create feature vector with high late night ratio and high session frequency
        anxious_features = FeatureVector(
            # Chat features
            chat_message_ratio=0.5,
            question_frequency=0.3,
            avg_message_length_norm=0.4,
            feedback_ratio=0.5,
            feedback_engagement=0.2,
            session_count_norm=0.3,
            messages_per_session=0.4,
            session_duration_norm=0.3,
            # Material features
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.6,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.3,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            # Activity features - key values for anxious detection
            active_days_ratio=0.6,
            session_frequency=0.5,  # High frequency (> 0.3 threshold)
            consistency_score=0.7,
            late_night_ratio=0.6,  # High late night (> 0.5 threshold)
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.5,
            # Quiz features
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
        
        prediction = create_classification_result(LearningPersona.ANXIOUS)
        
        result = guardrails.apply(
            prediction=prediction,
            features=anxious_features,
        )
        
        assert result.flags.potential_anxious is True
        assert len(result.flags.flag_reasons) > 0
        assert any("late-night" in reason.lower() for reason in result.flags.flag_reasons)

    def test_anxious_flag_not_set_when_low_late_night_ratio(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """
        Potential anxious flag should NOT be set when late_night_ratio <= 0.5.
        
        **Validates: Requirements 5.4**
        """
        # Create feature vector with low late night ratio but high session frequency
        low_late_night_features = FeatureVector(
            chat_message_ratio=0.5,
            question_frequency=0.3,
            avg_message_length_norm=0.4,
            feedback_ratio=0.5,
            feedback_engagement=0.2,
            session_count_norm=0.3,
            messages_per_session=0.4,
            session_duration_norm=0.3,
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.6,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.3,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            active_days_ratio=0.6,
            session_frequency=0.5,  # High frequency
            consistency_score=0.7,
            late_night_ratio=0.3,  # Low late night (< 0.5 threshold)
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.5,
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
        
        prediction = create_classification_result(LearningPersona.SOCIAL_LEARNER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=low_late_night_features,
        )
        
        assert result.flags.potential_anxious is False

    def test_anxious_flag_not_set_when_low_session_frequency(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """
        Potential anxious flag should NOT be set when session_frequency is low,
        even with high late_night_ratio.
        
        **Validates: Requirements 5.4**
        """
        # Create feature vector with high late night ratio but low session frequency
        low_frequency_features = FeatureVector(
            chat_message_ratio=0.5,
            question_frequency=0.3,
            avg_message_length_norm=0.4,
            feedback_ratio=0.5,
            feedback_engagement=0.2,
            session_count_norm=0.3,
            messages_per_session=0.4,
            session_duration_norm=0.3,
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.6,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.3,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            active_days_ratio=0.6,
            session_frequency=0.2,  # Low frequency (< 0.3 threshold)
            consistency_score=0.7,
            late_night_ratio=0.7,  # High late night (> 0.5 threshold)
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.5,
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
        
        prediction = create_classification_result(LearningPersona.SOCIAL_LEARNER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=low_frequency_features,
        )
        
        assert result.flags.potential_anxious is False

    def test_anxious_flag_boundary_late_night_exactly_0_5(
        self, guardrails: LogicGuardrails
    ):
        """
        Potential anxious flag should NOT be set when late_night_ratio == 0.5 (boundary).
        
        **Validates: Requirements 5.4**
        """
        boundary_features = FeatureVector(
            chat_message_ratio=0.5,
            question_frequency=0.3,
            avg_message_length_norm=0.4,
            feedback_ratio=0.5,
            feedback_engagement=0.2,
            session_count_norm=0.3,
            messages_per_session=0.4,
            session_duration_norm=0.3,
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.6,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.3,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            active_days_ratio=0.6,
            session_frequency=0.5,  # High frequency
            consistency_score=0.7,
            late_night_ratio=0.5,  # Exactly at threshold (should NOT trigger)
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.5,
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
        
        prediction = create_classification_result(LearningPersona.SOCIAL_LEARNER)
        guardrails = LogicGuardrails()
        
        result = guardrails.apply(
            prediction=prediction,
            features=boundary_features,
        )
        
        assert result.flags.potential_anxious is False

    def test_anxious_flag_boundary_session_frequency_exactly_0_3(
        self, guardrails: LogicGuardrails
    ):
        """
        Potential anxious flag should NOT be set when session_frequency == 0.3 (boundary).
        
        **Validates: Requirements 5.4**
        """
        boundary_features = FeatureVector(
            chat_message_ratio=0.5,
            question_frequency=0.3,
            avg_message_length_norm=0.4,
            feedback_ratio=0.5,
            feedback_engagement=0.2,
            session_count_norm=0.3,
            messages_per_session=0.4,
            session_duration_norm=0.3,
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.6,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.3,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            active_days_ratio=0.6,
            session_frequency=0.3,  # Exactly at threshold (should NOT trigger)
            consistency_score=0.7,
            late_night_ratio=0.7,  # High late night
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.5,
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
        
        prediction = create_classification_result(LearningPersona.SOCIAL_LEARNER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=boundary_features,
        )
        
        assert result.flags.potential_anxious is False


class TestBurnoutFlagRule:
    """Tests for potential_burnout flag rule (Requirement 5.5)."""

    def test_burnout_flag_set_when_declining_trend_from_master(
        self, guardrails: LogicGuardrails
    ):
        """
        Potential burnout flag should be set when engagement_trend is declining
        AND previous_persona was Master.
        
        **Validates: Requirements 5.5**
        """
        # Create feature vector with declining engagement trend
        declining_features = FeatureVector(
            chat_message_ratio=0.5,
            question_frequency=0.3,
            avg_message_length_norm=0.4,
            feedback_ratio=0.5,
            feedback_engagement=0.2,
            session_count_norm=0.3,
            messages_per_session=0.4,
            session_duration_norm=0.3,
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.6,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.3,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            active_days_ratio=0.6,
            session_frequency=0.4,
            consistency_score=0.7,
            late_night_ratio=0.1,
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.0,  # Declining (< 0.25)
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
        
        prediction = create_classification_result(LearningPersona.BURNOUT)
        
        result = guardrails.apply(
            prediction=prediction,
            features=declining_features,
            previous_persona=LearningPersona.MASTER,  # Was Master before
        )
        
        assert result.flags.potential_burnout is True
        assert len(result.flags.flag_reasons) > 0
        assert any("burnout" in reason.lower() for reason in result.flags.flag_reasons)

    def test_burnout_flag_not_set_when_stable_trend(
        self, guardrails: LogicGuardrails
    ):
        """
        Potential burnout flag should NOT be set when engagement_trend is stable,
        even if previous_persona was Master.
        
        **Validates: Requirements 5.5**
        """
        stable_features = FeatureVector(
            chat_message_ratio=0.5,
            question_frequency=0.3,
            avg_message_length_norm=0.4,
            feedback_ratio=0.5,
            feedback_engagement=0.2,
            session_count_norm=0.3,
            messages_per_session=0.4,
            session_duration_norm=0.3,
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.6,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.3,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            active_days_ratio=0.6,
            session_frequency=0.4,
            consistency_score=0.7,
            late_night_ratio=0.1,
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.5,  # Stable
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
        
        prediction = create_classification_result(LearningPersona.SOCIAL_LEARNER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=stable_features,
            previous_persona=LearningPersona.MASTER,
        )
        
        assert result.flags.potential_burnout is False

    def test_burnout_flag_not_set_when_increasing_trend(
        self, guardrails: LogicGuardrails
    ):
        """
        Potential burnout flag should NOT be set when engagement_trend is increasing.
        
        **Validates: Requirements 5.5**
        """
        increasing_features = FeatureVector(
            chat_message_ratio=0.5,
            question_frequency=0.3,
            avg_message_length_norm=0.4,
            feedback_ratio=0.5,
            feedback_engagement=0.2,
            session_count_norm=0.3,
            messages_per_session=0.4,
            session_duration_norm=0.3,
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.6,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.3,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            active_days_ratio=0.6,
            session_frequency=0.4,
            consistency_score=0.7,
            late_night_ratio=0.1,
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=1.0,  # Increasing
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
        
        prediction = create_classification_result(LearningPersona.MASTER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=increasing_features,
            previous_persona=LearningPersona.MASTER,
        )
        
        assert result.flags.potential_burnout is False

    def test_burnout_flag_not_set_when_previous_persona_not_master(
        self, guardrails: LogicGuardrails
    ):
        """
        Potential burnout flag should NOT be set when previous_persona was not Master,
        even with declining trend.
        
        **Validates: Requirements 5.5**
        """
        declining_features = FeatureVector(
            chat_message_ratio=0.5,
            question_frequency=0.3,
            avg_message_length_norm=0.4,
            feedback_ratio=0.5,
            feedback_engagement=0.2,
            session_count_norm=0.3,
            messages_per_session=0.4,
            session_duration_norm=0.3,
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.6,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.3,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            active_days_ratio=0.6,
            session_frequency=0.4,
            consistency_score=0.7,
            late_night_ratio=0.1,
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.0,  # Declining
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
        
        prediction = create_classification_result(LearningPersona.STRUGGLER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=declining_features,
            previous_persona=LearningPersona.STRUGGLER,  # Was Struggler, not Master
        )
        
        assert result.flags.potential_burnout is False

    def test_burnout_flag_not_set_when_no_previous_persona(
        self, guardrails: LogicGuardrails
    ):
        """
        Potential burnout flag should NOT be set when previous_persona is None.
        
        **Validates: Requirements 5.5**
        """
        declining_features = FeatureVector(
            chat_message_ratio=0.5,
            question_frequency=0.3,
            avg_message_length_norm=0.4,
            feedback_ratio=0.5,
            feedback_engagement=0.2,
            session_count_norm=0.3,
            messages_per_session=0.4,
            session_duration_norm=0.3,
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.6,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.3,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            active_days_ratio=0.6,
            session_frequency=0.4,
            consistency_score=0.7,
            late_night_ratio=0.1,
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.0,  # Declining
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
        
        prediction = create_classification_result(LearningPersona.BURNOUT)
        
        result = guardrails.apply(
            prediction=prediction,
            features=declining_features,
            previous_persona=None,  # No previous persona
        )
        
        assert result.flags.potential_burnout is False

    def test_burnout_flag_boundary_engagement_trend_exactly_0_25(
        self, guardrails: LogicGuardrails
    ):
        """
        Potential burnout flag should NOT be set when engagement_trend_encoded == 0.25 (boundary).
        
        **Validates: Requirements 5.5**
        """
        boundary_features = FeatureVector(
            chat_message_ratio=0.5,
            question_frequency=0.3,
            avg_message_length_norm=0.4,
            feedback_ratio=0.5,
            feedback_engagement=0.2,
            session_count_norm=0.3,
            messages_per_session=0.4,
            session_duration_norm=0.3,
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.6,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.3,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            active_days_ratio=0.6,
            session_frequency=0.4,
            consistency_score=0.7,
            late_night_ratio=0.1,
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.25,  # Exactly at threshold (should NOT trigger)
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
        
        prediction = create_classification_result(LearningPersona.SOCIAL_LEARNER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=boundary_features,
            previous_persona=LearningPersona.MASTER,
        )
        
        assert result.flags.potential_burnout is False


class TestFlagCombinations:
    """Tests for combined flag scenarios."""

    def test_both_flags_can_be_set_simultaneously(
        self, guardrails: LogicGuardrails
    ):
        """
        Both potential_anxious and potential_burnout flags can be set at the same time.
        
        **Validates: Requirements 5.4, 5.5**
        """
        # Create feature vector that triggers both flags
        both_flags_features = FeatureVector(
            chat_message_ratio=0.5,
            question_frequency=0.3,
            avg_message_length_norm=0.4,
            feedback_ratio=0.5,
            feedback_engagement=0.2,
            session_count_norm=0.3,
            messages_per_session=0.4,
            session_duration_norm=0.3,
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.6,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.3,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            active_days_ratio=0.6,
            session_frequency=0.5,  # High frequency for anxious
            consistency_score=0.7,
            late_night_ratio=0.7,  # High late night for anxious
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.0,  # Declining for burnout
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
        
        prediction = create_classification_result(LearningPersona.BURNOUT)
        
        result = guardrails.apply(
            prediction=prediction,
            features=both_flags_features,
            previous_persona=LearningPersona.MASTER,  # For burnout flag
        )
        
        assert result.flags.potential_anxious is True
        assert result.flags.potential_burnout is True
        assert result.flags.needs_attention is True
        assert len(result.flags.flag_reasons) >= 2

    def test_needs_attention_set_when_any_flag_is_true(
        self, guardrails: LogicGuardrails
    ):
        """
        needs_attention should be True when any flag is set.
        
        **Validates: Requirements 5.4, 5.5**
        """
        anxious_features = FeatureVector(
            chat_message_ratio=0.5,
            question_frequency=0.3,
            avg_message_length_norm=0.4,
            feedback_ratio=0.5,
            feedback_engagement=0.2,
            session_count_norm=0.3,
            messages_per_session=0.4,
            session_duration_norm=0.3,
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.6,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.3,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            active_days_ratio=0.6,
            session_frequency=0.5,  # High frequency
            consistency_score=0.7,
            late_night_ratio=0.7,  # High late night
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.5,  # Stable (no burnout)
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
        
        prediction = create_classification_result(LearningPersona.ANXIOUS)
        
        result = guardrails.apply(
            prediction=prediction,
            features=anxious_features,
        )
        
        assert result.flags.potential_anxious is True
        assert result.flags.potential_burnout is False
        assert result.flags.needs_attention is True

    def test_needs_attention_false_when_no_flags(
        self, guardrails: LogicGuardrails, sample_feature_vector: FeatureVector
    ):
        """
        needs_attention should be False when no flags are set and no override occurred.
        """
        prediction = create_classification_result(LearningPersona.SOCIAL_LEARNER)
        
        result = guardrails.apply(
            prediction=prediction,
            features=sample_feature_vector,
        )
        
        assert result.flags.potential_anxious is False
        assert result.flags.potential_burnout is False
        assert result.flags.needs_attention is False


class TestFlagThresholds:
    """Tests for flag threshold constants."""

    def test_anxious_late_night_threshold(self):
        """Anxious late night threshold should be 0.5."""
        assert LogicGuardrails.ANXIOUS_LATE_NIGHT_THRESHOLD == 0.5

    def test_high_session_frequency_threshold(self):
        """High session frequency threshold should be 3.0."""
        assert LogicGuardrails.HIGH_SESSION_FREQUENCY_THRESHOLD == 3.0
