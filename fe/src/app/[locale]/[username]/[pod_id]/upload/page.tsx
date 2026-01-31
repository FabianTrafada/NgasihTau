// fe/src/app/[locale]/[username]/[pod_id]/upload/page.tsx

"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { ChevronLeft, Loader, UploadCloud } from "lucide-react";
import { ProtectedRoute } from "@/components/auth";
import FileUploader from "@/components/dashboard/assets/FileUploader";
import { uploadMaterial } from "@/lib/api/uploadMaterial";
import { getPodDetail } from "@/lib/api/pod";
import { Pod } from "@/types/pod";

interface PageProps {
    params: Promise<{
        locale: string;
        username: string;
        pod_id: string;
    }>;
}

export default function UploadMaterialPage({ params }: PageProps) {
    const router = useRouter();
    const { locale, username, pod_id } = React.use(params);

    // States
    const [pod, setPod] = useState<Pod | null>(null);
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [title, setTitle] = useState("");
    const [description] = useState(""); // removed unused setter
    const [isUploading, setIsUploading] = useState(false);
    const [uploadProgress, setUploadProgress] = useState(0);
    const [error, setError] = useState<string | null>(null);
    const [loadingPod, setLoadingPod] = useState(true);

    // Fetch pod detail
    useEffect(() => {
        const fetchPod = async () => {
            try {
                setLoadingPod(true);
                const podData = await getPodDetail(pod_id);
                setPod(podData);
            } catch (err) {
                console.error("Failed to load pod:", err);
                setError("Failed to load pod details");
            } finally {
                setLoadingPod(false);
            }
        };

        fetchPod();
    }, [pod_id]);

    // Handle file selection
    const handleFileSelect = (file: File | null) => {
        if (file) {
            setSelectedFile(file);
            setError(null);
            // Auto-fill title from filename
            const fileName = file.name;
            const fileNameWithoutExt = fileName.substring(
                0,
                fileName.lastIndexOf(".")
            );
            setTitle(fileNameWithoutExt);
        } else {
            setSelectedFile(null);
        }
    };

    // Handle upload
    const handleUpload = async () => {
        if (!selectedFile || !title.trim()) {
            setError("Please select a file and enter a title");
            return;
        }

        try {
            setIsUploading(true);
            setError(null);
            setUploadProgress(0);

            const material = await uploadMaterial(
                selectedFile,
                pod_id,
                title,
                description,
                setUploadProgress
            );

            // Navigate to material detail page
            router.push(`/${locale}/${username}/${pod_id}/${material.id}`);
        } catch (err) {
            const errorMessage =
                err instanceof Error ? err.message : "Upload failed";
            setError(errorMessage);
            console.error("Upload error:", err);
            setIsUploading(false);
        }
    };

    // Handle back navigation
    const handleBack = () => {
        if (!isUploading) {
            router.push(`/${locale}/${username}/${pod_id}`);
        }
    };

    if (loadingPod) {
        return (
            <ProtectedRoute>
                <div className="flex h-screen items-center justify-center">
                    <Loader className="animate-spin text-orange-500" size={32} />
                </div>
            </ProtectedRoute>
        );
    }

    return (
        <ProtectedRoute>
            <div className="min-h-screen bg-[#FFFBF7]">
                {/* Header */}
                <header className="fixed top-0 w-full border-b-2 border-black bg-white shadow-sm z-50">
                    <div className="w-full px-4 py-3 md:px-6">
                        <div className="flex items-center gap-4">
                            <button
                                onClick={handleBack}
                                disabled={isUploading}
                                className="flex h-10 w-10 items-center justify-center rounded-lg border-2 border-black bg-white hover:bg-gray-50 transition-colors disabled:opacity-50"
                            >
                                <ChevronLeft size={20} />
                            </button>
                            <div className="flex-1">
                                <h1 className="text-xl font-black text-black md:text-2xl leading-none">
                                    Upload Material
                                </h1>
                                {pod && (
                                    <p className="text-sm text-gray-500 mt-0.5">
                                        to <span className="font-bold text-black">{pod.name}</span>
                                    </p>
                                )}
                            </div>
                        </div>
                    </div>
                </header>

                {/* Main Content */}
                <main className="mx-auto max-w-4xl px-4 py-6 md:px-6 md:py-8 flex items-center justify-center min-h-[calc(100vh-64px)]">
                    <div className="space-y-6 w-full">
                        {/* File Uploader */}
                        <div className="rounded-xl border-2 border-black bg-white p-6 shadow-[4px_4px_0_0_black]">
                            <label className="block text-sm font-bold text-black mb-4">
                                Select File <span className="text-red-500">*</span>
                            </label>
                            <FileUploader
                                onSingleFileSelect={handleFileSelect}
                                selectedFile={selectedFile}
                                multiple={false}
                                label="Click or drag"
                                description="Supported format: PDF, DOC, DOCX, PPT, PPTX (100MB max)"
                            />
                        </div>


                        {/* Progress Bar */}
                        {isUploading && (
                            <div className="rounded-xl border-2 border-black bg-white p-6 shadow-[4px_4px_0_0_black]">
                                <div className="space-y-3">
                                    <div className="flex items-center justify-between">
                                        <div className="flex items-center gap-2">
                                            <Loader className="animate-spin text-orange-500" size={20} />
                                            <span className="text-sm font-bold text-black">
                                                Uploading material...
                                            </span>
                                        </div>
                                        <span className="text-sm font-bold text-orange-500">
                                            {uploadProgress}%
                                        </span>
                                    </div>
                                    <div className="w-full h-3 border-2 border-black rounded-full overflow-hidden bg-gray-100">
                                        <div
                                            className="h-full bg-orange-500 transition-all duration-300"
                                            style={{ width: `${uploadProgress}%` }}
                                        />
                                    </div>
                                    <p className="text-xs text-gray-500 text-center">
                                        Please don't close this page while uploading
                                    </p>
                                </div>
                            </div>
                        )}

                        {/* Error Message */}
                        {error && (
                            <div className="rounded-xl border-2 border-red-500 bg-red-50 p-4 shadow-[4px_4px_0_0_red]">
                                <p className="text-sm font-bold text-red-600">{error}</p>
                            </div>
                        )}

                        {/* Action Buttons */}
                        <div className="flex flex-col-reverse sm:flex-row gap-3">
                            <button
                                onClick={handleBack}
                                disabled={isUploading}
                                className="flex-1 px-6 py-3 border-2 border-black rounded-lg font-bold bg-white hover:bg-gray-50 transition-colors disabled:opacity-50 text-sm md:text-base"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleUpload}
                                disabled={isUploading || !selectedFile || !title.trim()}
                                className="flex-1 sm:flex-2 px-6 py-3 border-2 border-black rounded-lg font-bold bg-[#FF8811] text-white shadow-[4px_4px_0_0_black] hover:shadow-[2px_2px_0_0_black] hover:translate-x-[2px] hover:translate-y-[2px] transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 text-sm md:text-base"
                            >
                                {isUploading ? (
                                    <>
                                        <Loader size={18} className="animate-spin" />
                                        Uploading...
                                    </>
                                ) : (
                                    <>
                                        <UploadCloud size={18} />
                                        Upload Material
                                    </>
                                )}
                            </button>
                        </div>
                    </div>
                </main>
            </div>
        </ProtectedRoute>
    );
}