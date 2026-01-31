"""
Tests for Learning Pulse Persona Classifier.

Tests persona classification including:
- Model loading (Requirements 1.1-1.5)
- Prediction with confidence scoring
- Low confidence flagging (Requirement 1.3)
- Probability distribution
"""

import pytest
from pathlib import Path

from learning_pulse.models import (
    LearningPersona,
    FeatureVector,
    ClassificationResult,
)
from learning_pulse.persona_classifier import PersonaClassifier


# Path to test models
MODELS_DIR = Path(__file__).parent.parent / "learning_pulse" / "models"
MODEL_PATH = MODELS_DIR / "persona_classifier.joblib"
SCALER_PATH = MODELS_DIR / "feature_scaler.joblib"
METADATA_PATH = MODELS_DIR / "model_metadata.json"


class TestPersonaClassifierLoading:
    """Tests for model loading functionality."""
    
    def test_classifier_not_loaded_initially(self):
        """Test that classifier is not loaded when no path is provided."""
        classifier = PersonaClassifier()
        assert not classifier.is_loaded()
    
    def test_classifier_loads_model(self):
        """Test that classifier loads model from file."""
        if not MODEL_PATH.exists():
            pytest.skip("Model file not found")
        
        classifier = PersonaClassifier(model_path=str(MODEL_PATH))
        assert classifier.is_loaded()
    
    def test_classifier_loads_scaler(self):
        """Test that classifier loads scaler from file."""
        if not SCALER_PATH.exists():
            pytest.skip("Scaler file not found")
        
        classifier = PersonaClassifier(
            model_path=str(MODEL_PATH),
            scaler_path=str(SCALER_PATH)
        )
        assert classifier.scaler is not None
    
    def test_classifier_loads_metadata(self):
        """Test that classifier loads metadata from file."""
        if not METADATA_PATH.exists():
            pytest.skip("Metadata file not found")
        
        classifier = PersonaClassifier(
            model_path=str(MODEL_PATH),
            metadata_path=str(METADATA_PATH)
        )
        assert classifier.model_version == "1.0.0"
        assert classifier.last_training_date != "unknown"
    
    def test_classifier_raises_on_missing_model(self):
        """Test that classifier raises FileNotFoundError for missing model."""
        with pytest.raises(FileNotFoundError):
            PersonaClassifier(model_path="/nonexistent/path/model.joblib")
    
    def test_classifier_raises_on_missing_scaler(self):
        """Test that classifier raises FileNotFoundError for missing scaler."""
        if not MODEL_PATH.exists():
            pytest.skip("Model file not found")
        
        with pytest.raises(FileNotFoundError):
            PersonaClassifier(
                model_path=str(MODEL_PATH),
                scaler_path="/nonexistent/path/scaler.joblib"
            )


