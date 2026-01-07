"use client";

import { useEffect, useState, useCallback, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { motion } from "framer-motion";
import { useAuth } from "@/lib/auth-context";

/**
 * Email Verification Waiting Page
 * 
 * FLOW:
 * 1. User lands here after sign-up
 * 2. Shows waiting UI with email sent confirmation
 * 3. On page refresh/focus, checks if email is verified
 * 4. If verified → redirect to onboarding
 * 5. If not verified → continue showing waiting UI
 */

type WaitingState = "waiting" | "checking" | "verified" | "redirecting";

function VerifyWaitingContent() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const { user, refreshUser, isAuthenticated, loading: authLoading } = useAuth();

    const [state, setState] = useState<WaitingState>("waiting");
    const [email, setEmail] = useState<string>("");
    const [resendState, setResendState] = useState<"idle" | "sending" | "success" | "error">("idle");
    const [resendMessage, setResendMessage] = useState<string>("");

    // Get email from query params or user
    useEffect(() => {
        const emailParam = searchParams.get("email");
        if (emailParam) {
            setEmail(emailParam);
        } else if (user?.email) {
            setEmail(user.email);
        }
    }, [searchParams, user]);

    // Check verification status
    const checkVerificationStatus = useCallback(async () => {
        if (!isAuthenticated || authLoading) return;

        try {
            setState("checking");
            await refreshUser();
        } catch (error) {
            console.error("Error checking verification status:", error);
            setState("waiting");
        }
    }, [isAuthenticated, authLoading, refreshUser]);

    // Resend verification email
    const handleResendEmail = async () => {
        if (resendState === "sending") return;

        try {
            setResendState("sending");
            setResendMessage("");

            const response = await fetch("/api/auth/resend-verification", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ email: email || user?.email }),
            });

            if (!response.ok) {
                throw new Error("Failed to resend verification email");
            }

            setResendState("success");
            setResendMessage("Verification email sent successfully!");

            // Reset success message after 5 seconds
            setTimeout(() => {
                setResendState("idle");
                setResendMessage("");
            }, 5000);
        } catch (error) {
            console.error("Error resending verification email:", error);
            setResendState("error");
            setResendMessage("Failed to resend email. Please try again.");

            // Reset error message after 5 seconds
            setTimeout(() => {
                setResendState("idle");
                setResendMessage("");
            }, 5000);
        }
    };

    // Check if user is verified after refresh
    useEffect(() => {
        if (authLoading) return;

        if (user?.email_verified) {
            setState("verified");

            // Short delay then redirect to onboarding
            setTimeout(() => {
                setState("redirecting");
                router.push("/onboarding");
            }, 1500);
        } else if (state === "checking") {
            setState("waiting");
        }
    }, [user?.email_verified, router, state, authLoading]);

    // Check on mount and on focus
    useEffect(() => {
        if (authLoading) return;

        // Check immediately on mount
        checkVerificationStatus();

        // Check when page regains focus (user returns from email)
        const handleFocus = () => {
            checkVerificationStatus();
        };

        // Check periodically (every 5 seconds)
        const intervalId = setInterval(() => {
            checkVerificationStatus();
        }, 5000);

        window.addEventListener("focus", handleFocus);

        return () => {
            window.removeEventListener("focus", handleFocus);
            clearInterval(intervalId);
        };
    }, [checkVerificationStatus, authLoading]);

    // Redirect if not authenticated (and auth is done loading)
    useEffect(() => {
        if (!authLoading && !isAuthenticated) {
            router.push("/sign-in");
        }
    }, [isAuthenticated, authLoading, router]);

    // Show loading while auth is loading
    if (authLoading) {
        return (
            <div className="w-full max-w-md">
                <div className="relative">
                    <div className="absolute top-3 left-3 w-full h-full bg-[#FF8811] rounded-2xl" />
                    <div className="bg-white rounded-2xl p-8 relative border-[2px] border-[#2B2D42] flex items-center justify-center min-h-[300px]">
                        <div className="w-8 h-8 border-4 border-[#FF8811] border-t-transparent rounded-full animate-spin" />
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="w-full max-w-md">
            {/* Card Container with Brutalist shadow */}
            <div className="relative">
                {/* Brutalist offset shadow */}
                <div className="absolute top-3 left-3 w-full h-full bg-[#FF8811] rounded-2xl" />

                {/* Main Card */}
                <div className="bg-white rounded-2xl p-8 relative border-[2px] border-[#2B2D42]">
                    {/* Logo */}
                    <div className="text-center mb-6">
                        <h1 className="text-2xl font-bold font-[(family-name:var(--font-plus-jakarta-sans))]">
                            <span className="text-[#FF8811]">Ngasih</span>
                            <span className="text-[#2B2D42]">Tau</span>
                        </h1>
                    </div>

                    {/* Content based on state */}
                    {state === "verified" || state === "redirecting" ? (
                        <VerifiedContent />
                    ) : (
                        <WaitingContent
                            email={email}
                            isChecking={state === "checking"}
                            onRefresh={checkVerificationStatus}
                            onResend={handleResendEmail}
                            resendState={resendState}
                            resendMessage={resendMessage}
                        />
                    )}
                </div>
            </div>
        </div>
    );
}

