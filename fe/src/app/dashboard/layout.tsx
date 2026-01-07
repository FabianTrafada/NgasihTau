'use client'
import RightSidebar from '@/components/dashboard/RightSidebar'
import Topbar from '@/components/dashboard/Topbar'
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar'
import React, { useState } from 'react'
import { usePathname } from 'next/navigation'
import DashboardSidebar from '@/components/dashboard/DashboardSidebar'
import { USER_SIDEBAR_GROUPS } from '@/lib/constants/navigation'



const DashboardLayout = ({ children }: { children: React.ReactNode }) => {
  const pathname = usePathname();
  const [rightSidebarOpen, setRightSidebarOpen] = useState(false);

  const hideRightSidebarPatterns = [
    /^\/dashboard\/my-pods\/[^/]+$/,
    /^\/dashboard\/pods\/[^/]+$/,
    /^\/dashboard\/pod\/create$/,
    /^\/dashboard\/my-pods$/,
  ];

  const hideSidebarPatterns = [
    /^\/dashboard\/pod\/create$/,
    /^\/dashboard\/pods\/[^/]+$/,
  ]


  const shouldHideRightSidebar = hideRightSidebarPatterns.some(pattern => pattern.test(pathname));
  const shouldHideSidebar = hideSidebarPatterns.some(pattern => pattern.test(pathname));
  return (
    <SidebarProvider defaultOpen={false}>
      <div className='flex min-h-screen w-full bg-[#FFFBF7] font-[family-name:var(--font-plus-jakarta-sans)]'>

        {/* 2. Masukkan USER_SIDEBAR_GROUPS ke dalam prop groups */}

        {!shouldHideSidebar && (
          <>
            <DashboardSidebar groups={USER_SIDEBAR_GROUPS} />
            <SidebarInset className="flex flex-col min-w-0 bg-[#FFFBF7]  ml-var(--sidebar-width)
">
              <Topbar
                onRightMenuClick={() => setRightSidebarOpen(!rightSidebarOpen)}
              />

              <div className="flex flex-1 overflow-hidden">
                <main className='flex-1 overflow-y-auto'>
                  {children}
                </main>

                {!shouldHideRightSidebar && (
                  <RightSidebar
                    isOpen={rightSidebarOpen}
                    onClose={() => setRightSidebarOpen(false)}
                  />
                )}
              </div>
            </SidebarInset>
          </>
        )}

        {shouldHideSidebar && (
          <div className="flex flex-col min-w-0 flex-1 bg-[#FFFBF7]">
            <Topbar onRightMenuClick={() => setRightSidebarOpen(!rightSidebarOpen)} />

            <div className="flex flex-1 overflow-hidden">
              <main className="flex-1 overflow-y-auto">
                {children}
              </main>

              {!shouldHideRightSidebar && (
                <RightSidebar
                  isOpen={rightSidebarOpen}
                  onClose={() => setRightSidebarOpen(false)}
                />
              )}
            </div>
          </div>

        )}

      </div>
    </SidebarProvider>
  )
}

export default DashboardLayout