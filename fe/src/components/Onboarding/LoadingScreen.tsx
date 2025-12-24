"use client";

import { motion } from "framer-motion";

interface LoadingScreenProps {
    message?: string;
}

export function LoadingScreen({ message = "Getting ready with your Workspace" }: LoadingScreenProps) {
    return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-[#FFFBF5]">
            <div className="text-center mb-12">
                <h1 className="text-4xl md:text-5xl font-bold text-[#1A1A1A] mb-2 tracking-tight">
                    Welcome to
                </h1>
                <h1 className="text-4xl md:text-5xl font-bold tracking-tight">
                    Ngasih<span className="text-[#FF8811]">Tau</span>
                </h1>
            </div>

            <div className="relative">
                {/* Main Card */}
                <div className="relative z-10 w-[350px] md:w-[400px] h-[250px] bg-white rounded-lg border-2 border-black flex flex-col items-center justify-center p-8">
                    <motion.div
                        animate={{ rotate: 360 }}
                        transition={{ duration: 1.5, repeat: Infinity, ease: "linear" }}
                        className="w-20 h-20 border-10 border-[#FF8811] border-t-transparent rounded-full mb-8"
                    />
                    <p className="text-sm font-bold text-black text-center">
                        {message}
                    </p>
                </div>

                {/* Orange Accent Background */}
                <div className="absolute top-2 -right-2 w-full h-full bg-[#FF8811] rounded-lg border-2 border-black z-0" />
            </div>
        </div>
    );
}