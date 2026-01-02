/**
 * API service untuk Material
 * Handle semua request ke backend terkait material
 */

import apiClient from "@/lib/api-client";
import { Material, ChatSession, ChatMessage } from "@/types/material";

// Use the configured apiClient which already has:
// - Correct base URL from NEXT_PUBLIC_API_URL
// - Token interceptor for auth
// - Response interceptor for token refresh

/**
 * Fetch detail material berdasarkan material_id
 * Endpoint: GET /api/v1/materials/{id}
 */
export async function getMaterialDetail(materialId: string): Promise<Material> {
  try {
    const response = await apiClient.get<{ data: Material }>(`/api/v1/materials/${materialId}`);
    return response.data.data || response.data;
  } catch (error) {
    console.error("Error fetching material detail:", error);
    throw error;
  }
}

/**
 * Get chat history untuk material
 * Endpoint: GET /api/v1/materials/{id}/chat/history (routed to AI Service via Traefik)
 */
export async function getMaterialChatHistory(materialId: string, limit: number = 20, offset: number = 0): Promise<ChatMessage[]> {
  try {
    const response = await apiClient.get<{ data: ChatMessage[] }>(`/api/v1/materials/${materialId}/chat/history`, { params: { limit, offset } });
    return response.data.data || [];
  } catch (error) {
    console.error("Error fetching chat history:", error);
    throw error;
  }
}

/**
 * Get preview URL untuk material
 * Endpoint: GET /api/v1/materials/{id}/preview
 */
export async function getMaterialPreviewUrl(materialId: string): Promise<string> {
  try {
    const response = await apiClient.get<{ data: { preview_url: string } }>(`/api/v1/materials/${materialId}/preview`);
    return response.data.data.preview_url || "";
  } catch (error) {
    console.error("Error fetching preview URL:", error);
    throw error;
  }
}

/**
 * Send chat message ke material
 * Endpoint: POST /api/v1/materials/{id}/chat (routed to AI Service via Traefik)
 */
export async function sendMaterialChatMessage(materialId: string, message: string): Promise<ChatMessage> {
  try {
    console.log("Sending chat message:", { materialId, message });
    const response = await apiClient.post<{ data: ChatMessage }>(`/api/v1/materials/${materialId}/chat`, { message });
    console.log("Chat message response:", response.data);
    return response.data.data || response.data;
  } catch (error) {
    console.error("Error sending chat message:", error);
    if (error instanceof Error) {
      console.error("Error details:", {
        message: error.message,
        status: (error as any).response?.status,
        data: (error as any).response?.data,
      });
    }
    throw error;
  }
}

/**
 * Bookmark material
 * Endpoint: POST /api/v1/materials/{id}/bookmark
 */
export async function bookmarkMaterial(materialId: string, folder?: string): Promise<void> {
  try {
    await apiClient.post(`/api/v1/materials/${materialId}/bookmark`, { folder });
  } catch (error) {
    console.error("Error bookmarking material:", error);
    throw error;
  }
}

/**
 * Rate material
 * Endpoint: POST /api/v1/materials/{id}/ratings
 */
export async function rateMaterial(materialId: string, score: number, review?: string): Promise<void> {
  try {
    await apiClient.post(`/api/v1/materials/${materialId}/ratings`, { score, review });
  } catch (error) {
    console.error("Error rating material:", error);
    throw error;
  }
}
