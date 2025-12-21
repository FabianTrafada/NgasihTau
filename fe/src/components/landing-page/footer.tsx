"use client";

import Link from "next/link";

export function Footer() {
    return (
        <footer className="py-10 px-6 md:px-12 max-w-7xl mx-auto w-full relative z-10 border-t-2 border-[#2B2D42]/10 mt-10">
            <div className="flex flex-col md:flex-row items-center justify-between gap-6">
                <div className="text-xl font-bold">
                    <span className="text-[#FF8811]">Ngasih</span>
                    <span className="text-[#2B2D42]">Tau</span>
                </div>
                <div className="flex items-center gap-8 text-sm font-bold text-[#2B2D42]">
                    <Link href="#" className="hover:text-[#FF8811]">Terms</Link>
                    <Link href="#" className="hover:text-[#FF8811]">Privacy</Link>
                    <Link href="#" className="hover:text-[#FF8811]">Contact</Link>
                </div>
                <p className="text-xs font-bold text-[#2B2D42]">© 2025 NgasihTau</p>
            </div>
        </footer>
    );
}
