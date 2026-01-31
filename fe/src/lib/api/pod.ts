import apiClient from "@/lib/api-client";
import { Pod, UpdatePodInput } from "@/types/pod";
import { Material } from "@/types/material";

/**
 * Fetch pod detail berdasarkan pod ID
 * Endpoint: GET /api/v1/pods/{id}
 */
export async function getPodDetail(podId: string): Promise<Pod> {
  // MOCK DATA for "Restore Dummy Data" request
  if (['1', '2', '3', '4'].includes(podId)) {
    const mockPods: Record<string, Pod> = {
      '1': {
        id: '1',
        name: 'Cara Belajar Mobil Kopling (99% bisa 1% nya hanya tuhan ....)',
        description: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. ...',
        visibility: 'public',
        owner_id: 'me',
        view_count: 150,
        star_count: 42,
        fork_count: 8,
        created_at: '2025-01-02T00:00:00Z',
        updated_at: '2025-01-02T00:00:00Z'
      },
      '2': {
        id: '2',
        name: 'Dasar Pemrograman React untuk Pemula',
        description: 'Panduan lengkap memahami component, props, dan state dalam React modern menggunakan TypeScript.',
        visibility: 'public',
        owner_id: 'me',
        view_count: 230,
        star_count: 67,
        fork_count: 15,
        created_at: '2025-01-05T00:00:00Z',
        updated_at: '2025-01-05T00:00:00Z'
      },
      '3': {
        id: '3',
        name: 'Manajemen Waktu untuk Mahasiswa Akhir',
        description: 'Tips dan trik mengelola jadwal skripsi tanpa mengabaikan kesehatan mental dan kehidupan sosial.',
        visibility: 'public',
        owner_id: 'me',
        view_count: 89,
        star_count: 23,
        fork_count: 3,
        created_at: '2025-01-10T00:00:00Z',
        updated_at: '2025-01-10T00:00:00Z'
      },
      '4': {
        id: '4',
        name: 'Food Photography dengan Smartphone',
        description: 'Cara mengambil foto makanan yang estetik hanya dengan modal kamera HP dan cahaya matahari.',
        visibility: 'public',
        owner_id: 'me',
        view_count: 178,
        star_count: 55,
        fork_count: 12,
        created_at: '2025-01-12T00:00:00Z',
        updated_at: '2025-01-12T00:00:00Z'
      }
    };
    return mockPods[podId];
  }

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
  // MOCK DATA for "Restore Dummy Data" request
  if (['1', '2', '3', '4'].includes(podId)) {
    // Return some dummy materials so the pod isn't empty
    return [
      {
        id: `mat-${podId}-1`,
        pod_id: podId,
        uploader_id: 'me',
        title: 'Pengenalan.pdf',
        description: 'Materi pengantar',
        file_type: 'pdf',
        file_url: 'dummy.pdf',
        file_size: 1024 * 1024 * 2,
        current_version: 1,
        status: 'ready',
        view_count: 10,
        download_count: 5,
        average_rating: 4.5,
        rating_count: 2,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      {
        id: `mat-${podId}-2`,
        pod_id: podId,
        uploader_id: 'me',
        title: 'Latihan_Praktek.docx',
        description: 'Soal latihan mingguan',
        file_type: 'docx',
        file_url: 'dummy.docx',
        file_size: 1024 * 500,
        current_version: 1,
        status: 'ready',
        view_count: 8,
        download_count: 3,
        average_rating: 4.8,
        rating_count: 4,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
    ];
  }

  try {
    const token = localStorage.getItem("access_token");
    const response = await apiClient.get<{ data: Material }>(`/api/v1/pods/${podId}/materials`, {
      params: { limit, offset },
      headers: {
        Authorization: `Bearer ${token}`,
      },
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

export async function getStarredPod(userId: string, limit: number = 20, offset: number = 0): Promise<Pod[]> {
  try {
    const token = localStorage.getItem("access_token");
    const page = Math.floor(offset / limit) + 1;
    const response = await apiClient.get<{ data: Pod }>(`/api/v1/users/${userId}/starred`, {
      params: { page, per_page: limit },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    // API returns paginated response with data array
    const data = response.data?.data || response.data;
    return Array.isArray(data) ? data : [];
  } catch (error) {
    console.error("Error fetching starred pod:", error);
    throw error;
  }
}

export async function starPod(podId: string): Promise<void> {
  try {
    const token = localStorage.getItem("access_token");
    await apiClient.post(
      `/api/v1/pods/${podId}/star`,
      {},
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
  } catch (error) {
    console.error("Error starring pod:", error);
    throw error;
  }
}

export async function unstarPod(podId: string): Promise<void> {
  try {
    const token = localStorage.getItem("access_token");
    await apiClient.delete(`/api/v1/pods/${podId}/star`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
  } catch (error) {
    console.error("Error unstarring pod:", error);
    throw error;
  }
}

export async function UpdatePod(podId: string, input: UpdatePodInput): Promise<Pod> {
  try {
    const token = localStorage.getItem("access_token");

    const response = await apiClient.put<{ data: Pod }>(`/api/v1/pods/${podId}`, input, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    return response.data.data || response.data;

  } catch (error) {
    console.error("Error updating pod:", error);
    throw error;
  }
}
