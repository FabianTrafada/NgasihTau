'use client';

import React, { createContext, useContext, useState, ReactNode } from 'react';

interface FormData {
  title?: string;
  description?: string;
  name?: string;
  visibility?: string;
  categories?: string[];
  tags?: string[];
  materialFiles?: File[];
  [key: string]: unknown;
}

interface FormContextType {
  formData: FormData;
  updateFormData: (data: Partial<FormData>) => void;
  resetForm: () => void;
  currentStep: number;
  setCurrentStep: (step: number) => void;
}

const FormContext = createContext<FormContextType | undefined>(undefined);

export const FormProvider = ({ children }: { children: ReactNode }) => {
  const [formData, setFormData] = useState<FormData>({});
  const [currentStep, setCurrentStep] = useState(1);

  const updateFormData = (data: Partial<FormData>) => {
    setFormData(prev => ({ ...prev, ...data }));
  };

  const resetForm = () => {
    setFormData({});
    setCurrentStep(1);
  };

  return (
    <FormContext.Provider value={{
      formData,
      updateFormData,
      resetForm,
      currentStep,
      setCurrentStep
    }}>
      {children}
    </FormContext.Provider>
  );
};

export const useFormContext = () => {
  const context = useContext(FormContext);
  if (!context) {
    throw new Error('useFormContext must be used within FormProvider');
  }
  return context;
};