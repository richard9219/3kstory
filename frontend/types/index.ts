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

// Video generation types
export interface VideoTask {
  id: number;
  user_id: number;
  project_id: number;
  scene_id?: number;
  task_type?: 'generate_video' | 'narration_video';
  title?: string;
  video_id: string;
  provider: 'runway' | 'pika' | 'local';
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
