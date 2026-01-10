"use client";

import React, { createContext, useContext, useState, ReactNode } from 'react';
import { KnowledgePodData, Step } from '../types';

interface FormContextType {
  formData: KnowledgePodData;
  currentStep: Step;
  updateFormData: (updates: Partial<KnowledgePodData>) => void;
  setStep: (step: Step) => void;
  resetForm: () => void;
}

const initialData: KnowledgePodData = {
  name: '',
  description: '',
  materialFiles: [],
  visibility: 'private',
};

const FormContext = createContext<FormContextType | undefined>(undefined);

export const FormProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [formData, setFormData] = useState<KnowledgePodData>(initialData);
  const [currentStep, setCurrentStep] = useState<Step>(Step.Material);

  const updateFormData = (updates: Partial<KnowledgePodData>) => {
    setFormData((prev) => ({ ...prev, ...updates }));
  };

  const setStep = (step: Step) => setCurrentStep(step);

  const resetForm = () => {
    setFormData(initialData);
    setCurrentStep(Step.General);
  };

  return (
    <FormContext.Provider value={{ formData, currentStep, updateFormData, setStep, resetForm }}>
      {children}
    </FormContext.Provider>
  );
};

export const useFormContext = () => {
  const context = useContext(FormContext);
  if (!context) {
    throw new Error('useFormContext must be used within a FormProvider');
  }
  return context;
};
