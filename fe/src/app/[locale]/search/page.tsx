"use client";

import { useSearchParams } from "next/navigation";
import { useEffect, Suspense } from "react";
import { motion } from "framer-motion";
import { useSearch, useTrendingContent } from "@/hooks/useSearch";
import { SearchBar, SearchResults, SearchFilters } from "@/components/search";
import { TrendingUp, Star } from "lucide-react";

function SearchPageContent() {
  const searchParams = useSearchParams();
  const initialQuery = searchParams.get("q") || "";

  const {
    query,
    results,
    isLoading,
    error,
    pagination,
    type,
    category,
    sortBy,
    verified,
    setQuery,
    setType,
    setCategory,
    setSortBy,
    setVerified,
    search,
    goToPage,
  } = useSearch(300);

  const { trending, popular } = useTrendingContent();

  // Initialize with URL query parameter
  useEffect(() => {
    if (initialQuery && initialQuery !== query) {
      setQuery(initialQuery);
    }
  }, [initialQuery]);

  const totalPages = Math.ceil(pagination.total / pagination.per_page);
  const showEmptyState = !query && results.length === 0 && !isLoading;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b border-gray-200 py-6">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-6">
            Search Knowledge
          </h1>

          {/* Search Bar */}
          <SearchBar
            placeholder="Search for knowledge pods, materials, topics..."
            showSuggestions={true}
            redirectToResults={false}
            onSearch={(q) => search(q)}
            autoFocus
          />
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Filters */}
        <SearchFilters
          type={type}
          category={category}
          sortBy={sortBy}
          verified={verified}
          categories={["Mathematics", "Science", "History", "Programming", "Language"]}
          onTypeChange={setType}
          onCategoryChange={setCategory}
          onSortChange={setSortBy}
          onVerifiedChange={setVerified}
          className="mb-6"
        />

        {/* Error State */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
            <p className="text-red-700">
              Something went wrong. Please try again.
            </p>
          </div>
        )}

        {/* Results */}
        {query ? (
          <div>
            <div className="flex items-center justify-between mb-4">
              <p className="text-sm text-gray-500">
                {isLoading
                  ? "Searching..."
                  : `${pagination.total} results for "${query}"`}
              </p>
            </div>

            <SearchResults results={results} isLoading={isLoading} />

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex justify-center gap-2 mt-8">
                <button
                  onClick={() => goToPage(pagination.page - 1)}
                  disabled={pagination.page <= 1}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
                >
                  Previous
                </button>
                <span className="flex items-center px-4 text-sm text-gray-600">
                  Page {pagination.page} of {totalPages}
                </span>
                <button
                  onClick={() => goToPage(pagination.page + 1)}
                  disabled={pagination.page >= totalPages}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
                >
                  Next
                </button>
              </div>
            )}
          </div>
        ) : showEmptyState ? (
          /* Empty State - Show Trending & Popular */
          <div className="space-y-12">
            {/* Trending Section */}
            {trending.length > 0 && (
              <motion.section
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
              >
                <div className="flex items-center gap-2 mb-4">
                  <TrendingUp className="h-5 w-5 text-orange-500" />
                  <h2 className="text-xl font-semibold">Trending Now</h2>
                </div>
                <SearchResults results={trending.slice(0, 5)} />
              </motion.section>
            )}

            {/* Popular Section */}
            {popular.length > 0 && (
              <motion.section
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1 }}
              >
                <div className="flex items-center gap-2 mb-4">
                  <Star className="h-5 w-5 text-yellow-500" />
                  <h2 className="text-xl font-semibold">Popular</h2>
                </div>
                <SearchResults results={popular.slice(0, 5)} />
              </motion.section>
            )}

            {/* Empty State Fallback */}
            {trending.length === 0 && popular.length === 0 && (
              <div className="text-center py-12">
                <p className="text-gray-500 text-lg">
                  Start searching to discover knowledge pods and materials
                </p>
              </div>
            )}
          </div>
        ) : null}
      </div>
    </div>
  );
}

export default function SearchPage() {
  return (
    <Suspense fallback={<SearchPageSkeleton />}>
      <SearchPageContent />
    </Suspense>
  );
}

function SearchPageSkeleton() {
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white border-b border-gray-200 py-6">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="h-10 w-48 bg-gray-200 rounded animate-pulse mb-6" />
          <div className="h-14 w-full bg-gray-200 rounded animate-pulse" />
        </div>
      </div>
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="h-16 w-full bg-gray-200 rounded animate-pulse mb-6" />
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="h-24 w-full bg-gray-200 rounded animate-pulse" />
          ))}
        </div>
      </div>
    </div>
  );
}
