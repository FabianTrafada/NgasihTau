'use client';

import FileListItem from '@/components/knowledge-pod/FileListItem';
import { KnowledgePod } from '@/types/knowledgePods';
import Link from 'next/link';
import React, { useState } from 'react';
const MyKnowledgePage: React.FC = () => {
  const [pods, setPods] = useState<KnowledgePod[]>([
    {
      id: '1',
      title: 'Cara Belajar Mobil Kopling (99% bisa 1% nya hanya tuhan ....)',
      description: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. ...',
      fileCount: 12,
      date: '2 Jan 25',
      isLiked: true,
    },
    {
      id: '2',
      title: 'Dasar Pemrograman React untuk Pemula',
      description: 'Panduan lengkap memahami component, props, dan state dalam React modern menggunakan TypeScript.',
      fileCount: 8,
      date: '5 Jan 25',
      isLiked: true,
    },
    {
      id: '3',
      title: 'Manajemen Waktu untuk Mahasiswa Akhir',
      description: 'Tips dan trik mengelola jadwal skripsi tanpa mengabaikan kesehatan mental dan kehidupan sosial.',
      fileCount: 5,
      date: '10 Jan 25',
      isLiked: true,
    },
    {
      id: '4',
      title: 'Food Photography dengan Smartphone',
      description: 'Cara mengambil foto makanan yang estetik hanya dengan modal kamera HP dan cahaya matahari.',
      fileCount: 20,
      date: '12 Jan 25',
      isLiked: true,
    }
  ]);

  const handleToggleLike = (id: string) => {
    setPods(prev => prev.map(pod =>
      pod.id === id ? { ...pod, isLiked: !pod.isLiked } : pod
    ));
  };



  return (
    <div className="min-h-screen mx-auto space-y-8 p-4 md:p-8">
      {/* Header Section */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 max-w-6xl">
        <div>
          <h1 className="text-1xl md:text-2xl font-family-name:var(--font-plus-jakarta-sans) sm:text-2xl font-bold text-[#2B2D42] mb-1">
            Knowledge Pods
          </h1>
          <p className="text-zinc-500 text-xs sm:text-base">
            Manage and explore your personal collection of knowledge repositories.
          </p>
        </div>
        <div className="flex gap-2 min-w-max">
          <button className="px-4 py-2  border-2 text-base border-black font-bold shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] hover:bg-[#FF8811] active:translate-x-[1px] hover:shadow-[2px_2px_0px_0px_#2B2D42] hover:text-white cursor-pointer hover:translate-x-[2px] hover:translate-y-[2px] transition-all group leading-none">
            Newest
          </button>
          <Link href="/dashboard/pod/create" className="px-4 py-2  border-2 border-black font-bold shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] active:translate-x-[1px] hover:shadow-[2px_2px_0px_0px_#2B2D42] hover:translate-x-[2px] cursor-pointer hover:bg-[#FF8811] hover:text-white hover:translate-y-[2px] transition-all group leading-none">
            Create Pod
          </Link>
        </div>
      </div>

      {/* Main Container - Industrial / Neo-brutalism Style */}
      <div className="bg-white border-2 border-r-4 border-black rounded-[12px] overflow-hidden shadow-[6px_6px_0px_0px_#FF8811]">
        <div className="flex flex-col">
          {pods.map((pod, index) => (
            <FileListItem
              key={pod.id}
              variant="pod"
              userId="me"
              podId={pod.id}
              title={pod.title}
              description={pod.description}
              date={pod.date}
              onToggleLike={handleToggleLike}
              isLast={index === pods.length - 1}
              isPersonal={true}
              isLiked={pod.isLiked}
            />
          ))}
        </div>
      </div>

      {/* Pagination / Industrial Footer */}
      <div className="flex justify-center pt-4">
        <nav className="flex items-center gap-1">
          {[1, 2, 3, '...', 10].map((page, i) => (
            <button
              key={i}
              className={`w-10 h-10 flex items-center justify-center border-2 border-black font-bold shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] transition-all ${page === 1 ? 'bg-orange-500 text-white' : 'bg-white text-black hover:bg-zinc-100'
                }`}
            >
              {page}
            </button>
          ))}
        </nav>
      </div>
    </div>
  );
};

export default MyKnowledgePage;