export default function VerifyWaitingPage() {
    return (
        <Suspense fallback={
            <div className="w-full max-w-md">
                <div className="relative">
                    <div className="absolute top-3 left-3 w-full h-full bg-[#FF8811] rounded-2xl" />
                    <div className="bg-white rounded-2xl p-8 relative border-[2px] border-[#2B2D42] flex items-center justify-center min-h-[300px]">
                        <div className="w-8 h-8 border-4 border-[#FF8811] border-t-transparent rounded-full animate-spin" />
                    </div>
                </div>
            </div>
        }>
            <VerifyWaitingContent />
        </Suspense>
    );
}

// ===========================================================================
// SUB-COMPONENTS
// ===========================================================================

function WaitingContent({
    email,
    isChecking,
    onRefresh,
    onResend,
    resendState,
    resendMessage
}: {
    email: string;
    isChecking: boolean;
    onRefresh: () => void;
    onResend: () => void;
    resendState: "idle" | "sending" | "success" | "error";
    resendMessage: string;
}) {
    return (
        <>
            {/* Title */}
            <div className="text-center mb-8">
                <h2 className="text-3xl font-bold font-[(family-name:var(--font-plus-jakarta-sans))]">
                    <span className="text-[#2B2D42]">Check Your </span>
                    <span className="text-[#FF8811]">Email</span>
                </h2>
                <p className="text-gray-500 text-sm mt-3 font-[(family-name:var(--font-inter))]">
                    We've sent a verification link to
                </p>
                {email && (
                    <p className="text-[#2B2D42] font-semibold mt-1 font-[(family-name:var(--font-inter))]">
                        {email}
                    </p>
                )}
            </div>

            {/* Mail Icon Animation */}
            <div className="flex justify-center mb-8">
                <motion.div
                    animate={{
                        y: [0, -10, 0],
                    }}
                    transition={{
                        duration: 2,
                        repeat: Infinity,
                        ease: "easeInOut"
                    }}
                    className="relative"
                >
                    {/* Mail envelope icon */}
                    <div className="w-24 h-24 bg-[#FFFBF5] rounded-2xl border-2 border-[#2B2D42] flex items-center justify-center shadow-[4px_4px_0px_0px_rgba(255,136,17,1)]">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            className="w-12 h-12 text-[#FF8811]"
                        >
                            <rect width="20" height="16" x="2" y="4" rx="2" />
                            <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" />
                        </svg>
                    </div>

                    {/* Floating dots animation */}
                    <motion.div
                        animate={{ opacity: [0, 1, 0], y: [0, -20] }}
                        transition={{ duration: 2, repeat: Infinity, delay: 0 }}
                        className="absolute -top-2 left-1/4 w-2 h-2 bg-[#FF8811] rounded-full"
                    />
                    <motion.div
                        animate={{ opacity: [0, 1, 0], y: [0, -20] }}
                        transition={{ duration: 2, repeat: Infinity, delay: 0.5 }}
                        className="absolute -top-2 left-1/2 w-2 h-2 bg-[#2B2D42] rounded-full"
                    />
                    <motion.div
                        animate={{ opacity: [0, 1, 0], y: [0, -20] }}
                        transition={{ duration: 2, repeat: Infinity, delay: 1 }}
                        className="absolute -top-2 right-1/4 w-2 h-2 bg-[#FF8811] rounded-full"
                    />
                </motion.div>
            </div>

            {/* Instructions */}
            <div className="bg-[#FFFBF5] rounded-lg p-4 mb-6 border border-[#FF8811]/20">
                <p className="text-sm text-gray-600 font-[(family-name:var(--font-inter))] text-center">
                    Click the link in your email to verify your account.
                    <br />
                    <span className="text-gray-500">
                        This page will automatically update once verified.
                    </span>
                </p>
            </div>

            {/* Refresh Button */}
            <button
                onClick={onRefresh}
                disabled={isChecking}
                className="w-full cursor-pointer bg-[#FF8811] text-white py-3 rounded-lg font-semibold transition-all font-[family-name:var(--font-plus-jakarta-sans)] disabled:opacity-50 disabled:cursor-not-allowed border-[2px] border-[#2B2D42] shadow-[4px_4px_0px_0px_rgba(43,45,66,1)] hover:shadow-[2px_2px_0px_0px_rgba(43,45,66,1)] hover:translate-x-[2px] hover:translate-y-[2px] active:shadow-none active:translate-x-[4px] active:translate-y-[4px]"
            >
                {isChecking ? (
                    <span className="flex items-center justify-center gap-2">
                        <motion.div
                            animate={{ rotate: 360 }}
                            transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
                            className="w-5 h-5 border-2 border-white border-t-transparent rounded-full"
                        />
                        Checking...
                    </span>
                ) : (
                    "I've Verified My Email"
                )}
            </button>

            {/* Resend link */}
            <div className="text-center mt-4">
                <p className="text-gray-500 text-sm font-[family-name:var(--font-inter)]">
                    Didn't receive the email?{" "}
                    <button
                        onClick={onResend}
                        disabled={resendState === "sending"}
                        className="text-[#FF8811] font-semibold hover:text-[#FF8811]/80 transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {resendState === "sending" ? "Sending..." : "Resend"}
                    </button>
                </p>

                {/* Resend feedback message */}
                {resendMessage && (
                    <motion.p
                        initial={{ opacity: 0, y: -10 }}
                        animate={{ opacity: 1, y: 0 }}
                        className={`text-sm mt-2 font-[family-name:var(--font-inter)] ${resendState === "success" ? "text-green-600" : "text-red-600"
                            }`}
                    >
                        {resendMessage}
                    </motion.p>
                )}
            </div>
        </>
    );
}

