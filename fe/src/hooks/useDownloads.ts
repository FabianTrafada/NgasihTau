
import { useState, useEffect } from 'react';
import { Material } from '@/types/material';

const STORAGE_KEY = 'downloaded_materials';

export interface DownloadedMaterial extends Material {
    downloadedAt: string;
}

export function useDownloads() {
    const [downloads, setDownloads] = useState<DownloadedMaterial[]>([]);

    useEffect(() => {
        const saved = localStorage.getItem(STORAGE_KEY);
        if (saved) {
            try {
                setDownloads(JSON.parse(saved));
            } catch (e) {
                console.error('Failed to parse downloaded materials', e);
            }
        }
    }, []);

    const saveToStorage = (materials: DownloadedMaterial[]) => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(materials));
        setDownloads(materials);
    };

    const addMaterial = (material: Material) => {
        if (downloads.some((d) => d.id === material.id)) return;

        const newDownload: DownloadedMaterial = {
            ...material,
            downloadedAt: new Date().toISOString(),
        };

        const newDownloads = [newDownload, ...downloads];
        saveToStorage(newDownloads);
    };

    const removeMaterial = (id: string) => {
        const newDownloads = downloads.filter((d) => d.id !== id);
        saveToStorage(newDownloads);
    };

    const isDownloaded = (id: string) => {
        return downloads.some((d) => d.id === id);
    };

    return {
        downloads,
        addMaterial,
        removeMaterial,
        isDownloaded,
    };
}
