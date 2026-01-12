"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useAuth } from "@/lib/auth-context";
import { PublicOnlyRoute } from "@/components/auth";
import { useTranslations } from "next-intl";

/**
 * Sign In Page Component
 * 
 * AUTHENTICATION FLOW:
 * 1. User enters email + password OR clicks Google sign-in
 * 2. Form submits to useAuth().login() OR initiateGoogleLogin()
 * 3. If 2FA enabled → Show 2FA input
 * 4. If no 2FA → Redirect to dashboard
 * 5. Errors are displayed from auth context
 */
export default function SignInPage() {
    const t = useTranslations('auth');
    const router = useRouter();
    const { login, verify2FA, initiateGoogleLogin, loading, error, clearError } = useAuth();

    // Form state
    const [showPassword, setShowPassword] = useState(false);
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");

    // 2FA state
    const [requires2FA, setRequires2FA] = useState(false);
    const [tempToken, setTempToken] = useState("");
    const [totpCode, setTotpCode] = useState("");

    // Local error state (for form validation)
    const [formError, setFormError] = useState("");

    // Translation errors
    const validCodeError = t('validCode');

    /**
     * Handle initial login form submission
     */
    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        console.log('[SignIn] ========== LOGIN FORM SUBMIT ==========');
        console.log('[SignIn] Email:', email);

        setFormError("");
        clearError();

        try {
            console.log('[SignIn] Calling login API...');
            const result = await login(email, password);
            console.log('[SignIn] Login result:', result);

            if (result.requires2FA && result.tempToken) {
                // User has 2FA enabled - show 2FA form
                console.log('[SignIn] 2FA required - showing 2FA form');
                setRequires2FA(true);
                setTempToken(result.tempToken);
            } else {
                // Login successful - redirect to dashboard
                console.log('[SignIn] ✅ Login successful! Redirecting to dashboard...');

                // Small delay to ensure auth context state is updated
                await new Promise(resolve => setTimeout(resolve, 100));

                console.log('[SignIn] Executing router.push("/dashboard")');
                router.push("/dashboard");
            }
        } catch (error) {
            console.error('[SignIn] ❌ Login failed:', error);
            // Error is already set in auth context
        }

        console.log('[SignIn] ========== END LOGIN SUBMIT ==========');
    };

    /**
     * Handle 2FA verification
     */
    const handle2FASubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        console.log('[SignIn] ========== 2FA SUBMIT ==========');
        console.log('[SignIn] TOTP Code:', totpCode);

        setFormError("");
        clearError();

        if (totpCode.length < 6) {
            setFormError(validCodeError);
            return;
        }

        try {
            console.log('[SignIn] Verifying 2FA...');
            await verify2FA(tempToken, totpCode);
            console.log('[SignIn] ✅ 2FA verified! Redirecting to dashboard...');

            // Small delay to ensure auth context state is updated
            await new Promise(resolve => setTimeout(resolve, 100));

            console.log('[SignIn] Executing router.push("/dashboard")');
            router.push("/dashboard");
        } catch (error) {
            console.error('[SignIn] ❌ 2FA verification failed:', error);
            // Error is already set in auth context
        }

        console.log('[SignIn] ========== END 2FA SUBMIT ==========');
    };

    /**
     * Display error message (from context or local validation)
     */
    const displayError = error?.message || formError;

    // =========================================================================
    // 2FA VERIFICATION UI
    // =========================================================================
    if (requires2FA) {
        return (
            <div className="w-full max-w-sm">
                <div className="relative">
                    <div className="absolute top-4 left-4 w-full h-full bg-[#FF8811] rounded-2xl" />
                    <div className="bg-white rounded-2xl p-8 relative border-2 border-[#2B2D42]">
                        {/* Logo */}
                        <div className="text-center mb-6">
                            <h1 className="text-2xl font-bold">
                                <span className="text-[#FF8811]">Ngasih</span>
                                <span className="text-[#2B2D42]">Tau</span>
                            </h1>
                        </div>

                        {/* 2FA Header */}
                        <div className="text-center mb-8">
                            <h2 className="text-3xl font-bold font-[family-name:var(--font-plus-jakarta-sans)]">
                                <span className="text-[#2B2D42]">{t('twoFactorAuth').split(' ')[0]} </span>
                                <span className="text-[#FF8811]">{t('twoFactorAuth').split(' ').slice(1).join(' ')}</span>
                            </h2>
                            <p className="text-gray-500 text-sm mt-2 font-[family-name:var(--font-inter)]">
                                {t('enterAuthCode')}
                            </p>
                        </div>

                        {/* Error Display */}
                        {displayError && (
                            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
                                {displayError}
                            </div>
                        )}

                        {/* 2FA Form */}
                        <form onSubmit={handle2FASubmit} className="space-y-5">
                            <div>
                                <label
                                    htmlFor="totp"
                                    className="block text-sm font-semibold text-[#2B2D42] mb-2"
                                >
                                    {t('verificationCode')}
                                </label>
                                <input
                                    type="text"
                                    id="totp"
                                    value={totpCode}
                                    onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                                    className="w-full text-[#2B2D42] px-4 py-3 border-[2px] border-gray-300 rounded-md bg-white focus:outline-none transition-all text-center text-2xl tracking-widest focus:border-[#FF8811] focus:shadow-[3px_3px_0px_0px_rgba(255,136,17,1)]"
                                    placeholder="000000"
                                    maxLength={6}
                                    autoComplete="one-time-code"
                                    required
                                />
                            </div>

                            <button
                                type="submit"
                                disabled={loading}
                                className="w-full cursor-pointer bg-[#FF8811] text-white py-3 rounded-lg font-bold border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[2px_2px_0px_0px_#2B2D42] active:shadow-none active:translate-x-[4px] active:translate-y-[4px] transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                {loading ? t('verifying') : t('verify')}
                            </button>
                        </form>

                        {/* Back Button */}
                        <button
                            onClick={() => {
                                setRequires2FA(false);
                                setTempToken("");
                                setTotpCode("");
                            }}
                            className="w-full mt-4 text-gray-500 text-sm hover:text-gray-700"
                        >
                            {t('backToLogin')}
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    // =========================================================================
    // LOGIN UI
    // =========================================================================
    return (
        <PublicOnlyRoute redirectTo="/dashboard">
            <div className="w-full max-w-sm">
                {/* Card Container with Brutalist shadow */}
                <div className="relative">
                    {/* Brutalist offset shadow */}
                    <div className="absolute top-3 left-3 w-full h-full bg-[#FF8811] rounded-2xl" />

                    {/* Main Card */}
                    <div className="bg-white rounded-2xl p-6 relative border-2 border-[#2B2D42]">

                        {/* Logo */}
                        <div className="text-center mb-4">
                            <h1 className="text-xl font-bold ">
                                <span className="text-[#FF8811] ">Ngasih</span>
                                <span className="text-[#2B2D42]">Tau</span>
                            </h1>
                        </div>

                        {/* Welcome Text */}
                        <div className="text-center mb-4">
                            <h2 className="text-2xl font-bold font-[family-name:var(--font-plus-jakarta-sans)]">
                                <span className="text-[#2B2D42] ">{t('welcomeBack').split(' ')[0]} </span>
                                <span className="text-[#FF8811]">{t('welcomeBack').split(' ').slice(1).join(' ')}</span>
                            </h2>
                            <p className="text-gray-500 text-xs mt-1 font-[family-name:var(--font-inter)]">
                                {t('enterCredentials')}
                            </p>
                        </div>

                        {/* Error Display */}
                        {displayError && (
                            <div className="mb-3 p-2 bg-red-50 border border-red-200 rounded-lg text-red-600 text-xs">
                                {displayError}
                            </div>
                        )}

                        {/* Form */}
                        <form onSubmit={handleSubmit} className="space-y-3">
                            {/* Email Field */}
                            <div>
                                <label
                                    htmlFor="email"
                                    className="block text-xs font-semibold text-[#2B2D42] mb-1 font-[family-name:var(--font-plus-jakarta-sans)]"
                                >
                                    {t('email')}
                                </label>
                                <input
                                    type="email"
                                    id="email"
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    className="w-full text-[#2B2D42] px-3 py-2 border-[2px] border-gray-300 rounded-md bg-white focus:outline-none transition-all font-[family-name:var(--font-inter)] focus:border-[#FF8811] focus:shadow-[3px_3px_0px_0px_rgba(255,136,17,1)] text-sm"
                                    placeholder=""
                                    required
                                />
                            </div>

                            {/* Password Field */}
                            <div>
                                <label
                                    htmlFor="password"
                                    className="block text-xs font-semibold text-[#2B2D42] mb-1 font-[family-name:var(--font-plus-jakarta-sans)]"
                                >
                                    {t('password')}
                                </label>
                                <div className="relative">
                                    <input
                                        type={showPassword ? "text" : "password"}
                                        id="password"
                                        value={password}
                                        onChange={(e) => setPassword(e.target.value)}
                                        className="w-full text-[#2B2D42] px-3 py-2 border-[2px] border-gray-300 rounded-md bg-white focus:outline-none transition-all pr-10 font-[family-name:var(--font-inter)] focus:border-[#FF8811] focus:shadow-[3px_3px_0px_0px_rgba(255,136,17,1)] text-sm"
                                        placeholder=""
                                        required
                                    />
                                    <button
                                        type="button"
                                        onClick={() => setShowPassword(!showPassword)}
                                        className="cursor-pointer absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 transition-colors"
                                    >
                                        {showPassword ? (
                                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                                <path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
                                                <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                            </svg>
                                        ) : (
                                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
                                                <path strokeLinecap="round" strokeLinejoin="round" d="M3.98 8.223A10.477 10.477 0 001.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.45 10.45 0 0112 4.5c4.756 0 8.773 3.162 10.065 7.498a10.523 10.523 0 01-4.293 5.774M6.228 6.228L3 3m3.228 3.228l3.65 3.65m7.894 7.894L21 21m-3.228-3.228l-3.65-3.65m0 0a3 3 0 10-4.243-4.243m4.242 4.242L9.88 9.88" />
                                            </svg>
                                        )}
                                    </button>
                                </div>
                            </div>

                            {/* Forgot Password Link */}
                            <div className="text-right">
                                <Link
                                    href="/forgot-password"
                                    className="text-[#FF8811] text-xs font-medium hover:text-[#FF8811]/80 transition-colors font-[family-name:var(--font-inter)] hover:underline hover:underline-offset-2"
                                >
                                    {t('forgotPassword')}
                                </Link>
                            </div>

                            <button
                                type="submit"
                                disabled={loading}
                                className="w-full cursor-pointer bg-[#FF8811] text-white py-2 rounded-lg font-bold border-2 border-[#2B2D42] shadow-[3px_3px_0px_0px_#2B2D42] hover:translate-x-[1px] hover:translate-y-[1px] hover:shadow-[2px_2px_0px_0px_#2B2D42] active:shadow-none active:translate-x-[3px] active:translate-y-[3px] transition-all font-[family-name:var(--font-plus-jakarta-sans)] disabled:opacity-50 disabled:cursor-not-allowed text-sm"
                            >
                                {loading ? t('loggingIn') : t('signInButton')}
                            </button>
                        </form>

                        {/* Divider */}
                        <div className="flex items-center my-4">
                            <div className="flex-1 border-t-2 border-gray-200"></div>
                            <span className="px-3 text-gray-500 text-xs font-bold font-[family-name:var(--font-inter)]">{t('orLoginWith')}</span>
                            <div className="flex-1 border-t-2 border-gray-200"></div>
                        </div>

                        {/* Google Button */}
                        <button
                            type="button"
                            onClick={initiateGoogleLogin}
                            disabled={loading}
                            className="w-full cursor-pointer flex items-center justify-center gap-2 px-4 py-2 bg-white border-2 border-[#2B2D42] rounded-lg shadow-[3px_3px_0px_0px_#2B2D42] hover:translate-x-[1px] hover:translate-y-[1px] hover:shadow-[2px_2px_0px_0px_#2B2D42] active:shadow-none active:translate-x-[3px] active:translate-y-[3px] transition-all font-[family-name:var(--font-inter)] disabled:opacity-50 disabled:cursor-not-allowed text-sm"
                        >
                            <svg width="18" height="18" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
                                <path d="M19.8055 10.2275C19.8055 9.51803 19.7477 8.83551 19.6399 8.17969H10.2002V12.0494H15.6006C15.3681 13.2994 14.6259 14.3578 13.5379 15.0672V17.5766H16.8055C18.7054 15.8369 19.8055 13.2713 19.8055 10.2275Z" fill="#4285F4" />
                                <path d="M10.2002 20.0003C12.9002 20.0003 15.1719 19.1044 16.8055 17.5763L13.5379 15.0669C12.6215 15.6669 11.4845 16.0213 10.2002 16.0213C7.59478 16.0213 5.38234 14.2631 4.57289 11.9004H1.20312V14.4907C2.83212 17.7594 6.26791 20.0003 10.2002 20.0003Z" fill="#34A853" />
                                <path d="M4.57289 11.9003C4.37289 11.3003 4.25952 10.659 4.25952 10.0003C4.25952 9.34155 4.37289 8.70023 4.57289 8.10023V5.50977H1.20312C0.437559 6.85977 0 8.38629 0 10.0003C0 11.6143 0.437559 13.1408 1.20312 14.4908L4.57289 11.9003Z" fill="#FBBC04" />
                                <path d="M10.2002 3.97898C11.5997 3.97898 12.8545 4.47898 13.8386 5.45896L16.8767 2.42091C15.168 0.919312 12.8963 0 10.2002 0C6.26791 0 2.83212 2.24091 1.20312 5.50958L4.57289 8.10004C5.38234 5.73725 7.59478 3.97898 10.2002 3.97898Z" fill="#E94235" />
                            </svg>
                            <span className="font-bold text-[#2B2D42]">Google</span>
                        </button>

                        {/* Register Link */}
                        <p className="text-center mt-4 text-gray-600 text-xs font-[family-name:var(--font-inter)]">
                            {t('noAccount')}{" "}
                            <Link
                                href="/sign-up"
                                className="text-[#FF8811] font-semibold hover:text-[#FF8811]/80 transition-colors hover:underline hover:underline-offset-2"
                            >
                                {t('registerNow')}
                            </Link>
                        </p>
                    </div>
                </div>
            </div>
        </PublicOnlyRoute>
    );
}