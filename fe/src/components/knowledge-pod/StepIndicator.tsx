"use client";

import React from 'react';
import { Step } from '@/types';
import { useFormContext } from '@/context/FormContext';

const StepIndicator = () => {
  const { currentStep } = useFormContext();
  
  const steps = [
    { id: Step.General, label: '01', title: 'CORE' },
    { id: Step.Material, label: '02', title: 'ASSETS' },
    { id: Step.Configurator, label: '03', title: 'DEPLOY' },
  ];

  return (
    <div className="flex flex-col items-center w-full mb-16 relative z-5">
      <div className="flex items-center justify-center w-full relative max-w-md">
        {/* Progress Line Background */}
        <div className="absolute top-1/2 -translate-y-1/2 left-0 w-full h-1 bg-gray-200 z-0"></div>
        
        {/* Active Progress Line */}
        <div 
          className="absolute top-1/2 -translate-y-1/2 left-0 h-1 bg-[#FF8A00] transition-all duration-700 ease-in-out z-0"
          style={{ 
            width: currentStep === Step.General ? '0%' : currentStep === Step.Material ? '50%' : '100%' 
          }}
        ></div>
        
        <div className="flex items-center justify-between w-full relative z-10">
          {steps.map((step) => {
            const isCompleted = currentStep > step.id;
            const isActive = currentStep === step.id;
            
            return (
              <div key={step.id} className="flex flex-col items-center">
                <div
                  className={`w-6 h-6 md:w-12 md:h-12 border-4 transition-all duration-500 flex items-center justify-center font-black text-xs md:text-xl ${
                    isCompleted 
                      ? 'bg-[#FF8A00] text-black border-black' 
                      : isActive 
                        ? 'bg-white text-black border-black scale-110 shadow-[6px_6px_0px_0px_rgba(0,0,0,1)]' 
                        : 'bg-gray-100 text-gray-400 border-gray-200'
                  }`}
                >
                  {step.label}
                </div>
                <span className={`absolute -bottom-8 font-mono text-[9px] font-black tracking-widest uppercase transition-colors duration-300 ${isActive ? 'text-black' : 'text-gray-400'}`}>
                  {step.title}
                </span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default StepIndicator;