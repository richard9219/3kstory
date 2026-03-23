import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const DEMO_MODE = String(process.env.NEXT_PUBLIC_DEMO_MODE || 'auto').toLowerCase();

// Create axios instance
export const apiClient = axios.create({
  baseURL: `${API_URL}/api/v1`,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

type FallbackAxiosResponse = {
  data: any;
  status: number;
  statusText: string;
  headers: Record<string, string>;
  config: any;
};

const FALLBACK_HEADER_KEY = 'x-frontend-fallback';
const FALLBACK_USER = {
  id: 1,
  email: 'demo@3kstory.local',
  username: 'Demo Operator',
  role: 'admin' as const,
  points: 999,
  created_at: new Date().toISOString(),
};

function isNetworkUnavailableError(error: any) {
  return !error?.response && (
    error?.code === 'ERR_NETWORK' ||
    error?.code === 'ECONNABORTED' ||
    error?.code === 'ECONNREFUSED' ||
    /network|fetch|timeout|ECONNREFUSED/i.test(String(error?.message || ''))
  );
}

function isDemoModeOn() {
  return DEMO_MODE === 'on' || DEMO_MODE === 'force';
}

function isDemoModeOff() {
  return DEMO_MODE === 'off';
}

function normalizeApiPath(url?: string) {
  if (!url) return '/';
  const noQuery = url.split('?')[0] || '/';
  const withoutBase = noQuery.replace(/^https?:\/\/[^/]+/, '');
  const clean = withoutBase.startsWith('/') ? withoutBase : `/${withoutBase}`;
  return clean.replace(/^\/api\/v1/, '') || '/';
}

function makeFallbackResponse(config: any, data: any, status = 200): FallbackAxiosResponse {
  return {
    data,
    status,
    statusText: 'OK',
    headers: { [FALLBACK_HEADER_KEY]: '1' },
    config,
  };
}

function nowMinusMinutes(minutes: number) {
  return new Date(Date.now() - minutes * 60000).toISOString();
}

function matchRoute(path: string, route: string) {
  if (route === path) return true;
  if (!route.includes('*')) return false;
  const escaped = route.replace(/[.+?^${}()|[\]\\]/g, '\\$&');
  const regex = new RegExp(`^${escaped.replace(/\*/g, '[^/]+')}$`);
  return regex.test(path);
}

function buildFallbackData(config: any): FallbackAxiosResponse | null {
  const method = String(config?.method || 'get').toLowerCase();
  const path = normalizeApiPath(config?.url);

  const demoProjects = [
    {
      id: 101,
      user_id: 1,
      title: '药命效应 - 第一集',
      prompt: '基于真实案件改编的影视解说内容',
      status: 'processing',
      created_at: nowMinusMinutes(2200),
      updated_at: nowMinusMinutes(12),
    },
    {
      id: 102,
      user_id: 1,
      title: '商业短剧试播片',
      prompt: 'AI 短剧 60 秒预告',
      status: 'completed',
      created_at: nowMinusMinutes(4100),
      updated_at: nowMinusMinutes(630),
    },
  ];

  const demoVideos = [
    {
      id: 9001,
      user_id: 1,
      project_id: 101,
      task_type: 'narration_video',
      title: '药命效应 EP01',
      video_id: 'demo_narration_9001',
      provider: 'runway',
      status: 'completed',
      created_at: nowMinusMinutes(420),
      completed_at: nowMinusMinutes(365),
      updated_at: nowMinusMinutes(365),
    },
    {
      id: 9002,
      user_id: 1,
      project_id: 101,
      task_type: 'generate_video',
      title: '镜头 03',
      video_id: 'demo_scene_9002',
      provider: 'minimax',
      status: 'processing',
      created_at: nowMinusMinutes(95),
      updated_at: nowMinusMinutes(9),
    },
    {
      id: 9003,
      user_id: 1,
      project_id: 102,
      task_type: 'generate_video',
      title: '镜头 12',
      video_id: 'demo_scene_9003',
      provider: 'comfy',
      status: 'failed',
      error_msg: '演示模式：后端未连接',
      created_at: nowMinusMinutes(720),
      updated_at: nowMinusMinutes(700),
    },
  ];

  const routes: Array<{
    method: string;
    route: string;
    handler: () => FallbackAxiosResponse;
  }> = [
    {
      method: 'post',
      route: '/auth/login',
      handler: () => makeFallbackResponse(config, { token: 'demo-token', user: FALLBACK_USER }),
    },
    {
      method: 'post',
      route: '/auth/register',
      handler: () => makeFallbackResponse(config, { token: 'demo-token', user: FALLBACK_USER }),
    },
    {
      method: 'get',
      route: '/users/me',
      handler: () => makeFallbackResponse(config, { data: FALLBACK_USER }),
    },
    {
      method: 'get',
      route: '/analytics/summary',
      handler: () =>
        makeFallbackResponse(config, {
          data: {
            total_projects: demoProjects.length,
            total_generated_videos: demoVideos.length,
            total_completed_videos: demoVideos.filter((v) => v.status === 'completed').length,
            total_interactions: 18230,
            bound_platforms: [
              {
                id: 1,
                platform: 'douyin',
                nickname: '三千视界官方号',
                avatar_url: '',
                created_at: nowMinusMinutes(2600),
              },
            ],
            configured_platforms: ['douyin', 'bilibili'],
          },
        }),
    },
    {
      method: 'get',
      route: '/analytics/videos',
      handler: () => makeFallbackResponse(config, { total: demoVideos.length, data: demoVideos }),
    },
    {
      method: 'get',
      route: '/projects',
      handler: () => makeFallbackResponse(config, demoProjects),
    },
    {
      method: 'post',
      route: '/projects/*/generate-narration',
      handler: () =>
        makeFallbackResponse(config, {
          status: 'queued',
          video_id: `demo_${Date.now()}`,
          note: '演示模式：后端不可达，任务仅在前端模拟。',
        }),
    },
    {
      method: 'get',
      route: '/platforms/configured',
      handler: () => makeFallbackResponse(config, { data: ['douyin', 'bilibili'] }),
    },
    {
      method: 'get',
      route: '/platforms',
      handler: () =>
        makeFallbackResponse(config, {
          data: [
            {
              id: 1,
              platform: 'douyin',
              nickname: '三千视界官方号',
              avatar_url: '',
              created_at: nowMinusMinutes(3800),
            },
          ],
        }),
    },
    {
      method: 'get',
      route: '/platforms/*/authorize',
      handler: () => makeFallbackResponse(config, { authorize_url: '/platform-bound?mock=1' }),
    },
    {
      method: 'delete',
      route: '/platforms/*',
      handler: () => makeFallbackResponse(config, { ok: true }),
    },
    {
      method: 'get',
      route: '/model-center/overview',
      handler: () =>
        makeFallbackResponse(config, {
          data: {
            preferred_video_provider: 'runway',
            video_providers: [
              {
                provider: 'runway',
                configured: true,
                healthy: true,
                message: '演示模式：服务正常',
                checked_at: nowMinusMinutes(2),
              },
              {
                provider: 'pika',
                configured: true,
                healthy: true,
                message: '演示模式：服务正常',
                checked_at: nowMinusMinutes(3),
              },
              {
                provider: 'local',
                configured: false,
                healthy: false,
                message: '演示模式：未配置',
                checked_at: nowMinusMinutes(3),
              },
            ],
            text_providers: [],
            image_providers: [],
            tts_providers: [],
            task_routes: [
              { task: 'narration_video', category: 'video', providers: ['runway', 'pika'] },
              { task: 'scene_video', category: 'video', providers: ['runway', 'pika', 'local'] },
            ],
            alerts: [],
            probe_task: {
              enabled: true,
              interval_seconds: 120,
              failure_threshold: 3,
              last_probe_at: nowMinusMinutes(2),
              next_probe_at: nowMinusMinutes(-2),
              running: false,
            },
          },
        }),
    },
    {
      method: 'post',
      route: '/model-center/probe',
      handler: () =>
        makeFallbackResponse(config, {
          data: {
            preferred_video_provider: 'runway',
            video_providers: [
              {
                provider: 'runway',
                configured: true,
                healthy: true,
                message: '演示模式：探活成功',
                checked_at: new Date().toISOString(),
              },
            ],
            text_providers: [],
            image_providers: [],
            tts_providers: [],
            task_routes: [],
            alerts: [],
            probe_task: {
              enabled: true,
              interval_seconds: 120,
              failure_threshold: 3,
              last_probe_at: new Date().toISOString(),
              next_probe_at: nowMinusMinutes(-2),
              running: false,
            },
          },
        }),
    },
    {
      method: 'get',
      route: '/assets/roles',
      handler: () => makeFallbackResponse(config, { total: 0, data: [] }),
    },
    {
      method: 'get',
      route: '/assets/prompt-templates',
      handler: () => makeFallbackResponse(config, { total: 0, data: [] }),
    },
    {
      method: 'get',
      route: '/projects/*/storyboard-shots',
      handler: () => makeFallbackResponse(config, { total: 0, data: [] }),
    },
    {
      method: 'get',
      route: '/projects/*/storyboard-shots/version-tree',
      handler: () => makeFallbackResponse(config, { total: 0, data: [] }),
    },
  ];

  const matched = routes.find((entry) => entry.method === method && matchRoute(path, entry.route));
  return matched ? matched.handler() : null;
}

export function isFallbackResponse(response: any) {
  return String(response?.headers?.[FALLBACK_HEADER_KEY] || '') === '1';
}

// Request interceptor - add auth token
apiClient.interceptors.request.use(
  (config) => {
    if (typeof window !== 'undefined' && isDemoModeOn()) {
      const fallback = buildFallbackData(config);
      const response = fallback || makeFallbackResponse(config, {
        error: '演示模式已开启，当前接口未配置兜底数据。',
      }, 503);
      sessionStorage.setItem('frontend-offline-fallback', '1');
      config.adapter = async () => response as any;
    }

    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor - handle errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (!isDemoModeOff() && isNetworkUnavailableError(error)) {
      const fallback = buildFallbackData(error?.config);
      if (fallback) {
        if (typeof window !== 'undefined') {
          sessionStorage.setItem('frontend-offline-fallback', '1');
        }
        return Promise.resolve(fallback);
      }
    }

    if (error.response?.status === 401) {
      // Token expired or invalid
      if (typeof window !== 'undefined') {
        localStorage.removeItem('token');
        window.location.href = '/';
      }
    }
    return Promise.reject(error);
  }
);

// Auth API
export const authAPI = {
  register: (data: { email: string; password: string; username: string }) =>
    apiClient.post('/auth/register', data),
    
  login: (data: { email: string; password: string }) =>
    apiClient.post('/auth/login', data),
    
  getProfile: () =>
    apiClient.get('/users/me'),
    
  updateProfile: (data: any) =>
    apiClient.put('/users/me', data),
};

// Project API
export const projectAPI = {
  create: (data: { title: string; prompt: string }) =>
    apiClient.post('/projects', data),
    
  list: (params?: { page?: number; limit?: number }) =>
    apiClient.get('/projects', { params }),
    
  get: (id: number) =>
    apiClient.get(`/projects/${id}`),
    
  update: (id: number, data: any) =>
    apiClient.put(`/projects/${id}`, data),
    
  delete: (id: number) =>
    apiClient.delete(`/projects/${id}`),
    
  getScenes: (id: number) =>
    apiClient.get(`/projects/${id}/scenes`),
    
  generateScenes: (id: number) =>
    apiClient.post(`/projects/${id}/generate`),
};

// Video API (Milestone 1.1)
export const videoAPI = {
  generate: (projectId: number, data: {
    scene_id: number;
    prompt: string;
    provider?: import('@/types').VideoProviderKind;
    model?: string;
    image_url?: string;
    resolution?: string;
    workflow_path?: string;
    workflow?: Record<string, any>;
    extra_data?: Record<string, any>;
    duration?: number;
    aspect_ratio?: '16:9' | '9:16';
  }) =>
    apiClient.post(`/projects/${projectId}/generate-video`, data),

  generateNarration: (projectId: number, data: {
    movie_title: string;
    synopsis?: string;
    style?: string;
    target_duration?: number;
    voice?: string;
    speed?: number;
    provider?: import('@/types').VideoProviderKind;
    aspect_ratio?: '16:9' | '9:16';
    source_video_path?: string;
    source_video_url?: string;
    creative_brief?: string;
  }) =>
    apiClient.post(`/projects/${projectId}/generate-narration`, data),
    
  getStatus: (projectId: number, data: {
    video_id: string;
    provider: import('@/types').VideoProviderKind;
  }) =>
    apiClient.post(`/projects/${projectId}/video-status`, data),
    
  list: (projectId: number, params?: {
    status?: string;
    limit?: number;
    offset?: number;
  }) =>
    apiClient.get(`/projects/${projectId}/videos`, { params }),
    
  cancel: (projectId: number, videoId: string) =>
    apiClient.delete(`/projects/${projectId}/video/${videoId}`),
};

// Platform API（里程碑一：第三方平台账号绑定）
export const platformAPI = {
  list: () =>
    apiClient.get<{ data: import('@/types').PlatformAccount[] }>('/platforms'),
  configured: () =>
    apiClient.get<{ data: import('@/types').PlatformKind[] }>('/platforms/configured'),
  getAuthorizeUrl: (platform: import('@/types').PlatformKind) =>
    apiClient.get<{ authorize_url: string }>(`/platforms/${platform}/authorize`),
  disconnect: (platform: import('@/types').PlatformKind) =>
    apiClient.delete(`/platforms/${platform}`),
};

// Analytics API（里程碑一：运营看板）
export const analyticsAPI = {
  summary: () =>
    apiClient.get<{ data: import('@/types').AnalyticsSummary }>('/analytics/summary'),
  videos: (params?: { limit?: number }) =>
    apiClient.get<import('@/types').ListResponse<import('@/types').VideoTask>>('/analytics/videos', { params }),
};

// Model Center API
export const modelCenterAPI = {
  overview: () =>
    apiClient.get<{ data: import('@/types').ModelCenterOverview }>('/model-center/overview'),
  triggerProbe: () =>
    apiClient.post<{ data: import('@/types').ModelCenterOverview }>('/model-center/probe'),
};

// Asset Center API
export const assetAPI = {
  listRoleAssets: (params?: { project_id?: number; q?: string; tags?: string }) =>
    apiClient.get<{ total: number; data: import('@/types').RoleAsset[] }>('/assets/roles', { params }),
  createRoleAsset: (data: {
    project_id?: number;
    name: string;
    role_type?: string;
    description?: string;
    avatar_url?: string;
    voice_preset?: string;
    style_prompt?: string;
    negative_hint?: string;
    tags?: string;
  }) => apiClient.post('/assets/roles', data),
  updateRoleAsset: (id: number, data: {
    project_id?: number;
    name?: string;
    role_type?: string;
    description?: string;
    avatar_url?: string;
    voice_preset?: string;
    style_prompt?: string;
    negative_hint?: string;
    tags?: string;
  }) => apiClient.put(`/assets/roles/${id}`, data),
  deleteRoleAsset: (id: number) => apiClient.delete(`/assets/roles/${id}`),

  listPromptTemplates: (params?: { project_id?: number; template_type?: string; q?: string; tags?: string }) =>
    apiClient.get<{ total: number; data: import('@/types').PromptTemplate[] }>('/assets/prompt-templates', { params }),
  createPromptTemplate: (data: {
    project_id?: number;
    name: string;
    template_type: string;
    provider_type?: string;
    content: string;
    variables?: string;
    tags?: string;
  }) => apiClient.post('/assets/prompt-templates', data),
  updatePromptTemplate: (id: number, data: {
    project_id?: number;
    name?: string;
    template_type?: string;
    provider_type?: string;
    content?: string;
    variables?: string;
    tags?: string;
  }) => apiClient.put(`/assets/prompt-templates/${id}`, data),
  deletePromptTemplate: (id: number) => apiClient.delete(`/assets/prompt-templates/${id}`),
};

// Storyboard API
export const storyboardAPI = {
  listProjectShots: (projectId: number) =>
    apiClient.get<{ total: number; data: import('@/types').StoryboardShot[] }>(`/projects/${projectId}/storyboard-shots`),
  importShots: (projectId: number, data: {
    shots: Array<{
      scene_id?: number;
      chapter?: string;
      shot_number?: number;
      sort_order?: number;
      title: string;
      description?: string;
      camera_language?: string;
      duration?: number;
      aspect_ratio?: '16:9' | '9:16';
      prompt?: string;
      negative_prompt?: string;
      reference_image_url?: string;
      status?: string;
    }>;
  }) =>
    apiClient.post(`/projects/${projectId}/storyboard-shots/import`, data),
  createShot: (projectId: number, data: {
    scene_id?: number;
    chapter?: string;
    shot_number?: number;
    sort_order?: number;
    title: string;
    description?: string;
    camera_language?: string;
    duration?: number;
    aspect_ratio?: '16:9' | '9:16';
    prompt?: string;
    negative_prompt?: string;
    reference_image_url?: string;
    status?: 'draft' | 'pending' | 'processing' | 'completed' | 'failed';
  }) => apiClient.post(`/projects/${projectId}/storyboard-shots`, data),
  reorderShots: (projectId: number, orderedIDs: number[]) =>
    apiClient.post<{ reordered: number }>(`/projects/${projectId}/storyboard-shots/reorder`, { ordered_ids: orderedIDs }),
  createShotVersion: (projectId: number, data: { source_shot_id: number; version_note?: string }) =>
    apiClient.post(`/projects/${projectId}/storyboard-shots/version`, data),
  getVersionTree: (projectId: number) =>
    apiClient.get<{ total: number; data: import('@/types').StoryboardVersionNode[] }>(`/projects/${projectId}/storyboard-shots/version-tree`),
  bootstrapFromScenes: (projectId: number) =>
    apiClient.post<{ bootstrapped: number }>(`/projects/${projectId}/storyboard-shots/bootstrap`),
};
