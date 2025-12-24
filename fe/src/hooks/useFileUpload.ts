import { useState, useCallback, useRef } from 'react';
import { uploadMaterial } from '@/lib/api/uploadMaterial';

// =============================================================================
// TYPES
// =============================================================================

export type UploadStatus = 'idle' | 'uploading' | 'success' | 'error';

export interface UploadItem {
    id: string;
    file: File;
    fileName: string;
    fileSize: string;
    progress: number;
    status: UploadStatus;
    error?: string;
    materialId?: number;
    type: 'video' | 'pdf' | 'image' | 'doc' | 'other';
}

export interface UploadOptions {
    podId: string;
    title?: string;
    description?: string;
}

// =============================================================================
// HOOK
// =============================================================================

/**
 * Custom hook for managing file uploads to MinIO via backend
 * 
 * Features:
 * - Multiple file uploads with individual progress tracking
 * - Automatic retry on failure
 * - Queue management (add, remove, clear)
 * - Upload status tracking per file
 * 
 * @example
 * ```tsx
 * const { uploads, addFiles, removeUpload, startUpload } = useFileUpload();
 * 
 * const handleFileSelect = (files: File[]) => {
 *   const uploadIds = addFiles(files);
 *   uploadIds.forEach((id, i) => startUpload(id, files[i], { podId: 1 }));
 * };
 * ```
 */
