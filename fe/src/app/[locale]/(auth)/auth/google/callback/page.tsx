// =============================================================================
// GOOGLE OAUTH CALLBACK PAGE
// =============================================================================
// This page handles the redirect from Google after user authorizes.
// Google redirects here with an authorization code in the URL.
//
// Flow:
// 1. User clicks "Sign in with Google" → Redirected to Google
// 2. User authorizes → Google redirects here with ?code=xxx
// 3. This page extracts the code → Sends to backend
// 4. Backend exchanges code for user data → Returns JWT tokens
// 5. User is redirected to dashboard

"use client";

import { useEffect, useState, useRef } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { useTranslations } from "next-intl";

export default function GoogleCallbackPage() {

    const t = useTranslations('auth.google.callback');

    const router = useRouter();
    const searchParams = useSearchParams();
    const { handleGoogleCallback, verify2FA, error } = useAuth();

    // State for UI
    const [status, setStatus] = useState<"loading" | "2fa" | "error">("loading");
    const [tempToken, setTempToken] = useState<string>("");
    const [twoFactorCode, setTwoFactorCode] = useState("");
    const [localError, setLocalError] = useState<string>("");

    // Prevent double execution in React Strict Mode
    const hasProcessed = useRef(false);

    // Process the OAuth callback
    useEffect(() => {
        const processCallback = async () => {
            // Prevent double execution
            if (hasProcessed.current) return;
            hasProcessed.current = true;

            // Extract code from URL
            const code = searchParams.get("code");
            const errorParam = searchParams.get("error");

            // Handle Google errors
            if (errorParam) {
                setLocalError(
                    errorParam === "access_denied"
                        ? t('cancelled')
                        : `Google sign-in failed: ${errorParam}`
                );
                setStatus("error");
                return;
            }

            // No code means invalid callback
            if (!code) {
                setLocalError(t('noCode'));
                setStatus("error");
                return;
            }

            try {
                // Exchange code for tokens
                const result = await handleGoogleCallback(code);

                if (result.requires2FA && result.tempToken) {
                    // User has 2FA enabled, show 2FA form
                    setTempToken(result.tempToken);
                    setStatus("2fa");
                } else {
                    // Success! Redirect to dashboard
                    router.push("/dashboard");
                }
            } catch (err) {
                setLocalError(
                    err instanceof Error ? err.message : t('authFailed')
                );
                setStatus("error");
            }
        };

        processCallback();
    }, [searchParams, handleGoogleCallback, router]);

    // Handle 2FA submission
    const handle2FASubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLocalError("");

        try {
            await verify2FA(tempToken, twoFactorCode);
            router.push("/dashboard");
        } catch (err) {
            setLocalError(
                err instanceof Error ? err.message : t('invalidCode')
            );
        }
    };

    // Loading state
    if (status === "loading") {
        return (
            <div className="min-h-screen flex items-center justify-center bg-[#F9F7F3]">
                <div className="text-center">
                    {/* Loading Spinner */}
                    <div className="inline-flex items-center justify-center w-16 h-16 mb-4">
                        <svg
                            className="animate-spin h-12 w-12 text-[#FF8811]"
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                        >
                            <circle
                                className="opacity-25"
                                cx="12"
                                cy="12"
                                r="10"
                                stroke="currentColor"
                                strokeWidth="4"
                            />
                            <path
                                className="opacity-75"
                                fill="currentColor"
                                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                            />
                        </svg>
                    </div>
                    <h2 className="text-xl font-bold text-[#2B2D42] font-[family-name:var(--font-plus-jakarta-sans)]">
                        {t('authenticating')}
                    </h2>
                    <p className="text-gray-500 mt-2 font-[family-name:var(--font-inter)]">
                        {t('pleaseWait')}
                    </p>
                </div>
            </div>
        );
    }

    // 2FA required state
    if (status === "2fa") {
        return (
            <div className="min-h-screen flex items-center justify-center bg-[#F9F7F3] px-4">
                <div className="w-full max-w-md bg-white rounded-2xl border-2 border-[#2B2D42] shadow-[8px_8px_0px_0px_#2B2D42] p-8">
                    <div className="text-center mb-6">
                        {/* Lock Icon */}
                        <div className="inline-flex items-center justify-center w-16 h-16 bg-[#FF8811]/10 rounded-full mb-4">
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                fill="none"
                                viewBox="0 0 24 24"
                                strokeWidth={1.5}
                                stroke="currentColor"
                                className="w-8 h-8 text-[#FF8811]"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z"
                                />
                            </svg>
                        </div>
                        <h2 className="text-2xl font-bold text-[#2B2D42] font-[family-name:var(--font-plus-jakarta-sans)]">
                            {t('twoFactorAuth')}
                        </h2>
                        <p className="text-gray-500 mt-2 font-[family-name:var(--font-inter)]">
                            {t('enterCode')}
                        </p>
                    </div>

                    {/* Error Message */}
                    {(localError || error) && (
                        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
                            <p className="text-red-600 text-sm font-[family-name:var(--font-inter)]">
                                {localError || error?.message}
                            </p>
                        </div>
                    )}

                    <form onSubmit={handle2FASubmit} className="space-y-4">
                        <div>
                            <input
                                type="text"
                                value={twoFactorCode}
                                onChange={(e) => setTwoFactorCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                                placeholder="000000"
                                className="w-full px-4 py-3 text-center text-2xl font-mono tracking-[0.5em] bg-white border-2 border-[#2B2D42] rounded-lg focus:outline-none focus:border-[#FF8811] transition-colors"
                                maxLength={6}
                                autoFocus
                            />
                        </div>

                        <button
                            type="submit"
                            disabled={twoFactorCode.length !== 6}
                            className="w-full cursor-pointer bg-[#FF8811] text-white py-3 rounded-lg font-bold border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[2px_2px_0px_0px_#2B2D42] active:shadow-none active:translate-x-[4px] active:translate-y-[4px] transition-all font-[family-name:var(--font-plus-jakarta-sans)] disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {t('verifyButton')}
                        </button>
                    </form>

                    <button
                        onClick={() => router.push("/sign-in")}
                        className="w-full mt-4 text-gray-500 hover:text-[#FF8811] font-[family-name:var(--font-inter)] transition-colors"
                    >
                        Cancel and go back
                    </button>
                </div>
            </div>
        );
    }

    // Error state
    return (
        <div className="min-h-screen flex items-center justify-center bg-[#F9F7F3] px-4">
            <div className="w-full max-w-md bg-white rounded-2xl border-2 border-[#2B2D42] shadow-[8px_8px_0px_0px_#2B2D42] p-8 text-center">
                {/* Error Icon */}
                <div className="inline-flex items-center justify-center w-16 h-16 bg-red-50 rounded-full mb-4">
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        strokeWidth={1.5}
                        stroke="currentColor"
                        className="w-8 h-8 text-red-500"
                    >
                        <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z"
                        />
                    </svg>
                </div>

                <h2 className="text-2xl font-bold text-[#2B2D42] mb-2 font-[family-name:var(--font-plus-jakarta-sans)]">
                    Sign In Failed
                </h2>
                <p className="text-gray-500 mb-6 font-[family-name:var(--font-inter)]">
                    {localError || error?.message || "Something went wrong"}
                </p>

                <button
                    onClick={() => router.push("/sign-in")}
                    className="w-full cursor-pointer bg-[#FF8811] text-white py-3 rounded-lg font-bold border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[2px_2px_0px_0px_#2B2D42] active:shadow-none active:translate-x-[4px] active:translate-y-[4px] transition-all font-[family-name:var(--font-plus-jakarta-sans)]"
                >
                    Try Again
                </button>
            </div>
        </div>
    );
}
