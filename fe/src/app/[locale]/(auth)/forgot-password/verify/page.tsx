"use client";

import Link from "next/link";
import { useState, useRef, useEffect } from "react";

export default function VerifyPage() {
    const [otp, setOtp] = useState<string[]>(["", "", "", ""]);
    const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

    useEffect(() => {
        // Focus on first input on mount
        inputRefs.current[0]?.focus();
    }, []);

    const handleChange = (index: number, value: string) => {
        // Only allow single digit
        if (value.length > 1) {
            value = value.slice(-1);
        }

        // Only allow numbers
        if (value && !/^\d$/.test(value)) {
            return;
        }

        const newOtp = [...otp];
        newOtp[index] = value;
        setOtp(newOtp);

        // Move to next input if value is entered
        if (value && index < 3) {
            inputRefs.current[index + 1]?.focus();
        }
    };

    const handleKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
        // Move to previous input on backspace if current is empty
        if (e.key === "Backspace" && !otp[index] && index > 0) {
            inputRefs.current[index - 1]?.focus();
        }
    };

    const handlePaste = (e: React.ClipboardEvent) => {
        e.preventDefault();
        const pastedData = e.clipboardData.getData("text").slice(0, 4);
        if (/^\d+$/.test(pastedData)) {
            const newOtp = [...otp];
            for (let i = 0; i < pastedData.length && i < 4; i++) {
                newOtp[i] = pastedData[i];
            }
            setOtp(newOtp);
            // Focus on the last filled input or the next empty one
            const nextIndex = Math.min(pastedData.length, 3);
            inputRefs.current[nextIndex]?.focus();
        }
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const code = otp.join("");
        console.log("Verification code:", code);
        // Handle verification logic here
    };

    const handleResend = () => {
        // Handle resend logic here
        console.log("Resending code...");
    };

    return (
        <div className="w-full max-w-sm">
            {/* Card Container with Brutalist shadow */}
            <div className="relative">
                {/* Brutalist offset shadow */}
                <div className="absolute top-3 left-3 w-full h-full bg-[#FF8811] rounded-2xl" />

                {/* Main Card */}
                <div className="bg-white rounded-2xl p-6 relative border-2 border-[#2B2D42] min-h-[400px] flex flex-col">

                    {/* Back Button */}
                    <Link
                        href="/forgot-password"
                        className="inline-flex items-center text-[#FF8811] mb-6 hover:text-[#FF8811]/80 transition-colors"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-4 h-4">
                            <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5L3 12m0 0l7.5-7.5M3 12h18" />
                        </svg>
                    </Link>

                    <div className="flex-1 flex flex-col justify-center">
                        {/* Verification Text */}
                        <div className="text-center mb-8">
                            <h2 className="text-2xl font-bold font-[family-name:var(--font-plus-jakarta-sans)]">
                                <span className="text-[#2B2D42] ">Verifi</span>
                                <span className="text-[#FF8811]">cation</span>
                            </h2>
                            <p className="text-gray-400 text-xs mt-2 font-[family-name:var(--font-inter)]">
                                We will email you a verification code to your email address
                            </p>
                        </div>

                        {/* OTP Input */}
                        <form onSubmit={handleSubmit} className="space-y-8">
                            <div className="flex justify-center gap-4" onPaste={handlePaste}>
                                {otp.map((digit, index) => (
                                    <input
                                        key={index}
                                        ref={(el) => { inputRefs.current[index] = el; }}
                                        type="text"
                                        inputMode="numeric"
                                        maxLength={1}
                                        value={digit}
                                        onChange={(e) => handleChange(index, e.target.value)}
                                        onKeyDown={(e) => handleKeyDown(index, e)}
                                        className="w-14 h-14 text-center text-2xl font-bold border-b-2 border-[#FF8811]/40 focus:border-[#FF8811] outline-none transition-colors bg-transparent font-[family-name:var(--font-plus-jakarta-sans)]"
                                    />
                                ))}
                            </div>

                            {/* Resend Link */}
                            <p className="text-center text-gray-600 text-sm font-[family-name:var(--font-inter)]">
                                If you didn&apos;t recieve a code!{" "}
                                <button
                                    type="button"
                                    onClick={handleResend}
                                    className="text-[#FF8811] cursor-pointer font-semibold hover:text-[#FF8811]/80 transition-colors"
                                >
                                    Resend
                                </button>
                            </p>

                            {/* Verify Button */}
                            <button
                                type="submit"
                                className="w-full bg-[#FF8811] cursor-pointer text-white py-3 rounded-sm font-semibold hover:bg-[#FF8811]/90 transition-colors  font-[family-name:var(--font-plus-jakarta-sans)]"
                            >
                                Verify
                            </button>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    );
}
