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
              file.name.split(".")[0], // Use filename without extension as title
              `Material ${i + 1}`,
              (progress) => {
                // Show progress for current file
                console.log(`Uploading ${file.name}: ${progress}%`);
              }
            );
            setUploadProgress({ current: i + 1, total: materialFiles.length });
          } catch (uploadError) {
            console.error(`Failed to upload material ${i + 1}:`, uploadError);
            // Continue with next material even if one fails
          }
        }
      }

      setIsSubmitting(false);
      setIsSuccess(true);
    } catch (err) {
      console.error("Error in handleComplete:", err);
      setError(err instanceof Error ? err.message : "Failed to create pod");
      setIsSubmitting(false);
    }
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
          <p className="text-xs font-mono font-bold uppercase">Pod: {formData.name || "Untitled"}</p>
          <p className="text-xs font-mono text-gray-600 mt-2">✓ Saved to database</p>
        </div>
        <button
          onClick={() => {
            resetForm();
            router.push("/dashboard/knowledge");
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

      {error && (
        <div className="mb-6 p-4 border-2 border-red-500 bg-red-50 flex gap-3">
          <AlertCircle size={20} className="text-red-500 flex-shrink-0 mt-0.5" />
          <div>
            <p className="font-bold text-red-900">Error</p>
            <p className="text-sm text-red-800">{error}</p>
          </div>
        </div>
      )}

      {uploadProgress && (
        <div className="mb-6 p-4 border-2 border-blue-500 bg-blue-50">
          <p className="font-mono text-sm font-bold uppercase mb-2">
            Uploading materials: {uploadProgress.current} / {uploadProgress.total}
          </p>
          <div className="w-full bg-gray-300 border border-black h-2">
            <div
              className="bg-[#FF8A00] h-full transition-all"
              style={{
                width: `${(uploadProgress.current / uploadProgress.total) * 100}%`,
              }}
            />
          </div>
        </div>
      )}

      <div className="flex-1 space-y-10">
        <div className="space-y-4">
          <label className="text-[10px] font-black font-mono uppercase tracking-[0.2em] text-gray-400">
            Node Visibility Protocols <span className="text-[#FF8A00]">*</span>
          </label>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <button
              onClick={() => updateFormData({ visibility: "private" })}
              className={`p-8 border-4 border-black text-left flex flex-col gap-4 transition-all ${
                formData.visibility === "private"
                  ? "bg-[#FF8A00] translate-x-1 translate-y-1 shadow-none"
                  : "bg-white shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] hover:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:translate-x-1 hover:translate-y-1"
              }`}
            >
              <Lock size={28} />
              <div>
                <p className="font-black text-xl uppercase italic">Private</p>
                <p className="text-[10px] font-mono mt-2 uppercase leading-relaxed opacity-60">Restricted access protocol. Only authorized entities may interface with this pod.</p>
              </div>
            </button>
            <button
              onClick={() => updateFormData({ visibility: "public" })}
              className={`p-8 border-4 border-black text-left flex flex-col gap-4 transition-all ${
                formData.visibility === "public"
                  ? "bg-[#FF8A00] translate-x-1 translate-y-1 shadow-none"
                  : "bg-white shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] hover:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:translate-x-1 hover:translate-y-1"
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
        <button disabled={isSubmitting} onClick={handleBack} className="flex items-center gap-3 font-black uppercase text-xs hover:text-[#FF8A00] transition-colors disabled:opacity-30">
          <ArrowLeft size={18} /> Revision required
        </button>
        <button disabled={isSubmitting} onClick={handleComplete} className="bg-black text-white font-black uppercase text-sm px-14 py-4 border-2 border-black neo-btn-shadow neo-btn-active transition-all disabled:opacity-50">
          {isSubmitting ? "Architecting..." : "Assemble Pod"}
        </button>
      </div>
    </div>
  );
};

export default ConfiguratorPage;
