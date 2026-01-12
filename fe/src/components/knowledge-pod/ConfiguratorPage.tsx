"use client";

import React, { useState } from "react";
import { useRouter } from "next/navigation";
import { Lock, Globe, CheckCircle, ArrowLeft, AlertCircle } from "lucide-react";
import { useFormContext } from "@/context/FormContext";
import { Step } from "@/types";
import { createPod, CreatePodInput } from "@/lib/api/pod";
import { uploadMaterial } from "@/lib/api/uploadMaterial";

const ConfiguratorPage: React.FC = () => {
  const { formData, updateFormData, resetForm, setCurrentStep } = useFormContext();
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [uploadProgress, setUploadProgress] = useState<{ current: number; total: number } | null>(null);

  const handleComplete = async () => {
    try {
      setIsSubmitting(true);
      setError(null);
      setUploadProgress(null);

      // Validate required fields
      if (!formData.name?.trim()) {
        throw new Error("Pod name is required");
      }
      if (!formData.description?.trim()) {
        throw new Error("Pod description is required");
      }
      if (!formData.visibility) {
        throw new Error("Pod visibility is required");
      }

      // Step 1: Create pod
      const podInput: CreatePodInput = {
        name: formData.name,
        description: formData.description,
        visibility: formData.visibility,
        categories: formData.categories || [],
        tags: formData.tags || [],
      };

      console.log("Creating pod with:", podInput);
      const newPod = await createPod(podInput);
      console.log("Pod created:", newPod);

      // Step 2: Upload materials if any
      const materialFiles = formData.materialFiles || [];
      if (materialFiles.length > 0) {
        setUploadProgress({ current: 0, total: materialFiles.length });

        for (let i = 0; i < materialFiles.length; i++) {
          const file = materialFiles[i];
          try {
            await uploadMaterial(
              file,
              newPod.id,
              file.name,
              `Material ${i + 1}`
            );
            setUploadProgress({ current: i + 1, total: materialFiles.length });
          } catch (uploadError) {
            console.error(`Failed to upload ${file.name}:`, uploadError);
            // Continue with other files even if one fails
          }
        }
      }

      setIsSubmitting(false);
      setIsSuccess(true);

      // Navigate to pod page after success
      setTimeout(() => {
        resetForm();
        router.push(`/dashboard/my-pods`);
      }, 2000);
    } catch (err: any) {
      console.error("Error creating pod:", err);
      setError(err.message || "Failed to create pod. Please try again.");
      setIsSubmitting(false);
    }
  };

  const handleBack = () => {
    setCurrentStep(Step.Material);
  };

  if (isSuccess) {
    return (
      <div className="h-full flex flex-col items-center justify-center space-y-6 py-6">
        <div className="w-16 h-16 bg-[#FF8A00] border-4 border-black flex items-center justify-center shadow-[6px_6px_0px_0px_rgba(0,0,0,1)]">
          <CheckCircle size={32} strokeWidth={3} />
        </div>
        <div className="text-center">
          <h2 className="text-2xl font-black uppercase tracking-tight">Pod Created!</h2>
          <p className="font-mono text-gray-400 text-[10px] mt-2 tracking-widest uppercase">
            {formData.name}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col min-h-screen">
      <div className="mb-6">
        <h2 className="text-2xl font-black uppercase tracking-tight mb-2">Configurator</h2>
        <div className="h-1 w-16 bg-[#FF8A00] border border-black"></div>
      </div>

      <div className="flex-1 space-y-6">
        {/* Error Display */}
        {error && (
          <div className="bg-red-50 border-2 border-red-500 p-3 flex items-start gap-3">
            <AlertCircle className="text-red-500 shrink-0 mt-0.5" size={18} />
            <div>
              <p className="font-bold text-red-700 text-xs">Error</p>
              <p className="text-red-600 text-[10px] mt-1">{error}</p>
            </div>
          </div>
        )}

        {/* Upload Progress */}
        {uploadProgress && (
          <div className="bg-[#FF8A00] border-2 border-black p-3">
            <p className="font-bold text-[10px] uppercase mb-2">
              Uploading materials... {uploadProgress.current}/{uploadProgress.total}
            </p>
            <div className="w-full bg-white border-2 border-black h-1.5">
              <div
                className="bg-black h-full transition-all"
                style={{
                  width: `${(uploadProgress.current / uploadProgress.total) * 100}%`,
                }}
              />
            </div>
          </div>
        )}

        {/* Visibility Selection */}
        <div className="space-y-3">
          <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-400">
            Node Visibility Protocols <span className="text-[#FF8A00]">*</span>
          </label>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <button
              onClick={() => updateFormData({ visibility: "private" })}
              className={`p-5 border-4 border-black text-left flex flex-col gap-3 transition-all ${formData.visibility === "private"
                ? "bg-[#FF8A00] translate-x-1 translate-y-1 shadow-none"
                : "bg-white shadow-[6px_6px_0px_0px_rgba(0,0,0,1)] hover:shadow-[3px_3px_0px_0px_rgba(0,0,0,1)] hover:translate-x-1 hover:translate-y-1"
                }`}
            >
              <div className={`p-1.5 rounded-md w-fit ${formData.visibility === 'private' ? 'bg-[#FF8811] text-white' : 'bg-gray-100 text-gray-600'}`}>
                <Lock size={16} />
              </div>
              <div>
                <p className="font-bold text-[#2B2D42] text-sm">Private</p>
                <p className="text-[10px] text-gray-500 mt-1 leading-relaxed">Only you and invited users can access this pod.</p>
              </div>
            </button>
            <button
              onClick={() => updateFormData({ visibility: "public" })}
              className={`p-5 border-4 border-black text-left flex flex-col gap-3 transition-all ${formData.visibility === "public"
                ? "bg-[#FF8A00] translate-x-1 translate-y-1 shadow-none"
                : "bg-white shadow-[6px_6px_0px_0px_rgba(0,0,0,1)] hover:shadow-[3px_3px_0px_0px_rgba(0,0,0,1)] hover:translate-x-1 hover:translate-y-1"
                }`}
            >
              <div className={`p-1.5 rounded-md w-fit ${formData.visibility === 'public' ? 'bg-[#FF8811] text-white' : 'bg-gray-100 text-gray-600'}`}>
                <Globe size={16} />
              </div>
              <div>
                <p className="font-bold text-[#2B2D42] text-sm">Public</p>
                <p className="text-[10px] text-gray-500 mt-1 leading-relaxed">Anyone can view and learn from this pod.</p>
              </div>
            </button>
          </div>
        </div>
      </div>

      <div className="pt-8 flex justify-between items-center">
        <button disabled={isSubmitting} onClick={handleBack} className="flex items-center gap-2 font-black uppercase text-[10px] hover:text-[#FF8A00] transition-colors disabled:opacity-30">
          <ArrowLeft size={16} /> Revision required
        </button>

        <button
          disabled={isSubmitting || !formData.visibility}
          onClick={handleComplete}
          className="bg-black text-white font-black uppercase text-xs px-8 py-3 border-2 border-black shadow-[4px_4px_0px_0px_#FF8A00] hover:shadow-none hover:translate-x-[4px] hover:translate-y-[4px] transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? 'Architecting...' : 'Assemble Pod'}
        </button>
      </div>
    </div>
  );
};

export default ConfiguratorPage;