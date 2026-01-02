'use client';

import React from 'react';
import { Eye, BadgeCheck, FileText } from 'lucide-react';
import FileListItem from '@/components/knowledge-pod/FileListItem'; // Sesuaikan path-nya
import { SearchSection } from '@/components/landing-page/search-section';

const KnowledgePodDetail = () => {
  const files = [
    {
      title: "Limit Turunan",
      description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.",
      likes: "2,3 K",
      date: "3 Jan 2025"
    },
    {
      title: "Limit Turunan",
      description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.",
      likes: "2,3 K",
      date: "3 Jan 2025"
    },
    {
      title: "Limit Turunan",
      description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.",
      likes: "2,3 K",
      date: "3 Jan 2025"
    }
  ];

  return (
    <div className="max-w-6xl mx-auto space-y-8 p-4 md:p-8">
      
      {/* Header Section */}
      <div className="flex justify-between items-center">
        <h1 className="text-4xl font-black text-black tracking-tight">
          Kalkulus
        </h1>
        <button className="px-10 py-2  border-2 border-black font-bold shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]  hover:bg-[#FF8811] hover:text-white cursor-pointer active:translate-x-[2px] active:translate-y-[2px] active:shadow-none hover:translate-x-[2px] hover:translate-y-[2px] transition-all group leading-none">
          Use
        </button>
      </div>

      {/* Description Card */}
      <div className="bg-white border-2 border-black  overflow-hidden shadow-[6px_6px_0px_0px_#FF8811] rounded-2xl p-6 pb-4">
        <p className='text-xs font-semibold text-zinc-400 mb-1'>Deskripsi</p>
        <p className="text-sm font-medium text-black mb-6">
          Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
        </p>
        
        <div className="flex justify-between items-center pt-4 border-t border-zinc-200">
          <div className='flex gap-6'>
            <div className="flex items-center gap-2 text-zinc-500">
              <Eye size={16} />
              <span className="text-xs font-mono font-bold">999K</span>
            </div>

            <div className="flex items-center gap-2 text-zinc-500">
              <FileText size={16} />
              <span className="text-xs font-mono font-bold">12 files</span>
            </div>
          </div>
          
          <div className="flex items-center gap-2 px-3 py-1 bg-zinc-100 border-2 border-black rounded-full">
            <BadgeCheck size={16} className="fill-black text-white" />
            <span className="font-bold text-xs">Guru Besar ITB</span>
          </div>
        </div>
      </div>

      {/* <SearchSection/> */}

      {/* Files List Container */}
      <div className="bg-white border-2 border-black rounded-2xl overflow-hidden shadow-[4px_4px_0px_0px_#FF8811]">
        <div className="flex flex-col">
          {files.map((file, index) => (
            <FileListItem 
              key={index}
              title={file.title}
              description={file.description}
              likes={file.likes}
              date={file.date}
              isLast={index === files.length - 1}
            />
          ))}
        </div>
      </div>

    </div>
  );
};

export default KnowledgePodDetail;