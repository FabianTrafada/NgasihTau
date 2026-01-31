"""
Persona Classifier for Learning Pulse Module

Machine learning model for classifying students into Learning Personas
using a Random Forest algorithm. Handles:
- Model loading and inference (Requirements 1.1-1.5)
- Confidence scoring and low-confidence flagging
- Probability distribution over all personas
"""

from typing import Optional, Dict
from pathlib import Path
import json
import logging

import joblib
import numpy as np

from .models import (
    LearningPersona,
    FeatureVector,
    ClassificationResult,
)

logger = logging.getLogger("learning_pulse.persona_classifier")


class PersonaClassifier:
    """
    Random Forest classifier for Learning Personas.
    
    Classifies students into one of ten Learning Personas based on
    their behavioral feature vector. Returns confidence scores and
    flags low-confidence predictions.
    
    Requirements:
    - 1.1: Classify into exactly one of ten Learning Personas
    - 1.2: Return confidence score between 0.0 and 1.0
    - 1.3: Flag classification as low confidence when confidence < 0.5
    """
    
    # Confidence threshold for low-confidence flagging (Requirement 1.3)
    LOW_CONFIDENCE_THRESHOLD = 0.5
    
    # Expected number of personas
    EXPECTED_NUM_CLASSES = 10
    
    def __init__(
        self, 
        model_path: Optional[str] = None,
        scaler_path: Optional[str] = None,
        metadata_path: Optional[str] = None
    ):
        """
        Initialize the persona classifier.
        
        Args:
            model_path: Path to the trained model file (.joblib)
            scaler_path: Path to the feature scaler file (.joblib)
            metadata_path: Path to the model metadata file (.json)
        """
        self.model = None
        self.scaler = None
        self._model_version = "0.0.0"
        self._last_training_date = "unknown"
        self._persona_names: list[str] = []
        
        if model_path:
            self._load_model(model_path)
        
        if scaler_path:
            self._load_scaler(scaler_path)
        
        if metadata_path:
            self._load_metadata(metadata_path)
    
    @property
    def model_version(self) -> str:
        """Get the model version string."""
        return self._model_version
    
    @property
    def last_training_date(self) -> str:
        """Get the last training date string."""
        return self._last_training_date
    
    def _load_model(self, model_path: str) -> None:
        """
        Load a trained model from file.
        
        Args:
            model_path: Path to the joblib-serialized model file
            
        Raises:
            FileNotFoundError: If model file doesn't exist
            ValueError: If model doesn't have expected number of classes
        """
        path = Path(model_path)
        if not path.exists():
            logger.error(f"Model file not found: {model_path}")
            raise FileNotFoundError(f"Model file not found: {model_path}")
        
        try:
            self.model = joblib.load(model_path)
            logger.info(f"Model loaded successfully from {model_path}")
            
            # Validate model has correct number of classes
            if hasattr(self.model, 'n_classes_'):
                n_classes = self.model.n_classes_
                if n_classes != self.EXPECTED_NUM_CLASSES:
                    logger.warning(
                        f"Model has {n_classes} classes, expected {self.EXPECTED_NUM_CLASSES}"
                    )
            
            # Note: model.classes_ contains integer indices (0-9), not persona names
            # Persona names are loaded from metadata file
                
        except Exception as e:
            logger.error(f"Failed to load model: {e}")
            self.model = None
            raise
    
    def _load_scaler(self, scaler_path: str) -> None:
        """
        Load a feature scaler from file.
        
        Args:
            scaler_path: Path to the joblib-serialized scaler file
            
        Raises:
            FileNotFoundError: If scaler file doesn't exist
        """
        path = Path(scaler_path)
        if not path.exists():
            logger.error(f"Scaler file not found: {scaler_path}")
            raise FileNotFoundError(f"Scaler file not found: {scaler_path}")
        
        try:
            self.scaler = joblib.load(scaler_path)
            logger.info(f"Scaler loaded successfully from {scaler_path}")
        except Exception as e:
            logger.error(f"Failed to load scaler: {e}")
            self.scaler = None
            raise
    
    def _load_metadata(self, metadata_path: str) -> None:
        """
        Load model metadata from JSON file.
        
        Args:
            metadata_path: Path to the metadata JSON file
        """
        path = Path(metadata_path)
        if not path.exists():
            logger.warning(f"Metadata file not found: {metadata_path}")
            return
        
        try:
            with open(metadata_path, 'r') as f:
                metadata = json.load(f)
            
            self._model_version = metadata.get('model_version', '0.0.0')
            self._last_training_date = metadata.get('training_date', 'unknown')
            
            # Use persona names from metadata if not already loaded from model
            if not self._persona_names and 'persona_names' in metadata:
                self._persona_names = metadata['persona_names']
            
            logger.info(
                f"Metadata loaded: version={self._model_version}, "
                f"training_date={self._last_training_date}"
            )
        except Exception as e:
            logger.warning(f"Failed to load metadata: {e}")
    
    def is_loaded(self) -> bool:
        """Check if the model is loaded and ready for inference."""
        return self.model is not None
    
    def predict(self, features: FeatureVector) -> ClassificationResult:
        """
        Predict persona from feature vector.
        
        Implements Requirements:
        - 1.1: Returns exactly one of ten Learning Personas
        - 1.2: Returns confidence score between 0.0 and 1.0
        - 1.3: Sets is_low_confidence flag when confidence < 0.5
        
        Args:
            features: Normalized feature vector
            
        Returns:
            Classification result with persona, confidence, and probabilities
            
        Raises:
            RuntimeError: If model is not loaded
        """
        if not self.is_loaded():
            raise RuntimeError("Model not loaded. Call _load_model() first.")
        
        # Convert FeatureVector to numpy array
        feature_array = np.array([self._feature_vector_to_array(features)])
        
        # Apply scaling if scaler is available
        if self.scaler is not None:
            feature_array = self.scaler.transform(feature_array)
        
        # Get prediction and probabilities from model
        predicted_class = self.model.predict(feature_array)[0]
        probabilities = self.model.predict_proba(feature_array)[0]
        
        # Build probability dictionary
        prob_dict = self._build_probability_dict(probabilities)
        
        # Determine confidence (max probability) - Requirement 1.2
        confidence = float(max(probabilities))
        
        # Convert predicted class to LearningPersona enum - Requirement 1.1
        # The model may return either a string persona name or an integer index
        persona = self._convert_to_persona(predicted_class)
        
        # Flag if confidence < 0.5 - Requirement 1.3
        is_low_confidence = confidence < self.LOW_CONFIDENCE_THRESHOLD
        
        if is_low_confidence:
            logger.warning(
                f"Low confidence prediction: persona={persona.value}, "
                f"confidence={confidence:.3f}"
            )
        else:
            logger.info(
                f"Prediction completed: persona={persona.value}, "
                f"confidence={confidence:.3f}"
            )
        
        return ClassificationResult(
            persona=persona,
            confidence=confidence,
            probabilities=prob_dict,
            is_low_confidence=is_low_confidence
        )
    
    def predict_proba(self, features: FeatureVector) -> Dict[str, float]:
        """
        Get probability distribution over all personas.
        
        Args:
            features: Normalized feature vector
            
        Returns:
            Dictionary mapping persona names to probabilities
            
        Raises:
            RuntimeError: If model is not loaded
        """
        if not self.is_loaded():
            raise RuntimeError("Model not loaded. Call _load_model() first.")
        
        # Convert FeatureVector to numpy array
        feature_array = np.array([self._feature_vector_to_array(features)])
        
        # Apply scaling if scaler is available
        if self.scaler is not None:
            feature_array = self.scaler.transform(feature_array)
        
        # Get probabilities from model
        probabilities = self.model.predict_proba(feature_array)[0]
        
        return self._build_probability_dict(probabilities)
    
    def _build_probability_dict(self, probabilities: np.ndarray) -> Dict[str, float]:
        """
        Build a dictionary mapping persona names to probabilities.
        
        Args:
            probabilities: Array of probabilities from model
            
        Returns:
            Dictionary mapping persona names to probabilities
        """
        # Use persona names from model classes or metadata
        if self._persona_names:
            return {
                name: float(prob) 
                for name, prob in zip(self._persona_names, probabilities)
            }
        
        # Fallback: use LearningPersona enum values in order
        persona_values = [p.value for p in LearningPersona]
        return {
            name: float(prob) 
            for name, prob in zip(persona_values, probabilities)
        }
    
    def _convert_to_persona(self, predicted_class) -> LearningPersona:
        """
        Convert a predicted class to a LearningPersona enum.
        
        The model returns integer indices (0-9) that map to persona names.
        Persona names are loaded from metadata file.
        
        Args:
            predicted_class: The predicted class from the model (integer index)
            
        Returns:
            LearningPersona enum value
        """
        # If it's already a string, convert directly
        if isinstance(predicted_class, str):
            return LearningPersona(predicted_class)
        
        # Convert numpy integer to Python int
        idx = int(predicted_class)
        
        # Use persona names from metadata if available
        if self._persona_names and 0 <= idx < len(self._persona_names):
            persona_name = self._persona_names[idx]
            # Ensure it's a string (not numpy type)
            if not isinstance(persona_name, str):
                persona_name = str(persona_name)
            return LearningPersona(persona_name)
        
        # Fallback: use LearningPersona enum values in order
        persona_values = list(LearningPersona)
        if 0 <= idx < len(persona_values):
            return persona_values[idx]
        
        raise ValueError(f"Cannot convert {predicted_class} to LearningPersona")
    
    def _feature_vector_to_array(self, features: FeatureVector) -> list:
        """
        Convert FeatureVector to a list for model input.
        
        Ensures consistent feature ordering matching the training data.
        The order must match the feature_names in model_metadata.json.
        """
        return [
            # Chat features (8)
            features.chat_message_ratio,
            features.question_frequency,
            features.avg_message_length_norm,
            features.feedback_ratio,
            features.feedback_engagement,
            features.session_count_norm,
            features.messages_per_session,
            features.session_duration_norm,
            # Material features (7)
            features.time_spent_norm,
            features.view_count_norm,
            features.material_diversity,
            features.avg_time_per_view_norm,
            features.bookmark_ratio,
            features.scroll_depth,
            features.material_engagement_score,
            # Activity features (7)
            features.active_days_ratio,
            features.session_frequency,
            features.consistency_score,
            features.late_night_ratio,
            features.weekend_ratio,
            features.peak_hour_norm,
            features.engagement_trend_encoded,
            # Quiz features (3)
            features.quiz_score_norm,
            features.quiz_completion_norm,
            features.quiz_attempt_frequency,
        ]
