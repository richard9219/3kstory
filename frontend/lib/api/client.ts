import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Create axios instance
export const apiClient = axios.create({
  baseURL: `${API_URL}/api/v1`,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor - add auth token
apiClient.interceptors.request.use(
  (config) => {
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
    if (error.response?.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('token');
      window.location.href = '/';
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
    provider: 'runway' | 'pika' | 'local';
    image_url?: string;
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
    provider?: 'runway' | 'pika' | 'local';
    aspect_ratio?: '16:9' | '9:16';
    source_video_path?: string;
    source_video_url?: string;
  }) =>
    apiClient.post(`/projects/${projectId}/generate-narration`, data),
    
  getStatus: (projectId: number, data: {
    video_id: string;
    provider: 'runway' | 'pika' | 'local';
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
