'use client'

import { cn } from '@/lib/utils';
import { div } from 'framer-motion/client';
import { LayoutDashboard, Sparkles, Users } from 'lucide-react';
import Link from 'next/link';
import { usePathname } from 'next/navigation'
import path from 'path';

const Sidebar = () => {


    const pathname = usePathname();

    const navItems = [
        {
            label: "Home",
            href: "/dashboard",
            icon: LayoutDashboard,
        },
        {
            label: "NgasihTau AI",
            href: "/dashboard/ngasihtau-ai",
            icon: Sparkles,
        },
    ]

   

    return (
        <aside className="w-64 h-screen bg-[#FFFBF7] border-r border-gray-200 flex flex-col p-6 fixed left-0 top-0 z-20">
            {/* Logo */}
            <div className="mb-10">
                <h1 className="text-2xl font-bold text-[#FF8811]">Logo</h1>
            </div>

            {/* Navigation */}
            <div className="space-y-6">
                <div>
                    <p className="text-xs font-semibold text-gray-400 uppercase mb-4 tracking-wider">
                        Navigation
                    </p>
                    <nav className="space-y-2">
                        {navItems.map((item) => {
                            const isActive = pathname === item.href || (item.href !== '/dashboard' && pathname?.startsWith(item.href));
                            return (
                                <Link
                                    key={item.href}
                                    href={item.href}
                                    className={cn(
                                        "flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 font-medium",
                                        isActive
                                            ? "bg-[#FF8811] text-white shadow-sm"
                                            : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                                    )}
                                >
                                    <item.icon className="w-5 h-5" />
                                    {item.label}
                                </Link>
                            );
                        })}
                    </nav>
                </div>

                {/* knowledge spot */}
{/* 
                <div>
                    <p className="text-xs font-semibold text-gray-400 uppercase mb-4 tracking-wider">
                        Knowldege pod
                    </p>


                    <nav className="space-y-2">
                        {knowledgePods.map((item, idx) => (
                            <Link
                                key={idx}
                                href={item.href}
                                className="flex items-center gap-3 px-4 py-3 rounded-xl text-gray-600 hover:bg-gray-100 hover:text-gray-900 transition-all duration-200 font-medium"
                            >
                                <item.icon className="w-5 h-5" />
                                {item.label}
                            </Link>
                        ))}
                    </nav>
                </div> */}
            </div>


        </aside>
    )
}

export default Sidebar