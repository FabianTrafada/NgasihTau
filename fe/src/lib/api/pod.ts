import apiClient from "@/lib/api-client";
import { Pod, PodMaterial } from "@/types/pod";

/**
 * Fetch pod detail berdasarkan pod ID
 * Endpoint: GET /api/v1/pods/{id}
 */
export async function getPodDetail(podId: string): Promise<Pod> {
  try {
    const response = await apiClient.get<{ data: Pod }>(`/api/v1/pods/${podId}`);
    return response.data.data || response.data;
  } catch (error) {
    console.error("Error fetching pod detail:", error);
    throw error;
  }
}

/**
 * Fetch materials dalam pod
 * Endpoint: GET /api/v1/pods/{id}/materials
 */
export async function getPodMaterials(podId: string, limit: number = 20, offset: number = 0): Promise<PodMaterial[]> {
  try {
    const token = localStorage.getItem("access_token");
    const response = await apiClient.get<any>(`/api/v1/pods/${podId}/materials`, {
      params: { limit, offset },
      headers: {
       Authorization: `Bearer ${token}`
      }
    });
    // API returns paginated response with data array
    const data = response.data?.data || response.data;
    return Array.isArray(data) ? data : [];
  } catch (error) {
    console.error("Error fetching pod materials:", error);
    throw error;
  }
}
