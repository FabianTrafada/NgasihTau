"use client";

import React from 'react';
import { Step } from '@/types';
import { useFormContext } from '@/context/FormContext';
import { useRouter } from 'next/navigation';

const StepIndicator = () => {
  const { currentStep } = useFormContext();
  const router = useRouter();

  const steps = [
    { id: Step.General, label: '01', title: 'General' },
    { id: Step.Material, label: '02', title: 'Material' },
    { id: Step.Configurator, label: '03', title: 'Configurator' },
  ];

  return (
    <div className="flex flex-col gap-10 relative mt-4 z-0 pointer-events-auto">
      {steps.map((step, index) => {
        const isCompleted = currentStep > step.id;
        const isActive = currentStep === step.id;
        const isLast = index === steps.length - 1;

        return (
          <div key={index} className="relative flex items-center gap-6 group">
            {/* Vertical Line Connector */}
            {!isLast && (
              <div
                className={`absolute left-[15px] top-[32px] w-[2px] h-[40px] z-0 transition-colors duration-500 ${
                  isCompleted ? 'bg-[#FF8A00]' : 'bg-gray-100'
                }`}
              />
            )}

            {/* Step Circle/Indicator */}
            <div
              className={`relative z-10 w-8 h-8 rounded-full border-2 transition-all duration-500 flex items-center justify-center font-black text-[10px] shrink-0 ${
                isCompleted
                  ? 'bg-[#FF8A00] text-black border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]'
                  : isActive
                    ? 'bg-black text-white border-black shadow-[4px_4px_0px_0px_rgba(255,138,0,0.4)] scale-110'
                    : 'bg-white text-gray-300 border-gray-100'
              }`}
            >
              {isCompleted ? '✓' : step.label}
            </div>

            {/* Step Label Content */}
            <div className="flex flex-col">
              <span
                className={`font-black uppercase tracking-widest text-[11px] transition-colors duration-300 ${isActive ? 'text-black' : 'text-gray-300'}`}
              >
                {step.title}
              </span>
              {isActive && (
                <span className="font-mono text-[7px] uppercase tracking-widest text-[#FF8A00] absolute -bottom-3 left-14 whitespace-nowrap">
                  [ ACTIVE_NODE ]
                </span>
              )}
            </div>
          </div>
        );
      })}

      {/* Button Back */}
      <button
        onClick={() => router.push('/pod')}
        className="mt-4 py-2 px-4 bg-white border-2 border-black text-black text-[11px] font-black uppercase tracking-widest hover:bg-gray-100 transition-colors self-start"
      >
        ← Back to Knowledge Spot
      </button>
    </div>
  );
};

export default StepIndicator;