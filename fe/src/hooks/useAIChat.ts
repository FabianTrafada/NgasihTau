/**
 * Custom hook untuk AI Chat functionality
 * Provides easy-to-use interface untuk material dan pod chat
 */

import { useState, useCallback, useEffect } from "react";
import {
    chatWithMaterial,
    chatWithPod,
    getMaterialChatHistory,
    getPodChatHistory,
    getMaterialChatSuggestions,
    getPodChatSuggestions,
    exportMaterialChatHistory,
    exportPodChatHistory,
    submitChatFeedback,
    downloadChatExport,
    ChatMessage,
    ChatRequest,
    FeedbackRequest,
} from "@/lib/api/ai";

// =============================================================================
// TYPES
// =============================================================================

interface UseAIChatOptions {
    type: "material" | "pod";
    id: string;
    autoLoadHistory?: boolean;
    historyLimit?: number;
}

interface UseAIChatReturn {
    // State
    messages: ChatMessage[];
    suggestions: string[];
    isLoading: boolean;
    isSending: boolean;
    isLoadingHistory: boolean;
    isLoadingSuggestions: boolean;
    error: string | null;
    sessionId: string | null;

    // Actions
    sendMessage: (message: string) => Promise<void>;
    loadHistory: (limit?: number, offset?: number) => Promise<void>;
    loadSuggestions: () => Promise<void>;
    exportHistory: (format?: string) => Promise<void>;
    submitFeedback: (messageId: string, feedback: FeedbackRequest) => Promise<void>;
    clearMessages: () => void;
    clearError: () => void;
}

// =============================================================================
// HOOK
// =============================================================================

export function useAIChat(options: UseAIChatOptions): UseAIChatReturn {
    const { type, id, autoLoadHistory = false, historyLimit = 20 } = options;

    // State
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [suggestions, setSuggestions] = useState<string[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [isSending, setIsSending] = useState(false);
    const [isLoadingHistory, setIsLoadingHistory] = useState(false);
    const [isLoadingSuggestions, setIsLoadingSuggestions] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [sessionId, setSessionId] = useState<string | null>(null);

    // =============================================================================
    // SEND MESSAGE
    // =============================================================================

    const sendMessage = useCallback(
        async (message: string) => {
            if (!message.trim()) return;

            setIsSending(true);
            setError(null);

            try {
                const request: ChatRequest = {
                    message: message.trim(),
                    session_id: sessionId || undefined,
                };

                // Add user message immediately for optimistic UI
                const userMessage: ChatMessage = {
                    id: `temp-${Date.now()}`,
                    role: "user",
                    content: message.trim(),
                    timestamp: new Date().toISOString(),
                };
                setMessages((prev) => [...prev, userMessage]);

                // Call API based on type
                const response =
                    type === "material"
                        ? await chatWithMaterial(id, request)
                        : await chatWithPod(id, request);

                // Update session ID if provided
                if (response.data.session_id) {
                    setSessionId(response.data.session_id);
                }

                // Add assistant message
                const assistantMessage: ChatMessage = {
                    id: `${response.data.session_id}-${Date.now()}`,
                    role: "assistant",
                    content: response.data.message,
                    timestamp: response.data.timestamp,
                    session_id: response.data.session_id,
                };

                setMessages((prev) => [...prev, assistantMessage]);
            } catch (err: any) {
                setError(err?.response?.data?.message || "Failed to send message");
                console.error("Error sending message:", err);
            } finally {
                setIsSending(false);
            }
        },
        [type, id, sessionId]
    );

    // =============================================================================
    // LOAD HISTORY
    // =============================================================================

    const loadHistory = useCallback(
        async (limit: number = historyLimit, offset: number = 0) => {
            setIsLoadingHistory(true);
            setError(null);

            try {
                const response =
                    type === "material"
                        ? await getMaterialChatHistory(id, limit, offset)
                        : await getPodChatHistory(id, limit, offset);

                setMessages(response.data || []);

                // Extract session ID from first message if available
                if (response.data && response.data.length > 0) {
                    const firstMessage = response.data[0];
                    if (firstMessage.session_id) {
                        setSessionId(firstMessage.session_id);
                    }
                }
            } catch (err: any) {
                setError(err?.response?.data?.message || "Failed to load history");
                console.error("Error loading history:", err);
            } finally {
                setIsLoadingHistory(false);
            }
        },
        [type, id, historyLimit]
    );

    // =============================================================================
    // LOAD SUGGESTIONS
    // =============================================================================

    const loadSuggestions = useCallback(async () => {
        setIsLoadingSuggestions(true);
        setError(null);

        try {
            const response =
                type === "material"
                    ? await getMaterialChatSuggestions(id)
                    : await getPodChatSuggestions(id);

            setSuggestions(
                response.data?.questions || []
            );
        } catch (err: any) {
            setError(err?.response?.data?.message || "Failed to load suggestions");
            console.error("Error loading suggestions:", err);
        } finally {
            setIsLoadingSuggestions(false);
        }
    }, [type, id]);

    // =============================================================================
    // EXPORT HISTORY
    // =============================================================================

    const exportHistory = useCallback(
        async (format: string = "pdf") => {
            setIsLoading(true);
            setError(null);

            try {
                const response =
                    type === "material"
                        ? await exportMaterialChatHistory(id, format)
                        : await exportPodChatHistory(id, format);

                downloadChatExport(
                    response.data,
                    `${type}-${id}-chat-history`
                );
            } catch (err: any) {
                setError(err?.response?.data?.message || "Failed to export history");
                console.error("Error exporting history:", err);
            } finally {
                setIsLoading(false);
            }
        },
        [type, id]
    );

    // =============================================================================
    // SUBMIT FEEDBACK
    // =============================================================================

    const submitFeedback = useCallback(
        async (messageId: string, feedback: FeedbackRequest) => {
            setError(null);

            try {
                await submitChatFeedback(messageId, feedback);
            } catch (err: any) {
                setError(err?.response?.data?.message || "Failed to submit feedback");
                console.error("Error submitting feedback:", err);
                throw err;
            }
        },
        []
    );

    // =============================================================================
    // UTILITY ACTIONS
    // =============================================================================

    const clearMessages = useCallback(() => {
        setMessages([]);
        setSessionId(null);
    }, []);

    const clearError = useCallback(() => {
        setError(null);
    }, []);

    // =============================================================================
    // EFFECTS
    // =============================================================================

    // Auto-load history on mount if enabled
    useEffect(() => {
        if (autoLoadHistory) {
            loadHistory();
        }
    }, [autoLoadHistory, loadHistory]);

    // =============================================================================
    // RETURN
    // =============================================================================

    return {
        // State
        messages,
        suggestions,
        isLoading,
        isSending,
        isLoadingHistory,
        isLoadingSuggestions,
        error,
        sessionId,

        // Actions
        sendMessage,
        loadHistory,
        loadSuggestions,
        exportHistory,
        submitFeedback,
        clearMessages,
        clearError,
    };
}
