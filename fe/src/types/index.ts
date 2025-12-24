export interface KnowledgePod {
  id: string;
  title: string;
  description: string;
  fileCount: number;
  date: string;
  isLiked?: boolean;
}

// Interest types
export interface Interest {
  id: string;
  name: string;
  slug?: string;
  description?: string;
  icon?: string;
  category: string;
  display_order?: number;
  is_active?: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface InterestCategory {
  name: string;
  interests: Interest[];
}

export interface UserInterest {
  interest_id: string;
  interest_name: string;
  category?: string;
}