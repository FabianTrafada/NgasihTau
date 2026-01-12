'use client';

import React, { useState } from 'react';
import { ArrowRight, X } from 'lucide-react';
import { useFormContext } from '@/context/FormContext';
import { Step } from '@/types';

const GeneralPage: React.FC = () => {
  const { formData, updateFormData, setCurrentStep } = useFormContext();
  const [tagInput, setTagInput] = useState('');
  const [categoryInput, setCategoryInput] = useState('');

  const handleNext = () => {
    const name = formData.name || '';
    const description = formData.description || '';

    if (name.trim() && description.trim()) {
      setCurrentStep(Step.Material);
    } else {
      alert('Pod Name and Description are required!');
    }
  };

  const addTag = () => {
    if (tagInput.trim()) {
      const currentTags = formData.tags || [];
      updateFormData({ tags: [...currentTags, tagInput.trim()] });
      setTagInput('');
    }
  };

  const removeTag = (index: number) => {
    const currentTags = formData.tags || [];
    const updatedTags = currentTags.filter((_: any, i: number) => i !== index);
    updateFormData({ tags: updatedTags });
  };

  const addCategory = () => {
    if (categoryInput.trim()) {
      const currentCategories = formData.categories || [];
      updateFormData({ categories: [...currentCategories, categoryInput.trim()] });
      setCategoryInput('');
    }
  };

  const removeCategory = (index: number) => {
    const currentCategories = formData.categories || [];
    const updatedCategories = currentCategories.filter((_: any, i: number) => i !== index);
    updateFormData({ categories: updatedCategories });
  };

  return (
    <div className="max-w-4xl mx-auto flex flex-col">
      <div className="mb-6">
        <h2 className="text-2xl font-black uppercase tracking-tight mb-2">General</h2>
        <div className="h-1 w-16 bg-[#FF8A00] border border-black"></div>
      </div>

      <div className="space-y-6 flex-1">
        {/* Pod Name */}
        <div className="space-y-2">
          <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-600 rounded-lg">
            Knowledge Pod Name
          </label>
          <input
            type="text"
            value={formData.name || ''}
            onChange={(e) => updateFormData({ name: e.target.value })}
            placeholder="Enter pod name..."
            className="w-full bg-white border-2 border-black p-4 shadow-[4px_4px_0px_0px_#FF8A00] focus:outline-none focus:shadow-none transition-all font-bold text-base"
          />
        </div>

        {/* Description */}
        <div className="space-y-2">
          <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-400">
            Pod Description <span className="text-[#FF8A00]">*</span>
          </label>
          <textarea
            value={formData.description || ''}
            onChange={(e) => updateFormData({ description: e.target.value })}
            placeholder="Provide architectural insights for this pod..."
            rows={4}
            className="w-full bg-white border-2 border-black p-4 shadow-[4px_4px_0px_0px_#FF8A00] focus:outline-none focus:shadow-none transition-all font-medium resize-none text-sm"
          />
        </div>

        {/* Categories */}
        <div className="space-y-2">
          <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-400">
            Categories
          </label>
          <div className="flex gap-2">
            <input
              type="text"
              value={categoryInput}
              onChange={(e) => setCategoryInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addCategory())}
              placeholder="Add category..."
              className="flex-1 bg-white border-2 border-black p-2.5 focus:outline-none focus:shadow-[2px_2px_0px_0px_#FF8A00] transition-all text-sm"
            />
            <button
              onClick={addCategory}
              className="bg-[#FF8A00] text-black font-bold uppercase text-[10px] px-4 border-2 border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] hover:shadow-none hover:translate-x-[2px] hover:translate-y-[2px] transition-all"
            >
              Add
            </button>
          </div>
          {formData.categories && formData.categories.length > 0 && (
            <div className="flex flex-wrap gap-2 mt-2">
              {formData.categories.map((category: string, idx: number) => (
                <div key={idx} className="bg-white border-2 border-black px-2 py-0.5 flex items-center gap-2 shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
                  <span className="text-[10px] font-bold">{category}</span>
                  <X size={12} className="cursor-pointer hover:text-red-500" onClick={() => removeCategory(idx)} />
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Tags */}
        <div className="space-y-2">
          <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-400">
            Tags
          </label>
          <div className="flex gap-2">
            <input
              type="text"
              value={tagInput}
              onChange={(e) => setTagInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addTag())}
              placeholder="Add tag..."
              className="flex-1 bg-white border-2 border-black p-2.5 focus:outline-none focus:shadow-[2px_2px_0px_0px_#FF8A00] transition-all text-sm"
            />
            <button
              onClick={addTag}
              className="bg-[#FF8A00] text-black font-bold uppercase text-[10px] px-4 border-2 border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] hover:shadow-none hover:translate-x-[2px] hover:translate-y-[2px] transition-all"
            >
              Add
            </button>
          </div>
          {formData.tags && formData.tags.length > 0 && (
            <div className="flex flex-wrap gap-2 mt-2">
              {formData.tags.map((tag: string, idx: number) => (
                <div key={idx} className="bg-white border-2 border-black px-2 py-0.5 flex items-center gap-2 shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
                  <span className="text-[10px] font-bold">{tag}</span>
                  <X size={12} className="cursor-pointer hover:text-red-500" onClick={() => removeTag(idx)} />
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="pt-8 flex justify-end">
        <button
          onClick={handleNext}
          className="bg-[#FF8A00] text-black font-black uppercase text-xs px-8 py-3 border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:shadow-none hover:translate-x-[4px] hover:translate-y-[4px] transition-all flex items-center gap-3"
        >
          Next: Build Materials <ArrowRight size={16} />
        </button>
      </div>
    </div>
  );
};

export default GeneralPage;


