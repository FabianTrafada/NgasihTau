/**
 * Custom hook for managing AI features state
 * 
 * Provides state management and handlers for:
 * - Suggested questions
 * - Chat export
 * - Feedback submission
 */

import { useState, useCallback } from "react";
import {
  getSuggestedQuestions,
  downloadChatExport,
  submitFeedback,
} from "@/lib/api/ai";
import { ExportFormat, FeedbackType } from "@/types/ai";

// ============================================================================
// HOOK INTERFACE
// ============================================================================

interface UseAIFeaturesProps {
  materialId: string;
}

interface UseAIFeaturesReturn {
  // Suggestions
  suggestions: string[];
  suggestionsLoading: boolean;
  suggestionsError: string | null;
  loadSuggestions: () => Promise<void>;
  
  // Export
  exportLoading: boolean;
  exportError: string | null;
  handleExport: (format: ExportFormat, filename?: string) => Promise<void>;
  
  // Feedback
  feedbackLoading: Record<string, boolean>;
  feedbackError: string | null;
  handleFeedback: (
    messageId: string,
    feedback: FeedbackType,
    feedbackText?: string
  ) => Promise<void>;
}

// ============================================================================
// HOOK IMPLEMENTATION
// ============================================================================

export function useAIFeatures({
  materialId,
}: UseAIFeaturesProps): UseAIFeaturesReturn {
  // Suggestions state
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [suggestionsLoading, setSuggestionsLoading] = useState(false);
  const [suggestionsError, setSuggestionsError] = useState<string | null>(null);

  // Export state
  const [exportLoading, setExportLoading] = useState(false);
  const [exportError, setExportError] = useState<string | null>(null);

  // Feedback state (per message)
  const [feedbackLoading, setFeedbackLoading] = useState<Record<string, boolean>>({});
  const [feedbackError, setFeedbackError] = useState<string | null>(null);

  // ============================================================================
  // SUGGESTIONS HANDLER
  // ============================================================================

  const loadSuggestions = useCallback(async () => {
    console.log("[useAIFeatures] loadSuggestions called for material:", materialId);
    setSuggestionsLoading(true);
    setSuggestionsError(null);

    try {
      console.log("[useAIFeatures] Calling getSuggestedQuestions...");
      const questions = await getSuggestedQuestions(materialId);
      console.log("[useAIFeatures] Received questions:", questions);
      setSuggestions(questions);
      console.log("[useAIFeatures] State updated with suggestions");
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to load suggestions";
      setSuggestionsError(errorMessage);
      console.error("[useAIFeatures] Error loading suggestions:", error);
    } finally {
      setSuggestionsLoading(false);
      console.log("[useAIFeatures] Loading complete");
    }
  }, [materialId]);

  // ============================================================================
  // EXPORT HANDLER
  // ============================================================================

  const handleExport = useCallback(
    async (format: ExportFormat, filename?: string) => {
      setExportLoading(true);
      setExportError(null);

      try {
        await downloadChatExport(materialId, format, filename);
      } catch (error) {
        const errorMessage =
          error instanceof Error ? error.message : "Failed to export chat";
        setExportError(errorMessage);
        console.error("Error exporting chat:", error);
        throw error; // Re-throw so caller can handle if needed
      } finally {
        setExportLoading(false);
      }
    },
    [materialId]
  );

  // ============================================================================
  // FEEDBACK HANDLER
  // ============================================================================

  const handleFeedback = useCallback(
    async (
      messageId: string,
      feedback: FeedbackType,
      feedbackText?: string
    ) => {
      // Set loading state for this specific message
      setFeedbackLoading((prev) => ({ ...prev, [messageId]: true }));
      setFeedbackError(null);

      try {
        await submitFeedback(messageId, feedback, feedbackText);
      } catch (error) {
        const errorMessage =
          error instanceof Error ? error.message : "Failed to submit feedback";
        setFeedbackError(errorMessage);
        console.error("Error submitting feedback:", error);
        throw error; // Re-throw so caller can handle if needed
      } finally {
        // Clear loading state for this message
        setFeedbackLoading((prev) => {
          const next = { ...prev };
          delete next[messageId];
          return next;
        });
      }
    },
    []
  );

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    // Suggestions
    suggestions,
    suggestionsLoading,
    suggestionsError,
    loadSuggestions,

    // Export
    exportLoading,
    exportError,
    handleExport,

    // Feedback
    feedbackLoading,
    feedbackError,
    handleFeedback,
  };
}
