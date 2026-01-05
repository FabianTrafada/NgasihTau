export interface KnowledgePod {
    id: string;
    title: string;
    description: string;
    fileCount: number;
    date: string;
    isLiked?: boolean;
    isPublic?: "public" | "private";

}