export interface Pod {
  id: string;
  owner_id: string;
  name: string;
  description?: string;
  visibility: "public" | "private";
  view_count: number;
  star_count: number;
  fork_count: number;
  categories?: string[];
  tags?: string[];
  slug?: string;
  created_at: string;
  updated_at: string;
  forked_from_id?: string;
}

export interface PodWithOwner extends Pod {
  owner_name: string;
  owner_title?: string;
}

export interface KnowledgePodCardProps {
  pod: Pod;
  userId: string;
  onToggleLike: (id: string) => void;
  isLast?: boolean;
  isPersonal?: boolean;
  isPublic?: "Public" | "Private";
}