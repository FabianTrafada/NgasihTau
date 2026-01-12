"use client";

import { motion } from "framer-motion";
import { useTranslations } from "next-intl";

export function SearchSection() {
    const t = useTranslations('landingPage.searchSection');
    return (
        <section className="py-10 px-6 md:px-12 max-w-4xl mx-auto w-full relative z-10">
            <motion.div
                initial={{ scale: 0.9, opacity: 0 }}
                whileInView={{ scale: 1, opacity: 1 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5 }}
                className="relative"
            >
                <input
                    type="text"
                    placeholder={t('placeholder')}
                    className="w-full bg-white border-2 border-[#2B2D42] rounded-md py-4 px-6 text-lg shadow-[4px_4px_0px_0px_#2B2D42] focus:outline-none focus:shadow-[2px_2px_0px_0px_#2B2D42] focus:translate-x-[2px] focus:translate-y-[2px] transition-all font-[family-name:var(--font-inter)]"
                />
            </motion.div>
        </section>
    );
}
