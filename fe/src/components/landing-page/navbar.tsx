"use client";

import Link from "next/link";
import { AnimatePresence, motion } from "framer-motion";
import { useState } from "react";
import { Menu, X } from "lucide-react";

export function Navbar() {

    const [isOpen, setIsOpen] = useState(false);
    return (
        <motion.nav
            initial={{ y: -100, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            transition={{ duration: 0.8, ease: "easeOut" }}
            className="w-full max-w-7xl mx-auto px-6 md:px-12 py-6 flex items-center justify-between relative z-50"
        >
            {/* Logo */}
            <Link href="/">
                <div className="text-2xl font-bold font-family-name:var(--font-plus-jakarta-sans) text-[#2B2D42] flex items-center gap-2">
                    <span className="text-[#FF8811]">
                        Ngasih
                        {""}
                        <span className="text-[#2B2D42]">Tau</span>
                    </span>
                </div>
            </Link>
            {/* Desktop Links */}
            <div className="hidden md:flex items-center gap-8 font-semibold text-[#2B2D42] font-family-name:var(--font-plus-jakarta-sans)">
                <Link href="#features" className="hover:text-[#FF8811] transition-colors">Features</Link>
                <Link href="/mentors" className="hover:text-[#FF8811] transition-colors">About</Link>
                <Link href="#testimonials" className="hover:text-[#FF8811] transition-colors">Testimonials</Link>
            </div>

            {/* CTA Button */}
            <Link
                href="/sign-up"
                className="hidden md:block bg-[#FF8811] text-white px-6 py-2.5 font-bold rounded-md border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] hover:shadow-[2px_2px_0px_0px_#2B2D42] hover:translate-x-[2px] hover:translate-y-[2px] transition-all"
            >
                Get Started
            </Link>


            {/* Mobile Button */}
            <button className="md:hidden text-[#2B2D42]"
                onClick={() => setIsOpen(!isOpen)}>

                {isOpen ? <X size={28} /> : <Menu size={28} />}
            </button>

            {/* Mobile Menu */}
            <AnimatePresence>
                {isOpen && (
                    <motion.div
                        initial={{ opacity: 0, y: -20 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -20 }}
                        className="absolute top-full left-0 w-full bg-[#FFFBF7] font-family-name:var(--font-plus-jakarta-sans) shadow-lg p-6 flex flex-col gap-4 md:hidden border-b border-gray-200"
                    >
                        <Link href="/explore" className="font-semibold text-[#2B2D42] hover:text-[#FF8811] transition-colors" onClick={() => setIsOpen(false)}>Features</Link>
                        <Link href="/mentors" className="font-semibold text-[#2B2D42] hover:text-[#FF8811] transition-colors" onClick={() => setIsOpen(false)}>About</Link>
                        <Link href="/community" className="font-semibold text-[#2B2D42] hover:text-[#FF8811] transition-colors" onClick={() => setIsOpen(false)}>testimonials</Link>
                        <Link
                            href="/sign-up"
                            className="bg-[#FF8811] text-white px-6 py-2.5 font-bold rounded-md border-2 border-[#2B2D42] shadow-[4px_4px_0px_0px_#2B2D42] hover:shadow-[2px_2px_0px_0px_#2B2D42] hover:translate-x-[2px] hover:translate-y-[2px] transition-all text-center"
                            onClick={() => setIsOpen(false)}
                        >
                            Get Started
                        </Link>
                    </motion.div>
                )}
            </AnimatePresence>
        </motion.nav>
        // mobile
    );
}
