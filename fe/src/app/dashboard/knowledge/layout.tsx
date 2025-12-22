
import React from 'react';
import Sidebar from '@/components/dashboard/Sidebar';

interface KnowledgeLayoutProps {
  children: React.ReactNode;
}

const KnowledgeLayout: React.FC<KnowledgeLayoutProps> = ({ children }) => {
  return (
    <div className="flex min-h-screen bg-[#FDFCF9]">
      {/* Sidebar - Fixed/Sticky on Large Screens */}
      <Sidebar />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0">
        
        <main className="flex-1 overflow-y-auto">
          {children}
        </main>
      </div>
    </div>
  );
};

export default KnowledgeLayout;
