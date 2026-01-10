"use client";

import React, { useState } from "react";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
} from "@/components/ui/dialog";
import FileUploader from "@/components/dashboard/assets/FileUploader";
import { uploadMaterial } from "@/lib/api/uploadMaterial";
import { Loader } from "lucide-react";

interface UploadMaterialModalProps {
    isOpen: boolean;
    onClose: () => void;
    podId: string;
    onSuccess?: (materialId: string) => void;
}

export default function UploadMaterialModal({
    isOpen,
    onClose,
    podId,
    onSuccess,
}: UploadMaterialModalProps) {
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [title, setTitle] = useState("");
    const [description, setDescription] = useState("");
    const [isUploading, setIsUploading] = useState(false);
    const [uploadProgress, setUploadProgress] = useState(0);
    const [error, setError] = useState<string | null>(null);

    // Handler untuk onSingleFileSelect - expects (file: File | null) => void
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
                podId,
                title,
                description,
                setUploadProgress
            );

            // Reset form
            setSelectedFile(null);
            setTitle("");
            setDescription("");
            setUploadProgress(0);

            onSuccess?.(material.id);
            onClose();
        } catch (err) {
            const errorMessage =
                err instanceof Error ? err.message : "Upload failed";
            setError(errorMessage);
            console.error("Upload error:", err);
        } finally {
            setIsUploading(false);
        }
    };

    const handleClose = () => {
        if (!isUploading) {
            setSelectedFile(null);
            setTitle("");
            setDescription("");
            setError(null);
            setUploadProgress(0);
            onClose();
        }
    };

    return (
        <Dialog open={isOpen} onOpenChange={handleClose}>
            <DialogContent className="sm:max-w-2xl border-2 border-black bg-[#FFFBF7] shadow-[6px_6px_0_0_black]">
                <DialogHeader>
                    <DialogTitle className="text-2xl font-black">
                        Upload Material
                    </DialogTitle>
                </DialogHeader>

                <div className="space-y-6">
                    {/* File Uploader */}
                    <div>
                        <label className="block text-sm font-bold text-black mb-3">
                            Select File
                        </label>
                        <FileUploader
                            onSingleFileSelect={handleFileSelect}
                            selectedFile={selectedFile}
                            multiple={false}
                            label="Click or drag"
                            description="Supported format: PDF, DOC, DOCX, PPT, PPTX (50mb max)"
                        />
                    </div>

                    {/* Title Input */}
                    <div>
                        <label className="block text-sm font-bold text-black mb-2">
                            Title
                        </label>
                        <input
                            type="text"
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            className="w-full px-4 py-2 border-2 border-black rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                            placeholder="Material title"
                            disabled={isUploading}
                        />
                    </div>

                    {/* Description Input */}
                    <div>
                        <label className="block text-sm font-bold text-black mb-2">
                            Description (Optional)
                        </label>
                        <textarea
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            className="w-full px-4 py-2 border-2 border-black rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500 resize-none"
                            rows={4}
                            placeholder="Material description..."
                            disabled={isUploading}
                        />
                    </div>

                    {/* Progress Bar */}
                    {isUploading && (
                        <div className="space-y-2">
                            <div className="flex items-center justify-between">
                                <span className="text-sm font-bold text-black">
                                    Uploading...
                                </span>
                                <span className="text-sm font-bold text-orange-500">
                                    {uploadProgress}%
                                </span>
                            </div>
                            <div className="w-full h-2 border-2 border-black rounded-full overflow-hidden bg-gray-200">
                                <div
                                    className="h-full bg-orange-500 transition-all duration-300"
                                    style={{ width: `${uploadProgress}%` }}
                                />
                            </div>
                        </div>
                    )}

                    {/* Error Message */}
                    {error && (
                        <div className="p-3 rounded-lg border-2 border-red-500 bg-red-50">
                            <p className="text-sm font-bold text-red-600">{error}</p>
                        </div>
                    )}
                </div>

                <DialogFooter className="gap-2">
                    <button
                        onClick={handleClose}
                        disabled={isUploading}
                        className="px-6 py-2 border-2 border-black rounded-lg font-bold bg-white hover:bg-gray-50 transition-colors disabled:opacity-50"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleUpload}
                        disabled={isUploading || !selectedFile || !title.trim()}
                        className="px-6 py-2 border-2 border-black rounded-lg font-bold bg-[#FF8811] text-white shadow-[2px_2px_0_0_black] hover:shadow-[1px_1px_0_0_black] hover:translate-x-[1px] hover:translate-y-[1px] transition-all disabled:opacity-50 flex items-center gap-2"
                    >
                        {isUploading && <Loader size={16} className="animate-spin" />}
                        {isUploading ? "Uploading..." : "Upload Material"}
                    </button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}