'use client';

import React, { useRef } from 'react';
import { Upload, X, FileText, ArrowLeft, ArrowRight } from 'lucide-react';
import { useFormContext } from '@/context/FormContext';
import { Step } from '@/types';

const MaterialPage: React.FC = () => {
  const { formData, updateFormData, setCurrentStep } = useFormContext();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const newFiles = Array.from(e.target.files);
      const existingFiles = formData.materialFiles || [];
      updateFormData({ materialFiles: [...existingFiles, ...newFiles] });
    }
  };

  const removeFile = (index: number) => {
    const updatedFiles = [...(formData.materialFiles || [])];
    updatedFiles.splice(index, 1);
    updateFormData({ materialFiles: updatedFiles });
  };

  const handleBack = () => {
    setCurrentStep(Step.General);
  };

  const handleNext = () => {
    setCurrentStep(Step.Configurator);
  };

  const materialFiles = formData.materialFiles || [];

  return (
    <div className="h-full flex flex-col">
      <div className="mb-6">
        <h2 className="text-2xl font-black uppercase tracking-tight mb-2">Material</h2>
        <div className="h-1 w-16 bg-[#FF8A00] border border-black"></div>
      </div>

      <div className="flex-1 space-y-6">
        <div className="space-y-3">
          <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-400">
            Resource Upload <span className="text-[#FF8A00]">*</span>
          </label>
          <div
            onClick={() => fileInputRef.current?.click()}
            className="border-4 border-dashed border-black bg-gray-50 p-8 flex flex-col items-center justify-center gap-3 cursor-pointer hover:bg-white transition-all shadow-[4px_4px_0px_0px_#FF8A00] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[2px_2px_0px_0px_#FF8A00]"
          >
            <div className="p-2 bg-white border-2 border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
              <Upload size={22} />
            </div>
            <p className="font-black uppercase text-xs">Drop files or click to build</p>
            <input type="file" multiple className="hidden" ref={fileInputRef} onChange={handleFileChange} />
          </div>
        </div>

        {materialFiles.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 max-h-48 overflow-y-auto pr-2 custom-scrollbar">
            {materialFiles.map((file: File, idx: number) => (
              <div key={idx} className="flex items-center justify-between p-3 bg-white border-2 border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
                <div className="flex items-center gap-2 overflow-hidden">
                  <FileText className="text-[#FF8A00] shrink-0" size={16} />
                  <span className="font-mono text-[9px] font-bold truncate">{file.name}</span>
                </div>
                <X size={12} className="cursor-pointer hover:text-red-500" onClick={() => removeFile(idx)} />
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="pt-8 flex justify-between items-center">
        <button
          onClick={handleBack}
          className="flex items-center gap-2 font-black uppercase text-[10px] hover:text-[#FF8A00] transition-colors"
        >
          <ArrowLeft size={16} /> Back to core
        </button>

        <button
          onClick={handleNext}
          className="bg-[#FF8A00] text-black font-black uppercase text-xs px-8 py-3 border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:shadow-none hover:translate-x-[4px] hover:translate-y-[4px] transition-all flex items-center gap-3"
        >
          Next: Configurator <ArrowRight size={16} />
        </button>
      </div>
    </div>
  );
};

export default MaterialPage;


