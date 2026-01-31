  /**
 * Search Result Types
 * Based on backend search service domain entities
 */

export interface SearchResult {
  id: string;
  type: "pod" | "material";
  title: string;
  description: string;
  score: number;
  highlights?: Record<string, string[]>;
  metadata?: Record<string, unknown>;
}

export interface SearchResponse {
  data: SearchResult[];
  pagination: {
    page: number;
    per_page: number;
    total: number;
  };
}

export interface SemanticSearchResponse {
  data: SearchResult[];
}

export interface SuggestionsResponse {
  data: {
    suggestions: string[];
  };
}

export interface TrendingResponse {
  data: SearchResult[];
}

export interface SearchHistoryItem {
  id: string;
  user_id: string;
  query: string;
  created_at: string;
}

export interface SearchHistoryResponse {
  data: {
    history: SearchHistoryItem[];
  };
}

/**
 * Search Query Parameters
 */
export type SortBy = "relevance" | "upvotes" | "trust_score" | "recent" | "popular";

export interface SearchParams {
  q: string;
  type?: "pod" | "material";
  category?: string;
  file_type?: string;
  pod_id?: string;
  verified?: boolean;
  sort?: SortBy;
  page?: number;
  per_page?: number;
}

export interface SemanticSearchParams {
  q: string;
  pod_id?: string;
  limit?: number;
  min_score?: number;
}

export interface HybridSearchParams extends SearchParams {
  semantic_weight?: number;
}

export interface SuggestionParams {
  q: string;
  limit?: number;
}

export interface TrendingParams {
  category?: string;
  limit?: number;
}

export interface PopularParams {
  category?: string;
  limit?: number;
}

export interface SearchHistoryParams {
  limit?: number;
}
