"use client";

import { LoadingScreen } from "@/components/Onboarding/LoadingScreen";
import { WorkspaceSetup } from "@/components/Onboarding/workspaceSetup";
import { getOnboardingStatus } from "@/lib/api/interests";
import { useAuth } from "@/lib/auth-context";
import { useRouter } from "next/navigation";
import { useState, useEffect, useCallback } from "react";
import { useLocale } from "next-intl";

type OnboardingState = "checking" | "loading" | "setup" | "completing";

export default function OnboardingPage() {
    const router = useRouter();
    const locale = useLocale();
    const { user, isAuthenticated, loading: authLoading } = useAuth();
    const [state, setState] = useState<OnboardingState>("checking");

    useEffect(() => {
        if (authLoading) return;

        if (!isAuthenticated) {
            router.replace('/sign-in');
            return;
        }

        // Redirect to verify-waiting if email not verified
        if (user && !user.email_verified) {
            console.warn('[Onboarding] Email not verified - redirecting to verify-waiting');
            // Use relative path - router handles locale prefix automatically
            router.replace(`/verify-waiting?email=${encodeURIComponent(user.email)}`);
            return;
        }

        // Check onboarding status
        const checkOnboardingStatus = async () => {
            try {
                const result = await getOnboardingStatus();

                // Check if completed is truthy (handle various response formats)
                if (result?.completed === true) {
                    // Already completed, redirect to dashboard
                    console.log("Onboarding completed, redirecting to dashboard");
                    // Use relative path - router handles locale prefix automatically
                    router.replace('/dashboard');
                    return; // Don't change state, just redirect
                }

                // Not completed - show initial loading for 2 seconds, then show setup
                console.log("Onboarding not completed, showing setup");
                setState("loading");
                setTimeout(() => {
                    setState("setup");
                }, 2000);
            } catch (error) {
                console.error("Error checking onboarding status:", error);
                // On error (likely 401 or network), redirect to dashboard
                // The dashboard will handle auth check properly
                // Use relative path - router handles locale prefix automatically
                router.replace('/dashboard');
            }
        };

        checkOnboardingStatus();
    }, [router, isAuthenticated, authLoading, user, locale]);

    // Handle completion - show loading for 4 seconds then redirect
    const handleComplete = useCallback(() => {
        setState("completing");

        setTimeout(() => {
            // Use relative path - router handles locale prefix automatically
            router.push('/dashboard');
        }, 4000);
    }, [router, locale]);

    // Render based on state
    switch (state) {
        case "checking":
            // Minimal loading while checking status
            return (
                <div className="flex items-center justify-center min-h-screen bg-[#FFFBF5]">
                    <div className="w-8 h-8 border-4 border-[#FF8811] border-t-transparent rounded-full animate-spin" />
                </div>
            );

        case "loading":
            return <LoadingScreen message="Getting ready with your Workspace" />;

        case "setup":
            return <WorkspaceSetup onComplete={handleComplete} />;

        case "completing":
            return <LoadingScreen message="Setting up your interests... Redirecting to dashboard" />;

        default:
            return null;
    }
}