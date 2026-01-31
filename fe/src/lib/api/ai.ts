/**
 * API service untuk AI Chat
 * Handle semua request ke backend terkait AI chat untuk materials dan pods
 */

import apiClient from "@/lib/api-client";

// =============================================================================
// TYPES
// =============================================================================

export interface ChatMessage {
    id: string;
    role: "user" | "assistant";
    content: string;
    timestamp: string;
    session_id?: string;
}

export interface ChatHistoryResponse {
    data: ChatMessage[];
    pagination?: {
        limit: number;
        offset: number;
        total: number;
    };
}

export interface ChatRequest {
    message: string;
    session_id?: string;
}

export interface ChatResponse {
    data: {
        message: string;
        session_id: string;
        timestamp: string;
    };
}

export interface SuggestedQuestion {
    question: string;
    category?: string;
}

export interface SuggestionsResponse {
    data: {
        questions: string[];
    };
}

export interface ExportChatResponse {
    data: {
        export_url?: string;
        content?: string;
        format: string;
    };
}

export interface FeedbackRequest {
    rating: number; // 1-5 or thumbs up/down
    comment?: string;
    feedback_type?: "helpful" | "not_helpful" | "incorrect" | "other";
}

export interface FeedbackResponse {
    data: {
        message: string;
        feedback_id: string;
    };
}

// =============================================================================
// MATERIAL CHAT APIs
// =============================================================================

/**
 * Chat with material
 * Endpoint: POST /api/v1/materials/{id}/chat
 */
export async function chatWithMaterial(
    materialId: string,
    request: ChatRequest
): Promise<ChatResponse> {
    try {
        const response = await apiClient.post<ChatResponse>(
            `/api/v1/materials/${materialId}/chat`,
            request
        );
        return response.data;
    } catch (error) {
        console.error("Error chatting with material:", error);
        throw error;
    }
}

/**
 * Get chat history for material
 * Endpoint: GET /api/v1/materials/{id}/chat/history
 */
export async function getMaterialChatHistory(
    materialId: string,
    limit: number = 20,
    offset: number = 0
): Promise<ChatHistoryResponse> {
    try {
        const response = await apiClient.get<ChatHistoryResponse>(
            `/api/v1/materials/${materialId}/chat/history`,
            {
                params: { limit, offset },
            }
        );
        return response.data;
    } catch (error) {
        console.error("Error fetching material chat history:", error);
        throw error;
    }
}

/**
 * Get suggested questions for material
 * Endpoint: GET /api/v1/materials/{id}/chat/suggestions
 */
export async function getMaterialChatSuggestions(
    materialId: string,
    locale?: string
): Promise<SuggestionsResponse> {
    try {
        const response = await apiClient.get<SuggestionsResponse>(
            `/api/v1/materials/${materialId}/chat/suggestions`,
            {
                params: locale ? { locale } : undefined
            }
        );
        console.log("response data:", response.data);
        return response.data;
    } catch (error) {
        console.error("Error fetching material chat suggestions:", error);
        throw error;
    }
}

/**
 * Export chat history for material
 * Endpoint: POST /api/v1/materials/{id}/chat/export
 */
export async function exportMaterialChatHistory(
    materialId: string,
    format: string = "pdf"
): Promise<ExportChatResponse> {
    try {
        const response = await apiClient.post<ExportChatResponse>(
            `/api/v1/materials/${materialId}/chat/export`,
            { format }
        );
        return response.data;
    } catch (error) {
        console.error("Error exporting material chat history:", error);
        throw error;
    }
}

// =============================================================================
// POD CHAT APIs
// =============================================================================

/**
 * Chat with pod
 * Endpoint: POST /api/v1/pods/{id}/chat
 */
export async function chatWithPod(
    podId: string,
    request: ChatRequest
): Promise<ChatResponse> {
    try {
        const response = await apiClient.post<ChatResponse>(
            `/api/v1/pods/${podId}/chat`,
            request
        );
        return response.data;
    } catch (error) {
        console.error("Error chatting with pod:", error);
        throw error;
    }
}

/**
 * Get chat history for pod
 * Endpoint: GET /api/v1/pods/{id}/chat/history
 */
export async function getPodChatHistory(
    podId: string,
    limit: number = 20,
    offset: number = 0
): Promise<ChatHistoryResponse> {
    try {
        const response = await apiClient.get<ChatHistoryResponse>(
            `/api/v1/pods/${podId}/chat/history`,
            {
                params: { limit, offset },
            }
        );
        return response.data;
    } catch (error) {
        console.error("Error fetching pod chat history:", error);
        throw error;
    }
}

/**
 * Get suggested questions for pod
 * Endpoint: GET /api/v1/pods/{id}/chat/suggestions
 */
export async function getPodChatSuggestions(
    podId: string
): Promise<SuggestionsResponse> {
    try {
        const response = await apiClient.get<SuggestionsResponse>(
            `/api/v1/pods/${podId}/chat/suggestions`
        );
        return response.data;
    } catch (error) {
        console.error("Error fetching pod chat suggestions:", error);
        throw error;
    }
}

/**
 * Export chat history for pod
 * Endpoint: POST /api/v1/pods/{id}/chat/export
 */
export async function exportPodChatHistory(
    podId: string,
    format: string = "pdf"
): Promise<ExportChatResponse> {
    try {
        const response = await apiClient.post<ExportChatResponse>(
            `/api/v1/pods/${podId}/chat/export`,
            { format }
        );
        return response.data;
    } catch (error) {
        console.error("Error exporting pod chat history:", error);
        throw error;
    }
}

// =============================================================================
// FEEDBACK API
// =============================================================================

/**
 * Submit feedback for a chat message
 * Endpoint: POST /api/v1/chat/{messageId}/feedback
 */
export async function submitChatFeedback(
    messageId: string,
    feedback: FeedbackRequest
): Promise<FeedbackResponse> {
    try {
        const response = await apiClient.post<FeedbackResponse>(
            `/api/v1/chat/${messageId}/feedback`,
            feedback
        );
        return response.data;
    } catch (error) {
        console.error("Error submitting chat feedback:", error);
        throw error;
    }
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

/**
 * Download exported chat history
 * Helper function to trigger download from URL or blob
 */
export function downloadChatExport(
    data: ExportChatResponse["data"],
    filename: string = "chat-history"
): void {
    try {
        if (data.export_url) {
            // If URL is provided, open in new tab or trigger download
            window.open(data.export_url, "_blank");
        } else if (data.content) {
            // If content is provided, create blob and download
            const blob = new Blob([data.content], {
                type: data.format === "pdf" ? "application/pdf" : "text/plain"
            });
            const url = URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = `${filename}.${data.format}`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
        }
    } catch (error) {
        console.error("Error downloading chat export:", error);
        throw error;
    }
}
