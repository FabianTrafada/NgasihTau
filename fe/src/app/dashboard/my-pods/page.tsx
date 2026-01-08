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
     <div className="mx-auto space-y-8 p-4 md:p-8">
      {/* Header Section */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 max-w-6xl">
        <div className="space-y-1">
          <h1 className="text-1xl md:text-2xl sm:text-2xl font-bold text-[#2B2D42]  tracking-tight font-family-name:var(--font-plus-jakarta-sans) mb-1">My Knowledge Pods</h1>
          <p className="text-zinc-500  font-medium text-xs sm:text-base max-w-md">
            Manage and explore your personal collection of knowledge repositories.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          {/* Added Filter Button to match screenshot */}
          <button className="px-5 py-2 border-2 border-black font-bold shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] bg-white hover:bg-zinc-50 transition-all text-sm leading-none">
            Newest
          </button>
          <Link href="/dashboard/pod/create" className="shrink-0">
            <button className="px-5 py-2 border-2 border-black font-bold shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] bg-[#FF8811] text-white hover:shadow-[4px_4px_0px_0px_#2B2D42] hover:-translate-x-[2px] hover:-translate-y-[2px] active:translate-x-0 active:translate-y-0 active:shadow-none transition-all text-sm leading-none cursor-pointer">
              Create Pod
            </button>
          </Link>
        </div>
      </div>

      {/* Loading State */}
      {loading && (
        <div className="flex justify-center items-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[#FF8811]"></div>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="bg-red-50 border-2 border-red-500 text-red-700 p-4 rounded">
          <p className="font-bold">Error</p>
          <p className="text-sm">{error}</p>
        </div>
      )}

      {/* Main Container - Industrial / Neo-brutalism Style */}
      {!loading && !error && currentUser && (
        <div className="bg-white border-2 border-r-4 border-black rounded-[12px] overflow-hidden shadow-[6px_6px_0px_0px_#FF8811]">
          <div className="flex flex-col">
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
                  visibility={pod.visibility}
                />
              ))
            ) : (
              <div className="p-8 text-center text-gray-500">
                <p>No knowledge pods yet. Create one to get started!</p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Pagination / Industrial Footer */}
      <div className="flex justify-center pt-4">
        <nav className="flex items-center gap-1">
          {[1, 2, 3, "...", 10].map((page, i) => (
            <button
              key={i}
              className={`w-10 h-10 flex items-center justify-center border-2 border-black font-bold shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] transition-all ${page === 1 ? "bg-orange-500 text-white" : "bg-white text-black hover:bg-zinc-100"
                }`}
            >
              {page}
            </button>
          ))}
        </nav>
      </div>
    </div>
  );
}
