'use client';

import FileListItem from '@/components/knowledge-pod/FileListItem';
import { KnowledgePod } from '@/types/knowledgePods';
import React, { useState, useEffect } from 'react';
import { useAuth } from '@/lib/auth-context';
import { getUserPods } from '@/lib/api/pod';
import { Loader } from 'lucide-react';
import Link from 'next/link';

const MyKnowledgePage: React.FC = () => {
  const { user } = useAuth();
  const [pods, setPods] = useState<KnowledgePod[]>([
    {
      id: '1',
      title: 'Cara Belajar Mobil Kopling (99% bisa 1% nya hanya tuhan ....)',
      description: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. ...',
      fileCount: 0,
      date: '1/2/2025',
      isLiked: false
    },
    {
      id: '2',
      title: 'Dasar Pemrograman React untuk Pemula',
      description: 'Panduan lengkap memahami component, props, dan state dalam React modern menggunakan TypeScript.',
      fileCount: 0,
      date: '1/5/2025',
      isLiked: false
    },
    {
      id: '3',
      title: 'Manajemen Waktu untuk Mahasiswa Akhir',
      description: 'Tips dan trik mengelola jadwal skripsi tanpa mengabaikan kesehatan mental dan kehidupan sosial.',
      fileCount: 0,
      date: '1/10/2025',
      isLiked: false
    },
    {
      id: '4',
      title: 'Food Photography dengan Smartphone',
      description: 'Cara mengambil foto makanan yang estetik hanya dengan modal kamera HP dan cahaya matahari.',
      fileCount: 0,
      date: '1/12/2025',
      isLiked: false
    }
  ]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  // Fetch user pods on mount or when user/page changes
  useEffect(() => {
    const fetchPods = async () => {
      if (!user) return;

      try {
        setLoading(true);
        // Note: getUserPods returns { data: Pod[], pagination: ... }
        // We need to map Pod[] to KnowledgePod[] if the types differ slightly, 
        // but let's assume for now we map what we need.
        const response = await getUserPods(user.id, page, 10);

        // Map API response to UI component interface
        const mappedPods: KnowledgePod[] = response.data.map((pod: any) => ({
          id: pod.id,
          title: pod.name,
          description: pod.description || "",
          fileCount: 0, // Backend might not return this yet, default to 0
          date: new Date(pod.created_at).toLocaleDateString(),
          isLiked: false, // Default
        }));

        setPods(prev => {
          // Combine initial dummy pods (ids 1-4) with fetched pods
          // Filter out duplicates if any fetched pod has same ID (unlikely but safe)
          const dummyIds = ['1', '2', '3', '4'];
          const newPods = mappedPods.filter(p => !dummyIds.includes(p.id));

          // Keep dummy pods at the top or bottom? Let's keep them at the top as requested.
          // Note: ensure we don't duplicate if useEffect runs twice
          const existingDummies = prev.filter(p => dummyIds.includes(p.id));
          return [...existingDummies, ...newPods];
        });
        // Calculate total pages if pagination is available
        if (response.pagination) {
          setTotalPages(Math.ceil(response.pagination.total / response.pagination.per_page));
        }
      } catch (err) {
        console.error("Failed to fetch pods:", err);
      } finally {
        setLoading(false);
      }
    };

    fetchPods();
  }, [user, page]);

  const handleToggleLike = (id: string) => {
    setPods(prev => prev.map(pod =>
      pod.id === id ? { ...pod, isLiked: !pod.isLiked } : pod
    ));
  };

  if (loading) {
    return (
      <div className="flex h-[50vh] items-center justify-center">
        <Loader className="animate-spin text-orange-500" size={32} />
      </div>
    );
  }

  return (
    <div className="mx-auto space-y-8 p-4 md:p-8">
      {/* Header Section */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-black text-black tracking-tight mb-1">
            My Knowledge Pods
          </h1>
          <p className="text-zinc-500 font-medium">
            Manage and explore your personal collection of knowledge repositories.
          </p>
        </div>
        <div className="flex gap-2 min-w-max">
          <button className="px-4 py-2  border-2 text-base border-black font-bold shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] hover:bg-[#FF8811] active:translate-x-[1px] hover:shadow-[2px_2px_0px_0px_#2B2D42] hover:text-white cursor-pointer hover:translate-x-[2px] hover:translate-y-[2px] transition-all group leading-none">
            Newest
          </button>
          {/* Assuming we have a create route, or this is just a button for now */}
          <button className="px-4 py-2  border-2 border-black font-bold shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] active:translate-x-[1px] hover:shadow-[2px_2px_0px_0px_#2B2D42] hover:translate-x-[2px] cursor-pointer hover:bg-[#FF8811] hover:text-white hover:translate-y-[2px] transition-all group leading-none">
            Upload Pod
          </button>
        </div>
      </div>

      {/* Main Container - Industrial / Neo-brutalism Style */}
      <div className="bg-white border-2 border-r-4 border-black rounded-[12px] overflow-hidden shadow-[6px_6px_0px_0px_#FF8811]">
        <div className="flex flex-col">
          {pods.length === 0 ? (
            <div className="p-8 text-center text-zinc-500">
              You haven't created any pods yet.
            </div>
          ) : (
            pods.map((pod, index) => (
              <FileListItem
                key={pod.id}
                variant="pod"
                userId="me" // This triggers the /me/... route which maps to [username]
                podId={pod.id}
                title={pod.title}
                description={pod.description}
                date={pod.date}
                onToggleLike={handleToggleLike}
                isLast={index === pods.length - 1}
                isPersonal={true}
                isLiked={pod.isLiked}
              />
            ))
          )}
        </div>
      </div>

      {/* Pagination / Industrial Footer */}
      {totalPages > 1 && (
        <div className="flex justify-center pt-4">
          <nav className="flex items-center gap-1">
            {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
              <button
                key={p}
                onClick={() => setPage(p)}
                className={`w-10 h-10 flex items-center justify-center border-2 border-black font-bold shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] transition-all ${page === p ? 'bg-orange-500 text-white' : 'bg-white text-black hover:bg-zinc-100'
                  }`}
              >
                {p}
              </button>
            ))}
          </nav>
        </div>
      )}
    </div>
  );
};

export default MyKnowledgePage;
