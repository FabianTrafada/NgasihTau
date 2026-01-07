"use client";

import React, { useState } from "react";
import { useRouter } from "next/navigation";
import { useFormContext } from "@/context/FormContext";
import { Step } from "@/types";
import { X } from "lucide-react";

const AVAILABLE_CATEGORIES = ["Technology", "Science", "Business", "Education", "Health", "Arts", "Design", "Marketing", "Development", "Data Science", "AI/ML", "Finance"];

const GeneralPage: React.FC = () => {
  const { formData, updateFormData, setCurrentStep } = useFormContext();
  const router = useRouter();
  const [showCategoryDropdown, setShowCategoryDropdown] = useState(false);

  const categories = formData.categories || [];

  const handleNext = () => {
    const name = formData.name || "";
    const description = formData.description || "";

    if (name.trim() && description.trim()) {
      setCurrentStep(Step.Material);
    } else {
      alert("Required: Knowledge Pod Name & Description.");
    }
  };

  const handleAddCategory = (category: string) => {
    if (!categories.includes(category)) {
      updateFormData({ categories: [...categories, category] });
    }
    setShowCategoryDropdown(false);
  };

  const handleRemoveCategory = (category: string) => {
    updateFormData({ categories: categories.filter((c: string) => c !== category) });
  };

  const availableForSelection = AVAILABLE_CATEGORIES.filter((cat) => !categories.includes(cat));

  return (
    <div className="h-full flex flex-col">
      <div className="mb-10">
        <h2 className="text-3xl font-black uppercase tracking-tight mb-2">General</h2>
        <div className="h-1 w-20 bg-[#FF8A00] border border-black"></div>
      </div>

      <div className="space-y-8 flex-1">
        {/* Knowledge Pod Name Input */}
        <div className="space-y-3">
          <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-400">
            Knowledge Pod Name <span className="text-[#FF8A00]">*</span>
          </label>
          <input
            type="text"
            value={formData.name || ""}
            onChange={(e) => updateFormData({ name: e.target.value })}
            placeholder="Enter pod name..."
            className="w-full bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#FF8A00] focus:outline-none focus:shadow-none transition-all font-bold text-lg"
          />
        </div>

        {/* Description Textarea */}
        <div className="space-y-3">
          <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-400">
            Pod Description <span className="text-[#FF8A00]">*</span>
          </label>
          <textarea
            value={formData.description || ""}
            onChange={(e) => updateFormData({ description: e.target.value })}
            placeholder="Provide architectural insights for this pod..."
            rows={6}
            className="w-full bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#FF8A00] focus:outline-none focus:shadow-none transition-all font-medium resize-none text-base"
          />
        </div>

        {/* Categories */}
        <div className="space-y-3">
          <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-400">Categories</label>

          {/* Selected Categories */}
          <div className="flex flex-wrap gap-2 min-h-[40px] p-3 bg-white border-2 border-black rounded">
            {categories.length === 0 ? (
              <span className="text-sm text-gray-400">Add categories...</span>
            ) : (
              categories.map((cat: string) => (
                <div key={cat} className="flex items-center gap-2 px-3 py-1 bg-[#FF8A00] text-black rounded-full text-sm font-bold">
                  {cat}
                  <button onClick={() => handleRemoveCategory(cat)} className="hover:opacity-70 transition-opacity">
                    <X size={14} />
                  </button>
                </div>
              ))
            )}
          </div>

          {/* Category Dropdown */}
          <div className="relative">
            <button onClick={() => setShowCategoryDropdown(!showCategoryDropdown)} className="w-full bg-white border-2 border-black p-3 text-left font-bold hover:bg-gray-50 transition-colors">
              {showCategoryDropdown ? "▼ Select categories..." : "▶ Add categories"}
            </button>

            {showCategoryDropdown && availableForSelection.length > 0 && (
              <div className="absolute top-full left-0 right-0 bg-white border-2 border-black border-t-0 z-10 shadow-[4px_4px_0px_0px_rgba(0,0,0,0.1)]">
                {availableForSelection.map((cat) => (
                  <button key={cat} onClick={() => handleAddCategory(cat)} className="w-full text-left px-4 py-2 hover:bg-[#FF8A00] hover:text-white transition-colors font-medium border-b border-gray-200 last:border-b-0">
                    {cat}
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>

      <div className="pt-12 flex justify-end">
        <button onClick={handleNext} className="bg-[#FF8A00] text-black font-black uppercase text-sm px-12 py-4 border-2 border-black neo-btn-shadow neo-btn-active transition-all">
          Next: Build Materials
        </button>
      </div>
    </div>
  );
};

export default GeneralPage;
