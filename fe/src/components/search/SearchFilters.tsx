"use client";

import { Check, ChevronDown, Filter, SlidersHorizontal } from "lucide-react";
import { useState } from "react";
import type { SortBy } from "@/types/search";
import { cn } from "@/lib/utils";

interface SearchFiltersProps {
  type?: "pod" | "material";
  category?: string;
  sortBy: SortBy;
  verified?: boolean;
  categories?: string[];
  onTypeChange: (type: "pod" | "material" | undefined) => void;
  onCategoryChange: (category: string | undefined) => void;
  onSortChange: (sort: SortBy) => void;
  onVerifiedChange: (verified: boolean | undefined) => void;
  className?: string;
}

const SORT_OPTIONS: { value: SortBy; label: string }[] = [
  { value: "relevance", label: "Most Relevant" },
  { value: "trust_score", label: "Trust Score" },
  { value: "upvotes", label: "Most Upvoted" },
  { value: "recent", label: "Most Recent" },
  { value: "popular", label: "Most Popular" },
];

const TYPE_OPTIONS = [
  { value: undefined, label: "All Types" },
  { value: "pod" as const, label: "Pods" },
  { value: "material" as const, label: "Materials" },
];

export function SearchFilters({
  type,
  category,
  sortBy,
  verified,
  categories = [],
  onTypeChange,
  onCategoryChange,
  onSortChange,
  onVerifiedChange,
  className,
}: SearchFiltersProps) {
  return (
    <div
      className={cn(
        "flex flex-wrap items-center gap-3 p-4 bg-gray-50 rounded-lg border border-gray-200",
        className
      )}
    >
      <div className="flex items-center gap-2 text-sm font-medium text-gray-700">
        <SlidersHorizontal className="h-4 w-4" />
        <span>Filters:</span>
      </div>

      {/* Type Filter */}
      <FilterDropdown
        label="Type"
        value={type ?? "all"}
        options={TYPE_OPTIONS.map((opt) => ({
          value: opt.value ?? "all",
          label: opt.label,
        }))}
        onChange={(val) =>
          onTypeChange(val === "all" ? undefined : (val as "pod" | "material"))
        }
      />

      {/* Category Filter */}
      {categories.length > 0 && (
        <FilterDropdown
          label="Category"
          value={category ?? "all"}
          options={[
            { value: "all", label: "All Categories" },
            ...categories.map((cat) => ({ value: cat, label: cat })),
          ]}
          onChange={(val) => onCategoryChange(val === "all" ? undefined : val)}
        />
      )}

      {/* Sort */}
      <FilterDropdown
        label="Sort by"
        value={sortBy}
        options={SORT_OPTIONS}
        onChange={(val) => onSortChange(val as SortBy)}
      />

      {/* Verified Toggle */}
      <button
        onClick={() =>
          onVerifiedChange(verified === undefined ? true : undefined)
        }
        className={cn(
          "flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors",
          verified
            ? "bg-green-100 text-green-700 border border-green-300"
            : "bg-white text-gray-600 border border-gray-300 hover:bg-gray-50"
        )}
      >
        <Check className={cn("h-4 w-4", verified && "text-green-600")} />
        Verified Only
      </button>
    </div>
  );
}

interface FilterDropdownProps {
  label: string;
  value: string;
  options: { value: string; label: string }[];
  onChange: (value: string) => void;
}

function FilterDropdown({
  label,
  value,
  options,
  onChange,
}: FilterDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const selectedOption = options.find((opt) => opt.value === value);

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          "flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-medium",
          "bg-white border border-gray-300 hover:bg-gray-50 transition-colors"
        )}
      >
        <span className="text-gray-500">{label}:</span>
        <span>{selectedOption?.label}</span>
        <ChevronDown
          className={cn("h-4 w-4 transition-transform", isOpen && "rotate-180")}
        />
      </button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-10"
            onClick={() => setIsOpen(false)}
          />
          <div className="absolute top-full left-0 mt-1 min-w-[160px] bg-white border border-gray-200 rounded-lg shadow-lg z-20">
            {options.map((option) => (
              <button
                key={option.value}
                onClick={() => {
                  onChange(option.value);
                  setIsOpen(false);
                }}
                className={cn(
                  "w-full flex items-center gap-2 px-3 py-2 text-sm text-left",
                  "hover:bg-gray-50 transition-colors first:rounded-t-lg last:rounded-b-lg",
                  value === option.value && "bg-blue-50 text-blue-700"
                )}
              >
                {value === option.value && <Check className="h-4 w-4" />}
                <span className={value !== option.value ? "ml-6" : ""}>
                  {option.label}
                </span>
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}

/**
 * Compact filter bar for mobile
 */
export function MobileSearchFilters({
  sortBy,
  onSortChange,
  verified,
  onVerifiedChange,
}: Pick<
  SearchFiltersProps,
  "sortBy" | "onSortChange" | "verified" | "onVerifiedChange"
>) {
  const [showFilters, setShowFilters] = useState(false);

  return (
    <div className="md:hidden">
      <button
        onClick={() => setShowFilters(!showFilters)}
        className="flex items-center gap-2 px-4 py-2 bg-gray-100 rounded-lg text-sm font-medium"
      >
        <Filter className="h-4 w-4" />
        Filters
        <ChevronDown
          className={cn(
            "h-4 w-4 transition-transform",
            showFilters && "rotate-180"
          )}
        />
      </button>

      {showFilters && (
        <div className="mt-3 p-4 bg-gray-50 rounded-lg space-y-3">
          <div>
            <label className="text-sm font-medium text-gray-700 mb-2 block">
              Sort by
            </label>
            <select
              value={sortBy}
              onChange={(e) => onSortChange(e.target.value as SortBy)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
            >
              {SORT_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="verified-mobile"
              checked={verified ?? false}
              onChange={(e) =>
                onVerifiedChange(e.target.checked ? true : undefined)
              }
              className="rounded border-gray-300"
            />
            <label htmlFor="verified-mobile" className="text-sm text-gray-700">
              Show verified only
            </label>
          </div>
        </div>
      )}
    </div>
  );
}
