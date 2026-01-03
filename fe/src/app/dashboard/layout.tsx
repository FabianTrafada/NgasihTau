'use client'
import Sidebar from '@/components/dashboard/Sidebar'
import RightSidebar from '@/components/dashboard/RightSidebar'
import Topbar from '@/components/dashboard/Topbar'
import { cn } from '@/lib/utils'
import React, { useState } from 'react'
import { usePathname } from 'next/navigation'



const DashboardLayout = ({ children }: { children: React.ReactNode }) => {
  const pathname = usePathname();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [rightSidebarOpen, setRightSidebarOpen] = useState(false);

  // Daftar pola path yang tidak menampilkan RightSidebar
  // Menggunakan Regex agar bisa menangani dynamic route seperti /dashboard/my-pods/123
  const hideRightSidebarPatterns = [
    /^\/dashboard\/my-pods\/[^/]+$/, // Matches /dashboard/my-pods/[id]
    /^\/dashboard\/pods\/[^/]+$/, // Matches /dashboard/my-pods/[id]
    /^\/dashboard\/pod\/create$/, // matches /dashboard/pod/create
    /^\/dashboard\/my-pods$/, // matches /dashboard/my-pods 
  ];

  const shouldHideRightSidebar = hideRightSidebarPatterns.some(pattern => pattern.test(pathname));

  return (
    <div className='flex min-h-screen bg-[#FFFBF7] font-[family-name:var(--font-plus-jakarta-sans)]'>
      {/* Left Sidebar */}
      <Sidebar
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
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

          {!shouldHideRightSidebar && (
            <RightSidebar
              // {/* Right Sidebar - Now part of layout */}
              isOpen={rightSidebarOpen}
              onClose={() => setRightSidebarOpen(false)}
            />

          )}
        </div>
      </div>
    </div>
  )
}

export default DashboardLayout