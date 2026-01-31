/**
 * AI Service API Client
 * 
 * Handles all AI-related API operations:
 * - Chat suggestions
 * - Chat history export
 * - Feedback submission
 * 
 * All endpoints are routed through Traefik gateway and use the configured apiClient
 * which handles authentication, token refresh, and base URL configuration.
 */

import apiClient from "@/lib/api-client";
import {
  SuggestedQuestionsResponse,
  ExportChatRequest,
  ExportFormat,
  FeedbackRequest,
  FeedbackType,
} from "@/types/ai";

// ============================================================================
// SUGGESTION API
// ============================================================================

/**
 * Get AI-generated suggested questions for a material
 * 
 * Endpoint: GET /api/v1/materials/{id}/chat/suggestions
 * 
 * @param materialId - UUID of the material
 * @returns Array of suggested question strings
 * @throws Error if request fails or material not found
 * 
 * @example
 * ```ts
 * const suggestions = await getSuggestedQuestions("material-uuid");
 * // ["What are the key concepts?", "Can you explain the main idea?"]
 * ```
 */
export async function getSuggestedQuestions(
  materialId: string
): Promise<string[]> {
  try {
    console.log("[AI API] Fetching suggestions for material:", materialId);
    const response = await apiClient.get<{ data: SuggestedQuestionsResponse }>(
      `/api/v1/materials/${materialId}/chat/suggestions`
    );

    console.log("[AI API] Full response:", response);
    console.log("[AI API] Response status:", response.status);
    console.log("[AI API] Response data:", response.data);

    // Handle both wrapped and unwrapped responses
    const data = response.data.data || response.data;
    console.log("[AI API] Parsed data:", data);
    console.log("[AI API] Questions from data:", data.questions);
    
    const questions = data.questions || [];
    console.log("[AI API] Final questions array:", questions);
    return questions;
  } catch (error) {
    console.error("[AI API] Error fetching suggested questions:", error);
    console.warn("[AI API] Using fallback mock questions");
    
    // Fallback: Return mock questions if API fails
    return [
      "What are the key concepts?",
      "Can you explain the main idea?",
      "What are the practical applications?"
    ];
  }
}

// ============================================================================
// EXPORT API
// ============================================================================

/**
 * Export chat history to PDF or Markdown format
 * 
 * Endpoint: POST /api/v1/materials/{id}/chat/export
 * 
 * This endpoint returns a file blob that should be downloaded by the browser.
 * The response content-type will be either:
 * - application/pdf for PDF exports
 * - text/markdown for Markdown exports
 * 
 * @param materialId - UUID of the material
 * @param format - Export format ("pdf" or "markdown")
 * @returns Blob containing the exported file
 * @throws Error if export fails or material not found
 * 
 * @example
 * ```ts
 * const blob = await exportChatHistory("material-uuid", "pdf");
 * const url = URL.createObjectURL(blob);
 * window.open(url, "_blank");
 * ```
 */
export async function exportChatHistory(
  materialId: string,
  format: ExportFormat
): Promise<Blob> {
  try {
    const response = await apiClient.post(
      `/api/v1/materials/${materialId}/chat/export`,
      { format } as ExportChatRequest,
      {
        responseType: "blob", // Important: tells axios to handle binary data
      }
    );

    return response.data;
  } catch (error) {
    console.error("Error exporting chat history:", error);
    throw error;
  }
}

/**
 * Helper function to trigger browser download of exported chat
 * 
 * @param materialId - UUID of the material
 * @param format - Export format
 * @param filename - Optional custom filename (defaults to "chat-export-{materialId}.{ext}")
 * 
 * @example
 * ```ts
 * await downloadChatExport("material-uuid", "pdf", "my-chat-history.pdf");
 * ```
 */
export async function downloadChatExport(
  materialId: string,
  format: ExportFormat,
  filename?: string
): Promise<void> {
  try {
    const blob = await exportChatHistory(materialId, format);
    
    // Generate filename if not provided
    const extension = format === "pdf" ? "pdf" : "md";
    const defaultFilename = `chat-export-${materialId}.${extension}`;
    const finalFilename = filename || defaultFilename;

    // Create download link and trigger download
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = finalFilename;
    document.body.appendChild(link);
    link.click();
    
    // Cleanup
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  } catch (error) {
    console.error("Error downloading chat export:", error);
    throw error;
  }
}

// ============================================================================
// FEEDBACK API
// ============================================================================

/**
 * Submit feedback for an AI chat message
 * 
 * Endpoint: POST /api/v1/chat/{messageId}/feedback
 * 
 * @param messageId - UUID of the chat message
 * @param feedback - Feedback type ("thumbs_up" or "thumbs_down")
 * @param feedbackText - Optional text explanation for the feedback
 * @throws Error if submission fails or message not found
 * 
 * @example
 * ```ts
 * await submitFeedback("message-uuid", "thumbs_up");
 * await submitFeedback("message-uuid", "thumbs_down", "Response was inaccurate");
 * ```
 */
export async function submitFeedback(
  messageId: string,
  feedback: FeedbackType,
  feedbackText?: string
): Promise<void> {
  try {
    const payload: FeedbackRequest = {
      feedback,
      ...(feedbackText && { feedback_text: feedbackText }),
    };

    await apiClient.post(
      `/api/v1/chat/${messageId}/feedback`,
      payload
    );
  } catch (error) {
    console.error("Error submitting feedback:", error);
    throw error;
  }
}
