import apiClient from "@/lib/api-client";
import { Pod } from "@/types/pod";
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

export interface PaginatedPodResponse {
  data: Pod[];
  pagination: {
    page: number;
    per_page: number;
    total: number;
  };
}

/**
 * Fetch pods milik user
 * Endpoint: GET /api/v1/users/{id}/pods
 */
export async function getUserPods(userId: string, page: number = 1, perPage: number = 20): Promise<PaginatedPodResponse> {
  try {
    const response = await apiClient.get<PaginatedPodResponse>(`/api/v1/users/${userId}/pods`, {
      params: { page, per_page: perPage },
    });
    return response.data;
  } catch (error) {
    console.error("Error fetching user pods:", error);
    throw error;
  }
}

export interface CreatePodInput {
  name: string;
  description: string;
  visibility: "public" | "private";
  categories?: string[];
  tags?: string[];
}

/**
 * Create a new knowledge pod
 * Endpoint: POST /api/v1/pods
 */
export async function createPod(input: CreatePodInput): Promise<Pod> {
  try {
    const response = await apiClient.post<{ data: Pod }>("/api/v1/pods", input);
    return response.data.data || response.data;
  } catch (error) {
    console.error("Error creating pod:", error);
    throw error;
  }
}

export async function DetailPod(id: string) {
  try {
    const response = await apiClient.get<{ data: Pod }>(`/api/v1/pods/${id}`);
    return response.data.data || response.data;
  } catch (error) {
    console.error("Error fetching pod detail:", error);
    throw error;
  }
}




