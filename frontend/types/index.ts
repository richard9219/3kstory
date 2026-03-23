// User types
export interface User {
  id: number;
  email: string;
  username: string;
  role: 'user' | 'admin';
  points: number;
  avatar?: string;
  created_at: string;
}

// Project types
export interface Project {
  id: number;
  user_id: number;
  title: string;
  prompt: string;
  status: 'draft' | 'processing' | 'completed' | 'failed';
  scenes?: Scene[];
  created_at: string;
  updated_at: string;
}

// Scene types
export interface Scene {
  id: number;
  project_id: number;
  scene_number: number;
  title: string;
  location: string;
  characters: string[];
  dialogue: string;
  shot_type: string;
  duration: number;
  image_url?: string;
  video_url?: string;
  created_at: string;
}

export type VideoProviderKind = 'runway' | 'pika' | 'local' | 'minimax' | 'seedance' | 'comfy';

// Video generation types
export interface VideoTask {
  id: number;
  user_id: number;
  project_id: number;
  scene_id?: number;
  task_type?: 'generate_video' | 'narration_video';
  title?: string;
  video_id: string;
  provider: VideoProviderKind;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  video_url?: string;
  error_msg?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

// API response types
export interface APIResponse<T> {
  data: T;
  message?: string;
}

export interface ListResponse<T> {
  total: number;
  data: T[];
}

// Auth types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  username: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// 第三方平台账号（里程碑一）
export type PlatformKind = 'douyin' | 'xiaohongshu' | 'bilibili' | 'weibo';

export interface PlatformAccount {
  id: number;
  platform: PlatformKind;
  nickname: string;
  avatar_url: string;
  created_at: string;
}

export interface AnalyticsSummary {
  total_projects: number;
  total_generated_videos: number;
  total_completed_videos: number;
  total_interactions: number;
  bound_platforms: PlatformAccount[];
  configured_platforms: PlatformKind[];
}

export interface ProviderHealth {
  provider: VideoProviderKind;
  configured: boolean;
  healthy: boolean;
  message: string;
  error_kind?: string;
  checked_at: string;
}

export interface ModelProviderStatus {
  name: string;
  category: 'text' | 'image' | 'tts' | 'video';
  configured: boolean;
  healthy: boolean;
  message: string;
  checked_at: string;
}

export interface ModelCenterOverview {
  preferred_video_provider: VideoProviderKind;
  video_providers: ProviderHealth[];
  text_providers: ModelProviderStatus[];
  image_providers: ModelProviderStatus[];
  tts_providers: ModelProviderStatus[];
  task_routes: {
    task: string;
    category: 'text' | 'video';
    providers: string[];
  }[];
  alerts: ModelProviderAlert[];
  probe_task: ModelProbeTaskStatus;
}

export interface ModelProviderAlert {
  name: string;
  category: 'text' | 'image' | 'tts' | 'video';
  failure_streak: number;
  failure_threshold: number;
  alerting: boolean;
  last_failure_at?: string;
  last_success_at?: string;
  last_message?: string;
  updated_at: string;
}

export interface ModelProbeTaskStatus {
  enabled: boolean;
  interval_seconds: number;
  failure_threshold: number;
  last_probe_at?: string;
  next_probe_at?: string;
  running: boolean;
}

export interface RoleAsset {
  id: number;
  user_id: number;
  project_id?: number;
  name: string;
  role_type?: string;
  description?: string;
  avatar_url?: string;
  voice_preset?: string;
  style_prompt?: string;
  negative_hint?: string;
  tags?: string;
  created_at: string;
  updated_at: string;
}

export interface PromptTemplate {
  id: number;
  user_id: number;
  project_id?: number;
  name: string;
  template_type: string;
  provider_type?: string;
  content: string;
  variables?: string;
  tags?: string;
  created_at: string;
  updated_at: string;
}

export interface StoryboardShot {
  id: number;
  user_id: number;
  project_id: number;
  scene_id?: number;
  chapter?: string;
  shot_number: number;
  sort_order: number;
  title: string;
  description?: string;
  camera_language?: string;
  duration: number;
  aspect_ratio: '16:9' | '9:16';
  prompt?: string;
  negative_prompt?: string;
  reference_image_url?: string;
  status: 'draft' | 'pending' | 'processing' | 'completed' | 'failed';
  version: number;
  parent_shot_id?: number;
  root_shot_id?: number;
  version_note?: string;
  created_at: string;
  updated_at: string;
}

export interface StoryboardVersionNode {
  root: StoryboardShot;
  versions: StoryboardShot[];
}
