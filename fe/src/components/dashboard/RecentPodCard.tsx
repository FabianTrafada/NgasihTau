import { Heart } from 'lucide-react';
import { title } from 'process'
import React from 'react'


interface RecentPodCardProps {
    title: string;
    description: string;
    fileCount: number;
    date: string;
}

const RecentPodCard = ({ title, description, fileCount, date }: RecentPodCardProps) => {
    return (
        <div className='bg-white border-gray-200 rounded-xl p-5 hover:shadow-sm transition-shadow duration-200 flex flex-col justify-between h-full relative group'>

            <div className="absolute inset-0 border-2 border-transparent group-hover:border-gray-900/5 rounded-xl pointer-events-none transition-colors" />

            <div>
                <h3 className='text-[#ff8811] font-bold text-lg mb-2 line-clamp-2'>
                    {title}
                </h3>
                <p className="text-gray-500 text-sm leading-relaxed line-clamp-3 mb-4 font-mono text-xs">
                    {description}
                </p>
            </div>

            <div className="flex items-center justify-between mt-auto pt-4 border-t border-gray-50">
                <div className="flex items-center gap-4 text-xs text-gray-400 font-medium">
                    <span className="flex items-center gap-1">
                        <span className="text-gray-900 font-bold text-sm">{fileCount}</span> files
                    </span>
                    <span>{date}</span>
                </div>

                <button className="flex items-center gap-1.5 text-xs font-medium text-gray-500 hover:text-red-500 transition-colors">
                    <Heart className="w-4 h-4" />
                    Liked
                </button>
            </div>

        </div>
    )
}

export default RecentPodCard