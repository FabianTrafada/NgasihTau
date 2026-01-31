import apiClient from "@/lib/api-client";
import { UserBehavior } from "@/types/userBehavior";

/**
 * Fetch the current user's behavior data
 * Endpoint: GET /api/v1/users/me/behavior
 */
export async function getUserBehavior(): Promise<UserBehavior> {
  try {
    const response = await apiClient.get<{ data: UserBehavior }>("/api/v1/users/me/behavior");
    return response.data.data || response.data;
  } catch (error) {
    console.error("Error fetching user behavior:", error);
    throw error;
  }
}
