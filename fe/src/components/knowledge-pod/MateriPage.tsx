'use client';

import React, { useRef } from 'react';
import { Upload, X, FileText, ArrowLeft } from 'lucide-react';
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

  const handleNext = () => {
    setCurrentStep(Step.Configurator);
  };

  const handleBack = () => {
    setCurrentStep(Step.General);
  };

  const materialFiles = formData.materialFiles || [];

  return (
    <div className="h-full flex flex-col">
      <div className="mb-10">
        <h2 className="text-3xl font-black uppercase tracking-tight mb-2">Material</h2>
        <div className="h-1 w-20 bg-[#FF8A00] border border-black"></div>
      </div>

      <div className="flex-1 space-y-8">
        <div className="space-y-3">
          <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-400">
            Resource Upload <span className="text-[#FF8A00]">*</span>
          </label>
          <div 
            onClick={() => fileInputRef.current?.click()}
            className="border-4 border-dashed border-black bg-gray-50 p-16 flex flex-col items-center justify-center gap-6 cursor-pointer hover:bg-white transition-all shadow-[6px_6px_0px_0px_#FF8A00]"
          >
            <div className="p-4 bg-white border-2 border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
              <Upload size={36} className="text-black" />
            </div>
            <div className="text-center">
              <p className="font-black text-lg uppercase">Architect Your Assets</p>
              <p className="text-[10px] font-mono text-gray-400 mt-2 uppercase tracking-widest">SVG, PDF, DOCS // SYSTEM LIMIT: 50MB</p>
            </div>
            <input type="file" multiple className="hidden" ref={fileInputRef} onChange={handleFileChange} />
          </div>
        </div>

        {/* File List */}
        {materialFiles.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-h-48 overflow-y-auto p-1">
            {materialFiles.map((file: File, idx: number) => (
              <div key={idx} className="flex items-center justify-between p-4 bg-white border-2 border-black shadow-[3px_3px_0px_0px_rgba(0,0,0,1)]">
                <div className="flex items-center gap-3 overflow-hidden">
                  <FileText className="text-[#FF8A00] flex-shrink-0" size={20} />
                  <span className="font-mono text-xs font-bold truncate">{file.name}</span>
                </div>
                <button onClick={() => removeFile(idx)} className="p-1 hover:bg-black hover:text-white transition-all border border-transparent hover:border-black">
                  <X size={16} />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="pt-12 flex justify-between items-center">
        <button
          onClick={handleBack}
          className="flex items-center gap-3 font-black uppercase text-xs hover:text-[#FF8A00] transition-colors"
        >
          <ArrowLeft size={18} /> Back to core
        </button>
        <button
          onClick={handleNext}
          className="bg-[#FF8A00] text-black font-black uppercase text-sm px-12 py-4 border-2 border-black neo-btn-shadow neo-btn-active transition-all"
        >
          Next: Configurator
        </button>
      </div>
    </div>
  );
};

export default MaterialPage;