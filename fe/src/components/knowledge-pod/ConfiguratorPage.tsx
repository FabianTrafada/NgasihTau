
import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Lock, Globe, CheckCircle, ArrowLeft } from 'lucide-react';
import { useFormContext } from '@/context/FormContext';
import { Button, FormField, StepContainer } from '@/components/ui/neobrutalism-components';

const ConfiguratorPage: React.FC = () => {
  const { formData, updateFormData, resetForm } = useFormContext();
  const router = useRouter();  
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const handleComplete = async () => {
    setIsSubmitting(true);
    await new Promise(r => setTimeout(r, 2000));
    setIsSubmitting(false);
    setIsSuccess(true);
  };

  if (isSuccess) {
    return (
      <div className="h-full flex flex-col items-center justify-center space-y-8 py-10">
        <div className="w-20 h-20 bg-[#FF8A00] border-4 border-black flex items-center justify-center shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]">
          <CheckCircle size={40} strokeWidth={3} />
        </div>
        <div className="text-center">
          <h2 className="text-3xl font-black uppercase italic">Architected</h2>
          <p className="font-mono text-gray-400 text-[10px] mt-2 tracking-widest uppercase">Pod: {formData.name}</p>
        </div>
        <Button variant="dark" onClick={() => { resetForm(); router.push('/'); }}>Initialize New Build</Button>
      </div>
    );
  }

  const Footer = (
    <div className="flex justify-between items-center w-full">
      <button
        disabled={isSubmitting}
        onClick={() => router.push('/material')}
        className="flex items-center gap-3 font-black uppercase text-xs hover:text-[#FF8A00] transition-colors disabled:opacity-30"
      >
        <ArrowLeft size={18} /> Revision required
      </button>
      <Button disabled={isSubmitting} variant="dark" onClick={handleComplete}>
        {isSubmitting ? 'Architecting...' : 'Assemble Pod'}
      </Button>
    </div>
  );

  return (
    <StepContainer title="Configurator" footer={Footer}>
      <FormField label="Node Visibility Protocols" required>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {[
            { id: 'private', icon: Lock, label: 'Private', desc: 'Restricted access protocol.' },
            { id: 'public', icon: Globe, label: 'Public', desc: 'Open-source architectural node.' }
          ].map((node) => (
            <button 
              key={node.id}
              onClick={() => updateFormData({ visibility: node.id as any })}
              className={`p-6 border-4 border-black text-left flex flex-col gap-3 transition-all ${
                formData.visibility === node.id 
                ? 'bg-[#FF8A00] translate-x-1 translate-y-1 shadow-none' 
                : 'bg-white shadow-[6px_6px_0px_0px_rgba(0,0,0,1)] hover:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]'
              }`}
            >
              <node.icon size={24} />
              <div>
                <p className="font-black text-lg uppercase italic">{node.label}</p>
                <p className="text-[10px] font-mono mt-1 uppercase opacity-60">{node.desc}</p>
              </div>
            </button>
          ))}
        </div>
      </FormField>
    </StepContainer>
  );
};

export default ConfiguratorPage;
