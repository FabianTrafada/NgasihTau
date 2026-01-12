"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import RecentPodCard from "@/components/dashboard/RecentPodCard";
import ChatbotLogCard from "@/components/dashboard/ChatbotLogCard";
import { ProtectedRoute } from "@/components/auth";
import { useAuth } from "@/lib/auth-context";
import { getOnboardingStatus } from "@/lib/api/interests";
import Link from "next/link";
import { useTranslations, useLocale } from "next-intl";

/**
 * Dashboard Page
 * 
 * This is a PROTECTED page - requires authentication and email verification.
 * The ProtectedRoute component handles:
 * 1. Showing loading state while checking auth
 * 2. Redirecting to /sign-in if not authenticated
 * 3. Redirecting to /verify-waiting if email not verified
 * 4. Rendering content if authenticated and verified
 * 
 * This component also checks onboarding status:
 * - If onboarding not complete, redirect to /onboarding
 */
export default function DashboardPage() {
    return (
        <ProtectedRoute>
            <DashboardContent />
        </ProtectedRoute>
    );
}

/**
 * Dashboard Content - only rendered when authenticated and email verified
 */
function DashboardContent() {
    const router = useRouter();
    const locale = useLocale();
    const { user } = useAuth();
    const t = useTranslations('dashboard');
    const [checkingOnboarding, setCheckingOnboarding] = useState(true);

    // Check onboarding status on mount
    useEffect(() => {
        const checkOnboarding = async () => {
            try {
                const result = await getOnboardingStatus();

                // If onboarding not completed, redirect
                if (result?.completed !== true) {
                    console.log('[Dashboard] Onboarding not completed - redirecting');
                    router.replace(`/${locale}/onboarding`);
                    return;
                }

                console.log('[Dashboard] Onboarding completed - showing dashboard');
                setCheckingOnboarding(false);
            } catch (error) {
                console.error('[Dashboard] Error checking onboarding:', error);
                // On error, assume onboarding needed or let user proceed
                // In production, you might want different error handling
                setCheckingOnboarding(false);
            }
        };

        checkOnboarding();
    }, [router, locale]);

    // Show loading while checking onboarding
    if (checkingOnboarding) {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <div className="w-8 h-8 border-4 border-[#FF8811] border-t-transparent rounded-full animate-spin" />
            </div>
        );
    }

    return (
        <div className="flex w-full ">
            {/* Main Content Area */}
            <div className="flex-1 w-full px-4 sm:px-6 lg:px-8 py-4 flex flex-col gap-4 max-w-none">

                {/* Welcome Section - Now uses real user name */}
                <div>
                    <h1 className="text-1xl md:text-2xl font-family-name:var(--font-plus-jakarta-sans) sm:text-2xl font-bold text-[#2B2D42] mb-1">
                        {t('welcomeMessage', { name: user?.name || 'User' })}
                    </h1>
                    <p className="text-xs sm:text-base text-gray-500 font-family-name:var(--font-inter)">{t('readyToLearn')}</p>
                </div>

                {/* Recent Pods */}
                <section>
                    <div className="flex items-center justify-between mb-4">
                        <h2 className="text-base sm:text-lg font-family-name:var(--font-plus-jakarta-sans)  font-bold text-[#2B2D42]">{t('recentPods')}</h2>

                        <Link href={'/dashboard/pod'}>
                            <button className="text-xs text-gray-400 hover:text-[#FF8811] cursor-pointer font-bold">{t('seeAll')}</button>
                        </Link>
                    </div>
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                        <RecentPodCard
                            title="Cara Belajar Mobil Kopling (99% bisa 1% nya hanya tuhan ....)"
                            description="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
                            fileCount={12}
                            date="2 Jan 25"
                        />
                        <RecentPodCard
                            title="Cara Belajar Mobil Kopling (99% bisa 1% nya hanya tuhan ....)"
                            description="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
                            fileCount={12}
                            date="2 Jan 25"
                        />
                    </div>
                </section>

                {/* Last Chatbot Logs */}
                <section>
                    <div className="flex items-center justify-between mb-4">
                        <h2 className="text-base sm:text-lg font-bold text-[#2B2D42]">{t('lastChatbotLogs')}</h2>
                        <Link href={'/dashboard/chatbot-logs'} className="text-xs text-gray-400 hover:text-[#FF8811] font-bold">{t('seeAll')}</Link>
                    </div>
                    <div className="space-y-3">
                        <ChatbotLogCard
                            title="Mempelajari materi limit dengan analogi"
                            snippet="Berikan saya Trik cepat pahan limit.."
                        />
                        <ChatbotLogCard
                            title="Mempelajari materi limit dengan analogi"
                            snippet="Berikan saya Trik cepat pahan limit.."
                        />
                        <ChatbotLogCard
                            title="Mempelajari materi limit dengan analogi"
                            snippet="Berikan saya Trik cepat pahan limit.."
                        />
                    </div>
                </section>
            </div>

            {/* Right Sidebar */}
        </div>
    );
}