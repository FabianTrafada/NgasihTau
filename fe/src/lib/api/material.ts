/**
 * API service untuk Material
 * Handle semua request ke backend terkait material
 */

import { Input } from "@/components/ui/input";
import apiClient from "@/lib/api-client";
import { UploadMaterialInput } from "@/types";
import { Material, ChatSession, ChatMessage, UploadUrlResponse } from "@/types/material";
import { resolve } from "path";
import { confirmUpload, ConfirmUploadRequest } from "./uploadMaterial";
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
    const token = localStorage.getItem("access_token");
    const response = await apiClient.get<{ data: Material }>(
      `/api/v1/materials/${materialId}`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
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
    const response = await apiClient.get<{ data: ChatMessage[] }>(
      `/api/v1/materials/${materialId}/chat/history`,
      { params: { limit, offset } }
    );
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
    const token = localStorage.getItem("access_token");
    const response = await apiClient.get<{ data: { preview_url: string } }>(
      `/api/v1/materials/${materialId}/preview`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
    // Return the presigned URL directly - do NOT modify it as it will break the signature
    return response.data.data?.preview_url || "";
  } catch (error) {
    console.error("Error fetching preview URL:", error);
    throw error;
  }
}

export async function getMaterialDownloadUrl(materialId: string): Promise<string> {
  try {
    const token = localStorage.getItem("access_token");
    const response = await apiClient.get<{ data: { download_url: string } }>(
      `/api/v1/materials/${materialId}/download`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
    // Return the presigned URL directly - do NOT modify it as it will break the signature
    const downloadUrl = response.data?.data?.download_url || "";
    console.log("Download URL fetched:", { materialId, url: downloadUrl });
    return downloadUrl;  } catch (error) {
    console.error("Error fetching download URL:", error);
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
    const response = await apiClient.post<{ data: ChatMessage }>(
      `/api/v1/materials/${materialId}/chat`,
      { message }
    );
    console.log("Chat message response:", response.data);
    return response.data.data || response.data;
  } catch (error) {
    console.error("Error sending chat message:", error);
    throw error;
  }
}

/**
 * Bookmark material
 * Endpoint: POST /api/v1/materials/{id}/bookmark
 */
export async function bookmarkMaterial(materialId: string, folder?: string): Promise<void> {
  try {
    await apiClient.post(`/api/v1/materials/${materialId}/bookmark`, {
      folder,
    });
  } catch (error) {
    console.error("Error bookmarking material:", error);
    throw error;
  }
}

/**
 * Rate material
 * Endpoint: POST /api/v1/materials/{id}/ratings
 */


export async function getUploadUrl(filename: string, content_type: string, size: number): Promise<UploadUrlResponse> {
  try {

    const token = localStorage.getItem("access_token");


    const response = await apiClient.post<{ data: UploadUrlResponse }>(
      `/api/v1/materials/upload-url`,
      {
        filenmae: filename,
        content_type,
        size
      },
      {
        headers: {
          Authorization: `Bearer ${token}`,
        }
      }
    );

    const data = response.data?.data || response.data;


    if (!data.upload_url || !data.object_key) {
      throw new Error("Invalid upload URL response");
    }
    return data;

  } catch (error) {
    console.error("Error getting upload URL:", error);
    throw error;
  }
}

export async function uploadToMinio(
  uploadUrl: string,
  file: File,
  onProgress?: (progress: number) => void
): Promise<void> {
  return new Promise(async (resolve, reject) => {
    const xhr = new XMLHttpRequest();

    xhr.upload.addEventListener("progress", (event) => {
      if (event.lengthComputable && onProgress) {
        const percentComplete = Math.round((event.loaded / event.total) * 100);
        onProgress(percentComplete);
      }
    });
    xhr.addEventListener("load", () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve();
      } else {
        reject(new Error(`Upload failed with status ${xhr.status}`));
      }
    });

    // Handle errors
    xhr.addEventListener("error", () => {
      reject(new Error("Upload failed - Network error"));
    });

    xhr.addEventListener("abort", () => {
      reject(new Error("Upload aborted"));
    });

    xhr.timeout = 5 * 60 * 1000; // 5 minutes
    xhr.open("PUT", uploadUrl);
    xhr.setRequestHeader("Content-Type", file.type);
    xhr.send(file);
  });
}

export async function confrimUpload(
  data: ConfirmUploadRequest
): Promise<Material> {

  try {
    const token = localStorage.getItem("access_token");
    const response = await apiClient.post<{ data: Material }>(
      `/api/v1/materials/confirm-upload`,
      data,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
    const material = response.data?.data || response.data;

    if (!material.id) {
      throw new Error("Invalid confirm upload response");
    }
    return material;


  } catch (error) {
    console.error("Error confirming upload:", error);
    throw error;
  }

}

export async function uploadMaterial(
  file: File,
  podId: string,
  title: string,
  description?: string,
  onProgress?: (progress: number) => void
): Promise<Material> {
  try {
    // Step 1: Get presigned URL
    onProgress?.(5);
    const { upload_url, object_key } = await getUploadUrl(
      file.name,
      file.type,
      file.size
    );
    onProgress?.(10);

    // Step 2: Upload to MinIO
    await uploadToMinio(upload_url, file, (uploadProgress) => {
      const mappedProgress = 10 + uploadProgress * 0.8;
      onProgress?.(Math.round(mappedProgress));
    });
    onProgress?.(95);

    // Step 3: Confirm upload
    const material = await confirmUpload({
      object_key,
      pod_id: podId,
      title,
      description,
    });
    onProgress?.(100);

    return material;
  } catch (error) {
    console.error("Upload flow failed:", error);
    throw error;
  }
}




export async function updateMaterial(materialId: string, input: UploadMaterialInput): Promise<Material> {
  try {
    const token = localStorage.getItem("access_token");
    const response = await apiClient.put<{ data: Material }>(
      `/api/v1/materials/${materialId}`,
      input,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
    return response.data.data || response.data;
  } catch (error) {
    console.error("Error updating material:", error);
    throw error;
  }
}

/**
 * Delete material
 * Endpoint: DELETE /api/v1/materials/{id}
 */
export async function deleteMaterial(materialId: string): Promise<void> {
  try {
    const token = localStorage.getItem("access_token");
    await apiClient.delete(`/api/v1/materials/${materialId}`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
  } catch (error) {
    console.error("Error deleting material:", error);
    throw error;
  }
}
