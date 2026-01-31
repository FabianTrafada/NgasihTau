'use client';

import React from 'react';
import { useRouter } from 'next/navigation';
import { useFormContext } from '@/context/FormContext';
import { Step } from '@/types';

const GeneralPage: React.FC = () => {
  const { formData, updateFormData, setCurrentStep } = useFormContext(); // ✅ Tambahin setCurrentStep
  const router = useRouter();

  const handleNext = () => {
    // ✅ Pake optional chaining atau default value
    const name = formData.name || '';
    const description = formData.description || '';
    
    if (name.trim() && description.trim()) {
      setCurrentStep(Step.Material); // ✅ Sekarang udah ada
    } else {
      alert('Required: Knowledge Pod Name & Description.');
    }
  };

  return (
    <div className="h-full flex flex-col">
      <div className="mb-10">
        <h2 className="text-3xl font-black uppercase font-[family-name:var(--font-plus-jakarta-sans)] mb-2">General</h2>
      </div>

      <div className="space-y-8 flex-1">
        {/* Knowledge Pod Name Input */}
        <div className="space-y-3">
          <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-400">
            Knowledge Pod Name <span className="text-[#FF8A00]">*</span>
          </label>
          <input
            type="text"
            value={formData.name || ''} // ✅ Tambahin || '' biar ga error
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
            value={formData.description || ''} // ✅ Tambahin || '' biar ga error
            onChange={(e) => updateFormData({ description: e.target.value })}
            placeholder="Provide architectural insights for this pod..."
            rows={6}
            className="w-full bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#FF8A00] focus:outline-none focus:shadow-none transition-all font-medium resize-none text-base"
          />
        </div>
      </div>

      <div className="pt-12 flex justify-end">
        <button
          onClick={handleNext}
          className="bg-[#FF8A00] text-black font-black uppercase text-sm px-12 py-4 border-2 border-black neo-btn-shadow neo-btn-active transition-all"
        >
          Next: Build Materials
        </button>
      </div>
    </div>
  );
};

export default GeneralPage;