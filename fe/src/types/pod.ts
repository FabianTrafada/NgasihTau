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

// Material dalam pod
export interface PodMaterial {
  id: string;
  title: string;
  description?: string;
  file_type: string;
  file_url: string;
  view_count: number;
  download_count: number;
  average_rating: number;
  created_at: string;
}
