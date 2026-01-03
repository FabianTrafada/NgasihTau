"use client";

import React from 'react';
import { FormProvider, useFormContext } from '@/context/FormContext';
import { Step } from '@/types';
import Layout from '@/components/knowledge-pod/Layout';
import GeneralStep from '@/components/knowledge-pod/GeneralPage';
import MaterialStep from '@/components/knowledge-pod/MateriPage';
import ConfiguratorStep from '@/components/knowledge-pod/ConfiguratorPage';

// Komponen internal untuk menangani logika perubahan tampilan antar langkah (steps)
const UploadFlow = () => {
  const { currentStep } = useFormContext();

  return (
    <>
      {currentStep === Step.General && <GeneralStep />}
      {currentStep === Step.Material && <MaterialStep />}
      {currentStep === Step.Configurator && <ConfiguratorStep />}
    </>
  );
};

// Halaman utama Upload yang menyediakan Context bagi semua komponen di dalamnya
const UploadPage = () => {
  return (
    <FormProvider>
      <Layout>
        <UploadFlow />
      </Layout>
    </FormProvider>
  );
};

export default UploadPage;
