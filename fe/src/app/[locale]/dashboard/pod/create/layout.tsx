"use client";

import React from 'react';
import { useTranslations } from 'next-intl';
import { FormProvider } from '@/context/FormContext';
import StepIndicator from '@/components/knowledge-pod/StepIndicator';

interface LayoutProps {
    children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
    const t = useTranslations('createPod');

    return (
        <FormProvider>
            <div className="min-h-screen flex flex-col md:flex-row bg-white">
                {/* SIDEBAR CONTAINER */}
                <aside className="hidden md:flex w-[380px] bg-white border-r-4 border-black p-10 flex-col shrink-0 h-screen top-0 z-20 overflow-hidden fixed left-0">
                    <div className="space-y-12 h-full flex flex-col">
                        {/* Logo & Brand Section */}
                        <div className="space-y-4 shrink-0">
                            <div className="w-12 h-12 bg-black flex items-center justify-center rotate-3 border-2 border-[#FF8A00] shadow-[4px_4px_0px_0px_rgba(255,138,0,1)]">
                                <span className="text-[#FF8A00] font-black text-2xl italic">K</span>
                            </div>
                            <div>
                                <h1 className="text-2xl font-black text-black tracking-tighter uppercase leading-none">
                                    Pod Architect
                                </h1>
                                <p className="mt-2 text-xs font-bold text-gray-400 uppercase leading-relaxed max-w-[240px]">
                                    Assemble your thoughts and materials into a structured Knowledge Pod.
                                </p>
                            </div>
                        </div>

                        {/* Vertical Step Indicator */}
                        <nav className="pl-2 flex-1 flex flex-col justify-center">
                            <StepIndicator />
                        </nav>

                      
                    </div>
                </aside>

                {/* MAIN CONTENT AREA */}
                <main className="flex-1 p-6 md:p-10 lg:p-12 flex justify-center items-start overflow-y-auto relative md:ml-[380px]">
                    <div className="w-full max-w-2xl relative z-10 py-10">
                        {/* Header Section */}
                        <header className="mb-10 text-left w-full">
                            <h1 className="text-3xl md:text-4xl font-black text-black tracking-tighter mb-2 font-[family-name:var(--font-plus-jakarta-sans)] uppercase">
                                {t('title')}
                            </h1>
                            <p className="text-sm font-bold text-gray-400 uppercase tracking-widest">System Configuration Module</p>
                        </header>

                        {/* Main Content Area */}
                        <div className="w-full flex flex-col relative">
                            <div className="flex-1 z-10">
                                {children}
                            </div>
                        </div>

                        {/* Footer Info */}
                        <div className="mt-20 flex flex-col md:flex-row justify-between items-center opacity-20 font-mono text-[9px] uppercase tracking-tighter px-2 gap-4">
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
        </FormProvider>
    );
};

export default Layout;

