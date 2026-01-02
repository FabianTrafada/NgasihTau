
export type Visibility = 'private' | 'public';

export interface KnowledgePodData {
  name: string;
  description: string;
  materialFiles: File[];
  visibility: Visibility;
}

export enum Step {
  General = 1,
  Material = 2,
  Configurator = 3
}