function VerifiedContent() {
    return (
        <>
            {/* Title */}
            <div className="text-center mb-8">
                <h2 className="text-3xl font-bold font-[(family-name:var(--font-plus-jakarta-sans))]">
                    <span className="text-[#2B2D42]">Email </span>
                    <span className="text-[#FF8811]">Verified!</span>
                </h2>
                <p className="text-gray-500 text-sm mt-3 font-[(family-name:var(--font-inter))]">
                    Your account has been successfully verified
                </p>
            </div>

            {/* Success Animation */}
            <div className="flex justify-center mb-8">
                <motion.div
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    transition={{ type: "spring", duration: 0.5 }}
                    className="w-24 h-24 bg-[#FF8811] rounded-full flex items-center justify-center border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_rgba(43,45,66,1)]"
                >
                    <motion.svg
                        initial={{ pathLength: 0 }}
                        animate={{ pathLength: 1 }}
                        transition={{ duration: 0.5, delay: 0.2 }}
                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="white"
                        strokeWidth="3"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        className="w-12 h-12"
                    >
                        <motion.path
                            initial={{ pathLength: 0 }}
                            animate={{ pathLength: 1 }}
                            transition={{ duration: 0.5, delay: 0.2 }}
                            d="M5 13l4 4L19 7"
                        />
                    </motion.svg>
                </motion.div>
            </div>

            {/* Redirecting message */}
            <div className="text-center">
                <p className="text-sm text-gray-600 font-[(family-name:var(--font-inter))] flex items-center justify-center gap-2">
                    <motion.div
                        animate={{ rotate: 360 }}
                        transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
                        className="w-4 h-4 border-2 border-[#FF8811] border-t-transparent rounded-full"
                    />
                    Redirecting to setup...
                </p>
            </div>
        </>
    );
}
