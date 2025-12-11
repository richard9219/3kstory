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
      window.location.href = '/login';
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
    provider: 'runway' | 'pika';
    image_url?: string;
    duration?: number;
    aspect_ratio?: '16:9' | '9:16';
  }) =>
    apiClient.post(`/projects/${projectId}/generate-video`, data),
    
  getStatus: (projectId: number, data: {
    video_id: string;
    provider: 'runway' | 'pika';
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
