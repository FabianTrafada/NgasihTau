import React from "react";
import { Heart } from "lucide-react";
import { KnowledgePodCardProps } from "@/types/pod";
import { useRouter } from "next/navigation";

const KnowledgePodCard: React.FC<KnowledgePodCardProps> = ({ pod, userId, onToggleLike, isLast, isPersonal }) => {
  const router = useRouter();

  const handleCardClick = () => {
    if (userId) {
      router.push(`/${userId}/${pod.id}`);
    }
  };

  return (
    <div onClick={handleCardClick} className={`p-6 flex flex-col md:flex-row md:items-start gap-4 transition-colors hover:bg-zinc-50/50 ${!isLast ? "border-b border-black" : ""}`}>
      {/* Content Section */}
      <div className="flex-1 space-y-2">
        <h3 className="text-xl font-bold text-[#FF8811] leading-tight">{pod.name}</h3>
        <p className="text-zinc-500 text-sm md:text-base leading-relaxed max-w-2xl">{pod.description}</p>

        {/* Interaction */}
        {isPersonal && (
          <button onClick={() => onToggleLike(pod.id)} className="flex items-center gap-2 group mt-4">
            {/* <Heart
              size={18}
              className={`transition-colors ${pod.isLiked ? 'fill-red-500 text-red-500' : 'text-zinc-400 group-hover:text-red-400'}`}
            /> */}
            <span className="text-sm font-medium text-zinc-600">Liked</span>
          </button>
        )}
      </div>

      {/* Metadata Section - Industrial / Mono Style */}
      <div className="flex items-center justify-between md:flex-col md:items-end md:justify-start gap-4 md:w-32 shrink-0 pt-1">
        <div className="flex items-center gap-2 md:text-right">{/* <span className="mono text-sm font-medium text-zinc-500">{pod.fileCount} files</span> */}</div>
        <div className="flex items-center gap-2 md:text-right">
          <span className="mono text-xs font-medium text-zinc-400">{pod.created_at}</span>
        </div>

        {!isPersonal && (
          <div className="flex items-center gap-2 md:text-right">
            <span className="mono text-xs font-medium text-zinc-400">{pod.visibility}</span>
          </div>
        )}
      </div>
    </div>
  );
};

export default KnowledgePodCard;
