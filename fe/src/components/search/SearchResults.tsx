"use client";

import { motion } from "framer-motion";
import { FileText, FolderOpen, Star, Eye, CheckCircle } from "lucide-react";
import { useRouter } from "next/navigation";
import type { SearchResult } from "@/types/search";
import { cn } from "@/lib/utils";

interface SearchResultsProps {
  results: SearchResult[];
  isLoading?: boolean;
  className?: string;
}

export function SearchResults({
  results,
  isLoading = false,
  className,
}: SearchResultsProps) {
  if (isLoading) {
    return (
      <div className={cn("space-y-4", className)}>
        {[...Array(3)].map((_, i) => (
          <SearchResultSkeleton key={i} />
        ))}
      </div>
    );
  }

  if (results.length === 0) {
    return (
      <div className={cn("text-center py-12", className)}>
        <FileText className="h-12 w-12 text-gray-300 mx-auto mb-4" />
        <p className="text-gray-500 text-lg">No results found</p>
        <p className="text-gray-400 text-sm mt-2">
          Try adjusting your search or filters
        </p>
      </div>
    );
  }

  return (
    <div className={cn("space-y-4", className)}>
      {results.map((result, index) => (
        <motion.div
          key={result.id}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: index * 0.05 }}
        >
          <SearchResultCard result={result} />
        </motion.div>
      ))}
    </div>
  );
}

interface SearchResultCardProps {
  result: SearchResult;
}

export function SearchResultCard({ result }: SearchResultCardProps) {
  const router = useRouter();

  const handleClick = () => {
    if (result.type === "pod") {
      // Navigate to pod page
      router.push(`/pods/${result.id}`);
    } else {
      // Navigate to material page (or pod containing the material)
      const podId = result.metadata?.pod_id as string;
      if (podId) {
        router.push(`/pods/${podId}/materials/${result.id}`);
      }
    }
  };

  const isPod = result.type === "pod";
  const isVerified = result.metadata?.is_verified as boolean;
  const starCount = result.metadata?.star_count as number;
  const viewCount = result.metadata?.view_count as number;
  const categories = result.metadata?.categories as string[];

  return (
    <div
      onClick={handleClick}
      className={cn(
        "bg-white border-2 border-[#2B2D42] rounded-lg p-4 cursor-pointer",
        "shadow-[4px_4px_0px_0px_#2B2D42] hover:shadow-[2px_2px_0px_0px_#2B2D42]",
        "hover:translate-x-[2px] hover:translate-y-[2px] transition-all"
      )}
    >
      <div className="flex items-start gap-4">
        {/* Icon */}
        <div
          className={cn(
            "flex-shrink-0 p-3 rounded-lg",
            isPod ? "bg-blue-100" : "bg-purple-100"
          )}
        >
          {isPod ? (
            <FolderOpen className={cn("h-6 w-6", "text-blue-600")} />
          ) : (
            <FileText className={cn("h-6 w-6", "text-purple-600")} />
          )}
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <h3 className="font-semibold text-lg truncate">{result.title}</h3>
            {isVerified && (
              <CheckCircle className="h-4 w-4 text-green-500 flex-shrink-0" />
            )}
            <span
              className={cn(
                "text-xs px-2 py-0.5 rounded-full font-medium flex-shrink-0",
                isPod
                  ? "bg-blue-100 text-blue-700"
                  : "bg-purple-100 text-purple-700"
              )}
            >
              {isPod ? "Pod" : "Material"}
            </span>
          </div>

          {result.description && (
            <p className="text-gray-600 text-sm line-clamp-2 mb-2">
              {result.description}
            </p>
          )}

          {/* Highlights */}
          {result.highlights && Object.keys(result.highlights).length > 0 && (
            <div className="mb-2">
              {Object.entries(result.highlights).map(([field, snippets]) => (
                <div key={field} className="text-sm text-gray-500">
                  {snippets.slice(0, 1).map((snippet, i) => (
                    <span
                      key={i}
                      dangerouslySetInnerHTML={{ __html: snippet }}
                      className="[&>em]:bg-yellow-200 [&>em]:font-semibold [&>em]:not-italic"
                    />
                  ))}
                </div>
              ))}
            </div>
          )}

          {/* Meta info */}
          <div className="flex items-center gap-4 text-sm text-gray-500">
            {starCount !== undefined && (
              <span className="flex items-center gap-1">
                <Star className="h-4 w-4" />
                {starCount}
              </span>
            )}
            {viewCount !== undefined && (
              <span className="flex items-center gap-1">
                <Eye className="h-4 w-4" />
                {viewCount}
              </span>
            )}
            {categories && categories.length > 0 && (
              <span className="truncate">
                {categories.slice(0, 2).join(", ")}
              </span>
            )}
          </div>
        </div>

        {/* Score badge */}
        {result.score > 0 && (
          <div className="flex-shrink-0 text-xs text-gray-400">
            {Math.round(result.score * 100)}% match
          </div>
        )}
      </div>
    </div>
  );
}

function SearchResultSkeleton() {
  return (
    <div className="bg-white border-2 border-gray-200 rounded-lg p-4 animate-pulse">
      <div className="flex items-start gap-4">
        <div className="w-12 h-12 bg-gray-200 rounded-lg" />
        <div className="flex-1">
          <div className="h-6 bg-gray-200 rounded w-1/3 mb-2" />
          <div className="h-4 bg-gray-200 rounded w-full mb-2" />
          <div className="h-4 bg-gray-200 rounded w-2/3" />
        </div>
      </div>
    </div>
  );
}
