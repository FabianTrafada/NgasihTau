import Sidebar from '@/components/dashboard/Sidebar'
import Topbar from '@/components/dashboard/Topbar'
import React from 'react'

const DashboardLayout = ({ children }: { children: React.ReactNode }) => {
  return (
    <div className='flex min-h-screen bg-[#FFFBF7] font-[family-name:var(--font-plus-jakarta-sans)]'>
      <Sidebar />
      <div className="flex-1 flex flex-col ml-64">
        <Topbar />

        <main className='flex-1 overflow-y-auto'>
          {children}
        </main>
      </div>
    </div>
  )
}

export default DashboardLayout