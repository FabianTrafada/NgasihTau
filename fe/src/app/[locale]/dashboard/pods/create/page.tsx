"use client";

import { useFormContext } from '@/context/FormContext';
import { Step } from '@/types';
import GeneralStep from '@/components/knowledge-pod/GeneralPage';
import MaterialStep from '@/components/knowledge-pod/MateriPage';
import ConfiguratorStep from '@/components/knowledge-pod/ConfiguratorPage';
import { ProtectedRoute } from '@/components/auth';

// Halaman utama Upload yang menampilkan step sesuai currentStep dari FormContext
export default function UploadPage() {
  return (
    <ProtectedRoute>
      <UploadPageContent />
    </ProtectedRoute>
  );
}

function UploadPageContent() {
  const { currentStep } = useFormContext();

  return (
    <>
      {currentStep === Step.General && <GeneralStep />}
      {currentStep === Step.Material && <MaterialStep />}
      {currentStep === Step.Configurator && <ConfiguratorStep />}
    </>
  );
}