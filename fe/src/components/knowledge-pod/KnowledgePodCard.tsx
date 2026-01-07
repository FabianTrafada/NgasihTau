import React, { useEffect } from "react";
import { Star, ArrowBigDownDash, ArrowBigUpDash } from "lucide-react";
import { KnowledgePodCardProps, Pod } from "@/types/pod";
import { useRouter } from "next/navigation";
import { getStarredPod, upvotePod, removeUpvotePod, downvotePod, removeDownvotePod } from "@/lib/api/pod";

const KnowledgePodCard: React.FC<KnowledgePodCardProps> = ({ pod, userId, userUsername, onToggleLike, isLast, isPersonal }) => {
  const router = useRouter();
  const [isStarred, setStarred] = React.useState(false);
  const [upvoteCount, setUpvoteCount] = React.useState(pod.upvote_count || 0);
  const [downvoteCount, setDownvoteCount] = React.useState(pod.downvote_count || 0);
  const [userVoteStatus, setUserVoteStatus] = React.useState<"upvote" | "downvote" | null>(null);
  const [isVoting, setIsVoting] = React.useState(false);

  const handleCardClick = () => {
    if (userUsername && pod.slug) {
      router.push(`/${userUsername}/${pod.slug}`);
    }
  };

  useEffect(() => {
    const fetchStarredPods = async () => {
      try {
        const starredPods = await getStarredPod(userId);
        const isStarred = starredPods.some((starredPod) => starredPod.id === pod.id);
        setStarred(isStarred);
        console.log(pod);
      } catch (error) {
        console.error("Error fetching starred pods:", error);
      }
    };

    fetchStarredPods();
  }, [pod.id, userId]);

  const handleUpvote = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (isVoting) return;

    try {
      setIsVoting(true);
      if (userVoteStatus === "upvote") {
        // Remove upvote
        await removeUpvotePod(pod.id);
        setUpvoteCount((prev) => Math.max(0, prev - 1));
        setUserVoteStatus(null);
      } else {
        // Add upvote (remove downvote if exists)
        if (userVoteStatus === "downvote") {
          await removeDownvotePod(pod.id);
          setDownvoteCount((prev) => Math.max(0, prev - 1));
        }
        await upvotePod(pod.id);
        setUpvoteCount((prev) => prev + 1);
        setUserVoteStatus("upvote");
      }
    } catch (error) {
      console.error("Error upvoting pod:", error);
    } finally {
      setIsVoting(false);
    }
  };

  const handleDownvote = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (isVoting) return;

    try {
      setIsVoting(true);
      if (userVoteStatus === "downvote") {
        // Remove downvote
        await removeDownvotePod(pod.id);
        setDownvoteCount((prev) => Math.max(0, prev - 1));
        setUserVoteStatus(null);
      } else {
        // Add downvote (remove upvote if exists)
        if (userVoteStatus === "upvote") {
          await removeUpvotePod(pod.id);
          setUpvoteCount((prev) => Math.max(0, prev - 1));
        }
        await downvotePod(pod.id);
        setDownvoteCount((prev) => prev + 1);
        setUserVoteStatus("downvote");
      }
    } catch (error) {
      console.error("Error downvoting pod:", error);
    } finally {
      setIsVoting(false);
    }
  };

  const year = pod.created_at.slice(0, 4);
  const month = pod.created_at.slice(5, 7);
  const day = pod.created_at.slice(8, 10);

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
            <Star size={18} className={`transition-colors ${isStarred ? "fill-[#FF8811] text-[#FF8811]" : "text-zinc-400 group-hover:text-red-400"}`} />
            <span className="text-sm font-medium text-zinc-600">{isStarred ? "Starred" : "Star"}</span>
          </button>
        )}
      </div>

      {/* Metadata Section - Industrial / Mono Style */}
      <div className="flex items-center justify-between md:flex-col md:items-end md:justify-start gap-4 md:w-32 shrink-0 pt-1">
        <div className="flex items-center gap-2 md:text-right">{/* <span className="mono text-sm font-medium text-zinc-500">{pod.fileCount} files</span> */}</div>
        <div className="flex items-center gap-2 md:text-right">
          <span className="mono text-xs font-medium text-zinc-400">
            {day} / <span>{month}</span> / {year}
          </span>
        </div>

        {!isPersonal && (
          <div className="flex items-center gap-2 md:text-right">
            <span className="mono text-xs font-medium text-zinc-400">{pod.visibility}</span>
          </div>
        )}

        <div className="flex border-2 border-zinc-500 cursor-pointer text-xs divide-x-2 rounded-sm">
          <button onClick={handleUpvote} disabled={isVoting} className={`p-1.5 px-3 flex items-center gap-1 hover:bg-gray-100 rounded-l-sm transition-colors disabled:opacity-50 ${userVoteStatus === "upvote" ? "bg-green-100" : ""}`}>
            <ArrowBigUpDash className={`size-4 ${userVoteStatus === "upvote" ? "fill-green-600 text-green-600" : ""}`} />
            <span>{upvoteCount}</span>
          </button>
          <button onClick={handleDownvote} disabled={isVoting} className={`p-1 px-3 flex items-center gap-1 hover:bg-gray-100 rounded-r-sm transition-colors disabled:opacity-50 ${userVoteStatus === "downvote" ? "bg-red-100" : ""}`}>
            <ArrowBigDownDash className={`size-4 ${userVoteStatus === "downvote" ? "fill-red-600 text-red-600" : ""}`} />
            <span>{downvoteCount}</span>
          </button>
        </div>
      </div>
    </div>
  );
};

export default KnowledgePodCard;
