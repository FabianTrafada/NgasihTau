'use client'
import Sidebar from '@/components/dashboard/Sidebar'
import RightSidebar from '@/components/dashboard/RightSidebar'
import Topbar from '@/components/dashboard/Topbar'
import { RoleProtectedRoute } from '@/components/auth/RoleProtectedRoute'
import { cn } from '@/lib/utils'
import React, { useState } from 'react'
import { TEACHER_SIDEBAR_GROUPS } from '@/lib/constants/navigation';

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
                    groups={TEACHER_SIDEBAR_GROUPS}
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