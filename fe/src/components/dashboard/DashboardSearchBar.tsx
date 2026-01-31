"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Search, X, Loader2, TrendingUp, FileText, FolderOpen } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";
import type { SearchResult } from "@/types/search";

interface DashboardSearchBarProps {
  className?: string;
}

/**
 * Dashboard Search Bar with integrated search API
 * Features:
 * - Debounced autocomplete suggestions
 * - Quick preview of search results
 * - Keyboard navigation
 * - Redirects to full search page on submit
 */
export function DashboardSearchBar({ className }: DashboardSearchBarProps) {
  const router = useRouter();
  const t = useTranslations("dashboard");
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const [isFocused, setIsFocused] = useState(false);
  const [inputValue, setInputValue] = useState("");
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [quickResults, setQuickResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(-1);

  const debounceRef = useRef<NodeJS.Timeout | undefined>(undefined);

  // Fetch suggestions and quick results
  const fetchSearchData = useCallback(async (query: string) => {
    if (query.length < 2) {
      setSuggestions([]);
      setQuickResults([]);
      return;
    }

    setIsLoading(true);
    try {
      const { getSuggestions, search } = await import("@/lib/api/search");
      
      // Fetch suggestions and quick results in parallel
      const [suggestionsData, searchData] = await Promise.all([
        getSuggestions({ q: query, limit: 5 }).catch(() => []),
        search({ q: query, per_page: 3 }).catch(() => ({ data: [] })),
      ]);

      setSuggestions(suggestionsData);
      setQuickResults(searchData.data || []);
    } catch (error) {
      console.error("Search error:", error);
      setSuggestions([]);
      setQuickResults([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Debounced input handler
  useEffect(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    debounceRef.current = setTimeout(() => {
      fetchSearchData(inputValue);
    }, 300);

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, [inputValue, fetchSearchData]);

  // Handle click outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsFocused(false);
        setSelectedIndex(-1);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Handle form submit
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputValue.trim()) return;

    setIsFocused(false);
    router.push(`/search?q=${encodeURIComponent(inputValue)}`);
  };

  // Handle suggestion click
  const handleSuggestionClick = (suggestion: string) => {
    setInputValue(suggestion);
    setIsFocused(false);
    router.push(`/search?q=${encodeURIComponent(suggestion)}`);
  };

  // Handle result click
  const handleResultClick = (result: SearchResult) => {
    setIsFocused(false);
    if (result.type === "pod") {
      router.push(`/dashboard/pods/${result.id}`);
    } else {
      router.push(`/search?q=${encodeURIComponent(inputValue)}`);
    }
  };

  // Clear input
  const handleClear = () => {
    setInputValue("");
    setSuggestions([]);
    setQuickResults([]);
    setSelectedIndex(-1);
    inputRef.current?.focus();
  };

  // Keyboard navigation
  const handleKeyDown = (e: React.KeyboardEvent) => {
    const totalItems = suggestions.length + quickResults.length;
    
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIndex((prev) => (prev < totalItems - 1 ? prev + 1 : prev));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex((prev) => (prev > 0 ? prev - 1 : -1));
    } else if (e.key === "Enter" && selectedIndex >= 0) {
      e.preventDefault();
      if (selectedIndex < suggestions.length) {
        handleSuggestionClick(suggestions[selectedIndex]);
      } else {
        const resultIndex = selectedIndex - suggestions.length;
        if (quickResults[resultIndex]) {
          handleResultClick(quickResults[resultIndex]);
        }
      }
    } else if (e.key === "Escape") {
      setIsFocused(false);
      setSelectedIndex(-1);
    }
  };

  const showDropdown =
    isFocused && (suggestions.length > 0 || quickResults.length > 0 || isLoading);

  return (
    <div ref={containerRef} className={cn("relative", className)}>
      <form onSubmit={handleSubmit}>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            ref={inputRef}
            type="text"
            value={inputValue}
            onChange={(e) => {
              setInputValue(e.target.value);
              setSelectedIndex(-1);
            }}
            onFocus={() => setIsFocused(true)}
            onKeyDown={handleKeyDown}
            placeholder={t("searchPlaceholder")}
            className="w-full h-8 pl-10 pr-8 py-2 lg:py-2.5 border-[#2B2D42] border-2 rounded-none focus:outline-none focus:shadow-[2px_2px_0px_0px_#FF8811] transition-all text-sm lg:text-base shadow-[4px_4px_0px_0px_#2B2D42]"
          />
          {inputValue && (
            <button
              type="button"
              onClick={handleClear}
              className="absolute right-2 top-1/2 -translate-y-1/2 p-1 hover:bg-gray-100 rounded-full transition-colors"
            >
              <X className="h-3 w-3 text-gray-400" />
            </button>
          )}
        </div>
      </form>

      {/* Dropdown */}
      <AnimatePresence>
        {showDropdown && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.15 }}
            className="absolute top-full left-0 right-0 mt-1 bg-white border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] overflow-hidden z-50 max-h-[400px] overflow-y-auto"
          >
            {isLoading ? (
              <div className="flex items-center justify-center py-4">
                <Loader2 className="h-5 w-5 animate-spin text-gray-400" />
              </div>
            ) : (
              <>
                {/* Suggestions Section */}
                {suggestions.length > 0 && (
                  <div className="border-b border-gray-100">
                    <div className="px-3 py-2 text-xs text-gray-500 font-medium uppercase tracking-wide">
                      Suggestions
                    </div>
                    <ul>
                      {suggestions.map((suggestion, index) => (
                        <li key={`suggestion-${index}`}>
                          <button
                            type="button"
                            onClick={() => handleSuggestionClick(suggestion)}
                            className={cn(
                              "w-full flex items-center gap-3 px-3 py-2 text-left hover:bg-gray-50 transition-colors text-sm",
                              selectedIndex === index && "bg-gray-100"
                            )}
                          >
                            <TrendingUp className="h-4 w-4 text-gray-400 flex-shrink-0" />
                            <span className="truncate">{suggestion}</span>
                          </button>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {/* Quick Results Section */}
                {quickResults.length > 0 && (
                  <div>
                    <div className="px-3 py-2 text-xs text-gray-500 font-medium uppercase tracking-wide">
                      Quick Results
                    </div>
                    <ul>
                      {quickResults.map((result, index) => {
                        const actualIndex = suggestions.length + index;
                        return (
                          <li key={`result-${result.id}`}>
                            <button
                              type="button"
                              onClick={() => handleResultClick(result)}
                              className={cn(
                                "w-full flex items-center gap-3 px-3 py-2 text-left hover:bg-gray-50 transition-colors",
                                selectedIndex === actualIndex && "bg-gray-100"
                              )}
                            >
                              {result.type === "pod" ? (
                                <FolderOpen className="h-4 w-4 text-[#FF8811] flex-shrink-0" />
                              ) : (
                                <FileText className="h-4 w-4 text-blue-500 flex-shrink-0" />
                              )}
                              <div className="flex-1 min-w-0">
                                <p className="text-sm font-medium text-gray-900 truncate">
                                  {result.title}
                                </p>
                                {result.description && (
                                  <p className="text-xs text-gray-500 truncate">
                                    {result.description}
                                  </p>
                                )}
                              </div>
                              <span className="text-xs text-gray-400 capitalize">
                                {result.type}
                              </span>
                            </button>
                          </li>
                        );
                      })}
                    </ul>
                  </div>
                )}

                {/* View All Results */}
                {inputValue && (suggestions.length > 0 || quickResults.length > 0) && (
                  <div className="border-t border-gray-100 p-2">
                    <button
                      type="button"
                      onClick={handleSubmit}
                      className="w-full text-center py-2 text-sm text-[#FF8811] font-medium hover:bg-gray-50 transition-colors"
                    >
                      View all results for &quot;{inputValue}&quot;
                    </button>
                  </div>
                )}
              </>
            )}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

export default DashboardSearchBar;
