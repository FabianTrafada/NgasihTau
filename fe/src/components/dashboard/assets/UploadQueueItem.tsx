import React from 'react';
import { FileText, Trash2, FileVideo, File, CheckCircle, AlertCircle, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';

export type UploadStatus = 'idle' | 'uploading' | 'success' | 'error';

interface UploadQueueItemProps {
    fileName: string;
    fileSize: string;
    progress: number;
    type: 'video' | 'pdf' | 'image' | 'doc' | 'other';
    status?: UploadStatus;
    error?: string;
    onDelete?: () => void;
    onRetry?: () => void;
}

const getFileIcon = (type: string) => {
    switch (type) {
        case 'video':
            return <FileVideo className="w-6 h-6 text-[#FF8811]" />;
        case 'pdf':
            return <FileText className="w-6 h-6 text-[#FF8811]" />;
        default:
            return <File className="w-6 h-6 text-[#FF8811]" />;
    }
};

const UploadQueueItem: React.FC<UploadQueueItemProps> = ({
    fileName,
    fileSize,
    progress,
    type,
    status = 'idle',
    error,
    onDelete,
    onRetry
}) => {
    const getStatusColor = () => {
        switch (status) {
            case 'success': return 'border-green-200 bg-green-50';
            case 'error': return 'border-red-200 bg-red-50';
            case 'uploading': return 'border-[#FF8811]/30 bg-[#FFFBF7]';
            default: return 'border-gray-200 bg-white';
        }
    };

    const getProgressColor = () => {
        switch (status) {
            case 'success': return 'bg-green-500';
            case 'error': return 'bg-red-500';
            default: return 'bg-[#FF8811]';
        }
    };

    return (
        <div className={cn(
            "border rounded-xl p-4 flex items-center gap-4 shadow-sm hover:shadow-md transition-all duration-200",
            getStatusColor()
        )}>
            {/* Icon Container */}
            <div className="w-12 h-12 rounded-full bg-[#FFFBF7] flex items-center justify-center flex-shrink-0 border border-gray-100">
                {getFileIcon(type)}
            </div>

            {/* Content */}
            <div className="flex-1 min-w-0">
                <div className="flex justify-between items-center mb-2">
                    <h3 className="font-bold text-[#2B2D42] truncate text-sm sm:text-base">
                        {fileName}
                    </h3>
                    <div className="flex items-center gap-2">
                        {/* Status Icon */}
                        {status === 'uploading' && (
                            <Loader2 className="w-4 h-4 text-[#FF8811] animate-spin" />
                        )}
                        {status === 'success' && (
                            <CheckCircle className="w-4 h-4 text-green-500" />
                        )}
                        {status === 'error' && (
                            <AlertCircle className="w-4 h-4 text-red-500" />
                        )}

                        {/* Action Buttons */}
                        {status === 'error' && onRetry && (
                            <button
                                onClick={onRetry}
                                className="text-xs font-bold text-[#FF8811] hover:text-[#ff6600] transition-colors px-2 py-1 rounded"
                            >
                                Retry
                            </button>
                        )}
                        <button
                            onClick={onDelete}
                            className="text-gray-400 hover:text-red-500 transition-colors"
                            disabled={status === 'uploading'}
                        >
                            <Trash2 className="w-4 h-4" />
                        </button>
                    </div>
                </div>

                {/* Progress Bar */}
                {status !== 'idle' && (
                    <div className="w-full h-2 bg-gray-100 rounded-full overflow-hidden">
                        <div
                            className={cn(
                                "h-full rounded-full transition-all duration-500 ease-out",
                                getProgressColor()
                            )}
                            style={{ width: `${progress}%` }}
                        />
                    </div>
                )}

                {/* Meta Info */}
                <div className="flex justify-between items-center mt-1">
                    <span className="text-xs text-gray-400 font-medium">{fileSize}</span>
                    {status === 'error' && error ? (
                        <span className="text-xs font-bold text-red-500">{error}</span>
                    ) : status === 'success' ? (
                        <span className="text-xs font-bold text-green-500">Complete</span>
                    ) : status === 'uploading' ? (
                        <span className="text-xs font-bold text-[#FF8811]">{progress}%</span>
                    ) : (
                        <span className="text-xs font-bold text-gray-400">Ready</span>
                    )}
                </div>
            </div>
        </div>
    );
};

export default UploadQueueItem;
