"use client";

import React, { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { getUserPods, PaginatedPodResponse } from "@/lib/api/pod";
import { Pod } from "@/types/pod";
import Link from "next/link";

export default function UserPublicProfilePage() {
  const params = useParams();
  const userId = params.username as string;

  const [pods, setPods] = useState<Pod[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState({
    page: 1,
    per_page: 20,
    total: 0,
  });

  useEffect(() => {
    const fetchUserPods = async () => {
      try {
        setLoading(true);
        setError(null);

        // userId is actually a UUID from the route parameter
        const response = await getUserPods(userId, pagination.page, pagination.per_page);

        setPods(response.data);
        setPagination(response.pagination);
      } catch (err) {
        console.error("Failed to fetch user pods:", err);
        setError("Failed to load knowledge pods");
      } finally {
        setLoading(false);
      }
    };

    if (userId) {
      fetchUserPods();
    }
  }, [userId, pagination.page]);

  const handleNextPage = () => {
    if (pagination.page * pagination.per_page < pagination.total) {
      setPagination((prev) => ({ ...prev, page: prev.page + 1 }));
    }
  };

  const handlePrevPage = () => {
    if (pagination.page > 1) {
      setPagination((prev) => ({ ...prev, page: prev.page - 1 }));
    }
  };

  return (
    <div className="min-h-screen bg-background">
      {/* User Info Section - TODO: Implement later */}
      <div className="border-b">
        <div className="max-w-6xl mx-auto px-4 py-8">
          <h1 className="text-3xl font-bold">User Profile</h1>
          <p className="text-muted-foreground mt-2">User ID: {userId}</p>
          <p className="text-sm text-muted-foreground mt-1">Note: Make sure to access with a valid UUID, e.g., /550e8400-e29b-41d4-a716-446655440000</p>
        </div>
      </div>

      {/* Knowledge Pods Section */}
      <div className="max-w-6xl mx-auto px-4 py-8">
        <h2 className="text-2xl font-bold mb-6">Knowledge Pods</h2>

        {error && <div className="bg-destructive/10 border border-destructive text-destructive p-4 rounded-lg mb-6">{error}</div>}

        {loading ? (
          <div className="flex justify-center items-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        ) : pods.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-muted-foreground">No knowledge pods yet</p>
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
              {pods.map((pod) => (
                <Link href={`/pods/${pod.id}`} key={pod.id}>
                  <div className="h-full border rounded-lg p-6 hover:shadow-lg hover:border-primary transition-all cursor-pointer">
                    <h3 className="font-semibold text-lg mb-2 line-clamp-2">{pod.name}</h3>
                    {pod.description && <p className="text-sm text-muted-foreground mb-4 line-clamp-2">{pod.description}</p>}

                    <div className="flex items-center gap-4 text-sm text-muted-foreground mb-4">
                      <div className="flex items-center gap-1">
                        <span>👁️</span>
                        <span>{pod.view_count}</span>
                      </div>
                      <div className="flex items-center gap-1">
                        <span>⭐</span>
                        <span>{pod.star_count}</span>
                      </div>
                      <div className="flex items-center gap-1">
                        <span>🔄</span>
                        <span>{pod.fork_count}</span>
                      </div>
                    </div>

                    {pod.tags && pod.tags.length > 0 && (
                      <div className="flex flex-wrap gap-2">
                        {pod.tags.slice(0, 3).map((tag) => (
                          <span key={tag} className="text-xs bg-muted px-2 py-1 rounded">
                            {tag}
                          </span>
                        ))}
                        {pod.tags.length > 3 && <span className="text-xs text-muted-foreground">+{pod.tags.length - 3}</span>}
                      </div>
                    )}

                    <div className="mt-4 pt-4 border-t text-xs text-muted-foreground">{pod.visibility === "public" ? "🌐 Public" : "🔒 Private"}</div>
                  </div>
                </Link>
              ))}
            </div>

            {/* Pagination */}
            {pagination.total > pagination.per_page && (
              <div className="flex justify-center items-center gap-4">
                <button onClick={handlePrevPage} disabled={pagination.page === 1} className="px-4 py-2 border rounded-lg hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed">
                  Previous
                </button>
                <span className="text-sm text-muted-foreground">
                  Page {pagination.page} of {Math.ceil(pagination.total / pagination.per_page)}
                </span>
                <button onClick={handleNextPage} disabled={pagination.page * pagination.per_page >= pagination.total} className="px-4 py-2 border rounded-lg hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed">
                  Next
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
