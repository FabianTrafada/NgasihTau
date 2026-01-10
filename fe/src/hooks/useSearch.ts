"use client";

import { useState, useCallback, useEffect, useRef } from "react";
import { searchApi } from "@/lib/api/search";
import type {
  SearchParams,
  SearchResult,
  SortBy,
  SearchHistoryItem,
} from "@/types/search";

/**
 * Custom hook for search functionality
 * Provides debounced search, suggestions, and search history management
 */
export function useSearch(debounceMs: number = 300) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [pagination, setPagination] = useState({
    page: 1,
    per_page: 20,
    total: 0,
  });

  // Filters
  const [type, setType] = useState<"pod" | "material" | undefined>();
  const [category, setCategory] = useState<string | undefined>();
  const [sortBy, setSortBy] = useState<SortBy>("relevance");
  const [verified, setVerified] = useState<boolean | undefined>();

  const debounceRef = useRef<NodeJS.Timeout>();

  /**
   * Perform a search with current filters
   */
  const performSearch = useCallback(
    async (searchQuery: string, page: number = 1) => {
      if (!searchQuery.trim()) {
        setResults([]);
        setPagination({ page: 1, per_page: 20, total: 0 });
        return;
      }

      setIsLoading(true);
      setError(null);

      try {
        const params: SearchParams = {
          q: searchQuery,
          type,
          category,
          sort: sortBy,
          verified,
          page,
          per_page: pagination.per_page,
        };

        const response = await searchApi.search(params);
        setResults(response.data);
        setPagination(response.pagination);
      } catch (err) {
        setError(err instanceof Error ? err : new Error("Search failed"));
        setResults([]);
      } finally {
        setIsLoading(false);
      }
    },
    [type, category, sortBy, verified, pagination.per_page]
  );

  /**
   * Fetch suggestions for autocomplete
   */
  const fetchSuggestions = useCallback(async (prefix: string) => {
    if (!prefix.trim() || prefix.length < 2) {
      setSuggestions([]);
      return;
    }

    try {
      const suggestions = await searchApi.getSuggestions({ q: prefix, limit: 5 });
      setSuggestions(suggestions);
    } catch {
      setSuggestions([]);
    }
  }, []);

  /**
   * Debounced search handler
   */
  const handleQueryChange = useCallback(
    (newQuery: string) => {
      setQuery(newQuery);

      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }

      debounceRef.current = setTimeout(() => {
        performSearch(newQuery);
        fetchSuggestions(newQuery);
      }, debounceMs);
    },
    [debounceMs, performSearch, fetchSuggestions]
  );

  /**
   * Go to a specific page
   */
  const goToPage = useCallback(
    (page: number) => {
      performSearch(query, page);
    },
    [query, performSearch]
  );

  /**
   * Clear search results and query
   */
  const clearSearch = useCallback(() => {
    setQuery("");
    setResults([]);
    setSuggestions([]);
    setPagination({ page: 1, per_page: 20, total: 0 });
    setError(null);
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, []);

  return {
    // State
    query,
    results,
    suggestions,
    isLoading,
    error,
    pagination,

    // Filters
    type,
    category,
    sortBy,
    verified,

    // Actions
    setQuery: handleQueryChange,
    setType,
    setCategory,
    setSortBy,
    setVerified,
    search: performSearch,
    goToPage,
    clearSearch,
  };
}

/**
 * Hook for trending and popular content
 */
export function useTrendingContent(category?: string) {
  const [trending, setTrending] = useState<SearchResult[]>([]);
  const [popular, setPopular] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const fetchContent = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const [trendingData, popularData] = await Promise.all([
        searchApi.getTrending({ category, limit: 10 }),
        searchApi.getPopular({ category, limit: 10 }),
      ]);

      setTrending(trendingData);
      setPopular(popularData);
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Failed to fetch content"));
    } finally {
      setIsLoading(false);
    }
  }, [category]);

  useEffect(() => {
    fetchContent();
  }, [fetchContent]);

  return {
    trending,
    popular,
    isLoading,
    error,
    refresh: fetchContent,
  };
}

/**
 * Hook for search history management
 */
export function useSearchHistory() {
  const [history, setHistory] = useState<SearchHistoryItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const fetchHistory = useCallback(async (limit?: number) => {
    setIsLoading(true);
    setError(null);

    try {
      const historyData = await searchApi.getSearchHistory({ limit });
      setHistory(historyData);
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Failed to fetch history"));
    } finally {
      setIsLoading(false);
    }
  }, []);

  const clearHistory = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      await searchApi.clearSearchHistory();
      setHistory([]);
    } catch (err) {
      setError(err instanceof Error ? err : new Error("Failed to clear history"));
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory]);

  return {
    history,
    isLoading,
    error,
    refresh: fetchHistory,
    clear: clearHistory,
  };
}

/**
 * Hook for semantic/hybrid search
 */
export function useSemanticSearch() {
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const performSemanticSearch = useCallback(
    async (query: string, podId?: string, limit?: number) => {
      if (!query.trim()) {
        setResults([]);
        return;
      }

      setIsLoading(true);
      setError(null);

      try {
        const data = await searchApi.semanticSearch({
          q: query,
          pod_id: podId,
          limit: limit ?? 10,
        });
        setResults(data);
      } catch (err) {
        setError(err instanceof Error ? err : new Error("Semantic search failed"));
        setResults([]);
      } finally {
        setIsLoading(false);
      }
    },
    []
  );

  const performHybridSearch = useCallback(
    async (
      query: string,
      options?: {
        type?: "pod" | "material";
        category?: string;
        semanticWeight?: number;
        page?: number;
        perPage?: number;
      }
    ) => {
      if (!query.trim()) {
        setResults([]);
        return { data: [], pagination: { page: 1, per_page: 20, total: 0 } };
      }

      setIsLoading(true);
      setError(null);

      try {
        const response = await searchApi.hybridSearch({
          q: query,
          type: options?.type,
          category: options?.category,
          semantic_weight: options?.semanticWeight ?? 0.3,
          page: options?.page ?? 1,
          per_page: options?.perPage ?? 20,
        });
        setResults(response.data);
        return response;
      } catch (err) {
        setError(err instanceof Error ? err : new Error("Hybrid search failed"));
        setResults([]);
        return { data: [], pagination: { page: 1, per_page: 20, total: 0 } };
      } finally {
        setIsLoading(false);
      }
    },
    []
  );

  const clearResults = useCallback(() => {
    setResults([]);
    setError(null);
  }, []);

  return {
    results,
    isLoading,
    error,
    semanticSearch: performSemanticSearch,
    hybridSearch: performHybridSearch,
    clearResults,
  };
}
