
import React, { useRef } from 'react';
import { useRouter } from 'next/navigation';
import { Upload, X, FileText, ArrowLeft } from 'lucide-react';
import { useFormContext } from '@/context/FormContext';
import { Button, FormField, StepContainer } from '@/components/ui/neobrutalism-components';

const MaterialPage: React.FC = () => {
  const { formData, updateFormData } = useFormContext();
  const router = useRouter();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const newFiles = Array.from(e.target.files);
      updateFormData({ materialFiles: [...formData.materialFiles, ...newFiles] });
    }
  };

  const removeFile = (index: number) => {
    const updatedFiles = [...formData.materialFiles];
    updatedFiles.splice(index, 1);
    updateFormData({ materialFiles: updatedFiles });
  };

  const Footer = (
    <div className="flex justify-between items-center w-full">
      <button
        onClick={() => router.push('/')}
        className="flex items-center gap-3 font-black uppercase text-xs hover:text-[#FF8A00] transition-colors"
      >
        <ArrowLeft size={18} /> Back to core
      </button>
      <Button onClick={() => router.push('/configurator')}>Next: Configurator</Button>
    </div>
  );

  return (
    <StepContainer title="Material" footer={Footer}>
      <div className="space-y-8">
        <FormField label="Resource Upload" required>
          <div 
            onClick={() => fileInputRef.current?.click()}
            className="border-4 border-dashed border-black bg-gray-50 p-12 flex flex-col items-center justify-center gap-4 cursor-pointer hover:bg-white transition-all shadow-[6px_6px_0px_0px_#FF8A00]"
          >
            <div className="p-3 bg-white border-2 border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
              <Upload size={28} />
            </div>
            <p className="font-black uppercase text-sm">Drop files or click to build</p>
            <input type="file" multiple className="hidden" ref={fileInputRef} onChange={handleFileChange} />
          </div>
        </FormField>

        {formData.materialFiles.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-h-40 overflow-y-auto pr-2">
            {formData.materialFiles.map((file, idx) => (
              <div key={idx} className="flex items-center justify-between p-4 bg-white border-2 border-black shadow-[3px_3px_0px_0px_rgba(0,0,0,1)]">
                <div className="flex items-center gap-3 overflow-hidden">
                  <FileText className="text-[#FF8A00] shrink-0" size={18} />
                  <span className="font-mono text-[10px] font-bold truncate">{file.name}</span>
                </div>
                <X size={14} className="cursor-pointer hover:text-red-500" onClick={() => removeFile(idx)} />
              </div>
            ))}
          </div>
        )}
      </div>
    </StepContainer>
  );
};

export default MaterialPage;
