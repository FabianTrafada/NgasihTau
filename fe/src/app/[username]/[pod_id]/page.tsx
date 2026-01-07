"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Eye, BadgeCheck, FileText, Loader } from "lucide-react";
import FileListItem from "@/components/knowledge-pod/FileListItem";
import { SearchSection } from "@/components/landing-page/search-section";
import { getPodDetail, getPodMaterials } from "@/lib/api/pod";
import { getUserDetail } from "@/lib/api/user";
import { Pod } from "@/types/pod";
import { Material } from "@/types/material";
import { ProtectedRoute } from "@/components/auth";

interface PageProps {
  params: Promise<{
    username: string;
    pod_id: string;
  }>;
}

export default function KnowledgePodDetail({ params }: PageProps) {
  const router = useRouter();
  const { username, pod_id } = React.use(params);

  // State untuk pod data
  const [pod, setPod] = useState<Pod | null>(null);
  const [materials, setMaterials] = useState<Material[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isNotFound, setIsNotFound] = useState(false);

  // Fetch pod detail dan materials
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);
        setIsNotFound(false);

        console.log("Fetching pod with ID:", pod_id);
        console.log("URL params - username:", username, "pod_id:", pod_id);

        // Fetch pod detail
        const podData = await getPodDetail(pod_id);
        console.log("Pod Detail API Response:", podData);

        // Fetch user detail untuk validasi username
        // const userData = await getUserDetail(podData.owner_id);
        // console.log("User Detail:", userData);

        // Validate: username dari URL harus sesuai dengan user.username
        // if (userData.username !== username) {
        //   console.warn("Username mismatch:", { urlUsername: username, userUsername: userData.username });
        //   setIsNotFound(true);
        //   setLoading(false);
        //   return;
        // }

        setPod(podData);

        // Fetch materials dalam pod
        try {
          const podMaterials = await getPodMaterials(pod_id);
          console.log("Pod Materials:", podMaterials);
          setMaterials(podMaterials);
        } catch (err) {
          console.warn("Failed to fetch pod materials:", err);
          setMaterials([]);
        }
      } catch (err) {
        console.error("Error loading pod:", err);
        setError(err instanceof Error ? err.message : "Failed to load pod");
        setIsNotFound(true);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [pod_id, username]);

  // Show loading state
  if (loading) {
    return (
      <ProtectedRoute>
        <div className="flex h-screen items-center justify-center">
          <Loader className="animate-spin text-orange-500" size={40} />
        </div>
      </ProtectedRoute>
    );
  }

  // Show 404 state
  if (isNotFound || !pod) {
    return (
      <ProtectedRoute>
        <div className="flex h-screen flex-col items-center justify-center gap-4 text-black">
          <div className="text-6xl font-bold">404</div>
          <div className="text-lg font-bold">Knowledge Pod Not Found</div>
          <div className="text-sm text-gray-500">{error || "The pod you are looking for does not exist or has been moved."}</div>
          <button onClick={() => router.back()} className="px-4 py-2 bg-orange-500 text-white rounded-lg font-bold hover:bg-orange-600">
            Go Back
          </button>
        </div>
      </ProtectedRoute>
    );
  }

  return (
    <ProtectedRoute>
      <div className="max-w-6xl mx-auto space-y-8 p-4 md:p-8">
        {/* Header Section */}
        <div className="flex justify-between items-center">
          <h1 className="text-4xl font-black text-black tracking-tight">{pod.name}</h1>
          <button className="px-10 py-2 bg-orange-500 text-white border-2 border-black font-bold shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:bg-orange-600 transition-all active:translate-x-[2px] active:translate-y-[2px] active:shadow-none">
            Use
          </button>
        </div>

        {/* Description Card */}
        <div className="bg-white border-2 border-black rounded-2xl p-6 pb-4 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
          <p className="text-xs font-semibold text-zinc-400 mb-1">Deskripsi</p>
          <p className="text-sm font-medium text-black mb-6">{pod.description || "No description available"}</p>

          <div className="flex justify-between items-center pt-4 border-t border-zinc-200">
            <div className="flex gap-6">
              <div className="flex items-center gap-2 text-zinc-500">
                <Eye size={16} />
                <span className="text-xs font-mono font-bold">{pod.view_count.toLocaleString()}</span>
              </div>

              <div className="flex items-center gap-2 text-zinc-500">
                <FileText size={16} />
                <span className="text-xs font-mono font-bold">{materials.length} files</span>
              </div>
            </div>

            <div className="flex items-center gap-2 px-3 py-1 bg-zinc-100 border-2 border-black rounded-full">
              <BadgeCheck size={16} className="fill-black text-white" />
              <span className="font-bold text-xs">Owner</span>
            </div>
          </div>
        </div>

        {/* <SearchSection/> */}

        {/* Files List Container */}
        <div className="bg-white border-2 border-black rounded-2xl overflow-hidden shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
          <div className="flex flex-col">
            {materials.length === 0 ? (
              <div className="p-6 text-center text-gray-500">
                <p className="text-sm">No materials in this pod yet</p>
              </div>
            ) : (
              materials.map((material, index) => (
                <FileListItem
                  key={material.id || index}
                  variant="file"
                  materialId={material.id}
                  userId={pod.owner_id || "User"}
                  podId={pod.id}
                  title={material.title}
                  description={material.description || ""}
                  likes={material.download_count.toString()}
                  date={new Date(material.created_at).toLocaleDateString()}
                  isLast={index === materials.length - 1}
                />
              ))
            )}
          </div>
        </div>
      </div>
    </ProtectedRoute>
  );
}
