import apiClient from "../api-client";

// =============================================================================
// TYPES
// =============================================================================

/**
 * Generic API response wrapper from backend
 */
export interface ApiResponse<T> {
    success: boolean;
    data: T;
    meta?: {
        timestamp: string;
        request_id: string;
    };
}

/**
 * Data from requesting presigned upload URL
 */
export interface UploadUrlData {
    upload_url: string;  // Presigned PUT URL for MinIO
    object_key: string;  // Object identifier in MinIO
    expires_at: string;  // Expiration timestamp
}

/**
 * Request body for getting upload URL
 */
export interface UploadUrlRequest {
    filename: string;
    content_type: string;
    size: number;
}

/**
 * Request body for confirming upload
 */
export interface ConfirmUploadRequest {
    object_key: string;
    pod_id: string;
    title: string;
    description?: string;
}

/**
 * Material object returned from backend
 */
export interface Material {
    id: string;
    pod_id: string;
    uploader_id: string;
    title: string;
    description?: string;
    file_type: string;
    file_url: string;
    file_size: number;
    current_version: number;
    status: string;
    view_count: number;
    download_count: number;
    average_rating: number;
    rating_count: number;
    created_at: string;
    updated_at: string;
}

/**
 * Response after confirming upload - uses the generic ApiResponse wrapper
 */
export type ConfirmUploadResponse = ApiResponse<Material>;

// =============================================================================
// API FUNCTIONS
// =============================================================================

/**
 * Step 1: Request presigned upload URL from backend
 * 
 * @param fileName - Original file name
 * @param contentType - MIME type (e.g., 'application/pdf')
 * @param fileSize - File size in bytes
 * @param podId - Pod UUID string
 * @returns Promise with upload_url and object_key
 */
export async function getUploadUrl(
    fileName: string,
    contentType: string,
    fileSize: number,
    podId: string
): Promise<UploadUrlData> {
    console.log('[getUploadUrl] Requesting presigned URL:', {
        fileName,
        contentType,
        fileSize,
        podId
    });

    try {
        // Note: podId is not needed for getting upload URL, only for confirm
        const response = await apiClient.post<ApiResponse<UploadUrlData>>(
            '/api/v1/materials/upload-url',
            {
                filename: fileName,
                content_type: contentType,
                size: fileSize
            }
        );

        console.log('[getUploadUrl] Response received:', response.data);

        // Validate response - backend wraps data in { success, data, meta }
        const data = response.data?.data;
        if (!data || !data.upload_url || !data.object_key) {
            console.error('[getUploadUrl] ❌ Invalid response from backend:', response.data);
            throw new Error('Backend returned invalid upload URL response');
        }

        console.log('[getUploadUrl] ✅ Valid presigned URL received');
        return data;
    } catch (error) {
        console.error('[getUploadUrl] ❌ Failed to get presigned URL:', error);
        throw error;
    }
}

/**
 * Step 2: Upload file directly to MinIO using presigned URL
 * 
 * @param uploadUrl - Presigned PUT URL from getUploadUrl
 * @param file - File object to upload
 * @param onProgress - Optional callback for upload progress (0-100)
 * @returns Promise that resolves when upload completes
 */
export async function uploadToMinio(
    uploadUrl: string,
    file: File,
    onProgress?: (progress: number) => void
): Promise<void> {
    console.log('[uploadToMinio] Starting upload to MinIO:', {
        uploadUrl,
        fileName: file.name,
        fileSize: file.size,
        fileType: file.type
    });

    // Validate presigned URL
    if (!uploadUrl || uploadUrl === 'undefined') {
        console.error('[uploadToMinio] ❌ Invalid upload URL:', uploadUrl);
        throw new Error('Invalid presigned upload URL');
    }

    // Check if URL is valid
    try {
        new URL(uploadUrl);
        console.log('[uploadToMinio] ✅ Upload URL is valid');
    } catch {
        console.error('[uploadToMinio] ❌ Malformed upload URL:', uploadUrl);
        throw new Error('Malformed presigned upload URL');
    }

    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();

        // Track upload progress
        xhr.upload.addEventListener('progress', (event) => {
            if (event.lengthComputable && onProgress) {
                const percentComplete = Math.round((event.loaded / event.total) * 100);
                console.log(`[uploadToMinio] Progress: ${percentComplete}% (${event.loaded}/${event.total} bytes)`);
                onProgress(percentComplete);
            }
        });

        // Handle completion
        xhr.addEventListener('load', () => {
            console.log('[uploadToMinio] Upload completed with status:', xhr.status);
            console.log('[uploadToMinio] Response headers:', xhr.getAllResponseHeaders());
            console.log('[uploadToMinio] Response text:', xhr.responseText);

            if (xhr.status >= 200 && xhr.status < 300) {
                console.log('[uploadToMinio] ✅ Upload successful');
                resolve();
            } else {
                console.error('[uploadToMinio] ❌ Upload failed with status:', xhr.status);
                console.error('[uploadToMinio] Error response:', xhr.responseText);
                reject(new Error(`Upload failed with status ${xhr.status}: ${xhr.responseText}`));
            }
        });

        // Handle errors
        xhr.addEventListener('error', (e) => {
            console.error('[uploadToMinio] ❌ Upload error occurred:', e);
            console.error('[uploadToMinio] XHR state:', {
                readyState: xhr.readyState,
                status: xhr.status,
                statusText: xhr.statusText
            });
            reject(new Error('Upload failed - Network error or CORS issue'));
        });

        xhr.addEventListener('abort', () => {
            console.error('[uploadToMinio] ❌ Upload aborted');
            reject(new Error('Upload aborted'));
        });

        xhr.addEventListener('timeout', () => {
            console.error('[uploadToMinio] ❌ Upload timeout');
            reject(new Error('Upload timeout'));
        });

        // Set timeout (5 minutes)
        xhr.timeout = 5 * 60 * 1000;

        // Open connection
        console.log('[uploadToMinio] Opening PUT request...');
        xhr.open('PUT', uploadUrl);

        // Set headers
        console.log('[uploadToMinio] Setting Content-Type:', file.type);
        xhr.setRequestHeader('Content-Type', file.type);

        // Send file
        console.log('[uploadToMinio] Sending file...');
        xhr.send(file);
    });
}

