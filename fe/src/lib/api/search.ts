import { apiClient } from "@/lib/api-client";
import type {
  SearchParams,
  SearchResponse,
  SemanticSearchParams,
  SemanticSearchResponse,
  HybridSearchParams,
  SuggestionParams,
  SuggestionsResponse,
  TrendingParams,
  PopularParams,
  TrendingResponse,
  SearchHistoryParams,
  SearchHistoryResponse,
  SearchResult,
} from "@/types/search";

/**
 * Search Service API
 * Implements all search endpoints from the backend search service
 */

/**
 * Full-text search for pods and materials
 * Endpoint: GET /api/v1/search
 *
 * @param params - Search parameters
 * @returns Paginated search results
 */
export async function search(params: SearchParams): Promise<SearchResponse> {
  try {
    const response = await apiClient.get<SearchResponse>("/api/v1/search", {
      params: {
        q: params.q,
        type: params.type,
        category: params.category,
        file_type: params.file_type,
        pod_id: params.pod_id,
        verified: params.verified,
        sort: params.sort,
        page: params.page ?? 1,
        per_page: params.per_page ?? 20,
      },
    });
    return response.data;
  } catch (error) {
    console.error("Error performing search:", error);
    throw error;
  }
}

/**
 * Semantic search using natural language with vector similarity
 * Endpoint: GET /api/v1/search/semantic
 *
 * @param params - Semantic search parameters
 * @returns Search results based on semantic similarity
 */
export async function semanticSearch(
  params: SemanticSearchParams
): Promise<SearchResult[]> {
  try {
    const response = await apiClient.get<SemanticSearchResponse>(
      "/api/v1/search/semantic",
      {
        params: {
          q: params.q,
          pod_id: params.pod_id,
          limit: params.limit ?? 10,
          min_score: params.min_score,
        },
      }
    );
    return response.data.data || [];
  } catch (error) {
    console.error("Error performing semantic search:", error);
    throw error;
  }
}

/**
 * Hybrid search combining keyword and semantic search
 * Endpoint: GET /api/v1/search/hybrid
 *
 * @param params - Hybrid search parameters
 * @returns Paginated search results combining both search methods
 */
export async function hybridSearch(
  params: HybridSearchParams
): Promise<SearchResponse> {
  try {
    const response = await apiClient.get<SearchResponse>(
      "/api/v1/search/hybrid",
      {
        params: {
          q: params.q,
          type: params.type,
          category: params.category,
          file_type: params.file_type,
          pod_id: params.pod_id,
          verified: params.verified,
          sort: params.sort,
          page: params.page ?? 1,
          per_page: params.per_page ?? 20,
          semantic_weight: params.semantic_weight ?? 0.3,
        },
      }
    );
    return response.data;
  } catch (error) {
    console.error("Error performing hybrid search:", error);
    throw error;
  }
}

/**
 * Get autocomplete suggestions based on query prefix
 * Endpoint: GET /api/v1/search/suggestions
 *
 * @param params - Suggestion parameters
 * @returns Array of suggestion strings
 */
export async function getSuggestions(
  params: SuggestionParams
): Promise<string[]> {
  try {
    const response = await apiClient.get<SuggestionsResponse>(
      "/api/v1/search/suggestions",
      {
        params: {
          q: params.q,
          limit: params.limit ?? 10,
        },
      }
    );
    return response.data.data?.suggestions || [];
  } catch (error) {
    console.error("Error fetching suggestions:", error);
    throw error;
  }
}

/**
 * Get trending materials (ranked by recent engagement in last 7 days)
 * Endpoint: GET /api/v1/search/trending
 *
 * @param params - Trending parameters
 * @returns Array of trending search results
 */
export async function getTrending(
  params?: TrendingParams
): Promise<SearchResult[]> {
  try {
    const response = await apiClient.get<TrendingResponse>(
      "/api/v1/search/trending",
      {
        params: {
          category: params?.category,
          limit: params?.limit ?? 20,
        },
      }
    );
    return response.data.data || [];
  } catch (error) {
    console.error("Error fetching trending:", error);
    throw error;
  }
}

/**
 * Get popular materials (ranked by all-time engagement)
 * Endpoint: GET /api/v1/search/popular
 *
 * @param params - Popular parameters
 * @returns Array of popular search results
 */
export async function getPopular(
  params?: PopularParams
): Promise<SearchResult[]> {
  try {
    const response = await apiClient.get<TrendingResponse>(
      "/api/v1/search/popular",
      {
        params: {
          category: params?.category,
          limit: params?.limit ?? 20,
        },
      }
    );
    return response.data.data || [];
  } catch (error) {
    console.error("Error fetching popular:", error);
    throw error;
  }
}

/**
 * Get the authenticated user's search history
 * Endpoint: GET /api/v1/search/history
 * Requires authentication
 *
 * @param params - History parameters
 * @returns Array of search history items
 */
export async function getSearchHistory(
  params?: SearchHistoryParams
): Promise<SearchHistoryResponse["data"]["history"]> {
  try {
    const response = await apiClient.get<SearchHistoryResponse>(
      "/api/v1/search/history",
      {
        params: {
          limit: params?.limit ?? 20,
        },
      }
    );
    return response.data.data?.history || [];
  } catch (error) {
    console.error("Error fetching search history:", error);
    throw error;
  }
}

/**
 * Clear the authenticated user's search history
 * Endpoint: DELETE /api/v1/search/history
 * Requires authentication
 */
export async function clearSearchHistory(): Promise<void> {
  try {
    await apiClient.delete("/api/v1/search/history");
  } catch (error) {
    console.error("Error clearing search history:", error);
    throw error;
  }
}

// Export all functions as a named object for convenience
export const searchApi = {
  search,
  semanticSearch,
  hybridSearch,
  getSuggestions,
  getTrending,
  getPopular,
  getSearchHistory,
  clearSearchHistory,
};

export default searchApi;
