'use client'

import { cn } from '@/lib/utils';
import { div, label } from 'framer-motion/client';
import { Book, BookOpen, LayoutDashboard, Sparkles, Users, X, Folder, LucideIcon } from 'lucide-react';
import Link from 'next/link';
import { usePathname } from 'next/navigation'

export interface NavItem {
    label: string;
    href: string;
    icon: LucideIcon;
}

interface sidebarProps {
    isOpen?: boolean;
    onClose?: () => void;
    navItems?: NavItem[];
    knowledgeItems?: NavItem[];
}

const Sidebar = ({ isOpen, onClose, navItems: propNavItems, knowledgeItems: propKnowledgeItems }: sidebarProps) => {


    const pathname = usePathname();

    const defaultNavItems = [
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
        {
            label: "Assets",
            href: "/dashboard/assets",
            icon: Folder,
        },
    ]

    const defaultKnowledgeItems = [
        {
            label: "Knowledge Spot",
            href: "/dashboard/knowledge",
            icon: BookOpen,
        },

    ]

    const navItems = propNavItems || defaultNavItems;
    const knowledgeItems = propKnowledgeItems || defaultKnowledgeItems;





    return (
        <>

            {isOpen && (
                <div className={cn('fixed inset-0  z-30 lg:hidden transition-opacity duration-300', isOpen ? "opacity-100" : "opacity-0 pointer-events-none")} onClick={onClose} />
            )}
            <aside className={cn("w-64 h-screen bg-[#FFFBF7] border-r border-gray-200 flex flex-col p-6 fixed left-0 top-0 z-40 transition-transform duration-300", "lg:translate-x-0", isOpen ? "translate-x-0 " : "-translate-x-full")}
            >
                <button onClick={onClose}
                    className='absolute top-4 right-4 p-2 rounded-lg hover:bg-gray-100 lg:hidden'
                >
                    <X className='w-5 h-5 text-gray-600' />
                </button>



                {/* Logo */}
                <div className="text-2xl font-bold mb-10">

                    <span className="text-[#FF8811]">Ngasih</span>
                    <span className="text-[#2B2D42]">Tau</span>
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


                    <div>
                        <p className="text-xs font-semibold text-gray-400 uppercase mb-4 tracking-wider">
                            Knowledge
                        </p>
                        <nav className="space-y-2">
                            {knowledgeItems.map((item) => {
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

                </div>


            </aside>
        </>


    )
}

export default Sidebar