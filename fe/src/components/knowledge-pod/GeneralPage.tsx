
import React from 'react';
import { useRouter } from 'next/navigation';
import { useFormContext } from '@/context/FormContext';
import { Button, FormField, StepContainer } from '@/components/ui/neobrutalism-components';

const GeneralPage: React.FC = () => {
  const { formData, updateFormData } = useFormContext();
  const router = useRouter();

  const handleNext = () => {
    if (formData.name.trim() && formData.description.trim()) {
      router.push('/dashboard/pod/create/material');
    } else {
      alert('Required: Knowledge Pod Name & Description.');
    }
  };

  const Footer = (
    <div className="flex justify-end">
      <Button onClick={handleNext}>Next: Build Materials</Button>
    </div>
  );

  return (
    <StepContainer title="General" footer={Footer}>
      <div className="space-y-8">
        <FormField label="Knowledge Pod Name" required>
          <input
            type="text"
            value={formData.name}
            onChange={(e) => updateFormData({ name: e.target.value })}
            placeholder="Enter pod name..."
            className="w-full bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#FF8A00] focus:outline-none focus:shadow-none transition-all font-bold text-lg"
          />
        </FormField>

        <FormField label="Pod Description" required>
          <textarea
            value={formData.description}
            onChange={(e) => updateFormData({ description: e.target.value })}
            placeholder="Provide architectural insights for this pod..."
            rows={6}
            className="w-full bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#FF8A00] focus:outline-none focus:shadow-none transition-all font-medium resize-none text-base"
          />
        </FormField>
      </div>
    </StepContainer>
  );
};

export default GeneralPage;
