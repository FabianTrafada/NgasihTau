"use client";

import { useTranslations } from "next-intl";
import { AnimatePresence, motion } from "framer-motion";
import { useState } from "react";
import { Menu, X } from "lucide-react";
import { LocalizedLink } from "../ui/LocalizedLink";
import { LanguageSwitcher } from "../ui/LanguageSwitcher";

export function Navbar() {
    const [isOpen, setIsOpen] = useState(false);
    const tNav = useTranslations('navigation');
    const tCommon = useTranslations('common');
    return (
        <motion.nav
            initial={{ y: -100, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            transition={{ duration: 0.8, ease: "easeOut" }}
            className="w-full max-w-7xl mx-auto px-6 md:px-12 py-6 flex items-center justify-between relative z-50"
        >
            {/* Logo */}
            <LocalizedLink href="/">
                <div className="text-2xl font-bold font-family-name:var(--font-plus-jakarta-sans) text-[#2B2D42] flex items-center gap-2">
                    <span className="text-[#FF8811]">
                        Ngasih
                        <span className="text-[#2B2D42]">Tau</span>
                    </span>
                </div>
            </LocalizedLink>
            {/* Desktop Links */}
            <div className="hidden md:flex items-center gap-8 font-semibold text-[#2B2D42] font-family-name:var(--font-plus-jakarta-sans)">
                <a href="#features" className="hover:text-[#FF8811] transition-colors">{tNav('features')}</a>
                <LocalizedLink href="/mentors" className="hover:text-[#FF8811] transition-colors">{tNav('about')}</LocalizedLink>
                <a href="#testimonials" className="hover:text-[#FF8811] transition-colors">{tNav('testimonials')}</a>
            </div>

            {/* Desktop Actions */}
            <div className="hidden md:flex items-center gap-4">
                <LanguageSwitcher />
                <LocalizedLink
                    href="/sign-up"
                    className="bg-[#FF8811] text-white px-6 py-2.5 font-bold rounded-md border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] hover:shadow-[2px_2px_0px_0px_#2B2D42] hover:translate-x-[2px] hover:translate-y-[2px] transition-all"
                >
                    {tCommon('getStarted')}
                </LocalizedLink>
            </div>


            {/* Mobile Button */}
            <div className="md:hidden flex items-center gap-2">
                <LanguageSwitcher />
                <button className="text-[#2B2D42] p-2"
                    onClick={() => setIsOpen(!isOpen)}
                    aria-label="Toggle menu"
                    aria-expanded={isOpen}>
                    {isOpen ? <X size={28} /> : <Menu size={28} />}
                </button>
            </div>

            {/* Mobile Menu */}
            <AnimatePresence>
                {isOpen && (
                    <motion.div
                        initial={{ opacity: 0, y: -20 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -20 }}
                        className="absolute top-full left-0 w-full bg-[#FFFBF7] font-family-name:var(--font-plus-jakarta-sans) shadow-lg p-6 flex flex-col gap-4 md:hidden border-b border-gray-200"
                    >
                        <a href="#features" className="font-semibold text-[#2B2D42] hover:text-[#FF8811] transition-colors" onClick={() => setIsOpen(false)}>{tNav('features')}</a>
                        <LocalizedLink href="/mentors" className="font-semibold text-[#2B2D42] hover:text-[#FF8811] transition-colors" onClick={() => setIsOpen(false)}>{tNav('about')}</LocalizedLink>
                        <a href="#testimonials" className="font-semibold text-[#2B2D42] hover:text-[#FF8811] transition-colors" onClick={() => setIsOpen(false)}>{tNav('testimonials')}</a>
                        <LocalizedLink
                            href="/sign-up"
                            className="bg-[#FF8811] text-white px-6 py-2.5 font-bold rounded-md border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] hover:shadow-[2px_2px_0px_0px_#2B2D42] hover:translate-x-[2px] hover:translate-y-[2px] transition-all text-center"
                            onClick={() => setIsOpen(false)}
                        >
                            {tCommon('getStarted')}
                        </LocalizedLink>
                    </motion.div>
                )}
            </AnimatePresence>
        </motion.nav>
        // mobile
    );
}
