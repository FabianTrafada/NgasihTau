"use client";

import React, { useEffect, useState, Suspense } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { useTranslations, useLocale } from "next-intl";

function VerifyEmailContent() {
  const t = useTranslations('auth.verifyEmail');
  const searchParams = useSearchParams();
  const router = useRouter();
  const locale = useLocale();
  const [isLoading, setIsLoading] = useState(true);
  const [isSuccess, setIsSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const verifyEmail = async () => {
      try {
        const token = searchParams.get("token");

        if (!token) {
          setError(t('error'));
          setIsLoading(false);
          return;
        }

        // Ensure API URL has trailing slash
        const apiUrl = process.env.NEXT_PUBLIC_API_URL?.endsWith('/')
          ? process.env.NEXT_PUBLIC_API_URL
          : `${process.env.NEXT_PUBLIC_API_URL}/`;

        const response = await fetch(`${apiUrl}api/v1/auth/verify-email`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ token }),
        });

        if (response.ok) {
          setIsSuccess(true);
          setError(null);

          // Redirect to sign-in page after 2 seconds
          // User needs to login once to get tokens, then auto-redirect to onboarding
          setTimeout(() => {
            router.push(`/sign-in?verified=true`);
          }, 2000);
        } else {
          const errorData = await response.json().catch(() => ({}));
          // Check nested error structure from API
          const errorMessage = errorData?.error?.message
            || errorData?.message
            || t('expired');
          setError(errorMessage);
          console.error("Verify email error:", errorData);
        }
      } catch (err) {
        console.error("Verify email exception:", err);
        setError(t('error'));
      } finally {
        setIsLoading(false);
      }
    };

    verifyEmail();
  }, [searchParams, router]);

  return (
    <div className="w-full max-w-md mx-auto">
      <div className="relative">
        <div className="absolute top-2 left-2 w-full h-full bg-[#FF8811] rounded-2xl" />

        {/* Main Card */}
        <div className="bg-white rounded-2xl p-8 relative border-2 border-[#2B2D42]">
          {/* Logo */}
          <div className="text-center mb-6">
            <h1 className="text-2xl font-bold ">
              <span className="text-[#FF8811] ">Ngasih</span>
              <span className="text-[#2B2D42]">Tau</span>
            </h1>
          </div>

          {/* Content */}
          <div className="text-center mb-8">
            {isLoading && (
              <>
                <h2 className="text-3xl font-bold font-[family-name:var(--font-plus-jakarta-sans)] mb-4">
                  <span className="text-[#2B2D42]">{t('verifying')}</span>
                  <span className="text-[#FF8811]"> {t('email')}</span>
                </h2>
                <p className="text-gray-500 text-sm font-[family-name:var(--font-inter)]">{t('pleaseWait')}</p>
                <div className="flex justify-center mt-6">
                  <div className="w-8 h-8 border-4 border-[#FF8811] border-t-[#2B2D42] rounded-full animate-spin" />
                </div>
              </>
            )}

            {isSuccess && !isLoading && (
              <>
                <h2 className="text-3xl font-bold font-[family-name:var(--font-plus-jakarta-sans)] mb-4">
                  <span className="text-[#2B2D42]">{t('email')} </span>
                  <span className="text-[#FF8811]">{t('verified')}!</span>
                </h2>
                <p className="text-gray-500 text-sm font-[family-name:var(--font-inter)] mb-6">
                  {t('success')} Redirecting to sign in...
                </p>
                <div className="flex justify-center">
                  <svg className="w-16 h-16 text-[#FF8811]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <div className="flex justify-center mt-4">
                  <div className="w-5 h-5 border-2 border-[#FF8811] border-t-transparent rounded-full animate-spin" />
                </div>
              </>
            )}

            {error && !isLoading && (
              <>
                <h2 className="text-3xl font-bold font-[family-name:var(--font-plus-jakarta-sans)] mb-4">
                  <span className="text-[#2B2D42]">Verifikasi Gagal</span>
                </h2>
                <p className="text-[#2B2D42] text-sm font-[family-name:var(--font-inter)] mb-6">{error}</p>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function LoadingFallback() {
  return (
    <div className="w-full max-w-md mx-auto">
      <div className="relative">
        <div className="absolute top-2 left-2 w-full h-full bg-[#FF8811] rounded-2xl" />
        <div className="bg-white rounded-2xl p-8 relative border-2 border-[#2B2D42]">
          <div className="text-center mb-6">
            <h1 className="text-2xl font-bold ">
              <span className="text-[#FF8811] ">Ngasih</span>
              <span className="text-[#2B2D42]">Tau</span>
            </h1>
          </div>
          <div className="flex justify-center py-8">
            <div className="w-8 h-8 border-4 border-[#FF8811] border-t-transparent rounded-full animate-spin" />
          </div>
        </div>
      </div>
    </div>
  );
}

export default function VerifyEmail() {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <VerifyEmailContent />
    </Suspense>
  );
}
