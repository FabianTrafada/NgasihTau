/** UI shape for version history (GitHub-style) */
export interface MaterialVersion {
  id: string;
  version_number: string; // e.g. "v1.0"
  version: number; // numeric for restore API
  created_at: string;
  commit_message: string;
  author_name: string;
  author_avatar?: string;
  uploader_id?: string;
}

/** Raw response from GET /api/v1/materials/:id/versions */
export interface MaterialVersionFromAPI {
  id: string;
  material_id: string;
  version: number;
  file_url: string;
  file_size: number;
  uploader_id: string;
  changelog: string | null;
  created_at: string;
}
