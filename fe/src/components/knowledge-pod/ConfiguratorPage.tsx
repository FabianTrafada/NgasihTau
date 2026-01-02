'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Lock, Globe, CheckCircle, ArrowLeft } from 'lucide-react';
import { useFormContext } from '@/context/FormContext';
import { Step } from '@/types';

const ConfiguratorPage: React.FC = () => {
  const { formData, updateFormData, resetForm, setCurrentStep } = useFormContext();
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const handleComplete = async () => {
    setIsSubmitting(true);
    await new Promise(r => setTimeout(r, 2000)); // Simulate "architecting" process
    setIsSubmitting(false);
    setIsSuccess(true);
  };

  const handleBack = () => {
    setCurrentStep(Step.Material);
  };

  if (isSuccess) {
    return (
      <div className="h-full flex flex-col items-center justify-center space-y-8 animate-in zoom-in-95 duration-700">
        <div className="w-24 h-24 bg-[#FF8A00] border-4 border-black flex items-center justify-center text-black shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]">
          <CheckCircle size={56} strokeWidth={3} />
        </div>
        <div className="text-center">
          <h2 className="text-4xl font-black mb-3 uppercase italic tracking-tighter">Architected</h2>
          <p className="font-mono text-gray-500 uppercase text-xs tracking-[0.3em]">POD_GEN_SUCCESS // SYNC_COMPLETE</p>
        </div>
        <div className="bg-gray-50 border-2 border-dashed border-black p-4 w-full max-w-sm text-center">
           <p className="text-xs font-mono font-bold uppercase">Pod: {formData.name || 'Untitled'}</p>
        </div>
        <button
          onClick={() => { 
            resetForm(); 
            router.push('/dashboard/knowledge'); 
          }}
          className="bg-black text-white font-black uppercase text-sm px-14 py-5 border-2 border-black neo-btn-shadow neo-btn-active transition-all"
        >
          Initialize New Build
        </button>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col">
      <div className="mb-10">
        <h2 className="text-3xl font-black uppercase tracking-tight mb-2">Configurator</h2>
        <div className="h-1 w-20 bg-[#FF8A00] border border-black"></div>
      </div>

      <div className="flex-1 space-y-10">
        <div className="space-y-4">
          <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-400">
            Node Visibility Protocols <span className="text-[#FF8A00]">*</span>
          </label>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <button 
              onClick={() => updateFormData({ visibility: 'private' })}
              className={`p-8 border-4 border-black text-left flex flex-col gap-4 transition-all ${
                formData.visibility === 'private' 
                ? 'bg-[#FF8A00] translate-x-1 translate-y-1 shadow-none' 
                : 'bg-white shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] hover:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:translate-x-1 hover:translate-y-1'
              }`}
            >
              <Lock size={28} />
              <div>
                <p className="font-black text-xl uppercase italic">Private</p>
                <p className="text-[10px] font-mono mt-2 uppercase leading-relaxed opacity-60">Restricted access protocol. Only authorized entities may interface with this pod.</p>
              </div>
            </button>
            <button 
              onClick={() => updateFormData({ visibility: 'public' })}
              className={`p-8 border-4 border-black text-left flex flex-col gap-4 transition-all ${
                formData.visibility === 'public' 
                ? 'bg-[#FF8A00] translate-x-1 translate-y-1 shadow-none' 
                : 'bg-white shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] hover:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:translate-x-1 hover:translate-y-1'
              }`}
            >
              <Globe size={28} />
              <div>
                <p className="font-black text-xl uppercase italic">Public</p>
                <p className="text-[10px] font-mono mt-2 uppercase leading-relaxed opacity-60">Open-source architectural node. Available for global network integration.</p>
              </div>
            </button>
          </div>
        </div>
      </div>

      <div className="pt-12 flex justify-between items-center">
        <button
          disabled={isSubmitting}
          onClick={handleBack}
          className="flex items-center gap-3 font-black uppercase text-xs hover:text-[#FF8A00] transition-colors disabled:opacity-30"
        >
          <ArrowLeft size={18} /> Revision required
        </button>
        <button
          disabled={isSubmitting}
          onClick={handleComplete}
          className="bg-black text-white font-black uppercase text-sm px-14 py-4 border-2 border-black neo-btn-shadow neo-btn-active transition-all disabled:opacity-50"
        >
          {isSubmitting ? 'Architecting...' : 'Assemble Pod'}
        </button>
      </div>
    </div>
  );
};

export default ConfiguratorPage;