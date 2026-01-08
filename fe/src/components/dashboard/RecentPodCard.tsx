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
        <div className='bg-white border-2 border-[#2B2D42] p-3 shadow-[4px_4px_0px_0px_#2B2D42] hover:shadow-[2px_2px_0px_0px_#FF8811] hover:translate-x-[2px] hover:translate-y-[2px] transition-all duration-200 flflex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 sm:gap-0 mt-auto pt-3
  border-t-2 '>
            <div>
                <h3 className='text-[#ff8811] font-bold text-base mb-1.5 line-clamp-1'>
                    {title}
                </h3>
                <p className="text-gray-500 leading-tight line-clamp-2 mb-3 font-mono text-[10px] sm:text-xs">
                    {description}
                </p>
            </div>

            <div className="flex items-center justify-between mt-auto pt-3 border-t-2 border-gray-100">
                <div className="flex items-center gap-3 text-[10px] sm:text-xs text-gray-500 font-medium">
                    <span className="flex items-center gap-1">
                        <span className="text-[#2B2D42] font-bold">{fileCount}</span> files
                    </span>
                    <span>{date}</span>
                </div>

                <button className="flex items-center gap-1 text-[10px] sm:text-xs font-bold text-[#2B2D42] hover:text-[#FF8811] transition-colors">
                    <Heart className="size-3.5" />
                    Liked
                </button>
            </div>

        </div>
    )
}

export default RecentPodCard