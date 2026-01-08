"use client";

import React, { useEffect, useState } from "react";
import FileListItem from "@/components/knowledge-pod/FileListItem";
import { KnowledgePod } from "@/types/knowledgePods";
import { getCurrentUser } from "@/lib/api/user";
import { getUserPods } from "@/lib/api/pod";
import { Pod } from "@/types/pod";
import Link from "next/link";

export default function KnowledgePage() {
  const [currentUser, setCurrentUser] = useState<{ id: string; name: string; username: string; email?: string } | null>(null);
  const [userPods, setUserPods] = useState<Pod[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        const userResponse = await getCurrentUser();
        setCurrentUser(userResponse);

        const userId = userResponse.id;

        const podsResponse = await getUserPods(userId, 1, 20);
        setUserPods(podsResponse.data);
      } catch (error) {
        console.error("Error fetching current user data:", error);
        setError("Failed to load pods");
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const handleToggleLike = async (id: string, isStarred: boolean): Promise<void> => {
    try {
      await import("@/lib/utils/starPod").then((m) => m.toggleStarPod(id, isStarred));
    } catch (error) {
      console.error("Error toggling star:", error);
      setError("Failed to toggle star");
    }
  };

  return (
    <div className="min-h-screen p-6">
      {/* 🔒 ONE GRID WRAPPER */}
      <div className="max-w-6xl mx-auto space-y-6">
        {/* ================= HEADER ================= */}
        <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold text-[#2B2D42] tracking-tight">
              My Knowledge Pods
            </h1>
            <p className="text-zinc-500 text-sm max-w-md">
              Manage and explore your personal collection of knowledge repositories.
            </p>
          </div>

          <div className="flex gap-3 md:ml-auto">
            <button className="px-5 py-2 border-2 border-black font-bold
              shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]
              bg-white hover:bg-zinc-50 transition-all text-sm">
              Newest
            </button>

            <Link href="/dashboard/pod/create">
              <button className="px-5 py-2 border-2 border-black font-bold
                bg-[#FF8811] text-white
                shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]
                hover:shadow-[4px_4px_0px_0px_#2B2D42]
                hover:-translate-x-[2px] hover:-translate-y-[2px]
                transition-all text-sm">
                Create Pod
              </button>
            </Link>
          </div>
        </div>

        {/* ================= LOADING ================= */}
        {loading && (
          <div className="flex justify-center py-12">
            <div className="h-8 w-8 animate-spin rounded-full border-b-2 border-[#FF8811]" />
          </div>
        )}

        {/* ================= ERROR ================= */}
        {error && (
          <div className="border-2 border-red-500 bg-red-50 p-4 rounded">
            <p className="font-bold text-red-700">Error</p>
            <p className="text-sm text-red-600">{error}</p>
          </div>
        )}

        {/* ================= LIST ================= */}
        {!loading && !error && currentUser && (
          <div className="bg-white border-2 border-black rounded-xl overflow-hidden
            shadow-[6px_6px_0px_0px_#FF8811]">
            {userPods.length > 0 ? (
              userPods.map((pod, index) => (
                <FileListItem 
                key={pod.id} 
                variant="pod" 
                userId={currentUser.id} 
                podId={pod.id} 
                title={pod.name} 
                description={pod.description || ""}
                date={new Date(pod.created_at).toLocaleDateString()} 
                onToggleLike={() => handleToggleLike} 
                isLast={index === userPods.length - 1} 
                isPersonal={true} 
                visibility={pod.visibility} />
              ))
            ) : (
              <div className="p-10 text-center text-zinc-500">
                No knowledge pods yet. Create one to get started!
              </div>
            )}
          </div>
        )}

        {/* ================= PAGINATION ================= */}
        <div className="flex justify-center pt-4">
          <nav className="flex gap-2">
            {[1, 2, 3, "...", 10].map((page, i) => (
              <button
                key={i}
                className={`w-10 h-10 border-2 border-black font-bold
                  shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]
                  ${page === 1
                    ? "bg-[#FF8811] text-white"
                    : "bg-white hover:bg-zinc-100"
                  }`}
              >
                {page}
              </button>
            ))}
          </nav>
        </div>
      </div>
    </div>
  );
}
