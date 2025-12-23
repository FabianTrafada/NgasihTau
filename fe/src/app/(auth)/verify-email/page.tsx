"use client";

import React, { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";

const VerifyEmail = () => {
  const searchParams = useSearchParams();
  const [isLoading, setIsLoading] = useState(true);
  const [isSuccess, setIsSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const verifyEmail = async () => {
      try {
        const token = searchParams.get("token");

        if (!token) {
          setError("Token tidak ditemukan");
          setIsLoading(false);
          return;
        }

        const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}api/v1/auth/verify-email`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ token }),
        });

        if (response.ok) {
          setIsSuccess(true);
          setError(null);
        } else {
          const errorData = await response.json().catch(() => ({}));
          setError(errorData.message || "Link sudah kadaluarsa atau tidak valid");
          console.log(errorData.message);
        }
      } catch (err) {
        setError("Terjadi kesalahan saat memverifikasi email");
      } finally {
        setIsLoading(false);
      }
    };

    verifyEmail();
  }, [searchParams]);

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
                  <span className="text-[#2B2D42]">Memverifikasi</span>
                  <span className="text-[#FF8811]"> Email</span>
                </h2>
                <p className="text-gray-500 text-sm font-[family-name:var(--font-inter)]">Mohon tunggu...</p>
                <div className="flex justify-center mt-6">
                  <div className="w-8 h-8 border-4 border-[#FF8811] border-t-[#2B2D42] rounded-full animate-spin" />
                </div>
              </>
            )}

            {isSuccess && !isLoading && (
              <>
                <h2 className="text-3xl font-bold font-[family-name:var(--font-plus-jakarta-sans)] mb-4">
                  <span className="text-[#2B2D42]">Email </span>
                  <span className="text-[#2B2D42]">Terverifikasi</span>
                </h2>
                <p className="text-gray-500 text-sm font-[family-name:var(--font-inter)] mb-6">Akun Anda telah berhasil diverifikasi. Anda dapat login sekarang.</p>
                <div className="flex justify-center">
                  <svg className="w-16 h-16 text-[#FF8811]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
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
};

export default VerifyEmail;
