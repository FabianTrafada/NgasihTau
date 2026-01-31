
"use client";

import { useDownloads } from "@/hooks/useDownloads";
import { ProtectedRoute } from "@/components/auth";
import Link from "next/link";
import { Download, FileText, Trash2 } from "lucide-react";

export default function DownloadedMaterialsPage() {
    const { downloads, removeMaterial } = useDownloads();

    return (
        <ProtectedRoute>
            <div className="p-4 sm:p-6 lg:p-8">
                <h1 className="text-2xl font-bold text-[#2B2D42] mb-6">Downloaded Materials</h1>

                {downloads.length === 0 ? (
                    <div className="flex flex-col items-center justify-center p-12 bg-white border-2 border-[#2B2D42] rounded-lg shadow-[4px_4px_0px_0px_#2B2D42]">
                        <Download size={48} className="text-[#FF8811] mb-4" />
                        <p className="text-lg font-bold text-[#2B2D42]">No materials downloaded yet</p>
                        <p className="text-sm text-gray-500 mb-6">Materials you download will appear here.</p>
                        <Link
                            href="/dashboard/pod"
                            className="px-6 py-2 bg-[#FF8811] text-white rounded-lg font-bold hover:bg-[#e67a0f] transition-all shadow-[2px_2px_0px_0px_#2B2D42]"
                        >
                            Browse Materials
                        </Link>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {downloads.map((material) => (
                            <div
                                key={material.id}
                                className="bg-white border-2 border-[#2B2D42] p-4 rounded-lg shadow-[4px_4px_0px_0px_#2B2D42] hover:translate-x-0.5 hover:translate-y-0.5 hover:shadow-[2px_2px_0px_0px_#2B2D42] transition-all"
                            >
                                <div className="flex items-start justify-between mb-2">
                                    <div className="flex items-center gap-2">
                                        <div className="p-2 bg-gray-100 rounded-lg border border-[#2B2D42]">
                                            <FileText size={20} className="text-[#FF8811]" />
                                        </div>
                                        <div>
                                            <h3 className="font-bold text-[#2B2D42] line-clamp-1" title={material.title}>
                                                {material.title}
                                            </h3>
                                            <p className="text-xs text-gray-500">{material.file_type.toUpperCase()} • {(material.file_size / 1024 / 1024).toFixed(2)} MB</p>
                                        </div>
                                    </div>
                                </div>

                                <div className="flex items-center justify-between mt-4">
                                    <Link
                                        href={`/dashboard/${material.uploader_id}/${material.pod_id}/${material.id}`}
                                        className="text-xs font-bold text-[#FF8811] hover:underline"
                                    >
                                        View Details
                                    </Link>
                                    <button
                                        onClick={() => removeMaterial(material.id)}
                                        className="p-2 text-red-500 hover:bg-red-50 rounded-lg transition-colors"
                                        title="Remove from downloads"
                                    >
                                        <Trash2 size={16} />
                                    </button>
                                </div>
                                <div className="mt-2 text-[10px] text-gray-400 text-right">
                                    Downloaded {new Date(material.downloadedAt).toLocaleDateString()}
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </ProtectedRoute>
    );
}
