"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Search, X, Clock, TrendingUp, Loader2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useSearch } from "@/hooks/useSearch";
import { cn } from "@/lib/utils";

interface SearchBarProps {
  placeholder?: string;
  className?: string;
  onSearch?: (query: string) => void;
  showSuggestions?: boolean;
  showHistory?: boolean;
  autoFocus?: boolean;
  redirectToResults?: boolean;
}

export function SearchBar({
  placeholder = "Search for knowledge pods, materials...",
  className,
  onSearch,
  showSuggestions = true,
  showHistory = false,
  autoFocus = false,
  redirectToResults = true,
}: SearchBarProps) {
  const router = useRouter();
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const [isFocused, setIsFocused] = useState(false);
  const [inputValue, setInputValue] = useState("");

  const {
    suggestions,
    isLoading,
  } = useSearch(300);

  const [localSuggestions, setLocalSuggestions] = useState<string[]>([]);

  // Fetch suggestions when input changes
  const fetchSuggestionsLocal = useCallback(async (value: string) => {
    if (value.length < 2) {
      setLocalSuggestions([]);
      return;
    }

    try {
      const { getSuggestions } = await import("@/lib/api/search");
      const data = await getSuggestions({ q: value, limit: 5 });
      setLocalSuggestions(data);
    } catch {
      setLocalSuggestions([]);
    }
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => {
      if (showSuggestions) {
        fetchSuggestionsLocal(inputValue);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [inputValue, showSuggestions, fetchSuggestionsLocal]);

  // Handle click outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsFocused(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputValue.trim()) return;

    setIsFocused(false);
    onSearch?.(inputValue);

    if (redirectToResults) {
      router.push(`/search?q=${encodeURIComponent(inputValue)}`);
    }
  };

  const handleSuggestionClick = (suggestion: string) => {
    setInputValue(suggestion);
    setIsFocused(false);
    onSearch?.(suggestion);

    if (redirectToResults) {
      router.push(`/search?q=${encodeURIComponent(suggestion)}`);
    }
  };

  const handleClear = () => {
    setInputValue("");
    setLocalSuggestions([]);
    inputRef.current?.focus();
  };

  const displaySuggestions = localSuggestions.length > 0 ? localSuggestions : suggestions;
  const showDropdown = isFocused && (displaySuggestions.length > 0 || isLoading);

  return (
    <div ref={containerRef} className={cn("relative w-full", className)}>
      <form onSubmit={handleSubmit}>
        <div className="relative">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
          <input
            ref={inputRef}
            type="text"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onFocus={() => setIsFocused(true)}
            placeholder={placeholder}
            autoFocus={autoFocus}
            className={cn(
              "w-full bg-white border-2 border-[#2B2D42] rounded-md py-4 pl-12 pr-12",
              "text-lg shadow-[4px_4px_0px_0px_#2B2D42]",
              "focus:outline-none focus:shadow-[2px_2px_0px_0px_#2B2D42]",
              "focus:translate-x-[2px] focus:translate-y-[2px]",
              "transition-all font-[family-name:var(--font-inter)]"
            )}
          />
          {inputValue && (
            <button
              type="button"
              onClick={handleClear}
              className="absolute right-4 top-1/2 -translate-y-1/2 p-1 hover:bg-gray-100 rounded-full transition-colors"
            >
              <X className="h-4 w-4 text-gray-400" />
            </button>
          )}
        </div>
      </form>

      {/* Suggestions Dropdown */}
      <AnimatePresence>
        {showDropdown && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.15 }}
            className="absolute top-full left-0 right-0 mt-2 bg-white border-2 border-[#2B2D42] rounded-md shadow-[4px_4px_0px_0px_#2B2D42] overflow-hidden z-50"
          >
            {isLoading ? (
              <div className="flex items-center justify-center py-4">
                <Loader2 className="h-5 w-5 animate-spin text-gray-400" />
              </div>
            ) : (
              <ul>
                {displaySuggestions.map((suggestion, index) => (
                  <li key={index}>
                    <button
                      type="button"
                      onClick={() => handleSuggestionClick(suggestion)}
                      className="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-gray-50 transition-colors"
                    >
                      <TrendingUp className="h-4 w-4 text-gray-400 flex-shrink-0" />
                      <span className="truncate">{suggestion}</span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

/**
 * Compact search bar variant for headers/navbars
 */
export function CompactSearchBar({
  placeholder = "Search...",
  className,
  onSearch,
}: Pick<SearchBarProps, "placeholder" | "className" | "onSearch">) {
  const router = useRouter();
  const [query, setQuery] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    onSearch?.(query);
    router.push(`/search?q=${encodeURIComponent(query)}`);
  };

  return (
    <form onSubmit={handleSubmit} className={cn("relative", className)}>
      <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder={placeholder}
        className={cn(
          "w-full bg-gray-100 border border-gray-200 rounded-lg py-2 pl-10 pr-4",
          "text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent",
          "transition-all"
        )}
      />
    </form>
  );
}
