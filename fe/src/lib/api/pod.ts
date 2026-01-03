import apiClient from "@/lib/api-client";
import { CreatePodInput, CreatePodResponse, Pod } from "@/types/pod";
import { Material } from "@/types/material";

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
export async function getPodMaterials(podId: string, limit: number = 20, offset: number = 0): Promise<Material[]> {
  try {
    const token = localStorage.getItem("access_token");
    const response = await apiClient.get<{ data: Material }>(`/api/v1/pods/${podId}/materials`, {
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


export async function createPod(input: CreatePodInput): Promise<Pod> {
  try {
    const token = localStorage.getItem("access_token");

    // Debug logging
    console.log("Creating pod with payload:", input);
    console.log("Token exists:", !!token);

    const response = await apiClient.post<CreatePodResponse>("/api/v1/pods", input, {
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json"
      }
    });

    console.log("Create pod response:", response.data);
    return response.data.data || response.data;
  } catch (error: any) {
    console.error("Error creating pod:", error);
    console.error("Error response:", error.response?.data);
    console.error("Error status:", error.response?.status);
    throw error;
  }
}

// export async function DetailPod(id) {
  
// }




