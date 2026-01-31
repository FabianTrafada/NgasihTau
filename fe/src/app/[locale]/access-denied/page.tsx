'use client'

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { Home, ArrowLeft, Lock } from 'lucide-react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';

export default function AccessDeniedPage() {
    const t = useTranslations('accessDenied');
    const router = useRouter();
    const searchParams = useSearchParams();
    const { user } = useAuth();
    const fromPath = searchParams.get('from');

    return (
        <div className="min-h-screen w-full flex flex-col items-center justify-center p-6 text-center bg-[#FFFBF7] font-[family-name:var(--font-plus-jakarta-sans)]">

            {/* Branding Logo */}
            <div className="absolute top-8 left-8 flex items-center gap-1">
                <span className="text-2xl font-extrabold text-[#FF8811]">Ngasih</span>
                <span className="text-2xl font-extrabold text-black">Tau</span>
            </div>

            <div className="max-w-lg w-full flex flex-col items-center animate-in fade-in zoom-in duration-500">

                {/* Restricted Icon */}
                <div className="mb-8 p-6 bg-red-400 border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] rounded-xl">
                    <Lock className="w-12 h-12 text-white" strokeWidth={2.5} />
                </div>

                {/* Permission Subtext */}
                <p className="text-lg font-semibold text-gray-500 mb-3 tracking-tight font-[family-name:var(--font-inter)]">
                    {t('subtitle')}
                </p>

                {/* Main Title */}
                <h1 className="text-5xl md:text-6xl font-[900] text-gray-900 tracking-tight mb-12 leading-tight">
                    {t('title')}
                </h1>

                {/* Action Buttons */}
                <div className="flex flex-col sm:flex-row items-center gap-4 w-full sm:w-auto">
                    <Link
                        href="/dashboard"
                        className="w-full sm:w-auto px-8 py-3.5 bg-[#FF8811] border-2 border-black text-white font-bold rounded-xl shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:translate-y-[2px] hover:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] active:translate-y-[4px] active:shadow-none transition-all flex items-center justify-center gap-2"
                    >
                        <Home className="w-5 h-5" />
                        {t('backToDashboard')}
                    </Link>
                </div>
            </div>

            {/* Footer */}
            <div className="text-center absolute bottom-6 ">
                <p className="text-xs text-gray-400 font-[family-name:var(--font-inter)]">
                    {t('contactSupport')} <a href="mailto:support@ngasihtau.com" className="text-[#FF8811] font-bold hover:underline">{t('support')}</a>.
                </p>
            </div>

        </div>

    );
}
