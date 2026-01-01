import { apiClient } from '@/lib/api-client';

/**
 * Fetch user detail berdasarkan user ID
 * Endpoint: GET /api/v1/users/{id}
 */

export async function getUserDetail(userId: string): Promise<{ id: string; name: string; username: string; email?: string }> {
  try {
    const response = await apiClient.get<{ data: { id: string; name: string; username: string; email?: string } }>(`/api/v1/users/${userId}`);
    return response.data.data || response.data;
  } catch (error) {
    console.error("Error fetching user detail:", error);
    throw error;
  }
}