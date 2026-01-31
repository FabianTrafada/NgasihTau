"use client";

import Link from "next/link";
import { useState } from "react";
import { useRouter } from "next/navigation";

export default function ForgotPasswordPage() {
    const [email, setEmail] = useState("");
    const router = useRouter();

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        // Handle forgot password logic here
        console.log({ email });
        // Navigate to verification page after sending email
        router.push("/forgot-password/verify");
    };

    return (
        <div className="w-full max-w-md">
            {/* Card Container with Brutalist shadow */}
            <div className="relative">
                {/* Brutalist offset shadow */}
                <div className="absolute top-2 left-2 w-full h-full bg-[#FF8811] rounded-2xl" />

                {/* Main Card */}
                <div className="bg-white rounded-2xl p-8 relative border-2 border-[#2B2D42]">

                    {/* Back Button */}
                    <Link
                        href="/sign-in"
                        className="inline-flex items-center text-[#FF8811] mb-8 hover:text-[#FF8811]/80 transition-colors"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-5 h-5">
                            <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5L3 12m0 0l7.5-7.5M3 12h18" />
                        </svg>
                    </Link>

                    {/* Forgot Password Text */}
                    <div className="text-center mb-8">
                        <h2 className="text-3xl font-bold font-[(family-name:var(--font-plus-jakarta-sans))]">
                            <span className="text-[#2B2D42] ">Forgot </span>
                            <span className="text-[#FF8811]">Password</span>
                        </h2>
                        <p className="text-[#2B2D42] font-medium mt-4 font-[(family-name:var(--font-inter))]">
                            Enter your email address<br />
                            associated with your account.
                        </p>
                        <p className="text-gray-500 text-sm mt-3 font-[(family-name:var(--font-inter))]">
                            We will email you a verification<br />
                            code to your email address
                        </p>
                    </div>

                    {/* Form */}
                    <form onSubmit={handleSubmit} className="space-y-6">
                        {/* Email Field */}
                        <div>
                            <label
                                htmlFor="email"
                                className="block text-sm font-semibold text-[#2B2D42] mb-2 font-[family-name:var(--font-plus-jakarta-sans)]"
                            >
                                Email
                            </label>
                            <input
                                type="email"
                                id="email"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                className="w-full text-[#2B2D42] px-4 py-3 border-2 border-gray-300 rounded-md bg-white focus:outline-none transition-all text-center text-2xl tracking-widest focus:border-[#FF8811] focus:shadow-[3px_3px_0px_0px_rgba(255,136,17,1)] font-(family-name:var(--font-inter))"
                                placeholder=""
                                required
                            />
                        </div>

                        {/* Send Button */}
                        <button
                            type="submit"
                            className="cursor-pointer w-full border-2  bg-[#FF8811] text-white py-3 rounded-sm font-semibold border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[2px_2px_0px_0px_#2B2D42] active:shadow-none active:translate-x-[4px] active:translate-y-[4px] transition-all disabled:opacity-50 disabled:cursor-not-allowe  font-(family-name:var(--font-plus-jakarta-sans))"
                        >
                            Send
                        </button>
                    </form>

                    {/* Create Account Link */}
                    <p className="text-center mt-8 text-gray-600 text-sm font-[family-name:var(--font-inter)]">
                        Don&apos;t Have An Account?{" "}
                        <Link
                            href="/sign-up"
                            className="text-[#FF8811] font-semibold hover:text-[#FF8811]/80 transition-colors"
                        >
                            Create Account
                        </Link>
                    </p>
                </div>
            </div>
        </div>
    );
}
