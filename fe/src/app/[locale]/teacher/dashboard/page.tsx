"use client";

import React, { useState } from "react";
import FileUploader from "@/components/dashboard/assets/FileUploader";
import UploadQueueItem from "@/components/dashboard/assets/UploadQueueItem";
import { useFileUpload } from "@/hooks/useFileUpload";
import { useTranslations } from "next-intl";

export default function TeacherDashboardPage() {
    const t = useTranslations('teacherDashboard');
    const { uploads, addFiles, removeUpload, startUpload, retryUpload, clearCompleted } = useFileUpload();
    // TODO: Get real pod UUID from pod selection or context
    const [podId] = useState("550e8400-e29b-41d4-a716-446655440000"); // Example UUID - replace with real pod ID

    const handleFileSelect = (files: File[]) => {


        // addFiles returns array of upload IDs immediately
        const uploadIds = addFiles(files);
        console.log('[TeacherDashboard] Files added with IDs:', uploadIds);

        // Start uploads immediately (no setTimeout needed - using ref now)
        console.log('[TeacherDashboard] Starting uploads with podId:', podId);

        uploadIds.forEach((uploadId, index) => {
            const file = files[index];
            console.log('[TeacherDashboard] Triggering upload:', {
                uploadId,
                fileName: file.name
            });

            // Start upload for each file
            startUpload(uploadId, {
                podId,
                title: file.name
            });
        });

        console.log('[TeacherDashboard] ========================================');
    };

    const handleRetry = (uploadId: string) => {
        const upload = uploads.find(u => u.id === uploadId);
        console.log('[TeacherDashboard] Retrying upload:', {
            uploadId,
            fileName: upload?.fileName
        });

        retryUpload(uploadId, {
            podId,
            title: upload?.fileName || 'Untitled'
        });
    };

    return (
        <div className="flex-1 p-4 sm:p-6 lg:p-8 max-w-5xl mx-auto flex flex-col gap-8">
            {/* Header */}
            <div>
                <h1 className="text-2xl sm:text-3xl font-bold text-[#2B2D42] mb-1">
                    {t('title')}
                </h1>
                <p className="text-sm sm:text-base text-gray-500">
                    {t('subtitle')}
                </p>
            </div>

            {/* Upload Section */}
            <section className="bg-white rounded-3xl p-1 shadow-[8px_8px_0px_0px_#FF8811] border-2 border-[#2B2D42]">
                <div className="bg-[#FFFBF7] rounded-[20px] p-6 sm:p-8 border border-[#2B2D42]/10">
                    <FileUploader onFileSelect={handleFileSelect} />
                </div>
            </section>

            {/* Queue Section */}
            <section>
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-lg font-bold text-[#2B2D42]">
                        {t('uploadQueue')} ({uploads.length})
                    </h2>
                    {uploads.some(u => u.status === 'success' || u.status === 'error') && (
                        <button
                            onClick={clearCompleted}
                            className="text-sm font-bold text-gray-400 hover:text-[#FF8811] transition-colors"
                        >
                            {t('clearCompleted')}
                        </button>
                    )}
                </div>
                <div className="space-y-4">
                    {uploads.length === 0 ? (
                        <div className="text-center py-8 text-gray-400">
                            <p>{t('noFiles')}</p>
                        </div>
                    ) : (
                        uploads.map((upload) => (
                            <UploadQueueItem
                                key={upload.id}
                                fileName={upload.fileName}
                                fileSize={upload.fileSize}
                                progress={upload.progress}
                                type={upload.type}
                                status={upload.status}
                                error={upload.error}
                                onDelete={() => removeUpload(upload.id)}
                                onRetry={() => handleRetry(upload.id)}
                            />
                        ))
                    )}
                </div>
            </section>
        </div>
    );
}
