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
        <div className='bg-white border-2 border-[#2B2D42] p-5 shadow-[6px_6px_0px_0px_#2B2D42] hover:shadow-[3px_3px_0px_0px_#2B2D42] hover:translate-x-[3px] hover:translate-y-[3px] transition-all duration-200 flex flex-col justify-between h-full relative group'>

            <div>
                <h3 className='text-[#ff8811] font-bold text-lg mb-2 line-clamp-2'>
                    {title}
                </h3>
                <p className="text-gray-500 text-sm leading-relaxed line-clamp-3 mb-4 font-mono text-xs">
                    {description}
                </p>
            </div>

            <div className="flex items-center justify-between mt-auto pt-4 border-t-2 border-gray-100">
                <div className="flex items-center gap-4 text-xs text-gray-500 font-medium">
                    <span className="flex items-center gap-1">
                        <span className="text-[#2B2D42] font-bold text-sm">{fileCount}</span> files
                    </span>
                    <span>{date}</span>
                </div>

                <button className="flex items-center gap-1.5 text-xs font-bold text-[#2B2D42] hover:text-[#FF8811] transition-colors">
                    <Heart className="w-4 h-4" />
                    Liked
                </button>
            </div>

        </div>
    )
}

export default RecentPodCard