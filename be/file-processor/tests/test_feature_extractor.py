"""
Tests for Learning Pulse Feature Extractor.

Tests feature extraction from behavior data including:
- Chat feature extraction (Requirements 2.1-2.8)
- Division by zero handling
- Feature normalization to [0, 1] range
"""

import pytest

from learning_pulse.models import (
    BehaviorData,
    ChatBehavior,
    MaterialInteraction,
    ActivityPattern,
)
from learning_pulse.feature_extractor import FeatureExtractor


class TestChatFeatureExtraction:
    """Tests for chat feature extraction (Requirements 2.1-2.8)."""
    
    @pytest.fixture
    def extractor(self) -> FeatureExtractor:
        """Create a feature extractor instance."""
        return FeatureExtractor()
    
    def test_chat_message_ratio_normal(self, extractor: FeatureExtractor):
        """Test chat_message_ratio calculation with normal values."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                total_messages=100,
                user_messages=60,
                assistant_messages=40,
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        # user_messages / total_messages = 60 / 100 = 0.6
        assert features["chat_message_ratio"] == pytest.approx(0.6)
    
    def test_chat_message_ratio_zero_total_messages(self, extractor: FeatureExtractor):
        """Test chat_message_ratio defaults to 0.5 when total_messages is 0."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                total_messages=0,
                user_messages=0,
                assistant_messages=0,
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        # Default to 0.5 when no messages
        assert features["chat_message_ratio"] == 0.5
    
    def test_question_frequency_normal(self, extractor: FeatureExtractor):
        """Test question_frequency calculation with normal values."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                total_messages=100,
                user_messages=50,
                question_count=25,
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        # question_count / user_messages = 25 / 50 = 0.5
        assert features["question_frequency"] == pytest.approx(0.5)
    
    def test_question_frequency_zero_user_messages(self, extractor: FeatureExtractor):
        """Test question_frequency defaults to 0 when user_messages is 0."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                total_messages=50,
                user_messages=0,
                question_count=0,
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        # Default to 0 when no user messages
        assert features["question_frequency"] == 0.0
    
    def test_feedback_ratio_normal(self, extractor: FeatureExtractor):
        """Test feedback_ratio calculation with normal values."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                total_messages=100,
                thumbs_up_count=8,
                thumbs_down_count=2,
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        # thumbs_up / (thumbs_up + thumbs_down) = 8 / 10 = 0.8
        assert features["feedback_ratio"] == pytest.approx(0.8)
    
    def test_feedback_ratio_no_feedback(self, extractor: FeatureExtractor):
        """Test feedback_ratio defaults to 0.5 when no feedback exists."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                total_messages=100,
                thumbs_up_count=0,
                thumbs_down_count=0,
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        # Default to 0.5 when no feedback
        assert features["feedback_ratio"] == 0.5
    
    def test_feedback_ratio_all_positive(self, extractor: FeatureExtractor):
        """Test feedback_ratio is 1.0 when all feedback is positive."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                total_messages=100,
                thumbs_up_count=10,
                thumbs_down_count=0,
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        assert features["feedback_ratio"] == 1.0
    
    def test_feedback_ratio_all_negative(self, extractor: FeatureExtractor):
        """Test feedback_ratio is 0.0 when all feedback is negative."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                total_messages=100,
                thumbs_up_count=0,
                thumbs_down_count=10,
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        assert features["feedback_ratio"] == 0.0
    
    def test_messages_per_session_normal(self, extractor: FeatureExtractor):
        """Test messages_per_session calculation with normal values."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                total_messages=100,
                unique_sessions=10,
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        # total_messages / unique_sessions = 100 / 10 = 10
        # Normalized: 10 / 20 = 0.5
        assert features["messages_per_session"] == pytest.approx(0.5)
    
    def test_messages_per_session_zero_sessions(self, extractor: FeatureExtractor):
        """Test messages_per_session defaults to 0 when no sessions exist."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                total_messages=100,
                unique_sessions=0,
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        # Default to 0 when no sessions
        assert features["messages_per_session"] == 0.0
    
    def test_avg_message_length_norm(self, extractor: FeatureExtractor):
        """Test avg_message_length_norm is normalized correctly."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                avg_message_length=250.0,  # Half of MAX_MESSAGE_LENGTH (500)
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        # 250 / 500 = 0.5
        assert features["avg_message_length_norm"] == pytest.approx(0.5)
    
    def test_session_count_norm(self, extractor: FeatureExtractor):
        """Test session_count_norm is normalized correctly."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                unique_sessions=50,  # Half of MAX_SESSIONS (100)
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        # 50 / 100 = 0.5
        assert features["session_count_norm"] == pytest.approx(0.5)
    
    def test_session_duration_norm(self, extractor: FeatureExtractor):
        """Test session_duration_norm is normalized correctly."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                total_session_duration_minutes=300.0,  # Half of MAX_SESSION_DURATION (600)
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        # 300 / 600 = 0.5
        assert features["session_duration_norm"] == pytest.approx(0.5)
    
    def test_feedback_engagement(self, extractor: FeatureExtractor):
        """Test feedback_engagement calculation."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                total_messages=100,
                thumbs_up_count=5,
                thumbs_down_count=5,
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        # total_feedback / total_messages = 10 / 100 = 0.1
        assert features["feedback_engagement"] == pytest.approx(0.1)
    
    def test_all_features_in_range(self, extractor: FeatureExtractor):
        """Test that all chat features are in [0, 1] range."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior(
                total_messages=150,
                user_messages=80,
                assistant_messages=70,
                question_count=45,
                avg_message_length=85.5,
                thumbs_up_count=12,
                thumbs_down_count=3,
                unique_sessions=25,
                total_session_duration_minutes=180.5,
            )
        )
        
        features = extractor._extract_chat_features(data)
        
        for key, value in features.items():
            assert 0.0 <= value <= 1.0, f"{key} = {value} is out of range [0, 1]"
    
    def test_all_zero_values(self, extractor: FeatureExtractor):
        """Test extraction with all zero values (edge case)."""
        data = BehaviorData(
            user_id="test-user",
            chat=ChatBehavior()  # All defaults to 0
        )
        
        features = extractor._extract_chat_features(data)
        
        # Verify defaults are applied correctly
        assert features["chat_message_ratio"] == 0.5  # Default for ratio
        assert features["question_frequency"] == 0.0  # Default for count-based
        assert features["feedback_ratio"] == 0.5  # Default for ratio
        assert features["messages_per_session"] == 0.0  # Default for count-based
        
        # All features should be in valid range
        for key, value in features.items():
            assert 0.0 <= value <= 1.0, f"{key} = {value} is out of range [0, 1]"


class TestSafeDivide:
    """Tests for the _safe_divide helper method."""
    
    @pytest.fixture
    def extractor(self) -> FeatureExtractor:
        """Create a feature extractor instance."""
        return FeatureExtractor()
    
    def test_safe_divide_normal(self, extractor: FeatureExtractor):
        """Test safe_divide with normal values."""
        assert extractor._safe_divide(10, 2) == 5.0
    
    def test_safe_divide_zero_denominator(self, extractor: FeatureExtractor):
        """Test safe_divide returns default when denominator is zero."""
        assert extractor._safe_divide(10, 0, default=0.5) == 0.5
    
    def test_safe_divide_custom_default(self, extractor: FeatureExtractor):
        """Test safe_divide with custom default value."""
        assert extractor._safe_divide(10, 0, default=0.0) == 0.0
        assert extractor._safe_divide(10, 0, default=1.0) == 1.0


class TestNormalize:
    """Tests for the _normalize helper method."""
    
    @pytest.fixture
    def extractor(self) -> FeatureExtractor:
        """Create a feature extractor instance."""
        return FeatureExtractor()
    
    def test_normalize_normal(self, extractor: FeatureExtractor):
        """Test normalize with normal values."""
        assert extractor._normalize(50, 100) == 0.5
    
    def test_normalize_at_max(self, extractor: FeatureExtractor):
        """Test normalize at max value."""
        assert extractor._normalize(100, 100) == 1.0
    
    def test_normalize_above_max(self, extractor: FeatureExtractor):
        """Test normalize caps at 1.0 when value exceeds max."""
        assert extractor._normalize(150, 100) == 1.0
    
    def test_normalize_zero(self, extractor: FeatureExtractor):
        """Test normalize with zero value."""
        assert extractor._normalize(0, 100) == 0.0
    
    def test_normalize_zero_max(self, extractor: FeatureExtractor):
        """Test normalize returns 0 when max is zero."""
        assert extractor._normalize(50, 0) == 0.0


class TestMaterialFeatureExtraction:
    """Tests for material feature extraction (Requirements 3.1-3.7)."""
    
    @pytest.fixture
    def extractor(self) -> FeatureExtractor:
        """Create a feature extractor instance."""
        return FeatureExtractor()
    
    def test_time_spent_norm_normal(self, extractor: FeatureExtractor):
        """Test time_spent_norm calculation with normal values."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_time_spent_seconds=18000,  # Half of MAX_TIME_SPENT (36000)
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # 18000 / 36000 = 0.5
        assert features["time_spent_norm"] == pytest.approx(0.5)
    
    def test_time_spent_norm_zero(self, extractor: FeatureExtractor):
        """Test time_spent_norm is 0 when no time spent."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_time_spent_seconds=0,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        assert features["time_spent_norm"] == 0.0
    
    def test_view_count_norm_normal(self, extractor: FeatureExtractor):
        """Test view_count_norm calculation with normal values."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_views=100,  # Half of MAX_VIEWS (200)
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # 100 / 200 = 0.5
        assert features["view_count_norm"] == pytest.approx(0.5)
    
    def test_view_count_norm_zero(self, extractor: FeatureExtractor):
        """Test view_count_norm is 0 when no views."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_views=0,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        assert features["view_count_norm"] == 0.0
    
    def test_material_diversity_normal(self, extractor: FeatureExtractor):
        """Test material_diversity calculation with normal values."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_views=100,
                unique_materials_viewed=25,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # unique_materials / total_views = 25 / 100 = 0.25
        assert features["material_diversity"] == pytest.approx(0.25)
    
    def test_material_diversity_zero_views(self, extractor: FeatureExtractor):
        """Test material_diversity defaults to 0 when total_views is 0."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_views=0,
                unique_materials_viewed=0,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # Default to 0 when no views
        assert features["material_diversity"] == 0.0
    
    def test_material_diversity_all_unique(self, extractor: FeatureExtractor):
        """Test material_diversity is 1.0 when all views are unique."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_views=50,
                unique_materials_viewed=50,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # 50 / 50 = 1.0
        assert features["material_diversity"] == 1.0
    
    def test_avg_time_per_view_norm_normal(self, extractor: FeatureExtractor):
        """Test avg_time_per_view_norm calculation with normal values."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_time_spent_seconds=3000,  # 3000 seconds total
                total_views=10,  # 10 views
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # avg_time_per_view = 3000 / 10 = 300 seconds
        # Normalized: 300 / 600 = 0.5
        assert features["avg_time_per_view_norm"] == pytest.approx(0.5)
    
    def test_avg_time_per_view_norm_zero_views(self, extractor: FeatureExtractor):
        """Test avg_time_per_view_norm defaults to 0 when no views."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_time_spent_seconds=1000,
                total_views=0,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # Default to 0 when no views
        assert features["avg_time_per_view_norm"] == 0.0
    
    def test_avg_time_per_view_norm_high_engagement(self, extractor: FeatureExtractor):
        """Test avg_time_per_view_norm caps at 1.0 for high engagement."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_time_spent_seconds=12000,  # 12000 seconds total
                total_views=10,  # 10 views = 1200 sec/view
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # avg_time_per_view = 1200 seconds (exceeds 600 max)
        # Should cap at 1.0
        assert features["avg_time_per_view_norm"] == 1.0
    
    def test_bookmark_ratio_normal(self, extractor: FeatureExtractor):
        """Test bookmark_ratio calculation with normal values."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_views=100,
                bookmark_count=20,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # bookmark_count / total_views = 20 / 100 = 0.2
        assert features["bookmark_ratio"] == pytest.approx(0.2)
    
    def test_bookmark_ratio_zero_views(self, extractor: FeatureExtractor):
        """Test bookmark_ratio defaults to 0 when no views."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_views=0,
                bookmark_count=0,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # Default to 0 when no views
        assert features["bookmark_ratio"] == 0.0
    
    def test_bookmark_ratio_all_bookmarked(self, extractor: FeatureExtractor):
        """Test bookmark_ratio is 1.0 when all views are bookmarked."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_views=50,
                bookmark_count=50,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # 50 / 50 = 1.0
        assert features["bookmark_ratio"] == 1.0
    
    def test_scroll_depth_normal(self, extractor: FeatureExtractor):
        """Test scroll_depth is passed through correctly."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                avg_scroll_depth=0.75,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        assert features["scroll_depth"] == pytest.approx(0.75)
    
    def test_scroll_depth_default(self, extractor: FeatureExtractor):
        """Test scroll_depth uses default value of 0.5."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction()  # Uses default avg_scroll_depth=0.5
        )
        
        features = extractor._extract_material_features(data)
        
        assert features["scroll_depth"] == pytest.approx(0.5)
    
    def test_scroll_depth_zero(self, extractor: FeatureExtractor):
        """Test scroll_depth handles zero value."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                avg_scroll_depth=0.0,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        assert features["scroll_depth"] == 0.0
    
    def test_scroll_depth_max(self, extractor: FeatureExtractor):
        """Test scroll_depth handles max value of 1.0."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                avg_scroll_depth=1.0,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        assert features["scroll_depth"] == 1.0
    
    def test_material_engagement_score_composite(self, extractor: FeatureExtractor):
        """Test material_engagement_score is calculated as weighted average."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_time_spent_seconds=36000,  # time_spent_norm = 1.0
                total_views=100,
                bookmark_count=100,  # bookmark_ratio = 1.0
                avg_scroll_depth=1.0,  # scroll_depth = 1.0
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # material_engagement_score = 0.4 * time_spent_norm + 0.4 * scroll_depth + 0.2 * bookmark_ratio
        # = 0.4 * 1.0 + 0.4 * 1.0 + 0.2 * 1.0 = 1.0
        assert features["material_engagement_score"] == pytest.approx(1.0)
    
    def test_material_engagement_score_zero(self, extractor: FeatureExtractor):
        """Test material_engagement_score is 0 when all components are 0."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_time_spent_seconds=0,  # time_spent_norm = 0
                total_views=0,  # bookmark_ratio = 0 (default)
                bookmark_count=0,
                avg_scroll_depth=0.0,  # scroll_depth = 0
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # material_engagement_score = 0.4 * 0 + 0.4 * 0 + 0.2 * 0 = 0
        assert features["material_engagement_score"] == pytest.approx(0.0)
    
    def test_material_engagement_score_mixed(self, extractor: FeatureExtractor):
        """Test material_engagement_score with mixed values."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_time_spent_seconds=18000,  # time_spent_norm = 0.5
                total_views=100,
                bookmark_count=50,  # bookmark_ratio = 0.5
                avg_scroll_depth=0.5,  # scroll_depth = 0.5
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # material_engagement_score = 0.4 * 0.5 + 0.4 * 0.5 + 0.2 * 0.5 = 0.5
        assert features["material_engagement_score"] == pytest.approx(0.5)
    
    def test_all_features_in_range(self, extractor: FeatureExtractor):
        """Test that all material features are in [0, 1] range."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_time_spent_seconds=7200,
                total_views=45,
                unique_materials_viewed=12,
                bookmark_count=8,
                avg_scroll_depth=0.75,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        for key, value in features.items():
            assert 0.0 <= value <= 1.0, f"{key} = {value} is out of range [0, 1]"
    
    def test_all_zero_values(self, extractor: FeatureExtractor):
        """Test extraction with all zero values (edge case)."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_time_spent_seconds=0,
                total_views=0,
                unique_materials_viewed=0,
                bookmark_count=0,
                avg_scroll_depth=0.0,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # Verify defaults are applied correctly
        assert features["time_spent_norm"] == 0.0
        assert features["view_count_norm"] == 0.0
        assert features["material_diversity"] == 0.0  # Default for division by zero
        assert features["avg_time_per_view_norm"] == 0.0  # Default for division by zero
        assert features["bookmark_ratio"] == 0.0  # Default for division by zero
        assert features["scroll_depth"] == 0.0
        assert features["material_engagement_score"] == 0.0
        
        # All features should be in valid range
        for key, value in features.items():
            assert 0.0 <= value <= 1.0, f"{key} = {value} is out of range [0, 1]"
    
    def test_high_values_capped(self, extractor: FeatureExtractor):
        """Test that values exceeding max are capped at 1.0."""
        data = BehaviorData(
            user_id="test-user",
            material=MaterialInteraction(
                total_time_spent_seconds=100000,  # Exceeds MAX_TIME_SPENT
                total_views=500,  # Exceeds MAX_VIEWS
                unique_materials_viewed=500,
                bookmark_count=500,
                avg_scroll_depth=1.0,
            )
        )
        
        features = extractor._extract_material_features(data)
        
        # All normalized values should be capped at 1.0
        assert features["time_spent_norm"] == 1.0
        assert features["view_count_norm"] == 1.0
        assert features["material_diversity"] == 1.0
        assert features["bookmark_ratio"] == 1.0
        assert features["material_engagement_score"] <= 1.0


class TestActivityFeatureExtraction:
    """Tests for activity feature extraction (Requirements 4.1-4.7)."""
    
    @pytest.fixture
    def extractor(self) -> FeatureExtractor:
        """Create a feature extractor instance."""
        return FeatureExtractor()
    
    # =========================================================================
    # Requirement 4.1: active_days_ratio tests
    # =========================================================================
    
    def test_active_days_ratio_normal(self, extractor: FeatureExtractor):
        """Test active_days_ratio calculation with normal values."""
        data = BehaviorData(
            user_id="test-user",
            analysis_period_days=30,
            activity=ActivityPattern(
                active_days=15,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # active_days / analysis_period_days = 15 / 30 = 0.5
        assert features["active_days_ratio"] == pytest.approx(0.5)
    
    def test_active_days_ratio_zero_active_days(self, extractor: FeatureExtractor):
        """Test active_days_ratio is 0 when no active days."""
        data = BehaviorData(
            user_id="test-user",
            analysis_period_days=30,
            activity=ActivityPattern(
                active_days=0,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        assert features["active_days_ratio"] == 0.0
    
    def test_active_days_ratio_all_days_active(self, extractor: FeatureExtractor):
        """Test active_days_ratio is 1.0 when all days are active."""
        data = BehaviorData(
            user_id="test-user",
            analysis_period_days=30,
            activity=ActivityPattern(
                active_days=30,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # 30 / 30 = 1.0
        assert features["active_days_ratio"] == 1.0
    
    def test_active_days_ratio_capped_at_one(self, extractor: FeatureExtractor):
        """Test active_days_ratio is capped at 1.0 even if active_days > period."""
        data = BehaviorData(
            user_id="test-user",
            analysis_period_days=30,
            activity=ActivityPattern(
                active_days=40,  # Invalid but should be handled gracefully
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # Should be capped at 1.0
        assert features["active_days_ratio"] == 1.0
    
    # =========================================================================
    # Requirement 4.2: session_frequency tests
    # =========================================================================
    
    def test_session_frequency_normal(self, extractor: FeatureExtractor):
        """Test session_frequency calculation with normal values."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                active_days=10,
                total_sessions=50,  # 5 sessions per day
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # total_sessions / active_days = 50 / 10 = 5
        # Normalized: 5 / 10 (MAX_SESSIONS_PER_DAY) = 0.5
        assert features["session_frequency"] == pytest.approx(0.5)
    
    def test_session_frequency_zero_active_days(self, extractor: FeatureExtractor):
        """Test session_frequency defaults to 0 when no active days."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                active_days=0,
                total_sessions=0,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # Default to 0 when no active days
        assert features["session_frequency"] == 0.0
    
    def test_session_frequency_high_frequency(self, extractor: FeatureExtractor):
        """Test session_frequency caps at 1.0 for high frequency."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                active_days=5,
                total_sessions=100,  # 20 sessions per day (exceeds max)
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # 100 / 5 = 20 sessions per day, exceeds MAX_SESSIONS_PER_DAY (10)
        # Should cap at 1.0
        assert features["session_frequency"] == 1.0
    
    def test_session_frequency_one_session_per_day(self, extractor: FeatureExtractor):
        """Test session_frequency with exactly one session per day."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                active_days=10,
                total_sessions=10,  # 1 session per day
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # 10 / 10 = 1 session per day
        # Normalized: 1 / 10 = 0.1
        assert features["session_frequency"] == pytest.approx(0.1)
    
    # =========================================================================
    # Requirement 4.3: consistency_score tests
    # =========================================================================
    
    def test_consistency_score_zero_variance(self, extractor: FeatureExtractor):
        """Test consistency_score is 1.0 when variance is 0 (perfectly consistent)."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                daily_activity_variance=0.0,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # consistency_score = 1 - normalized_variance = 1 - 0 = 1.0
        assert features["consistency_score"] == 1.0
    
    def test_consistency_score_high_variance(self, extractor: FeatureExtractor):
        """Test consistency_score is 0 when variance is at max (very inconsistent)."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                daily_activity_variance=10.0,  # MAX_VARIANCE
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # consistency_score = 1 - normalized_variance = 1 - 1.0 = 0.0
        assert features["consistency_score"] == 0.0
    
    def test_consistency_score_medium_variance(self, extractor: FeatureExtractor):
        """Test consistency_score with medium variance."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                daily_activity_variance=5.0,  # Half of MAX_VARIANCE
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # consistency_score = 1 - (5 / 10) = 1 - 0.5 = 0.5
        assert features["consistency_score"] == pytest.approx(0.5)
    
    def test_consistency_score_exceeds_max_variance(self, extractor: FeatureExtractor):
        """Test consistency_score is 0 when variance exceeds max."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                daily_activity_variance=20.0,  # Exceeds MAX_VARIANCE
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # normalized_variance would be > 1.0, but capped
        # consistency_score = 1 - 1.0 = 0.0
        assert features["consistency_score"] == 0.0
    
    # =========================================================================
    # Requirement 4.4: peak_hour_norm tests
    # =========================================================================
    
    def test_peak_hour_norm_midnight(self, extractor: FeatureExtractor):
        """Test peak_hour_norm is 0 at midnight (hour 0)."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                peak_hour=0,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # 0 / 23 = 0.0
        assert features["peak_hour_norm"] == 0.0
    
    def test_peak_hour_norm_end_of_day(self, extractor: FeatureExtractor):
        """Test peak_hour_norm is 1.0 at hour 23."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                peak_hour=23,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # 23 / 23 = 1.0
        assert features["peak_hour_norm"] == 1.0
    
    def test_peak_hour_norm_noon(self, extractor: FeatureExtractor):
        """Test peak_hour_norm at noon (hour 12)."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                peak_hour=12,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # 12 / 23 ≈ 0.5217
        assert features["peak_hour_norm"] == pytest.approx(12 / 23)
    
    def test_peak_hour_norm_default(self, extractor: FeatureExtractor):
        """Test peak_hour_norm uses default value of 12."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern()  # Uses default peak_hour=12
        )
        
        features = extractor._extract_activity_features(data)
        
        # 12 / 23 ≈ 0.5217
        assert features["peak_hour_norm"] == pytest.approx(12 / 23)
    
    # =========================================================================
    # Requirement 4.5: late_night_ratio tests
    # =========================================================================
    
    def test_late_night_ratio_normal(self, extractor: FeatureExtractor):
        """Test late_night_ratio calculation with normal values."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                total_sessions=100,
                late_night_sessions=25,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # late_night_sessions / total_sessions = 25 / 100 = 0.25
        assert features["late_night_ratio"] == pytest.approx(0.25)
    
    def test_late_night_ratio_zero_sessions(self, extractor: FeatureExtractor):
        """Test late_night_ratio defaults to 0 when no sessions."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                total_sessions=0,
                late_night_sessions=0,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # Default to 0 when no sessions
        assert features["late_night_ratio"] == 0.0
    
    def test_late_night_ratio_all_late_night(self, extractor: FeatureExtractor):
        """Test late_night_ratio is 1.0 when all sessions are late night."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                total_sessions=50,
                late_night_sessions=50,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # 50 / 50 = 1.0
        assert features["late_night_ratio"] == 1.0
    
    def test_late_night_ratio_no_late_night(self, extractor: FeatureExtractor):
        """Test late_night_ratio is 0 when no late night sessions."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                total_sessions=50,
                late_night_sessions=0,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # 0 / 50 = 0.0
        assert features["late_night_ratio"] == 0.0
    
    # =========================================================================
    # Requirement 4.6: weekend_ratio tests
    # =========================================================================
    
    def test_weekend_ratio_normal(self, extractor: FeatureExtractor):
        """Test weekend_ratio calculation with normal values."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                total_sessions=100,
                weekend_sessions=30,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # weekend_sessions / total_sessions = 30 / 100 = 0.3
        assert features["weekend_ratio"] == pytest.approx(0.3)
    
    def test_weekend_ratio_zero_sessions(self, extractor: FeatureExtractor):
        """Test weekend_ratio defaults to 0 when no sessions."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                total_sessions=0,
                weekend_sessions=0,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # Default to 0 when no sessions
        assert features["weekend_ratio"] == 0.0
    
    def test_weekend_ratio_all_weekend(self, extractor: FeatureExtractor):
        """Test weekend_ratio is 1.0 when all sessions are on weekends."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                total_sessions=50,
                weekend_sessions=50,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # 50 / 50 = 1.0
        assert features["weekend_ratio"] == 1.0
    
    def test_weekend_ratio_no_weekend(self, extractor: FeatureExtractor):
        """Test weekend_ratio is 0 when no weekend sessions."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                total_sessions=50,
                weekend_sessions=0,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # 0 / 50 = 0.0
        assert features["weekend_ratio"] == 0.0
    
    # =========================================================================
    # Requirement 4.7: engagement_trend_encoded tests
    # =========================================================================
    
    def test_engagement_trend_stable_default(self, extractor: FeatureExtractor):
        """Test engagement_trend_encoded is 0.5 (stable) for default values."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern()  # All defaults
        )
        
        features = extractor._extract_activity_features(data)
        
        # Default should be stable (0.5)
        assert features["engagement_trend_encoded"] == 0.5
    
    def test_engagement_trend_increasing(self, extractor: FeatureExtractor):
        """Test engagement_trend_encoded is 1.0 for increasing engagement."""
        data = BehaviorData(
            user_id="test-user",
            analysis_period_days=30,
            activity=ActivityPattern(
                active_days=25,  # High active days ratio (> 0.6)
                total_sessions=75,  # 3 sessions per day (> 2)
                daily_activity_variance=1.0,  # Low variance (< 2.0)
                late_night_sessions=5,  # Low late night ratio
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # Should be increasing (1.0)
        assert features["engagement_trend_encoded"] == 1.0
    
    def test_engagement_trend_declining(self, extractor: FeatureExtractor):
        """Test engagement_trend_encoded is 0.0 for declining engagement."""
        data = BehaviorData(
            user_id="test-user",
            analysis_period_days=30,
            activity=ActivityPattern(
                active_days=5,  # Low active days ratio (< 0.3)
                total_sessions=10,  # Some sessions but low frequency
                daily_activity_variance=8.0,  # High variance (> 5.0)
                late_night_sessions=5,  # 50% late night ratio (> 0.4)
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # Should be declining (0.0)
        assert features["engagement_trend_encoded"] == 0.0
    
    def test_engagement_trend_no_sessions(self, extractor: FeatureExtractor):
        """Test engagement_trend_encoded is 0.5 (stable) when no sessions."""
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(
                total_sessions=0,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # No sessions = stable (0.5)
        assert features["engagement_trend_encoded"] == 0.5
    
    # =========================================================================
    # General tests
    # =========================================================================
    
    def test_all_features_in_range(self, extractor: FeatureExtractor):
        """Test that all activity features are in [0, 1] range."""
        data = BehaviorData(
            user_id="test-user",
            analysis_period_days=30,
            activity=ActivityPattern(
                active_days=18,
                total_sessions=35,
                peak_hour=14,
                late_night_sessions=5,
                weekend_sessions=8,
                total_weekday_sessions=27,
                daily_activity_variance=2.5,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        for key, value in features.items():
            assert 0.0 <= value <= 1.0, f"{key} = {value} is out of range [0, 1]"
    
    def test_all_zero_values(self, extractor: FeatureExtractor):
        """Test extraction with all zero values (edge case)."""
        data = BehaviorData(
            user_id="test-user",
            analysis_period_days=30,
            activity=ActivityPattern(
                active_days=0,
                total_sessions=0,
                peak_hour=0,
                late_night_sessions=0,
                weekend_sessions=0,
                total_weekday_sessions=0,
                daily_activity_variance=0.0,
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # Verify defaults are applied correctly
        assert features["active_days_ratio"] == 0.0
        assert features["session_frequency"] == 0.0  # Default for division by zero
        assert features["consistency_score"] == 1.0  # 1 - 0 = 1.0 (perfectly consistent)
        assert features["late_night_ratio"] == 0.0  # Default for division by zero
        assert features["weekend_ratio"] == 0.0  # Default for division by zero
        assert features["peak_hour_norm"] == 0.0  # 0 / 23 = 0
        assert features["engagement_trend_encoded"] == 0.5  # Stable (no sessions)
        
        # All features should be in valid range
        for key, value in features.items():
            assert 0.0 <= value <= 1.0, f"{key} = {value} is out of range [0, 1]"
    
    def test_high_values_capped(self, extractor: FeatureExtractor):
        """Test that values exceeding max are capped at 1.0."""
        data = BehaviorData(
            user_id="test-user",
            analysis_period_days=30,
            activity=ActivityPattern(
                active_days=50,  # Exceeds analysis_period_days
                total_sessions=500,  # Very high session count (10+ per day)
                peak_hour=23,
                late_night_sessions=500,  # All sessions are late night
                weekend_sessions=500,  # All sessions are weekend
                daily_activity_variance=20.0,  # Exceeds MAX_VARIANCE
            )
        )
        
        features = extractor._extract_activity_features(data)
        
        # All values should be capped at 1.0 or 0.0 (for consistency_score)
        assert features["active_days_ratio"] == 1.0
        # 500 sessions / 50 active days = 10 sessions/day, normalized: 10/10 = 1.0
        assert features["session_frequency"] == 1.0
        assert features["consistency_score"] == 0.0  # High variance = low consistency
        assert features["late_night_ratio"] == 1.0
        assert features["weekend_ratio"] == 1.0
        assert features["peak_hour_norm"] == 1.0


class TestEngagementTrendCalculation:
    """Tests for the _calculate_engagement_trend helper method."""
    
    @pytest.fixture
    def extractor(self) -> FeatureExtractor:
        """Create a feature extractor instance."""
        return FeatureExtractor()
    
    def test_stable_with_no_sessions(self, extractor: FeatureExtractor):
        """Test that no sessions results in stable trend."""
        from learning_pulse.models import EngagementTrend
        
        data = BehaviorData(
            user_id="test-user",
            activity=ActivityPattern(total_sessions=0)
        )
        
        trend = extractor._calculate_engagement_trend(data)
        
        assert trend == EngagementTrend.STABLE
    
    def test_increasing_with_high_engagement(self, extractor: FeatureExtractor):
        """Test increasing trend with high engagement indicators."""
        from learning_pulse.models import EngagementTrend
        
        data = BehaviorData(
            user_id="test-user",
            analysis_period_days=30,
            activity=ActivityPattern(
                active_days=25,  # High active days ratio
                total_sessions=75,  # High session frequency
                daily_activity_variance=1.0,  # Low variance
                late_night_sessions=5,  # Low late night ratio
            )
        )
        
        trend = extractor._calculate_engagement_trend(data)
        
        assert trend == EngagementTrend.INCREASING
    
    def test_declining_with_low_engagement(self, extractor: FeatureExtractor):
        """Test declining trend with low engagement indicators."""
        from learning_pulse.models import EngagementTrend
        
        data = BehaviorData(
            user_id="test-user",
            analysis_period_days=30,
            activity=ActivityPattern(
                active_days=5,  # Low active days ratio
                total_sessions=10,  # Low total sessions
                daily_activity_variance=8.0,  # High variance
                late_night_sessions=5,  # High late night ratio (50%)
            )
        )
        
        trend = extractor._calculate_engagement_trend(data)
        
        assert trend == EngagementTrend.DECLINING
    
    def test_stable_with_mixed_indicators(self, extractor: FeatureExtractor):
        """Test stable trend with mixed indicators."""
        from learning_pulse.models import EngagementTrend
        
        data = BehaviorData(
            user_id="test-user",
            analysis_period_days=30,
            activity=ActivityPattern(
                active_days=15,  # Medium active days ratio
                total_sessions=30,  # Medium session count
                daily_activity_variance=3.0,  # Medium variance
                late_night_sessions=3,  # Low late night ratio
            )
        )
        
        trend = extractor._calculate_engagement_trend(data)
        
        assert trend == EngagementTrend.STABLE
