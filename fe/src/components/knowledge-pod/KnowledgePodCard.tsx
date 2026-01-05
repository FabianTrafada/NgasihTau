import React, { useEffect } from "react";
import { Heart } from "lucide-react";
import { KnowledgePodCardProps, Pod } from "@/types/pod";
import { useRouter } from "next/navigation";
import { getStarredPod } from "@/lib/api/pod";

const KnowledgePodCard: React.FC<KnowledgePodCardProps> = ({ pod, userId, onToggleLike, isLast, isPersonal }) => {
  const router = useRouter();
  const [isStarred, setStarred] = React.useState(false);

  const handleCardClick = () => {
    if (userId) {
      router.push(`/${userId}/${pod.id}`);
    }
  };

  useEffect(() => {
    const fetchStarredPods = async () => {
      try {
        const starredPods = await getStarredPod(userId);
        const isStarred = starredPods.some((starredPod) => starredPod.id === pod.id);
        setStarred(isStarred);
      } catch (error) {
        console.error("Error fetching starred pods:", error);
      }
    };

    fetchStarredPods();
  }, [pod.id, userId]);

  return (
    <div onClick={handleCardClick} className={`p-6 flex flex-col md:flex-row md:items-start gap-4 transition-colors hover:bg-zinc-50/50 ${!isLast ? "border-b border-black" : ""}`}>
      {/* Content Section */}
      <div className="flex-1 space-y-2">
        <h3 className="text-xl font-bold text-[#FF8811] leading-tight">{pod.name}</h3>
        <p className="text-zinc-500 text-sm md:text-base leading-relaxed max-w-2xl">{pod.description}</p>

        {/* Interaction */}
        {isPersonal && (
          <button
            onClick={async (e) => {
              e.stopPropagation();
              try {
                await onToggleLike(pod.id, isStarred);
                setStarred(!isStarred);
                console.log("Toggled like for pod:", pod.id, "isStarred:", !isStarred);
              } catch (error) {
                console.error("Error toggling star:", error);
              }
            }}
            className="flex items-center gap-2 group mt-4"
          >
            <Heart size={18} className={`transition-colors ${isStarred ? "fill-red-500 text-red-500" : "text-zinc-400 group-hover:text-red-400"}`} />
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
