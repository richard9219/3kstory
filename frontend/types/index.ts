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
  job_id?: string;
  scene_id?: number;
  candidate_no?: number;
  task_type?: 'generate_video' | 'narration_video' | 'storyboard_shot_video' | 'director_cut';
  title?: string;
  video_id: string;
  provider: VideoProviderKind;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  video_url?: string;
  score?: number;
  score_detail?: Record<string, any>;
  rank?: number;
  is_selected?: boolean;
  error_msg?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface VideoJob {
  id: number;
  job_id: string;
  user_id: number;
  project_id: number;
  pipeline_type: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  queue_status?: 'queued' | 'processing' | 'done' | 'failed';
  candidate_count: number;
  quality_mode: 'fast' | 'quality';
  score_profile: 'default' | 'short_drama' | 'movie_narration';
  provider_mode: 'single' | 'multi';
  auto_pick: boolean;
  publish_threshold?: number;
  publish_gate_passed?: boolean;
  publish_block_reason?: string;
  started_at?: string;
  selected_task_id?: number;
  selected_video_id?: string;
  request_data?: Record<string, any>;
  result_data?: Record<string, any>;
  error_msg?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface VideoJobDetail {
  job: VideoJob;
  pipeline_status: Record<string, string>;
  candidates: VideoTask[];
  selected_video_id?: string;
  selected_task_id?: number;
  top_score?: number;
  score_profile?: string;
  quality_mode?: string;
  provider_mode?: string;
  candidate_summary?: Array<{ candidate_no: number; score: number; rank: number }>;
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
  track_index?: number;
  chapter?: string;
  shot_number: number;
  sort_order: number;
  title: string;
  description?: string;
  camera_language?: string;
  emotion_tone?: string;
  timeline_start_ms?: number;
  timeline_duration_ms?: number;
  transition_type?: 'cut' | 'fade' | 'wipe' | 'match';
  transition_duration_ms?: number;
  duration: number;
  aspect_ratio: '16:9' | '9:16';
  prompt?: string;
  negative_prompt?: string;
  reference_image_url?: string;
  clip_provider?: string;
  clip_video_id?: string;
  clip_video_url?: string;
  clip_status?: 'draft' | 'pending' | 'processing' | 'completed' | 'failed';
  clip_score?: number;
  clip_notes?: string;
  locked?: boolean;
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

export interface StoryboardTimeline {
  project_id: number;
  total_duration_ms: number;
  ready_shot_count: number;
  total_shot_count: number;
  latest_export?: VideoTask;
  shots: StoryboardShot[];
}

export interface DirectorPublishRecord {
  id: number;
  user_id: number;
  project_id: number;
  video_task_id: number;
  export_id: string;
  platform: PlatformKind;
  status: 'pending' | 'success' | 'failed';
  attempt_no: number;
  retried_from_id?: number;
  receipt_id?: string;
  remote_video_id?: string;
  remote_url?: string;
  request_payload?: Record<string, any>;
  response_payload?: Record<string, any>;
  error_msg?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface DirectorTemplate {
  id: number;
  user_id: number;
  project_id: number;
  name: string;
  slug: string;
  sample_frame_url?: string;
  sample_video_url?: string;
  prompt_prefix?: string;
  camera_language?: string;
  emotion_tone?: string;
  transition_type?: 'cut' | 'fade' | 'wipe' | 'match';
  transition_duration_ms?: number;
  genre_keywords?: string;
  weight_narrative: number;
  weight_visual: number;
  weight_emotion: number;
  weight_rhythm: number;
  weight_continuity: number;
  is_builtin?: boolean;
  created_at: string;
  updated_at: string;
}

export interface DirectorAutoStrategyResult {
  genre: string;
  applied: boolean;
  tune_percent: number;
  selected: DirectorTemplate;
  predicted_score: number;
  shot_updates: number;
}

export interface DirectorABCompareResult {
  genre: string;
  template_a: DirectorTemplate;
  template_b: DirectorTemplate;
  score_a: number;
  score_b: number;
  winner_template_id: number;
  winner_template: string;
  applied: boolean;
  rendered_export_id?: string;
  predicted_gain: number;
  compared_shot_count: number;
}

export interface StoryboardShotUpdatePayload {
  track_index?: number;
  chapter?: string;
  title?: string;
  description?: string;
  camera_language?: string;
  emotion_tone?: string;
  duration?: number;
  aspect_ratio?: '16:9' | '9:16';
  prompt?: string;
  negative_prompt?: string;
  reference_image_url?: string;
  timeline_start_ms?: number;
  timeline_duration_ms?: number;
  transition_type?: 'cut' | 'fade' | 'wipe' | 'match';
  transition_duration_ms?: number;
  locked?: boolean;
  status?: 'draft' | 'pending' | 'processing' | 'completed' | 'failed';
}

export interface StoryboardShotGeneratePayload {
  provider?: VideoProviderKind;
  model?: string;
  duration?: number;
  aspect_ratio?: '16:9' | '9:16';
  prompt?: string;
  negative_prompt?: string;
  reference_image_url?: string;
  workflow_path?: string;
  version_note?: string;
}
