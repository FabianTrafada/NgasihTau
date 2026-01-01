'use client';

import React, { useState } from 'react';
import Dropzone, { FileRejection, Accept } from 'react-dropzone';
import { UploadCloud, AlertCircle, Check } from 'lucide-react';
import { cn } from '@/lib/utils';

interface FileUploaderProps {
    onFileSelect?: (files: File[]) => void;
    onSingleFileSelect?: (file: File | null) => void;
    accept?: Accept;
    maxSize?: number;
    multiple?: boolean;
    label?: string;
    description?: string;
    selectedFile?: File | null;
    className?: string;
}

const DEFAULT_MAX_SIZE = 50 * 1024 * 1024; // 50MB
const DEFAULT_ACCEPTED_TYPES: Accept = {
    'application/pdf': ['.pdf'],
    'application/msword': ['.doc'],
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx'],
    'application/vnd.ms-powerpoint': ['.ppt'],
    'application/vnd.openxmlformats-officedocument.presentationml.presentation': ['.pptx']
};

const FileUploader: React.FC<FileUploaderProps> = ({
    onFileSelect,
    onSingleFileSelect,
    accept = DEFAULT_ACCEPTED_TYPES,
    maxSize = DEFAULT_MAX_SIZE,
    multiple = true,
    label = "Click here",
    description = "Supported format: SVG, PNG, JPG, PDF, DOCS, PPT (50mb each)",
    selectedFile,
    className
}) => {
    const [error, setError] = useState<string | null>(null);

    const handleDrop = (acceptedFiles: File[], fileRejections: FileRejection[]) => {
        console.log('[FileUploader] ==========================================');
        console.log('[FileUploader] handleDrop triggered!');
        console.log('[FileUploader] Accepted files:', acceptedFiles.length);
        console.log('[FileUploader] Rejected files:', fileRejections.length);

        setError(null);

        if (fileRejections.length > 0) {
            const rejection = fileRejections[0];
            console.log('[FileUploader] File rejected:', rejection.errors);
            if (rejection.errors[0].code === 'file-too-large') {
                setError(`File is too large. Max size is ${Math.round(maxSize / 1024 / 1024)}MB.`);
            } else if (rejection.errors[0].code === 'file-invalid-type') {
                setError(`Invalid file type.`);
            } else {
                setError(rejection.errors[0].message);
            }
            return;
        }

        if (acceptedFiles.length > 0) {
            if (onFileSelect) {
                console.log('[FileUploader] ✅ Calling onFileSelect with files:',
                    acceptedFiles.map(f => ({ name: f.name, size: f.size }))
                );
                onFileSelect(acceptedFiles);
            }
            if (onSingleFileSelect) {
                console.log('[FileUploader] ✅ Calling onSingleFileSelect with file:', acceptedFiles[0].name);
                onSingleFileSelect(acceptedFiles[0]);
            }
        } else {
            console.warn('[FileUploader] ⚠️ onFileSelect not called:', {
                hasCallback: !!onFileSelect,
                filesCount: acceptedFiles.length
            });
        }

        console.log('[FileUploader] ==========================================');
    };

    return (
        <Dropzone
            onDrop={handleDrop}
            multiple={multiple}
            maxSize={maxSize}
            accept={accept}
        >
            {({ getRootProps, getInputProps, isDragActive }) => (
                <div
                    {...getRootProps()}
                    className={cn(
                        "w-full h-64 rounded-2xl border-2 border-dashed transition-all duration-200 cursor-pointer flex flex-col items-center justify-center gap-4 bg-[#FFFBF7]",
                        isDragActive
                            ? "border-[#FF8811] bg-[#FFFBF7]/80 scale-[1.01]"
                            : error
                                ? "border-red-500 bg-red-50"
                                : "border-[#2B2D42] hover:border-[#FF8811]",
                        className
                    )}
                >
                    <input {...getInputProps()} />

                    {selectedFile ? (
                        <div className="flex flex-col items-center justify-center gap-4">
                            <div className="w-16 h-16 rounded-full border-2 border-green-200 bg-green-100 flex items-center justify-center shadow-sm">
                                <Check className="w-8 h-8 text-green-600" />
                            </div>
                            <div className="text-center space-y-1 px-4">
                                <p className="text-[#2B2D42] font-bold text-lg">
                                    {selectedFile.name}
                                </p>
                                <p className="text-gray-500 text-sm">
                                    {(selectedFile.size / 1024 / 1024).toFixed(2)} MB
                                </p>
                                <p className="text-[#FF8811] text-sm font-bold mt-2">
                                    Click or drag to replace
                                </p>
                            </div>
                        </div>
                    ) : (
                        <>
                            <div className={cn(
                                "w-16 h-16 rounded-full border-2 flex items-center justify-center shadow-sm",
                                error
                                    ? "bg-red-100 border-red-200"
                                    : "bg-[#FFFBF7] border-[#FF8811]/20"
                            )}>
                                {error ? (
                                    <AlertCircle className="w-8 h-8 text-red-500" />
                                ) : (
                                    <UploadCloud className="w-8 h-8 text-[#FF8811]" />
                                )}
                            </div>

                            <div className="text-center space-y-2 px-4">
                                {error ? (
                                    <p className="text-red-500 font-bold text-lg">
                                        {error}
                                    </p>
                                ) : (
                                    <p className="text-[#2B2D42] font-bold text-lg">
                                        <span className="text-[#FF8811]">{label}</span> to upload your file.
                                    </p>
                                )}
                                <p className="text-gray-400 text-sm font-medium">
                                    {description}
                                </p>
                            </div>
                        </>
                    )}
                </div>
            )}
        </Dropzone>
    );
};

export default FileUploader;
