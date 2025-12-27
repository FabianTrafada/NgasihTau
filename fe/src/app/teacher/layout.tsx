'use client'
import Sidebar from '@/components/dashboard/Sidebar'
import RightSidebar from '@/components/dashboard/RightSidebar'
import Topbar from '@/components/dashboard/Topbar'
import { RoleProtectedRoute } from '@/components/auth/RoleProtectedRoute'
import { cn } from '@/lib/utils'
import React, { useState } from 'react'
import { LayoutDashboard, Users, BookOpen, Settings } from 'lucide-react';

const teacherNavItems = [
    {
        label: "Dashboard",
        href: "/teacher/dashboard",
        icon: LayoutDashboard,
    },
    {
        label: "My Classes",
        href: "/teacher/classes",
        icon: Users,
    },
    {
        label: "Materials",
        href: "/teacher/materials",
        icon: BookOpen,
    },
    {
        label: "Settings",
        href: "/teacher/settings",
        icon: Settings,
    }
];

const TeacherLayout = ({ children }: { children: React.ReactNode }) => {
    const [sidebarOpen, setSidebarOpen] = useState(false);
    const [rightSidebarOpen, setRightSidebarOpen] = useState(false);

    return (
        <RoleProtectedRoute requiredRole="teacher">
            <div className='flex min-h-screen bg-[#FFFBF7] font-[family-name:var(--font-plus-jakarta-sans)]'>
                {/* Left Sidebar */}
                <Sidebar
                    isOpen={sidebarOpen}
                    onClose={() => setSidebarOpen(false)}
                    navItems={teacherNavItems}
                    knowledgeItems={[]}
                />

                {/* Main Content Wrapper */}
                <div className={cn(
                    "flex-1 flex flex-col transition-all duration-300",
                    "ml-0 lg:ml-64"
                )}>

                    <Topbar
                        onMenuClick={() => setSidebarOpen(!sidebarOpen)}
                        onRightMenuClick={() => setRightSidebarOpen(!rightSidebarOpen)}
                        sidebarOpen={sidebarOpen}
                    />

                    <div className="flex flex-1 overflow-hidden">
                        <main className='flex-1 overflow-y-auto'>
                            {children}
                        </main>
                    </div>
                </div>
            </div>
        </RoleProtectedRoute>
    )
}

export default TeacherLayout