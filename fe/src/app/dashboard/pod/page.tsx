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
     <div className="min-h-screen p-6">
      {/* 🔒 ONE GRID WRAPPER */}
      <div className="max-w-6xl mx-auto space-y-6">
        <div className="flex items-end justify-between">
          <div>
            <h1 className="text-2xl font-bold text-[#2B2D42]">
              Knowledge Pods
            </h1>
            <p className="text-zinc-500 text-xs sm:text-base">
              Learn and collaborate with your personal knowledge repositories.
            </p>
          </div>

          <div className="flex gap-2">
            <button className="px-4 py-2 border-2 border-black font-bold
              shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]
              hover:bg-[#FF8811] hover:text-white
              hover:translate-x-[2px] hover:translate-y-[2px]
              transition-all">
              Newest
            </button>

            <Link href="/dashboard/pod/create">
              <button className="px-4 py-2 border-2 border-black font-bold
                shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]
                hover:bg-[#FF8811] hover:text-white
                hover:translate-x-[2px] hover:translate-y-[2px]
                transition-all">
                Create Pod
              </button>
            </Link>
          </div>
        </div>

        <div className="bg-white border-2 border-black rounded-xl overflow-hidden
          shadow-[6px_6px_0px_0px_#FF8811]">
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
              isPersonal
              isLiked={pod.isLiked}
            />
          ))}
        </div>

        <div className="flex justify-center pt-4">
          <nav className="flex items-center gap-2">
            {[1, 2, 3, '...', 10].map((page, i) => (
              <button
                key={i}
                className={`w-10 h-10 border-2 border-black font-bold
                  shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]
                  transition-all
                  ${
                    page === 1
                      ? 'bg-[#FF8811] text-white'
                      : 'bg-white hover:bg-zinc-100'
                  }`}
              >
                {page}
              </button>
            ))}
          </nav>
        </div>
      </div>
    </div>
  );
};

export default MyKnowledgePage;
