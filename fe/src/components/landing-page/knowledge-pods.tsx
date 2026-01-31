"use client";

import { motion } from "framer-motion";
import { fadeInUp, staggerContainer } from "@/lib/animations";
import { useTranslations } from "next-intl";

export function KnowledgePods() {
    const t = useTranslations('landingPage.knowledgePods');
    return (
        <section className="py-20 px-6 md:px-12 max-w-7xl mx-auto w-full relative z-10">
            <div className="flex flex-col md:flex-row items-start justify-between gap-12">
                {/* Left Content */}
                <motion.div
                    initial="hidden"
                    whileInView="visible"
                    viewport={{ once: true }}
                    variants={fadeInUp}
                    className="md:w-1/2 pt-10"
                >
                    <h2 className="text-4xl md:text-5xl font-bold text-[#2B2D42] mb-6 leading-tight">
                        {t('title')}
                    </h2>
                    <p className="text-gray-600 font-[family-name:var(--font-inter)] text-lg leading-relaxed max-w-md">
                        {t('description')}
                    </p>
                </motion.div>

                {/* Right Content - Neo-brutalist Cards */}
                <motion.div
                    initial="hidden"
                    whileInView="visible"
                    viewport={{ once: true }}
                    variants={staggerContainer}
                    className="md:w-1/2 flex flex-col gap-6 w-full"
                >
                    {[1, 2].map((item) => (
                        <motion.div
                            key={item}
                            variants={fadeInUp}
                            className="bg-white border-2 border-[#2B2D42] p-6 shadow-[6px_6px_0px_0px_#2B2D42] hover:shadow-[3px_3px_0px_0px_#2B2D42] hover:translate-x-[3px] hover:translate-y-[3px] transition-all relative group"
                        >
                            {/* Brutalist Tag */}
                            <div className="absolute -top-3 -right-3 bg-[#FF8811] text-white text-xs font-bold px-2 py-1 border-2 border-[#2B2D42] shadow-[2px_2px_0px_0px_#2B2D42]">
                                {t('popularTag')}
                            </div>

                            <h3 className="text-[#FF8811] font-bold text-sm mb-2">Cara Belajar Mobil Kopling (99% bisa 1% nya hanya tuhan ....)</h3>
                            <p className="text-xs text-gray-500 font-[family-name:var(--font-inter)]">
                                Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. ...
                            </p>
                        </motion.div>
                    ))}
                </motion.div>
            </div>
        </section>
    );
}
