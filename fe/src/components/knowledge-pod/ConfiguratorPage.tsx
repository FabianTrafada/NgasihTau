'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Lock, Globe, CheckCircle, ArrowLeft, Loader2 } from 'lucide-react';
import { useFormContext } from '@/context/FormContext';
import { Step } from '@/types';
import { createPod } from '@/lib/api/pod';

const ConfiguratorPage: React.FC = () => {
  const { formData, updateFormData, resetForm, setCurrentStep } = useFormContext();
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleComplete = async () => {
    setIsSubmitting(true);
    setError(null);
    try {
      await createPod({
        name: formData.name || 'Untitled Pod',
        description: formData.description || '',
        visibility: (formData.visibility as 'public' | 'private') || 'public',
        categories: ['Uncategorized'], 
        tags: ['general'],
      });
      setIsSuccess(true);
    } catch (err) {
      console.error(err);
      setError("Failed to create pod. Please try again.");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleBack = () => {
    setCurrentStep(Step.Material);
  };

  if (isSuccess) {
    return (
      <div className="h-full flex flex-col items-center justify-center space-y-6 animate-in zoom-in-95 duration-500">
        <div className="w-20 h-20 bg-green-100 rounded-full flex items-center justify-center text-green-600">
          <CheckCircle size={40} />
        </div>
        <div className="text-center space-y-2">
          <h2 className="text-2xl font-bold text-[#2B2D42]">Pod Created Successfully!</h2>
          <p className="text-gray-500">Your knowledge pod is now ready.</p>
        </div>
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 w-full max-w-sm text-center">
           <p className="text-sm font-medium text-[#2B2D42]">{formData.name || 'Untitled'}</p>
        </div>
        <button
          onClick={() => { 
            resetForm(); 
            router.push('/dashboard/pods'); 
          }}
          className="bg-[#2B2D42] text-white font-medium text-sm px-8 py-3 rounded-md hover:bg-opacity-90 transition-all"
        >
          Go to Dashboard
        </button>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col">
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-[#2B2D42]">Configuration</h2>
        <p className="text-gray-500 text-sm mt-1">Finalize your pod settings before creation.</p>
      </div>

      <div className="flex-1 space-y-8">
        <div className="space-y-3">
          <label className="text-sm font-bold text-[#2B2D42]">
            Visibility <span className="text-[#FF8811]">*</span>
          </label>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <button 
              onClick={() => updateFormData({ visibility: 'private' })}
              className={`p-6 border-[2px] rounded-lg text-left flex flex-col gap-3 transition-all ${
                formData.visibility === 'private' 
                ? 'border-[#FF8811] bg-orange-50 shadow-[3px_3px_0px_0px_rgba(255,136,17,1)]' 
                : 'border-gray-200 bg-white hover:border-[#FF8811] hover:shadow-sm'
              }`}
            >
              <div className={`p-2 rounded-md w-fit ${formData.visibility === 'private' ? 'bg-[#FF8811] text-white' : 'bg-gray-100 text-gray-600'}`}>
                <Lock size={20} />
              </div>
              <div>
                <p className="font-bold text-[#2B2D42]">Private</p>
                <p className="text-xs text-gray-500 mt-1 leading-relaxed">Only you and invited users can access this pod.</p>
              </div>
            </button>
            
            <button 
              onClick={() => updateFormData({ visibility: 'public' })}
              className={`p-6 border-[2px] rounded-lg text-left flex flex-col gap-3 transition-all ${
                formData.visibility === 'public' 
                ? 'border-[#FF8811] bg-orange-50 shadow-[3px_3px_0px_0px_rgba(255,136,17,1)]' 
                : 'border-gray-200 bg-white hover:border-[#FF8811] hover:shadow-sm'
              }`}
            >
              <div className={`p-2 rounded-md w-fit ${formData.visibility === 'public' ? 'bg-[#FF8811] text-white' : 'bg-gray-100 text-gray-600'}`}>
                <Globe size={20} />
              </div>
              <div>
                <p className="font-bold text-[#2B2D42]">Public</p>
                <p className="text-xs text-gray-500 mt-1 leading-relaxed">Anyone can view and learn from this pod.</p>
              </div>
            </button>
          </div>
        </div>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 text-red-600 text-sm rounded-md border border-red-100">
          {error}
        </div>
      )}

      <div className="pt-8 flex justify-between items-center border-t border-gray-100 mt-auto">
        <button
          disabled={isSubmitting}
          onClick={handleBack}
          className="flex items-center gap-2 text-sm font-bold text-gray-500 hover:text-[#2B2D42] transition-colors disabled:opacity-50 px-4 py-2 rounded-md hover:bg-gray-100"
        >
          <ArrowLeft size={16} /> Back
        </button>
        <button
          disabled={isSubmitting}
          onClick={handleComplete}
          className="bg-[#2B2D42] text-white font-bold text-sm px-8 py-3 rounded-md hover:bg-opacity-90 transition-all disabled:opacity-70 flex items-center gap-2 shadow-sm"
        >
          {isSubmitting ? (
            <>
              <Loader2 size={16} className="animate-spin" /> Creating...
            </>
          ) : (
            'Create Pod'
          )}
        </button>
      </div>
    </div>
  );
};

export default ConfiguratorPage;