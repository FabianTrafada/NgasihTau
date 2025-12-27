"use client";

import Link from "next/link";
import { motion } from "framer-motion";
import { fadeInUp, staggerContainer } from "@/lib/animations";
import { Highlighter } from "../ui/highlighter";

export function Hero() {
    return (
        <main className="flex-1 flex flex-col items-center justify-center text-center px-4 relative z-10 mt-20 md:mt-32 pb-20">
            {/* Decorative Circles */}
            {/* Left Circle */}
            <motion.div
                initial={{ x: -200, opacity: 0 }}
                animate={{ x: 0, opacity: 0.5 }}
                transition={{ duration: 1, delay: 0.2 }}
                className="absolute left-0 top-1/2 -translate-y-1/2 -translate-x-1/2 w-[300px] h-[300px] md:w-[600px] md:h-[600px] rounded-full border-[40px] md:border-[80px] border-[#FFE5C8] bg-transparent -z-10 pointer-events-none"
            ></motion.div>
            {/* Right Circle */}
            <motion.div
                initial={{ x: 200, opacity: 0 }}
                animate={{ x: 0, opacity: 0.5 }}
                transition={{ duration: 1, delay: 0.2 }}
                className="absolute right-0 top-1/2 -translate-y-1/2 translate-x-1/2 w-[300px] h-[300px] md:w-[600px] md:h-[600px] rounded-full border-[40px] md:border-[80px] border-[#FFE5C8] bg-transparent -z-10 pointer-events-none"
            ></motion.div>

            <motion.div
                initial="hidden"
                animate="visible"
                variants={staggerContainer}
                className="flex flex-col items-center"
            >
                {/* Headline */}
                <motion.h1 variants={fadeInUp} className="text-5xl md:text-7xl font-bold text-[#2B2D42] leading-[1.1] tracking-tight mb-6">
                    Ubah Rasa <br />
                    Keingintahuan <br />

                    <span className="relative inline-block">
                        <Highlighter action="underline" color="#FF8811" strokeWidth={5}>

                            menjadi {""}
                            <span className="text-[#FF8811]">Tahu</span>
                        </Highlighter>
                    </span>
                </motion.h1>

                {/* Subtext */}
                <motion.p variants={fadeInUp} className="text-lg md:text-xl text-gray-600 max-w-2xl mb-12 font-[family-name:var(--font-inter)] leading-relaxed">
                    Unlock the power of collaborative learning. Create Knowledge Pods, <br className="hidden md:block" />
                    interact with AI-driven insights, and engage in meaningful <br className="hidden md:block" />
                    conversations to master any subject.
                </motion.p>

                {/* Buttons */}
                <motion.div variants={fadeInUp} className="flex flex-col md:flex-row items-center gap-6">
                    <Link
                        href="/sign-up"
                        className="bg-[#FF8811] text-white px-10 py-4 text-lg font-bold rounded-md border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] hover:shadow-[2px_2px_0px_0px_#2B2D42] hover:translate-x-[2px] hover:translate-y-[2px] transition-all"
                    >
                        Start Learn
                    </Link>

                    <button
                        className="bg-white text-[#2B2D42] px-10 py-4 text-lg font-bold rounded-md border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] hover:shadow-[2px_2px_0px_0px_#2B2D42] hover:translate-x-[2px] hover:translate-y-[2px] transition-all"
                    >
                        Watch Demo
                    </button>
                </motion.div>
            </motion.div>
        </main>
    );
}
