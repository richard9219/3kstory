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
  project_id: number;
  scene_id: number;
  video_id: string;
  provider: 'runway' | 'pika';
  status: 'pending' | 'processing' | 'completed' | 'failed';
  video_url?: string;
  error_msg?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
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
