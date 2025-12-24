"use client";

import React, { useEffect, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { sendVerificationEmail, getCurrentUser } from "@/lib/auth";

const VerifyEmail = () => {
  const searchParams = useSearchParams();
  const router = useRouter();

  const [isLoading, setIsLoading] = useState(true);
  const [isVerified, setIsVerified] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [resendSent, setResendSent] = useState(false);
  const [resendLoading, setResendLoading] = useState(false);
  const [email, setEmail] = useState("");

  // Handle token-based verification (from email link)
  useEffect(() => {
    const verifyEmailFromToken = async () => {
      try {
        const token = searchParams.get("token");
        const emailParam = searchParams.get("email");

        if (emailParam) {
          setEmail(emailParam);
        }

        if (!token) {
          // No token, check if email is already verified via user data
          try {
            const user = await getCurrentUser();
            if (user.email_verified) {
              setIsVerified(true);
              setTimeout(() => {
                router.push("/dashboard");
              }, 2000);
            }
          } catch (err) {
            // User might not be authenticated yet, just show waiting screen
            console.log("Could not fetch user data");
          }
          setIsLoading(false);
          return;
        }

        // Token found, verify it
        const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}api/v1/auth/verify-email`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ token }),
        });

        if (response.ok) {
          setIsVerified(true);
          setError(null);
          // Auto redirect to dashboard after 2 seconds
          setTimeout(() => {
            router.push("/dashboard");
          }, 2000);
        } else {
          const errorData = await response.json().catch(() => ({}));
          setError(errorData.message || "Link sudah kadaluarsa atau tidak valid");
        }
      } catch (err) {
        setError("Terjadi kesalahan saat memverifikasi email");
      } finally {
        setIsLoading(false);
      }
    };

    verifyEmailFromToken();
  }, [searchParams, router]);

  // Send verification email
  const handleResendEmail = async () => {
    setResendLoading(true);
    setError(null);

    try {
      // Use the sendVerificationEmail function from auth.ts
      // This already includes the correct Authorization header via apiClient
      await sendVerificationEmail();

      setResendSent(true);
      // Reset the resend success message after 5 seconds
      setTimeout(() => {
        setResendSent(false);
      }, 5000);
    } catch (err) {
      console.error("Resend email error:", err);

      const errorMessage = err instanceof Error ? err.message : "Gagal mengirim email verifikasi";

      // Check if email is already verified
      if (errorMessage.includes("already verified")) {
        setIsVerified(true);
        setError(null);
        // Auto redirect to dashboard after 2 seconds
        setTimeout(() => {
          router.push("/dashboard");
        }, 2000);
      } else {
        setError(errorMessage);
      }
    } finally {
      setResendLoading(false);
    }
  };

  return (
    <div className="w-full max-w-md mx-auto">
      <div className="relative">
        <div className="absolute top-3 left-3 w-full h-full bg-[#FF8811] rounded-2xl" />

        {/* Main Card */}
        <div className="bg-white rounded-2xl p-8 relative border-[2px] border-[#2B2D42]">
          {/* Logo */}
          <div className="text-center mb-6">
            <h1 className="text-2xl font-bold">
              <span className="text-[#FF8811]">Ngasih</span>
              <span className="text-[#2B2D42]">Tau</span>
            </h1>
          </div>

          {/* Content */}
          <div className="text-center mb-8">
            {/* Loading with token - verifying */}
            {isLoading && searchParams.get("token") && (
              <>
                <h2 className="text-3xl font-bold font-[family-name:var(--font-plus-jakarta-sans)] mb-4">
                  <span className="text-[#2B2D42]">Memverifikasi</span>
                  <span className="text-[#FF8811]"> Email</span>
                </h2>
                <p className="text-gray-500 text-sm font-[family-name:var(--font-inter)]">Mohon tunggu...</p>
                <div className="flex justify-center mt-6">
                  <div className="w-8 h-8 border-4 border-[#FF8811] border-t-[#2B2D42] rounded-full animate-spin" />
                </div>
              </>
            )}

            {/* Waiting state - no token */}
            {!isLoading && !isVerified && !error && (
              <>
                <h2 className="text-3xl font-bold font-[family-name:var(--font-plus-jakarta-sans)] mb-4">
                  <span className="text-[#2B2D42]">Verifikasi </span>
                  <span className="text-[#FF8811]">Email Anda</span>
                </h2>
                <p className="text-gray-500 text-sm font-[family-name:var(--font-inter)] mb-6">Kami telah mengirimkan email verifikasi. Silakan buka email Anda dan klik link verifikasi untuk melanjutkan.</p>

                {/* Email display */}
                {email && (
                  <div className="bg-gray-50 border border-gray-200 rounded-lg p-3 mb-6">
                    <p className="text-sm text-gray-600">
                      Email: <strong>{email}</strong>
                    </p>
                  </div>
                )}

                {/* Instruction icon */}
                <div className="flex justify-center mb-6">
                  <svg className="w-16 h-16 text-[#FF8811]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                </div>

                {/* Resend button */}
                <div className="space-y-3">
                  <button
                    onClick={handleResendEmail}
                    disabled={resendLoading}
                    className="w-full cursor-pointer bg-white text-[#2B2D42] py-3 rounded-lg font-semibold transition-all font-[family-name:var(--font-plus-jakarta-sans)] disabled:opacity-50 disabled:cursor-not-allowed border-[2px] border-[#FF8811] hover:bg-[#FF8811]/10"
                  >
                    {resendLoading ? "Mengirim..." : "Kirim Ulang Email"}
                  </button>

                  {resendSent && <p className="text-green-600 text-xs font-[family-name:var(--font-inter)]">Email verifikasi telah dikirim ulang</p>}
                </div>
              </>
            )}

            {/* Success state */}
            {isVerified && !isLoading && (
              <>
                <h2 className="text-3xl font-bold font-[family-name:var(--font-plus-jakarta-sans)] mb-4">
                  <span className="text-[#2B2D42]">Email </span>
                  <span className="text-[#2B2D42]">Terverifikasi</span>
                </h2>
                <p className="text-gray-500 text-sm font-[family-name:var(--font-inter)] mb-6">Akun Anda telah berhasil diverifikasi. Mengarahkan ke dashboard...</p>
                <div className="flex justify-center">
                  <svg className="w-16 h-16 text-[#FF8811]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
              </>
            )}

            {/* Error state */}
            {error && !isLoading && !isVerified && (
              <>
                <h2 className="text-3xl font-bold font-[family-name:var(--font-plus-jakarta-sans)] mb-4">
                  <span className="text-[#2B2D42]">Verifikasi Gagal</span>
                </h2>
                <p className="text-[#2B2D42] text-sm font-[family-name:var(--font-inter)] mb-6">{error}</p>
                <div className="flex justify-center">
                  <svg className="w-16 h-16 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </div>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default VerifyEmail;