export function useFileUpload() {
    const [uploads, setUploads] = useState<UploadItem[]>([]);

    // Use ref to track uploads for avoiding stale closure issues
    const uploadsRef = useRef<Map<string, UploadItem>>(new Map());

    /**
     * Add files to upload queue and return their IDs
     */
    const addFiles = useCallback((files: File[]): string[] => {
        console.log('[useFileUpload] Adding files to queue:', files.length);

        const newUploads: UploadItem[] = files.map((file) => {
            console.log('[useFileUpload] Processing file:', {
                name: file.name,
                size: file.size,
                type: file.type
            });

            // Determine file type from MIME type
            const type = file.type.includes('video') ? 'video'
                : file.type.includes('pdf') ? 'pdf'
                    : file.type.includes('image') ? 'image'
                        : file.type.includes('document') || file.type.includes('word') || file.type.includes('presentation') ? 'doc'
                            : 'other';

            const uploadItem: UploadItem = {
                id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
                file,
                fileName: file.name,
                fileSize: formatFileSize(file.size),
                progress: 0,
                status: 'idle',
                type
            };

            // Store in ref immediately (no async delay)
            uploadsRef.current.set(uploadItem.id, uploadItem);
            console.log('[useFileUpload] Stored in ref:', uploadItem.id);

            return uploadItem;
        });

        // Update state for UI
        setUploads(prev => [...newUploads, ...prev]);
        console.log('[useFileUpload] Total uploads in ref:', uploadsRef.current.size);

        return newUploads.map(u => u.id);
    }, []);

    /**
     * Start upload for a specific file
     * Uses ref to get file data (avoids stale closure)
     */
    const startUpload = useCallback(async (
        uploadId: string,
        options: UploadOptions
    ) => {
        console.log('[useFileUpload] ========== START UPLOAD ==========');
        console.log('[useFileUpload] Upload ID:', uploadId);
        console.log('[useFileUpload] Options:', options);
        console.log('[useFileUpload] Ref size:', uploadsRef.current.size);

        // Get upload from ref (always current)
        const upload = uploadsRef.current.get(uploadId);

        if (!upload) {
            console.error('[useFileUpload] ❌ Upload NOT FOUND in ref!');
            console.error('[useFileUpload] Available IDs:', Array.from(uploadsRef.current.keys()));
            return;
        }

        console.log('[useFileUpload] ✅ Found upload in ref:', {
            id: upload.id,
            fileName: upload.fileName,
            fileSize: upload.fileSize,
            fileExists: !!upload.file
        });

        // Update status to uploading
        setUploads(prev => prev.map(u =>
            u.id === uploadId
                ? { ...u, status: 'uploading' as UploadStatus, progress: 0, error: undefined }
                : u
        ));

        try {
            console.log('[useFileUpload] 🚀 Calling uploadMaterial API...');

            const materialId = await uploadMaterial(
                upload.file,
                options.podId,
                options.title || upload.fileName,
                options.description,
                (progress) => {
                    console.log(`[useFileUpload] Progress: ${progress}%`);
                    setUploads(prev => prev.map(u =>
                        u.id === uploadId
                            ? { ...u, progress }
                            : u
                    ));
                }
            );

            console.log('[useFileUpload] ✅ Upload successful! Material ID:', materialId);

            // Update ref
            const updatedItem = { ...upload, status: 'success' as UploadStatus, progress: 100, materialId };
            uploadsRef.current.set(uploadId, updatedItem);

            // Update state
            setUploads(prev => prev.map(u =>
                u.id === uploadId
                    ? updatedItem
                    : u
            ));
        } catch (error) {
            console.error('[useFileUpload] ❌ Upload failed:', error);
            const errorMessage = error instanceof Error
                ? error.message
                : 'Upload failed';
            console.error('[useFileUpload] Error message:', errorMessage);

            // Update ref
            const updatedItem = { ...upload, status: 'error' as UploadStatus, error: errorMessage };
            uploadsRef.current.set(uploadId, updatedItem);

            // Update state
            setUploads(prev => prev.map(u =>
                u.id === uploadId
                    ? updatedItem
                    : u
            ));
        }

        console.log('[useFileUpload] ========== END UPLOAD ==========');
    }, []);

    /**
     * Start all pending uploads
     */
    const startAllUploads = useCallback(async (options: UploadOptions) => {
        const pendingUploads = Array.from(uploadsRef.current.values())
            .filter(u => u.status === 'idle');

        console.log('[useFileUpload] Starting all pending uploads:', pendingUploads.length);

        // Upload files sequentially to avoid overwhelming the server
        for (const upload of pendingUploads) {
            await startUpload(upload.id, options);
        }
    }, [startUpload]);

    /**
     * Remove an upload from queue
     */
    const removeUpload = useCallback((uploadId: string) => {
        uploadsRef.current.delete(uploadId);
        setUploads(prev => prev.filter(u => u.id !== uploadId));
    }, []);

    /**
     * Clear all uploads
     */
    const clearAll = useCallback(() => {
        uploadsRef.current.clear();
        setUploads([]);
    }, []);

    /**
     * Clear completed/failed uploads
     */
    const clearCompleted = useCallback(() => {
        const toRemove = Array.from(uploadsRef.current.values())
            .filter(u => u.status === 'success' || u.status === 'error')
            .map(u => u.id);

        toRemove.forEach(id => uploadsRef.current.delete(id));

        setUploads(prev => prev.filter(u =>
            u.status !== 'success' && u.status !== 'error'
        ));
    }, []);

    /**
     * Retry failed upload
     */
    const retryUpload = useCallback((uploadId: string, options: UploadOptions) => {
        const upload = uploadsRef.current.get(uploadId);
        if (upload) {
            const resetItem = { ...upload, status: 'idle' as UploadStatus, progress: 0, error: undefined };
            uploadsRef.current.set(uploadId, resetItem);
        }

        setUploads(prev => prev.map(u =>
            u.id === uploadId
                ? { ...u, status: 'idle' as UploadStatus, progress: 0, error: undefined }
                : u
        ));

        // Start upload after state update
        setTimeout(() => startUpload(uploadId, options), 0);
    }, [startUpload]);

    return {
        uploads,
        addFiles,
        startUpload,
        startAllUploads,
        removeUpload,
        clearAll,
        clearCompleted,
        retryUpload
    };
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

/**
 * Format file size to human-readable string
 */
function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';

    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}
