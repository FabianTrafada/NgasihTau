export interface KnowledgePod {
    id: string;
    title: string;
    description: string;
    fileCount: number;
    date: string;
    isLiked?: boolean;
    isPublic?: "public" | "private";

}

export interface KnowledgePodCardProps {
  pod: KnowledgePod;
  onToggleLike: (id: string) => void;
  isLast?: boolean;
  isPersonal?: boolean;
}