import React from 'react';
import { FileText, Heart } from 'lucide-react';
import { useRouter } from 'next/navigation';

export interface BaseListItemProps {
  title: string;
  description: string;
  date: string;
  isLast?: boolean;
  onBoxClick?: () => void;
}

export interface FileItemVariant extends BaseListItemProps {
  variant: 'file';
  materialId?: string;
  podId?: string;
  userId?: string;
  likes?: string | number;
}

export interface PodItemVariant extends BaseListItemProps {
  variant: 'pod';
  podId: string;
  userId: string;
  visibility?: 'public' | 'private';
  isPersonal?: boolean;
  onToggleLike?: (id: string) => void;
  isLiked?: boolean;
  fileCount?: number;
}

export type FileListItemProps = FileItemVariant | PodItemVariant;

const FileListItem: React.FC<FileListItemProps> = (props) => {
  const router = useRouter();
  const { title, description, date, isLast, onBoxClick } = props;

  const handleCardClick = () => {
    if (onBoxClick) {
      onBoxClick();
      return;
    }

    if (props.variant === 'file') {
      if (props.userId && props.podId && props.materialId) {
        router.push(`/${props.userId}/${props.podId}/${props.materialId}`);
      }
    } else if (props.variant === 'pod') {
      if (props.userId && props.podId) {
        router.push(`/${props.userId}/${props.podId}`);
      }
    }
  };

  const isFile = props.variant === 'file';

  return (
    <div
      onClick={handleCardClick}
      className={`p-4 py-4 flex flex-col md:flex-row md:items-start gap-4 cursor-pointer transition-colors hover:bg-zinc-50 ${!isLast ? 'border-b border-black' : ''}`}
    >
      {/* Icon (Only for File) */}
      {isFile && (
        <div className="shrink-0">
          <div className="w-8 h-10 border-2 border-black rounded-sm flex items-center justify-center bg-white shadow-[3px_3px_0px_0px_rgba(0,0,0,1)]">
            <FileText size={32} strokeWidth={1.5} />
          </div>
        </div>
      )}

      {/* Content */}
      <div className="flex-1 flex flex-col justify-between space-y-2">
        <div>
          <h4 className={`font-bold ${isFile ? 'text-md text-orange-500' : 'text-xl text-[#FF8811] leading-tight'}`}>
            {title || "Untitled"}
          </h4>
          <p className={`text-zinc-600 leading-relaxed max-w-3xl ${isFile ? 'text-xs' : 'text-sm md:text-base'}`}>
            {description || "No description available."}
          </p>
        </div>

        {/* Pod Like Button */}
        {!isFile && (props as PodItemVariant).isPersonal && (props as PodItemVariant).onToggleLike && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              (props as PodItemVariant).onToggleLike?.((props as PodItemVariant).podId);
            }}
            className="flex items-center gap-2 group mt-4 w-fit"
          >
            <span className="text-sm font-medium text-zinc-600">
              Liked
            </span>
          </button>
        )}
      </div>

      {/* Metadata */}
      <div className="flex justify-end items-end gap-6 self-end md:flex-col md:items-end md:justify-start md:gap-2 shrink-0">
        {isFile ? (
          <>
            <div className='flex gap-1 items-center'>
              <Heart size={12} className='text-zinc-400 stroke-3' />
              <span className="font-mono text-xs font-bold text-zinc-400">{(props as FileItemVariant).likes || 0}</span>
            </div>
            <span className="font-mono text-xs font-bold text-zinc-400">{date}</span>
          </>
        ) : (
          // Pod Metadata
          <>
            <div className="flex items-center gap-2 md:text-right">
              <span className="font-mono text-xs font-medium text-zinc-400">{date}</span>
            </div>
            {!(props as PodItemVariant).isPersonal && (props as PodItemVariant).visibility && (
              <div className="flex items-center gap-2 md:text-right">
                <span className="font-mono text-xs font-medium text-zinc-400 uppercase">{(props as PodItemVariant).visibility}</span>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

export default FileListItem;