class TestPersonaClassifierPrediction:
    """Tests for prediction functionality (Requirements 1.1-1.3)."""
    
    @pytest.fixture
    def classifier(self) -> PersonaClassifier:
        """Create a loaded classifier instance."""
        if not MODEL_PATH.exists():
            pytest.skip("Model file not found")
        
        return PersonaClassifier(
            model_path=str(MODEL_PATH),
            scaler_path=str(SCALER_PATH) if SCALER_PATH.exists() else None,
            metadata_path=str(METADATA_PATH) if METADATA_PATH.exists() else None,
        )
    
    @pytest.fixture
    def sample_feature_vector(self) -> FeatureVector:
        """Create a sample feature vector for testing."""
        return FeatureVector(
            # Chat features (8)
            chat_message_ratio=0.6,
            question_frequency=0.3,
            avg_message_length_norm=0.5,
            feedback_ratio=0.7,
            feedback_engagement=0.2,
            session_count_norm=0.4,
            messages_per_session=0.5,
            session_duration_norm=0.4,
            # Material features (7)
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.3,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.2,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            # Activity features (7)
            active_days_ratio=0.6,
            session_frequency=0.5,
            consistency_score=0.7,
            late_night_ratio=0.1,
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.5,
            # Quiz features (3)
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
    
    def test_predict_returns_classification_result(
        self, classifier: PersonaClassifier, sample_feature_vector: FeatureVector
    ):
        """Test that predict returns a ClassificationResult."""
        result = classifier.predict(sample_feature_vector)
        assert isinstance(result, ClassificationResult)
    
    def test_predict_returns_valid_persona(
        self, classifier: PersonaClassifier, sample_feature_vector: FeatureVector
    ):
        """Test that predict returns one of the ten valid personas (Requirement 1.1)."""
        result = classifier.predict(sample_feature_vector)
        assert result.persona in LearningPersona
    
    def test_predict_returns_confidence_in_range(
        self, classifier: PersonaClassifier, sample_feature_vector: FeatureVector
    ):
        """Test that confidence is between 0.0 and 1.0 (Requirement 1.2)."""
        result = classifier.predict(sample_feature_vector)
        assert 0.0 <= result.confidence <= 1.0
    
    def test_predict_returns_probabilities_for_all_personas(
        self, classifier: PersonaClassifier, sample_feature_vector: FeatureVector
    ):
        """Test that probabilities are returned for all personas."""
        result = classifier.predict(sample_feature_vector)
        assert len(result.probabilities) == 10
        
        # All probabilities should sum to approximately 1.0
        total_prob = sum(result.probabilities.values())
        assert pytest.approx(total_prob, abs=0.01) == 1.0
    
    def test_predict_probabilities_in_range(
        self, classifier: PersonaClassifier, sample_feature_vector: FeatureVector
    ):
        """Test that all probabilities are between 0.0 and 1.0."""
        result = classifier.predict(sample_feature_vector)
        for persona, prob in result.probabilities.items():
            assert 0.0 <= prob <= 1.0, f"Probability for {persona} is {prob}"
    
    def test_predict_confidence_matches_max_probability(
        self, classifier: PersonaClassifier, sample_feature_vector: FeatureVector
    ):
        """Test that confidence equals the maximum probability."""
        result = classifier.predict(sample_feature_vector)
        max_prob = max(result.probabilities.values())
        assert result.confidence == pytest.approx(max_prob)
    
    def test_predict_raises_when_model_not_loaded(
        self, sample_feature_vector: FeatureVector
    ):
        """Test that predict raises RuntimeError when model is not loaded."""
        classifier = PersonaClassifier()
        with pytest.raises(RuntimeError, match="Model not loaded"):
            classifier.predict(sample_feature_vector)


class TestLowConfidenceFlag:
    """Tests for low confidence flagging (Requirement 1.3)."""
    
    @pytest.fixture
    def classifier(self) -> PersonaClassifier:
        """Create a loaded classifier instance."""
        if not MODEL_PATH.exists():
            pytest.skip("Model file not found")
        
        return PersonaClassifier(
            model_path=str(MODEL_PATH),
            scaler_path=str(SCALER_PATH) if SCALER_PATH.exists() else None,
            metadata_path=str(METADATA_PATH) if METADATA_PATH.exists() else None,
        )
    
    def test_low_confidence_threshold_is_0_5(self):
        """Test that the low confidence threshold is 0.5."""
        assert PersonaClassifier.LOW_CONFIDENCE_THRESHOLD == 0.5
    
    def test_is_low_confidence_flag_set_correctly(
        self, classifier: PersonaClassifier
    ):
        """Test that is_low_confidence flag matches confidence < 0.5."""
        # Create a feature vector
        features = FeatureVector(
            chat_message_ratio=0.5,
            question_frequency=0.5,
            avg_message_length_norm=0.5,
            feedback_ratio=0.5,
            feedback_engagement=0.5,
            session_count_norm=0.5,
            messages_per_session=0.5,
            session_duration_norm=0.5,
            time_spent_norm=0.5,
            view_count_norm=0.5,
            material_diversity=0.5,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.5,
            scroll_depth=0.5,
            material_engagement_score=0.5,
            active_days_ratio=0.5,
            session_frequency=0.5,
            consistency_score=0.5,
            late_night_ratio=0.5,
            weekend_ratio=0.5,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.5,
            quiz_score_norm=0.5,
            quiz_completion_norm=0.5,
            quiz_attempt_frequency=0.5,
        )
        
        result = classifier.predict(features)
        
        # Verify flag matches threshold
        expected_low_confidence = result.confidence < 0.5
        assert result.is_low_confidence == expected_low_confidence


class TestPredictProba:
    """Tests for predict_proba functionality."""
    
    @pytest.fixture
    def classifier(self) -> PersonaClassifier:
        """Create a loaded classifier instance."""
        if not MODEL_PATH.exists():
            pytest.skip("Model file not found")
        
        return PersonaClassifier(
            model_path=str(MODEL_PATH),
            scaler_path=str(SCALER_PATH) if SCALER_PATH.exists() else None,
            metadata_path=str(METADATA_PATH) if METADATA_PATH.exists() else None,
        )
    
    @pytest.fixture
    def sample_feature_vector(self) -> FeatureVector:
        """Create a sample feature vector for testing."""
        return FeatureVector(
            chat_message_ratio=0.6,
            question_frequency=0.3,
            avg_message_length_norm=0.5,
            feedback_ratio=0.7,
            feedback_engagement=0.2,
            session_count_norm=0.4,
            messages_per_session=0.5,
            session_duration_norm=0.4,
            time_spent_norm=0.5,
            view_count_norm=0.4,
            material_diversity=0.3,
            avg_time_per_view_norm=0.5,
            bookmark_ratio=0.2,
            scroll_depth=0.7,
            material_engagement_score=0.5,
            active_days_ratio=0.6,
            session_frequency=0.5,
            consistency_score=0.7,
            late_night_ratio=0.1,
            weekend_ratio=0.3,
            peak_hour_norm=0.5,
            engagement_trend_encoded=0.5,
            quiz_score_norm=0.7,
            quiz_completion_norm=0.8,
            quiz_attempt_frequency=0.3,
        )
    
    def test_predict_proba_returns_dict(
        self, classifier: PersonaClassifier, sample_feature_vector: FeatureVector
    ):
        """Test that predict_proba returns a dictionary."""
        result = classifier.predict_proba(sample_feature_vector)
        assert isinstance(result, dict)
    
    def test_predict_proba_returns_all_personas(
        self, classifier: PersonaClassifier, sample_feature_vector: FeatureVector
    ):
        """Test that predict_proba returns probabilities for all 10 personas."""
        result = classifier.predict_proba(sample_feature_vector)
        assert len(result) == 10
    
    def test_predict_proba_sums_to_one(
        self, classifier: PersonaClassifier, sample_feature_vector: FeatureVector
    ):
        """Test that probabilities sum to approximately 1.0."""
        result = classifier.predict_proba(sample_feature_vector)
        total = sum(result.values())
        assert pytest.approx(total, abs=0.01) == 1.0
    
    def test_predict_proba_raises_when_model_not_loaded(
        self, sample_feature_vector: FeatureVector
    ):
        """Test that predict_proba raises RuntimeError when model is not loaded."""
        classifier = PersonaClassifier()
        with pytest.raises(RuntimeError, match="Model not loaded"):
            classifier.predict_proba(sample_feature_vector)
    
    def test_predict_proba_matches_predict_probabilities(
        self, classifier: PersonaClassifier, sample_feature_vector: FeatureVector
    ):
        """Test that predict_proba returns same probabilities as predict."""
        proba_result = classifier.predict_proba(sample_feature_vector)
        predict_result = classifier.predict(sample_feature_vector)
        
        for persona, prob in proba_result.items():
            assert prob == pytest.approx(predict_result.probabilities[persona])


class TestModelProperties:
    """Tests for model properties."""
    
    def test_model_version_default(self):
        """Test that model_version defaults to '0.0.0'."""
        classifier = PersonaClassifier()
        assert classifier.model_version == "0.0.0"
    
    def test_last_training_date_default(self):
        """Test that last_training_date defaults to 'unknown'."""
        classifier = PersonaClassifier()
        assert classifier.last_training_date == "unknown"
    
    def test_model_version_from_metadata(self):
        """Test that model_version is loaded from metadata."""
        if not MODEL_PATH.exists() or not METADATA_PATH.exists():
            pytest.skip("Model or metadata file not found")
        
        classifier = PersonaClassifier(
            model_path=str(MODEL_PATH),
            metadata_path=str(METADATA_PATH)
        )
        assert classifier.model_version == "1.0.0"
    
    def test_last_training_date_from_metadata(self):
        """Test that last_training_date is loaded from metadata."""
        if not MODEL_PATH.exists() or not METADATA_PATH.exists():
            pytest.skip("Model or metadata file not found")
        
        classifier = PersonaClassifier(
            model_path=str(MODEL_PATH),
            metadata_path=str(METADATA_PATH)
        )
        assert "2026" in classifier.last_training_date
