'use client';

import React, { useState } from 'react';
import Dropzone, { FileRejection } from 'react-dropzone';
import { UploadCloud, AlertCircle } from 'lucide-react';
import { cn } from '@/lib/utils';

interface FileUploaderProps {
    onFileSelect?: (files: File[]) => void;
}

const MAX_SIZE = 50 * 1024 * 1024; // 50MB
const ACCEPTED_TYPES = {
    'application/pdf': ['.pdf'],
    'application/msword': ['.doc'],
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx'],
    'application/vnd.ms-powerpoint': ['.ppt'],
    'application/vnd.openxmlformats-officedocument.presentationml.presentation': ['.pptx']
};

const FileUploader: React.FC<FileUploaderProps> = ({ onFileSelect }) => {
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
                setError(`File is too large. Max size is 50MB.`);
            } else if (rejection.errors[0].code === 'file-invalid-type') {
                setError(`Invalid file type. Only SVG, PNG, JPG, PDF, DOCS, PPT are allowed.`);
            } else {
                setError(rejection.errors[0].message);
            }
            return;
        }

        if (onFileSelect && acceptedFiles.length > 0) {
            console.log('[FileUploader] ✅ Calling onFileSelect with files:',
                acceptedFiles.map(f => ({ name: f.name, size: f.size }))
            );
            onFileSelect(acceptedFiles);
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
            multiple
            maxSize={MAX_SIZE}
            accept={ACCEPTED_TYPES}
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
                                : "border-[#2B2D42] hover:border-[#FF8811]"
                    )}
                >
                    <input {...getInputProps()} />

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
                                <span className="text-[#FF8811]">Click here</span> to upload your file.
                            </p>
                        )}
                        <p className="text-gray-400 text-sm font-medium">
                            Supported format: SVG, PNG, JPG, PDF, DOCS, PPT (50mb each)
                        </p>
                    </div>
                </div>
            )}
        </Dropzone>
    );
};

export default FileUploader;
