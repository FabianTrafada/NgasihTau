"""
Recommendation Engine for Learning Pulse Module

Maps Learning Personas to actionable, pedagogically-aware recommendations.
Each persona receives tailored suggestions to improve learning outcomes.
Handles Requirements 6.1-6.11.
"""

from typing import Dict
import logging

from .models import (
    LearningPersona,
    Recommendation,
    PersonaRecommendations,
)

logger = logging.getLogger("learning_pulse.recommendations")


class RecommendationEngine:
    """
    Maps personas to actionable recommendations.
    
    Provides at least 3 recommendations per persona, each with:
    - Unique identifier
    - Human-readable title and description
    - Action type (content, ui, notification, feature)
    - Priority level (1 = highest)
    """
    
    def __init__(self):
        """Initialize the recommendation engine with persona mappings."""
        self._persona_recommendations: Dict[LearningPersona, PersonaRecommendations] = {
            LearningPersona.STRUGGLER: self._get_struggler_recommendations(),
            LearningPersona.SKIMMER: self._get_skimmer_recommendations(),
            LearningPersona.ANXIOUS: self._get_anxious_recommendations(),
            LearningPersona.BURNOUT: self._get_burnout_recommendations(),
            LearningPersona.MASTER: self._get_master_recommendations(),
            LearningPersona.PROCRASTINATOR: self._get_procrastinator_recommendations(),
            LearningPersona.DEEP_DIVER: self._get_deep_diver_recommendations(),
            LearningPersona.SOCIAL_LEARNER: self._get_social_learner_recommendations(),
            LearningPersona.PERFECTIONIST: self._get_perfectionist_recommendations(),
            LearningPersona.LOST: self._get_lost_recommendations(),
        }
    
    def get_recommendations(self, persona: LearningPersona) -> PersonaRecommendations:
        """
        Get recommendations for a persona.
        
        Args:
            persona: The classified Learning Persona
            
        Returns:
            PersonaRecommendations with at least 3 actionable recommendations
        """
        if persona not in self._persona_recommendations:
            logger.error(f"Unknown persona: {persona}")
            raise ValueError(f"Unknown persona: {persona}")
        
        recommendations = self._persona_recommendations[persona]
        logger.info(f"Retrieved {len(recommendations.recommendations)} recommendations for {persona.value}")
        return recommendations
    
    def _get_struggler_recommendations(self) -> PersonaRecommendations:
        """
        Recommendations for Struggler persona.
        
        Struggler: High AI chat usage, repeated questions, low comprehension.
        Focus: Simplified materials, more examples, LLM-generated explanations.
        Validates: Requirement 6.1
        """
        return PersonaRecommendations(
            persona=LearningPersona.STRUGGLER,
            summary="You seem to be working hard but facing challenges. Let's simplify things and provide more support.",
            recommendations=[
                Recommendation(
                    id="struggler_simplified_materials",
                    title="Access Simplified Materials",
                    description="We've prepared easier-to-understand versions of the content. Start with these simplified materials to build a stronger foundation before tackling advanced topics.",
                    action_type="content",
                    priority=1
                ),
                Recommendation(
                    id="struggler_more_examples",
                    title="Explore More Examples",
                    description="Practice makes perfect! Check out additional worked examples that break down complex concepts step-by-step.",
                    action_type="content",
                    priority=2
                ),
                Recommendation(
                    id="struggler_llm_explanations",
                    title="Ask AI for Explanations",
                    description="Use the AI assistant to get personalized explanations. Try asking 'Can you explain this concept in simpler terms?' or 'Give me an analogy for this.'",
                    action_type="feature",
                    priority=3
                ),
                Recommendation(
                    id="struggler_review_basics",
                    title="Review Foundational Concepts",
                    description="Sometimes going back to basics helps. We've identified prerequisite topics that might help strengthen your understanding.",
                    action_type="content",
                    priority=4
                ),
            ],
            ui_hints={
                "highlight_ai_chat": "true",
                "show_difficulty_filter": "true",
                "suggest_prerequisites": "true"
            }
        )
    
    def _get_skimmer_recommendations(self) -> PersonaRecommendations:
        """
        Recommendations for Skimmer persona.
        
        Skimmer: Quick browsing, low engagement, jumps between materials.
        Focus: Engagement prompts, quizzes, summary highlights.
        Validates: Requirement 6.2
        """
        return PersonaRecommendations(
            persona=LearningPersona.SKIMMER,
            summary="You're covering a lot of ground quickly! Let's help you engage more deeply with the material.",
            recommendations=[
                Recommendation(
                    id="skimmer_engagement_prompts",
                    title="Try Interactive Checkpoints",
                    description="We've added quick comprehension checks throughout the material. These short prompts help ensure you're absorbing key concepts as you go.",
                    action_type="ui",
                    priority=1
                ),
                Recommendation(
                    id="skimmer_quizzes",
                    title="Take Quick Quizzes",
                    description="Test your understanding with short quizzes after each section. They only take a few minutes and help reinforce what you've learned.",
                    action_type="feature",
                    priority=2
                ),
                Recommendation(
                    id="skimmer_summary_highlights",
                    title="Review Key Highlights",
                    description="Check out the highlighted summaries and key takeaways for each material. These capture the most important points you shouldn't miss.",
                    action_type="content",
                    priority=3
                ),
                Recommendation(
                    id="skimmer_slow_down_reminder",
                    title="Pace Yourself",
                    description="Quality over quantity! Try spending at least 5 minutes on each material to fully understand the concepts before moving on.",
                    action_type="notification",
                    priority=4
                ),
            ],
            ui_hints={
                "show_reading_progress": "true",
                "enable_comprehension_checks": "true",
                "highlight_key_points": "true"
            }
        )
    
    def _get_anxious_recommendations(self) -> PersonaRecommendations:
        """
        Recommendations for Anxious persona.
        
        Anxious: Erratic patterns, late-night activity, high question frequency.
        Focus: Calming UI elements, progress indicators, encouragement messages.
        Validates: Requirement 6.3
        """
        return PersonaRecommendations(
            persona=LearningPersona.ANXIOUS,
            summary="Learning can feel overwhelming sometimes. Let's create a calmer, more supportive environment for you.",
            recommendations=[
                Recommendation(
                    id="anxious_calming_ui",
                    title="Enable Calm Mode",
                    description="Switch to our calming interface with softer colors and reduced visual clutter. This helps create a more relaxed learning environment.",
                    action_type="ui",
                    priority=1
                ),
                Recommendation(
                    id="anxious_progress_indicators",
                    title="Track Your Progress",
                    description="See how far you've come! Our progress tracker shows your achievements and helps you visualize your learning journey.",
                    action_type="ui",
                    priority=2
                ),
                Recommendation(
                    id="anxious_encouragement",
                    title="You're Doing Great!",
                    description="Remember: learning is a journey, not a race. Every question you ask and every material you review brings you closer to mastery.",
                    action_type="notification",
                    priority=3
                ),
                Recommendation(
                    id="anxious_breathing_break",
                    title="Take a Mindful Break",
                    description="Feeling stressed? Try our 2-minute breathing exercise to reset and refocus before continuing your studies.",
                    action_type="feature",
                    priority=4
                ),
            ],
            ui_hints={
                "enable_calm_mode": "true",
                "show_encouragement_messages": "true",
                "reduce_notifications": "true",
                "soft_color_scheme": "true"
            }
        )
    
    def _get_burnout_recommendations(self) -> PersonaRecommendations:
        """
        Recommendations for Burnout persona.
        
        Burnout: Declining engagement over time, fatigue indicators.
        Focus: Break reminders, lighter content, achievement celebrations.
        Validates: Requirement 6.4
        """
        return PersonaRecommendations(
            persona=LearningPersona.BURNOUT,
            summary="You've been working hard! It's important to rest and recharge. Let's help you maintain a sustainable learning pace.",
            recommendations=[
                Recommendation(
                    id="burnout_break_reminder",
                    title="Take Regular Breaks",
                    description="Schedule 5-minute breaks every 25 minutes using the Pomodoro technique. Your brain needs rest to consolidate learning.",
                    action_type="notification",
                    priority=1
                ),
                Recommendation(
                    id="burnout_lighter_content",
                    title="Try Lighter Content",
                    description="Start with summary materials or video overviews before diving into detailed content. Ease back into learning gradually.",
                    action_type="content",
                    priority=2
                ),
                Recommendation(
                    id="burnout_celebrate_achievements",
                    title="Celebrate Your Progress",
                    description="Look at how much you've accomplished! Review your achievements from the past week and give yourself credit for your hard work.",
                    action_type="ui",
                    priority=3
                ),
                Recommendation(
                    id="burnout_reduce_workload",
                    title="Set Realistic Goals",
                    description="Consider reducing your daily learning goals temporarily. It's better to learn consistently at a sustainable pace than to burn out.",
                    action_type="feature",
                    priority=4
                ),
            ],
            ui_hints={
                "enable_break_reminders": "true",
                "show_achievements": "true",
                "suggest_lighter_content": "true",
                "reduce_daily_goals": "true"
            }
        )
    
    def _get_master_recommendations(self) -> PersonaRecommendations:
        """
        Recommendations for Master persona.
        
        Master: High comprehension, good quiz scores, efficient learning.
        Focus: Advanced materials, peer tutoring opportunities, challenge content.
        Validates: Requirement 6.5
        """
        return PersonaRecommendations(
            persona=LearningPersona.MASTER,
            summary="Excellent work! You're mastering the material. Let's challenge you further and help you share your knowledge.",
            recommendations=[
                Recommendation(
                    id="master_advanced_materials",
                    title="Explore Advanced Content",
                    description="Ready for more? Access advanced materials and deep-dive resources that go beyond the basics.",
                    action_type="content",
                    priority=1
                ),
                Recommendation(
                    id="master_peer_tutoring",
                    title="Help Fellow Learners",
                    description="Your understanding is strong! Consider joining our peer tutoring program to help others while reinforcing your own knowledge.",
                    action_type="feature",
                    priority=2
                ),
                Recommendation(
                    id="master_challenge_content",
                    title="Take on Challenges",
                    description="Test your limits with challenging problems and advanced exercises. Push yourself to the next level!",
                    action_type="content",
                    priority=3
                ),
                Recommendation(
                    id="master_explore_related",
                    title="Explore Related Topics",
                    description="Broaden your expertise by exploring related subjects and interdisciplinary connections.",
                    action_type="content",
                    priority=4
                ),
            ],
            ui_hints={
                "show_advanced_filter": "true",
                "highlight_challenges": "true",
                "enable_tutoring_badge": "true"
            }
        )
    
    def _get_procrastinator_recommendations(self) -> PersonaRecommendations:
        """
        Recommendations for Procrastinator persona.
        
        Procrastinator: Last-minute cramming, irregular patterns.
        Focus: Deadline reminders, micro-tasks, progress tracking.
        Validates: Requirement 6.6
        """
        return PersonaRecommendations(
            persona=LearningPersona.PROCRASTINATOR,
            summary="We notice you tend to study in bursts. Let's help you build more consistent habits with small, manageable steps.",
            recommendations=[
                Recommendation(
                    id="procrastinator_deadline_reminders",
                    title="Set Up Deadline Alerts",
                    description="Enable smart reminders that notify you well before deadlines. Get gentle nudges to start early and avoid last-minute stress.",
                    action_type="notification",
                    priority=1
                ),
                Recommendation(
                    id="procrastinator_micro_tasks",
                    title="Break It Into Micro-Tasks",
                    description="Large tasks feel overwhelming. We've broken down your learning into small, 10-minute chunks that are easy to start.",
                    action_type="feature",
                    priority=2
                ),
                Recommendation(
                    id="procrastinator_progress_tracking",
                    title="Track Daily Progress",
                    description="Build momentum with our daily streak tracker. Even 15 minutes a day adds up to significant progress over time.",
                    action_type="ui",
                    priority=3
                ),
                Recommendation(
                    id="procrastinator_schedule_sessions",
                    title="Schedule Study Sessions",
                    description="Block out specific times for studying in your calendar. Treating learning like an appointment makes it harder to skip.",
                    action_type="feature",
                    priority=4
                ),
            ],
            ui_hints={
                "show_deadline_countdown": "true",
                "enable_streak_tracker": "true",
                "show_micro_tasks": "true",
                "calendar_integration": "true"
            }
        )
    
    def _get_deep_diver_recommendations(self) -> PersonaRecommendations:
        """
        Recommendations for Deep_Diver persona.
        
        Deep_Diver: Long sessions, thorough material consumption.
        Focus: Supplementary resources, deep-dive content, research opportunities.
        Validates: Requirement 6.7
        """
        return PersonaRecommendations(
            persona=LearningPersona.DEEP_DIVER,
            summary="You love going deep into topics! Let's fuel your curiosity with more resources and research opportunities.",
            recommendations=[
                Recommendation(
                    id="deep_diver_supplementary_resources",
                    title="Access Supplementary Resources",
                    description="Explore additional readings, research papers, and external resources that expand on the topics you're studying.",
                    action_type="content",
                    priority=1
                ),
                Recommendation(
                    id="deep_diver_deep_content",
                    title="Unlock Deep-Dive Content",
                    description="Access our extended materials with detailed explanations, case studies, and comprehensive analyses.",
                    action_type="content",
                    priority=2
                ),
                Recommendation(
                    id="deep_diver_research_opportunities",
                    title="Join Research Projects",
                    description="Your thorough approach is perfect for research! Explore opportunities to contribute to ongoing projects or start your own investigation.",
                    action_type="feature",
                    priority=3
                ),
                Recommendation(
                    id="deep_diver_expert_connections",
                    title="Connect with Experts",
                    description="Get access to expert Q&A sessions and office hours where you can discuss advanced topics in depth.",
                    action_type="feature",
                    priority=4
                ),
            ],
            ui_hints={
                "show_related_resources": "true",
                "enable_research_mode": "true",
                "show_citation_links": "true"
            }
        )
    
    def _get_social_learner_recommendations(self) -> PersonaRecommendations:
        """
        Recommendations for Social_Learner persona.
        
        Social_Learner: High collaboration, peer interactions.
        Focus: Study groups, discussion forums, collaborative activities.
        Validates: Requirement 6.8
        """
        return PersonaRecommendations(
            persona=LearningPersona.SOCIAL_LEARNER,
            summary="You thrive when learning with others! Let's connect you with study groups and collaborative opportunities.",
            recommendations=[
                Recommendation(
                    id="social_learner_study_groups",
                    title="Join Study Groups",
                    description="Connect with peers studying the same material. Join or create study groups to learn together and share insights.",
                    action_type="feature",
                    priority=1
                ),
                Recommendation(
                    id="social_learner_discussion_forums",
                    title="Participate in Discussions",
                    description="Join our discussion forums to ask questions, share perspectives, and learn from diverse viewpoints.",
                    action_type="feature",
                    priority=2
                ),
                Recommendation(
                    id="social_learner_collaborative_activities",
                    title="Try Collaborative Projects",
                    description="Work on group projects and collaborative exercises. Learning together often leads to deeper understanding.",
                    action_type="feature",
                    priority=3
                ),
                Recommendation(
                    id="social_learner_peer_review",
                    title="Exchange Peer Feedback",
                    description="Share your work and get feedback from peers. Reviewing others' work also helps reinforce your own learning.",
                    action_type="feature",
                    priority=4
                ),
            ],
            ui_hints={
                "show_active_groups": "true",
                "enable_chat_features": "true",
                "highlight_collaborative_pods": "true"
            }
        )
    
    def _get_perfectionist_recommendations(self) -> PersonaRecommendations:
        """
        Recommendations for Perfectionist persona.
        
        Perfectionist: Excessive review, high self-correction.
        Focus: Time-boxing, good-enough thresholds, progress celebration.
        Validates: Requirement 6.9
        """
        return PersonaRecommendations(
            persona=LearningPersona.PERFECTIONIST,
            summary="Your attention to detail is admirable! Let's help you balance thoroughness with progress.",
            recommendations=[
                Recommendation(
                    id="perfectionist_time_boxing",
                    title="Try Time-Boxing",
                    description="Set time limits for each topic or task. When the timer ends, move on. You can always revisit later if needed.",
                    action_type="feature",
                    priority=1
                ),
                Recommendation(
                    id="perfectionist_good_enough",
                    title="Embrace 'Good Enough'",
                    description="Aim for 80% understanding before moving on. Perfection isn't required for progress—you'll reinforce concepts through practice.",
                    action_type="notification",
                    priority=2
                ),
                Recommendation(
                    id="perfectionist_celebrate_progress",
                    title="Celebrate Your Progress",
                    description="Focus on how far you've come, not just what's left. Every completed section is an achievement worth recognizing.",
                    action_type="ui",
                    priority=3
                ),
                Recommendation(
                    id="perfectionist_limit_reviews",
                    title="Limit Review Cycles",
                    description="Set a maximum of 2 review passes per material. Trust your understanding and move forward with confidence.",
                    action_type="feature",
                    priority=4
                ),
            ],
            ui_hints={
                "show_time_tracker": "true",
                "enable_progress_milestones": "true",
                "limit_review_prompts": "true"
            }
        )
    
    def _get_lost_recommendations(self) -> PersonaRecommendations:
        """
        Recommendations for Lost persona.
        
        Lost: Random navigation, no clear learning path.
        Focus: Guided learning paths, mentor matching, foundational materials.
        Validates: Requirement 6.10
        """
        return PersonaRecommendations(
            persona=LearningPersona.LOST,
            summary="Finding your way can be challenging. Let's create a clear path and connect you with support.",
            recommendations=[
                Recommendation(
                    id="lost_guided_paths",
                    title="Follow a Guided Learning Path",
                    description="We've created a structured learning path just for you. Follow the recommended sequence to build knowledge step-by-step.",
                    action_type="content",
                    priority=1
                ),
                Recommendation(
                    id="lost_mentor_matching",
                    title="Get Matched with a Mentor",
                    description="Connect with an experienced mentor who can guide you, answer questions, and help you stay on track.",
                    action_type="feature",
                    priority=2
                ),
                Recommendation(
                    id="lost_foundational_materials",
                    title="Start with Foundations",
                    description="Begin with our foundational materials that cover the basics. Building a strong foundation makes everything else easier.",
                    action_type="content",
                    priority=3
                ),
                Recommendation(
                    id="lost_goal_setting",
                    title="Set Clear Learning Goals",
                    description="Define what you want to achieve. Having clear goals helps you focus and measure your progress.",
                    action_type="feature",
                    priority=4
                ),
            ],
            ui_hints={
                "show_learning_path": "true",
                "enable_mentor_chat": "true",
                "highlight_prerequisites": "true",
                "show_goal_tracker": "true"
            }
        )
