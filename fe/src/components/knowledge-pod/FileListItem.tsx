import React from 'react';
import { FileText, Heart } from 'lucide-react';

interface FileListItemProps {
  title: string;
  description: string;
  likes: string;
  date: string;
  isLast?: boolean;
}

const FileListItem: React.FC<FileListItemProps> = ({ title, description, likes, date, isLast }) => {
  return (
    <a className={`p-4 py-4 flex items-center gap-4 cursor-pointer transition-colors hover:bg-zinc-50 ${!isLast ? 'border-b border-black' : ''}`}>
      {/* File Icon */}
      <div className="shrink-0">
        <div className="w-8 h-10 border-2 border-black rounded-sm flex items-center justify-center bg-white shadow-[3px_3px_0px_0px_rgba(0,0,0,1)]">
          <FileText size={32} strokeWidth={1.5} />
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 flex flex-col justify-between">
        <div>
          <h4 className="text-md font-bold text-orange-500">{title}</h4>
          <p className="text-zinc-600 text-xs leading-relaxed max-w-3xl">
            {description}
          </p>
        </div>
      </div>

      {/* Metadata */}
      <div className="flex justify-end items-end gap-6 self-end">
        <div className='flex gap-1 items-center'>
          <Heart size={12} className='text-zinc-400 stroke-3'/>
          <span className="font-mono text-xs font-bold text-zinc-400">{likes}</span>
        </div>
        <span className="font-mono text-xs font-bold text-zinc-400">{date}</span>
      </div>
    </a>
  );
};

export default FileListItem;