'use client';

import React, { useState } from 'react';
import KnowledgePodCard from '@/components/knowledge-pod/KnowledgePodCard';
import { KnowledgePod } from '@/types';

const KnowledgePage: React.FC = () => {
  const [pods, setPods] = useState<KnowledgePod[]>([
    {
      id: '1',
      title: 'Cara Belajar Mobil Kopling (99% bisa 1% nya hanya tuhan ....)',
      description: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. ...',
      fileCount: 12,
      date: '2 Jan 25',
      isLiked: false
    },
    {
      id: '2',
      title: 'Dasar Pemrograman React untuk Pemula',
      description: 'Panduan lengkap memahami component, props, dan state dalam React modern menggunakan TypeScript.',
      fileCount: 8,
      date: '5 Jan 25',
      isLiked: true
    },
    {
      id: '3',
      title: 'Manajemen Waktu untuk Mahasiswa Akhir',
      description: 'Tips dan trik mengelola jadwal skripsi tanpa mengabaikan kesehatan mental dan kehidupan sosial.',
      fileCount: 5,
      date: '10 Jan 25',
      isLiked: false
    },
    {
      id: '4',
      title: 'Food Photography dengan Smartphone',
      description: 'Cara mengambil foto makanan yang estetik hanya dengan modal kamera HP dan cahaya matahari.',
      fileCount: 20,
      date: '12 Jan 25',
      isLiked: false
    }
  ]);

  const handleToggleLike = (id: string) => {
    setPods(prev => prev.map(pod => 
      pod.id === id ? { ...pod, isLiked: !pod.isLiked } : pod
    ));
  };

  return (
    <div className="max-w-6xl mx-auto space-y-8 p-4 md:p-8">
      {/* Header Section */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <h1 className="text-4xl font-black text-black tracking-tight mb-2">
            Top Knowledge Pod
          </h1>
          <p className="text-zinc-500 font-medium">
            Explore the most shared and updated knowledge repositories.
          </p>
        </div>
        <div className="flex gap-2">
          <button className="px-4 py-2 bg-white border-2 border-black font-bold shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] hover:bg-zinc-50 transition-all active:translate-x-[1px] active:translate-y-[1px]">
            Newest
          </button>
          <button className="px-4 py-2 bg-orange-500 text-white border-2 border-black font-bold shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] hover:bg-orange-600 transition-all active:translate-x-[1px] active:translate-y-[1px]">
            Upload Pod
          </button>
        </div>
      </div>

      {/* Main Container - Industrial / Neo-brutalism Style */}
      <div className="bg-white border-2 border-black rounded-[12px] overflow-hidden shadow-[6px_6px_0px_0px_rgba(0,0,0,1)]">
        <div className="flex flex-col">
          {pods.map((pod, index) => (
            <KnowledgePodCard 
              key={pod.id} 
              pod={pod} 
              onToggleLike={handleToggleLike}
              isLast={index === pods.length - 1}
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
              className={`w-10 h-10 flex items-center justify-center border-2 border-black font-bold shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] transition-all ${
                page === 1 ? 'bg-orange-500 text-white' : 'bg-white text-black hover:bg-zinc-100'
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

export default KnowledgePage;