/**
 * Step 3: Confirm upload and save metadata to PostgreSQL
 * 
 * @param data - Confirmation data with object_key, pod_id, title
 * @returns Promise with material_id and confirmation message
 */
export async function confirmUpload(
    data: ConfirmUploadRequest
): Promise<Material> {
    console.log('[confirmUpload] Confirming upload with data:', data);

    try {
        const response = await apiClient.post<ConfirmUploadResponse>(
            '/api/v1/materials/confirm',
            data
        );

        console.log('[confirmUpload] Confirmation response:', response.data);

        if (!response.data || !response.data.data || !response.data.data.id) {
            console.error('[confirmUpload] ❌ Invalid confirmation response:', response.data);
            throw new Error('Backend returned invalid confirmation response');
        }

        console.log('[confirmUpload] ✅ Upload confirmed successfully');
        return response.data.data;
    } catch (error) {
        console.error('[confirmUpload] ❌ Failed to confirm upload:', error);
        throw error;
    }
}

// =============================================================================
// HELPER FUNCTION - COMPLETE UPLOAD FLOW
// =============================================================================

/**
 * Complete upload flow: Get URL -> Upload to MinIO -> Confirm
 * 
 * @param file - File to upload
 * @param podId - Pod ID to associate material with
 * @param title - Material title
 * @param description - Optional material description
 * @param onProgress - Optional progress callback (0-100)
 * @returns Promise with Material object
 */
export async function uploadMaterial(
    file: File,
    podId: string,
    title: string,
    description?: string,
    onProgress?: (progress: number) => void
): Promise<Material> {
    console.log('[uploadMaterial] 🚀 Starting complete upload flow:', {
        fileName: file.name,
        fileSize: file.size,
        fileType: file.type,
        podId,
        title,
        description
    });

    try {
        // Step 1: Get presigned URL (0-10% progress)
        console.log('[uploadMaterial] Step 1: Requesting presigned URL...');
        onProgress?.(5);
        const { upload_url, object_key } = await getUploadUrl(
            file.name,
            file.type,
            file.size,
            podId
        );
        console.log('[uploadMaterial] ✅ Step 1 complete - Got presigned URL');
        onProgress?.(10);

        // Step 2: Upload to MinIO (10-90% progress)
        console.log('[uploadMaterial] Step 2: Uploading to MinIO...');
        await uploadToMinio(upload_url, file, (uploadProgress) => {
            // Map upload progress from 10% to 90%
            const mappedProgress = 10 + (uploadProgress * 0.8);
            onProgress?.(Math.round(mappedProgress));
        });
        console.log('[uploadMaterial] ✅ Step 2 complete - File uploaded to MinIO');

        // Step 3: Confirm upload (90-100% progress)
        console.log('[uploadMaterial] Step 3: Confirming upload...');
        onProgress?.(95);
        const material = await confirmUpload({
            object_key,
            pod_id: podId,
            title,
            description
        });
        console.log('[uploadMaterial] ✅ Step 3 complete - Upload confirmed');
        onProgress?.(100);

        console.log('[uploadMaterial] 🎉 Complete upload flow finished! Material ID:', material.id);
        return material;
    } catch (error) {
        console.error('[uploadMaterial] ❌ Upload flow failed:', error);
        if (error instanceof Error) {
            console.error('[uploadMaterial] Error details:', {
                message: error.message,
                stack: error.stack
            });
        }
        throw error;
    }
}
