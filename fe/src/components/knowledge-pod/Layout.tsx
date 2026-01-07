"use client";

import React from 'react';
import StepIndicator from './StepIndicator';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  return (
    <div className="min-h-screen industrial-grid flex flex-col items-center py-8 md:py-16 px-4">
      <div className="w-full max-w-4xl flex flex-col items-center">
        {/* Header Section */}
        <header className="mb-10 text-center w-full">
          <h1 className="text-4xl md:text-5xl font-black text-black tracking-tighter mb-3 font-[family-name:var(--font-plus-jakarta-sans)]">
            Knowledge Pods
          </h1>
        </header>

        {/* Step Indicator - Sekarang otomatis sinkron dengan context */}
        <StepIndicator />

        {/* Main Content Area */}
        <div className="w-full bg-white border-4 border-black p-8 md:p-12 shadow-[12px_12px_0px_0px_rgba(0,0,0,1)] min-h-[500px] flex flex-col relative overflow-hidden">
          {/* Decorative Corner Element */}
          <div className="absolute top-0 right-0 w-16 h-16 bg-[#FF8A00] border-l-4 border-b-4 border-black flex items-center justify-center rotate-45 translate-x-8 -translate-y-8"></div>
          
          <main className="flex-1 z-10">
            {children}
          </main>
        </div>

        {/* Technical Footer */}
      </div>
    </div>
  );
};

export default Layout;