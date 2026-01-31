export interface Material {
  id: string;
  pod_id: string;
  uploader_id: string;
  title: string;
  description?: string;
  file_type: string; // "pdf" | "docx" | "pptx"
  file_url: string;
  file_size: number;
  current_version: number;
  status: string; // "processing" | "ready" | "processing_failed"
  view_count: number;
  download_count: number;
  average_rating: number;
  rating_count: number;
  created_at: string;
  updated_at: string;
}

// Jika backend return dengan uploader info (optional)
export interface MaterialWithUploader extends Material {
  uploader_name: string;
  uploader_avatar_url?: string;
}

// Chat session untuk material
export interface ChatSession {
  id: string;
  user_id: string;
  material_id?: string;
  pod_id?: string;
  mode: "material" | "pod";
  created_at: string;
  updated_at: string;
}

// Chat message
export interface ChatMessage {
  id: string;
  session_id: string;
  role: "user" | "assistant";
  content: string;
  sources?: ChunkSource[];
  feedback?: "thumbs_up" | "thumbs_down";
  feedback_text?: string;
  created_at: string;
}

// Chunk source dari material untuk AI references
export interface ChunkSource {
  material_id: string;
  chunk_index: number;
  text: string;
  score: number;
}



export interface EditingMaterial {
  id: string;
  title: string;
  description: string;
}

export interface UploadUrlRequest {
  filenmae: string;
  content_type: string;
  size: number; 
}

export interface UploadUrlResponse {
  upload_url: string;
  object_key: string;
  expires_at: string;
}

export interface UpdateMaterialInput {
  object_key: string;
  pod_id: string;
  title: string;
  description?: string;
}