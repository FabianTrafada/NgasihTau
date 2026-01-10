"use client";

import React from 'react';
import StepIndicator from './StepIndicator';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  return (
    <div className="min-h-screen flex flex-col md:flex-row bg-[#FFF8F0]">
      {/* SIDEBAR CONTAINER */}
      <aside className="w-full md:w-[380px] bg-white border-r-4 border-black p-10 flex flex-col shrink-0 h-auto md:h-screen md:sticky md:top-0 z-20 shadow-[8px_0px_0px_0px_rgba(0,0,0,0.05)]">
        <div className="space-y-12">
          {/* Logo & Brand Section */}
          <div className="space-y-4">
            <div className="w-12 h-12 bg-black flex items-center justify-center rotate-3 border-2 border-[#FF8A00] shadow-[4px_4px_0px_0px_rgba(255,138,0,1)]">
              <span className="text-[#FF8A00] font-black text-2xl italic">K</span>
            </div>
            <div>
              <h1 className="text-2xl font-black text-black tracking-tighter uppercase leading-none">
                Find Customer
              </h1>
              <p className="mt-2 text-xs font-bold text-gray-400 uppercase leading-relaxed max-w-[240px]">
                Please select the customer and confirm their details to proceed.
              </p>
            </div>
          </div>

          {/* Vertical Step Indicator */}
          <nav className="pl-2">
            <StepIndicator />
          </nav>
        </div>
        
        {/* Sidebar Footer Accent */}
        <div className="mt-auto pt-10 opacity-10 font-mono text-[8px] tracking-[0.3em] uppercase">
          Build_v2.5_Stable
        </div>
      </aside>

      {/* MAIN CONTENT AREA - NO CARD CONTAINER */}
      <main className="flex-1 p-6 md:p-16 lg:p-24 flex justify-center items-start overflow-y-auto industrial-grid relative">
        <div className="w-full max-w-3xl relative z-10">
          {/* Konten langsung di sini tanpa pembungkus kartu putih */}
          <div className="relative">
            {children}
          </div>

          {/* Secondary Footer Info - Tetap ada di bawah */}
          <div className="mt-20 flex flex-col md:flex-row justify-between items-center opacity-40 font-mono text-[9px] uppercase tracking-tighter px-2 gap-4">
            <div className="flex gap-6">
              <span className="flex items-center gap-2"><div className="w-1.5 h-1.5 bg-green-500 rounded-full"></div> SYSTEM_ID: POD-ARCH-01</span>
              <span>LATENCY: 14MS</span>
              <span>ENCRYPTION: AES-256</span>
            </div>
            <div className="font-bold">© {new Date().getFullYear()} // KNOWLEDGE_ARCHITECT</div>
          </div>
        </div>
      </main>
    </div>
  );
};

export default Layout;