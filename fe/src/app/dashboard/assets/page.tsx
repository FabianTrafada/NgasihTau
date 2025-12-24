"use client";

import React from "react";
import { ProtectedRoute } from "@/components/auth";
import FileUploader from "@/components/dashboard/assets/FileUploader";
import UploadQueueItem from "@/components/dashboard/assets/UploadQueueItem";

export default function AssetsPage() {
    return (
        <ProtectedRoute>
            <AssetsContent />
        </ProtectedRoute>
    );
}

function AssetsContent() {
    // Mock data matching the design
    const queueItems = [
        {
            id: 1,
            fileName: "Vidio Hot 2 Jam Jule.mp4",
            fileSize: "34.4mb",
            progress: 38,
            type: "video" as const
        },
        {
            id: 2,
            fileName: "Vidio SBY Garuk Biji.mp4",
            fileSize: "3.4mb",
            progress: 93,
            type: "video" as const
        },
        {
            id: 3,
            fileName: "Sejarah MBG.pdf",
            fileSize: "10mb",
            progress: 75,
            type: "pdf" as const
        }
    ];

    return (
        <div className="flex-1 p-4 sm:p-6 lg:p-8 max-w-5xl mx-auto flex flex-col gap-8">
            {/* Header */}
            <div>
                <h1 className="text-2xl sm:text-3xl font-bold text-[#2B2D42] mb-1">
                    Update Assets File
                </h1>
                <p className="text-sm sm:text-base text-gray-500">
                    Here you can manage your files.
                </p>
            </div>

            {/* Upload Section */}
            <section className="bg-white rounded-3xl p-1 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] border-2 border-[#2B2D42]">
                <div className="bg-[#FFFBF7] rounded-[20px] p-6 sm:p-8 border border-[#2B2D42]/10">
                    <FileUploader />
                </div>
            </section>

            {/* Queue Section */}
            <section>
                <h2 className="text-lg font-bold text-[#2B2D42] mb-4">
                    Upload Queue
                </h2>
                <div className="space-y-4">
                    {queueItems.map((item) => (
                        <UploadQueueItem
                            key={item.id}
                            fileName={item.fileName}
                            fileSize={item.fileSize}
                            progress={item.progress}
                            type={item.type}
                            onDelete={() => console.log(`Delete ${item.id}`)}
                        />
                    ))}
                </div>
            </section>
        </div>
    );
}
