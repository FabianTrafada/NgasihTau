'use client'
import RightSidebar from '@/components/dashboard/RightSidebar'
import Topbar from '@/components/dashboard/Topbar'
import { RoleProtectedRoute } from '@/components/auth/RoleProtectedRoute'
import { cn } from '@/lib/utils'
import React, { useState } from 'react'
import { TEACHER_SIDEBAR_GROUPS } from '@/lib/constants/navigation';
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar'
import DashboardSidebar from '@/components/dashboard/DashboardSidebar'

const TeacherLayout = ({ children }: { children: React.ReactNode }) => {
    const [rightSidebarOpen, setRightSidebarOpen] = useState(false);

    return (
        <RoleProtectedRoute requiredRole="teacher">
            <SidebarProvider>
                <div className='flex min-h-screen w-full bg-[#FFFBF7] font-[family-name:var(--font-plus-jakarta-sans)]'>
                    {/* 2. Masukkan USER_SIDEBAR_GROUPS ke dalam prop groups */}
                    <DashboardSidebar groups={TEACHER_SIDEBAR_GROUPS} />

                    <SidebarInset className="flex flex-col min-w-0 bg-[#FFFBF7]">
                        <Topbar
                            onRightMenuClick={() => setRightSidebarOpen(!rightSidebarOpen)}
                        />

                        <div className="flex flex-1 overflow-hidden">
                            <main className='flex-1 overflow-y-auto'>
                                {children}
                            </main>
                        </div>
                    </SidebarInset>
                </div>
            </SidebarProvider>
        </RoleProtectedRoute>
    )
}

export default TeacherLayout