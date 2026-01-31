/**
 * Type definitions for AI-related API operations
 * Based on Swagger documentation at /api/v1
 */

// ============================================================================
// SUGGESTION TYPES
// ============================================================================

/**
 * Response from GET /materials/{id}/chat/suggestions
 * Contains AI-generated suggested questions based on material content
 */
export interface SuggestedQuestionsResponse {
  questions: string[];
}

// ============================================================================
// EXPORT TYPES
// ============================================================================

/**
 * Supported export formats for chat history
 */
export type ExportFormat = "pdf" | "markdown";

/**
 * Request body for POST /materials/{id}/chat/export
 */
export interface ExportChatRequest {
  format: ExportFormat;
}

/**
 * Export operation state for UI handling
 */
export interface ExportState {
  isExporting: boolean;
  error: string | null;
}

// ============================================================================
// FEEDBACK TYPES
// ============================================================================

/**
 * Feedback types for AI responses
 */
export type FeedbackType = "thumbs_up" | "thumbs_down";

/**
 * Request body for POST /chat/{messageId}/feedback
 */
export interface FeedbackRequest {
  feedback: FeedbackType;
  feedback_text?: string;
}

/**
 * Feedback submission state for UI handling
 */
export interface FeedbackState {
  isSubmitting: boolean;
  error: string | null;
  submitted: boolean;
}
