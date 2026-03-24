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

  const demoTimelineShots = [
    {
    id: 7001,
    user_id: 1,
    project_id: 101,
    scene_id: 1,
    track_index: 1,
    chapter: '第一章 隆中',
    shot_number: 1,
    sort_order: 1,
    title: '草庐夜读',
    description: '诸葛亮在烛光下翻阅天下地图，镜头缓慢推进。',
    camera_language: '慢推近景',
    emotion_tone: '沉静',
    timeline_start_ms: 0,
    timeline_duration_ms: 6000,
    transition_type: 'fade',
    transition_duration_ms: 400,
    duration: 6,
    aspect_ratio: '16:9',
    prompt: 'ancient chinese strategist reading map by candlelight, cinematic close-up, warm amber lighting',
    clip_provider: 'seedance',
    clip_video_id: 'demo_shot_clip_7001',
    clip_video_url: 'https://www.w3schools.com/html/mov_bbb.mp4',
    clip_status: 'completed',
    clip_score: 0.88,
    clip_notes: '角色稳定，适合作为开场建立镜头',
    locked: true,
    status: 'completed',
    version: 1,
    created_at: nowMinusMinutes(120),
    updated_at: nowMinusMinutes(40),
    },
    {
    id: 7002,
    user_id: 1,
    project_id: 101,
    scene_id: 2,
    track_index: 1,
    chapter: '第一章 隆中',
    shot_number: 2,
    sort_order: 2,
    title: '扇柄开机',
    description: '羽扇划过画面，引出谋略推进。',
    camera_language: '横移特写',
    emotion_tone: '机锋',
    timeline_start_ms: 5600,
    timeline_duration_ms: 5000,
    transition_type: 'cut',
    transition_duration_ms: 0,
    duration: 5,
    aspect_ratio: '16:9',
    prompt: 'feather fan swipes across frame, cinematic macro shot, elegant motion blur',
    clip_status: 'processing',
    clip_score: 0.0,
    clip_notes: '正在生成局部重生版本',
    locked: false,
    status: 'processing',
    version: 2,
    created_at: nowMinusMinutes(115),
    updated_at: nowMinusMinutes(3),
    },
    {
    id: 7003,
    user_id: 1,
    project_id: 101,
    scene_id: 3,
    track_index: 2,
    chapter: '第二章 赤壁',
    shot_number: 3,
    sort_order: 3,
    title: '江面火光',
    description: '夜色中火光映在江面，营造大战前夕氛围。',
    camera_language: '无人机俯视',
    emotion_tone: '压迫',
    timeline_start_ms: 10600,
    timeline_duration_ms: 7000,
    transition_type: 'fade',
    transition_duration_ms: 500,
    duration: 7,
    aspect_ratio: '16:9',
    prompt: 'river surface reflecting flames at night, epic battle atmosphere, aerial cinematic shot',
    clip_status: 'draft',
    clip_score: 0.0,
    clip_notes: '待生成',
    locked: false,
    status: 'draft',
    version: 1,
    created_at: nowMinusMinutes(90),
    updated_at: nowMinusMinutes(20),
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
      method: 'post',
      route: '/projects/*/generate-narration-advanced',
      handler: () => {
        const jobId = `demo_job_${Date.now()}`;
        return makeFallbackResponse(config, {
          job_id: jobId,
          status: 'completed',
          selected_video_id: 'demo_scene_9002',
          detail: {
            job: {
              id: 1,
              job_id: jobId,
              user_id: 1,
              project_id: 101,
              pipeline_type: 'narration_advanced',
              status: 'completed',
              queue_status: 'done',
              candidate_count: 3,
              quality_mode: 'fast',
              score_profile: 'default',
              provider_mode: 'multi',
              auto_pick: true,
              publish_threshold: 0.72,
              publish_gate_passed: true,
              publish_block_reason: '',
              selected_video_id: 'demo_scene_9002',
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            },
            pipeline_status: {
              script: 'completed',
              tts: 'completed',
              render: 'completed',
              score: 'completed',
              selection: 'completed',
            },
            selected_video_id: 'demo_scene_9002',
            top_score: 0.91,
            candidates: [
              {
                id: 9002,
                user_id: 1,
                project_id: 101,
                job_id: jobId,
                candidate_no: 1,
                task_type: 'narration_video',
                title: '药命效应 EP01',
                video_id: 'demo_scene_9002',
                provider: 'runway',
                status: 'completed',
                score: 0.91,
                rank: 1,
                is_selected: true,
                video_url: 'https://example.com/demo.mp4',
                created_at: nowMinusMinutes(30),
                updated_at: nowMinusMinutes(20),
              },
              {
                id: 9004,
                user_id: 1,
                project_id: 101,
                job_id: jobId,
                candidate_no: 2,
                task_type: 'narration_video',
                title: '药命效应 EP01',
                video_id: 'demo_scene_9004',
                provider: 'minimax',
                status: 'completed',
                score: 0.84,
                rank: 2,
                is_selected: false,
                video_url: 'https://example.com/demo2.mp4',
                created_at: nowMinusMinutes(30),
                updated_at: nowMinusMinutes(19),
              },
            ],
          },
        });
      },
    },
    {
      method: 'get',
      route: '/projects/*/video-jobs/*',
      handler: () =>
        makeFallbackResponse(config, {
          data: {
            job: {
              id: 1,
              job_id: 'demo_job_fixed',
              user_id: 1,
              project_id: 101,
              pipeline_type: 'narration_advanced',
              status: 'completed',
              queue_status: 'done',
              candidate_count: 2,
              quality_mode: 'fast',
              score_profile: 'default',
              provider_mode: 'multi',
              auto_pick: true,
              publish_threshold: 0.72,
              publish_gate_passed: true,
              publish_block_reason: '',
              selected_video_id: 'demo_scene_9002',
              created_at: nowMinusMinutes(40),
              updated_at: nowMinusMinutes(18),
            },
            pipeline_status: {
              script: 'completed',
              tts: 'completed',
              render: 'completed',
              score: 'completed',
              selection: 'completed',
            },
            selected_video_id: 'demo_scene_9002',
            top_score: 0.91,
            candidates: [
              {
                id: 9002,
                user_id: 1,
                project_id: 101,
                job_id: 'demo_job_fixed',
                candidate_no: 1,
                task_type: 'narration_video',
                title: '药命效应 EP01',
                video_id: 'demo_scene_9002',
                provider: 'runway',
                status: 'completed',
                score: 0.91,
                rank: 1,
                is_selected: true,
                video_url: 'https://example.com/demo.mp4',
                created_at: nowMinusMinutes(30),
                updated_at: nowMinusMinutes(20),
              },
              {
                id: 9004,
                user_id: 1,
                project_id: 101,
                job_id: 'demo_job_fixed',
                candidate_no: 2,
                task_type: 'narration_video',
                title: '药命效应 EP01',
                video_id: 'demo_scene_9004',
                provider: 'minimax',
                status: 'completed',
                score: 0.84,
                rank: 2,
                is_selected: false,
                video_url: 'https://example.com/demo2.mp4',
                created_at: nowMinusMinutes(30),
                updated_at: nowMinusMinutes(19),
              },
            ],
          },
        }),
    },
    {
      method: 'post',
      route: '/projects/*/video-jobs/*/select',
      handler: () => makeFallbackResponse(config, { status: 'selected', ok: true }),
    },
    {
      method: 'post',
      route: '/projects/*/video-jobs/*/auto-publish',
      handler: () => makeFallbackResponse(config, {
        status: 'passed',
        message: 'quality gate passed; ready for publish workflow',
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
      handler: () => makeFallbackResponse(config, { total: demoTimelineShots.length, data: demoTimelineShots }),
    },
    {
      method: 'get',
      route: '/projects/*/storyboard-shots/version-tree',
      handler: () => makeFallbackResponse(config, {
      total: 2,
      data: [
        { root: demoTimelineShots[0], versions: [] },
        {
        root: demoTimelineShots[1],
        versions: [
          {
          ...demoTimelineShots[1],
          id: 7012,
          version: 1,
          clip_status: 'completed',
          status: 'completed',
          clip_notes: '旧版镜头，羽扇动作略僵',
          },
        ],
        },
      ],
    }),
    },
    {
    method: 'get',
    route: '/projects/*/storyboard-timeline',
    handler: () => makeFallbackResponse(config, {
      project_id: 101,
      total_duration_ms: 17600,
      ready_shot_count: 1,
      total_shot_count: demoTimelineShots.length,
      latest_export: {
      id: 9991,
      user_id: 1,
      project_id: 101,
      task_type: 'director_cut',
      title: 'Director Cut v1',
      video_id: 'director_demo_1',
      provider: 'local',
      status: 'completed',
      video_url: 'https://www.w3schools.com/html/mov_bbb.mp4',
      created_at: nowMinusMinutes(15),
      updated_at: nowMinusMinutes(15),
      },
      shots: demoTimelineShots,
    }),
    },
    {
    method: 'patch',
    route: '/projects/*/storyboard-shots/*',
    handler: () => makeFallbackResponse(config, {
      ...demoTimelineShots[0],
      ...(config?.data ? JSON.parse(config.data) : {}),
      updated_at: new Date().toISOString(),
    }),
    },
    {
    method: 'post',
    route: '/projects/*/storyboard-shots/*/generate',
    handler: () => makeFallbackResponse(config, {
      ...demoTimelineShots[0],
      clip_status: 'completed',
      status: 'completed',
      clip_video_id: `demo_generated_${Date.now()}`,
      clip_video_url: 'https://www.w3schools.com/html/mov_bbb.mp4',
      clip_score: 0.86,
      updated_at: new Date().toISOString(),
    }),
    },
    {
    method: 'post',
    route: '/projects/*/storyboard-shots/*/regenerate',
    handler: () => makeFallbackResponse(config, {
      ...demoTimelineShots[1],
      id: 7020,
      version: 3,
      clip_status: 'completed',
      status: 'completed',
      clip_video_id: `demo_regenerated_${Date.now()}`,
      clip_video_url: 'https://www.w3schools.com/html/movie.mp4',
      clip_score: 0.9,
      updated_at: new Date().toISOString(),
    }),
    },
    {
    method: 'post',
    route: '/projects/*/storyboard-timeline/render',
    handler: () => makeFallbackResponse(config, {
      id: 9992,
      user_id: 1,
      project_id: 101,
      task_type: 'director_cut',
      title: 'Director Cut v2',
      video_id: `director_demo_${Date.now()}`,
      provider: 'local',
      status: 'completed',
      video_url: 'https://www.w3schools.com/html/movie.mp4',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }),
    },
    {
    method: 'post',
    route: '/projects/*/storyboard-timeline/exports/*/auto-publish',
    handler: () => makeFallbackResponse(config, {
      status: 'passed',
      reason: '',
      score: 0.86,
      threshold: 0.72,
    }),
    },
    {
    method: 'post',
    route: '/projects/*/storyboard-timeline/exports/*/publish',
    handler: () => makeFallbackResponse(config, {
      status: 'published',
      platform: (config?.data ? JSON.parse(config.data)?.platform : '') || 'douyin',
      video_id: 'director_demo_published',
      task_id: 9992,
    }),
    },
    {
    method: 'get',
    route: '/projects/*/storyboard-timeline/publish-history',
    handler: () => makeFallbackResponse(config, {
      total: 3,
      data: [
        {
          id: 1,
          user_id: 1,
          project_id: 101,
          video_task_id: 9992,
          export_id: 'director_demo_1',
          platform: 'douyin',
          status: 'success',
          attempt_no: 1,
          receipt_id: 'dy_receipt_demo_1',
          remote_video_id: 'dy_video_demo_1',
          remote_url: 'https://www.douyin.com/video/demo',
          response_payload: { request_id: 'dy-req-1', http_status: 200 },
          completed_at: nowMinusMinutes(55),
          created_at: nowMinusMinutes(56),
          updated_at: nowMinusMinutes(55),
        },
        {
          id: 2,
          user_id: 1,
          project_id: 101,
          video_task_id: 9992,
          export_id: 'director_demo_1',
          platform: 'bilibili',
          status: 'failed',
          attempt_no: 1,
          receipt_id: 'bili_receipt_demo_1',
          error_msg: 'token expired',
          response_payload: { request_id: 'bili-req-1', http_status: 401 },
          completed_at: nowMinusMinutes(48),
          created_at: nowMinusMinutes(48),
          updated_at: nowMinusMinutes(48),
        },
        {
          id: 3,
          user_id: 1,
          project_id: 101,
          video_task_id: 9992,
          export_id: 'director_demo_1',
          platform: 'bilibili',
          status: 'success',
          attempt_no: 2,
          retried_from_id: 2,
          receipt_id: 'bili_receipt_demo_2',
          remote_video_id: 'BV1xxxDemo',
          response_payload: { request_id: 'bili-req-2', http_status: 200 },
          completed_at: nowMinusMinutes(42),
          created_at: nowMinusMinutes(42),
          updated_at: nowMinusMinutes(42),
        },
      ],
    }),
    },
    {
    method: 'post',
    route: '/projects/*/storyboard-timeline/publish-history/*/retry',
    handler: () => makeFallbackResponse(config, {
      status: 'retried',
      video_id: 'director_demo_published_retry',
      task_id: 9992,
      publish_record: {
        id: 4,
        user_id: 1,
        project_id: 101,
        video_task_id: 9992,
        export_id: 'director_demo_1',
        platform: 'bilibili',
        status: 'success',
        attempt_no: 3,
        retried_from_id: 2,
        receipt_id: 'bili_receipt_demo_3',
        remote_video_id: 'BV1retryDemo',
        response_payload: { request_id: 'bili-req-3', http_status: 200 },
        completed_at: new Date().toISOString(),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
    }),
    },
    {
    method: 'get',
    route: '/projects/*/director-templates',
    handler: () => makeFallbackResponse(config, {
      total: 4,
      data: [
        {
          id: 101,
          user_id: 1,
          project_id: 101,
          name: '黑泽明',
          slug: 'kurosawa',
          sample_frame_url: 'https://images.unsplash.com/photo-1522156373667-4c7234bbd804?auto=format&fit=crop&w=1200&q=80',
          sample_video_url: 'https://www.w3schools.com/html/mov_bbb.mp4',
          prompt_prefix: 'high-contrast monochrome mood',
          camera_language: '长焦压缩 + 风雨动势',
          emotion_tone: '肃杀',
          transition_type: 'wipe',
          transition_duration_ms: 260,
          genre_keywords: '历史,战争,宿命',
          weight_narrative: 0.18,
          weight_visual: 0.28,
          weight_emotion: 0.2,
          weight_rhythm: 0.22,
          weight_continuity: 0.12,
          is_builtin: true,
          created_at: nowMinusMinutes(1000),
          updated_at: nowMinusMinutes(1000),
        },
        {
          id: 102,
          user_id: 1,
          project_id: 101,
          name: '张艺谋',
          slug: 'zhangyimou',
          sample_frame_url: 'https://images.unsplash.com/photo-1478720568477-152d9b164e26?auto=format&fit=crop&w=1200&q=80',
          sample_video_url: 'https://www.w3schools.com/html/movie.mp4',
          prompt_prefix: 'bold color choreography',
          camera_language: '色块构图 + 仪式化调度',
          emotion_tone: '浓烈',
          transition_type: 'fade',
          transition_duration_ms: 320,
          genre_keywords: '史诗,情感,古装',
          weight_narrative: 0.16,
          weight_visual: 0.34,
          weight_emotion: 0.24,
          weight_rhythm: 0.16,
          weight_continuity: 0.1,
          is_builtin: true,
          created_at: nowMinusMinutes(1000),
          updated_at: nowMinusMinutes(1000),
        },
      ],
    }),
    },
    {
    method: 'post',
    route: '/projects/*/director-templates',
    handler: () => makeFallbackResponse(config, {
      id: 333,
      user_id: 1,
      project_id: 101,
      ...(config?.data ? JSON.parse(config.data) : {}),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }, 201),
    },
    {
    method: 'put',
    route: '/projects/*/director-templates/*',
    handler: () => makeFallbackResponse(config, {
      id: 333,
      user_id: 1,
      project_id: 101,
      ...(config?.data ? JSON.parse(config.data) : {}),
      updated_at: new Date().toISOString(),
    }),
    },
    {
    method: 'delete',
    route: '/projects/*/director-templates/*',
    handler: () => makeFallbackResponse(config, { deleted: true }),
    },
    {
    method: 'post',
    route: '/projects/*/director-agent/auto-strategy',
    handler: () => makeFallbackResponse(config, {
      genre: (config?.data ? JSON.parse(config.data)?.genre : '') || '历史',
      applied: Boolean(config?.data ? JSON.parse(config.data)?.apply : false),
      tune_percent: (config?.data ? JSON.parse(config.data)?.tune_percent : 70) || 70,
      selected: {
        id: 101,
        user_id: 1,
        project_id: 101,
        name: '黑泽明',
        slug: 'kurosawa',
        weight_narrative: 0.18,
        weight_visual: 0.28,
        weight_emotion: 0.2,
        weight_rhythm: 0.22,
        weight_continuity: 0.12,
        created_at: nowMinusMinutes(1000),
        updated_at: nowMinusMinutes(1000),
      },
      predicted_score: 0.842,
      shot_updates: 6,
    }),
    },
    {
    method: 'post',
    route: '/projects/*/director-agent/ab-compare',
    handler: () => makeFallbackResponse(config, {
      genre: (config?.data ? JSON.parse(config.data)?.genre : '') || '历史',
      template_a: {
        id: 101,
        user_id: 1,
        project_id: 101,
        name: '黑泽明',
        slug: 'kurosawa',
        weight_narrative: 0.18,
        weight_visual: 0.28,
        weight_emotion: 0.2,
        weight_rhythm: 0.22,
        weight_continuity: 0.12,
        created_at: nowMinusMinutes(1000),
        updated_at: nowMinusMinutes(1000),
      },
      template_b: {
        id: 102,
        user_id: 1,
        project_id: 101,
        name: '张艺谋',
        slug: 'zhangyimou',
        weight_narrative: 0.16,
        weight_visual: 0.34,
        weight_emotion: 0.24,
        weight_rhythm: 0.16,
        weight_continuity: 0.1,
        created_at: nowMinusMinutes(1000),
        updated_at: nowMinusMinutes(1000),
      },
      score_a: 0.842,
      score_b: 0.811,
      winner_template_id: 101,
      winner_template: '黑泽明',
      applied: Boolean(config?.data ? JSON.parse(config.data)?.apply_best : false),
      rendered_export_id: Boolean(config?.data ? JSON.parse(config.data)?.render_best_cut : false) ? `director_demo_ab_${Date.now()}` : '',
      predicted_gain: 0.031,
      compared_shot_count: 8,
    }),
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

  generateNarrationAdvanced: (projectId: number, data: {
    movie_title: string;
    synopsis?: string;
    style?: string;
    target_duration?: number;
    voice?: string;
    speed?: number;
    provider?: import('@/types').VideoProviderKind;
    provider_mode?: 'single' | 'multi';
    providers?: import('@/types').VideoProviderKind[];
    candidate_count?: number;
    quality_mode?: 'fast' | 'quality';
    score_profile?: 'default' | 'short_drama' | 'movie_narration';
    auto_pick?: boolean;
    aspect_ratio?: '16:9' | '9:16';
    source_video_path?: string;
    source_video_url?: string;
    creative_brief?: string;
  }) =>
    apiClient.post(`/projects/${projectId}/generate-narration-advanced`, data),

  getJob: (projectId: number, jobId: string) =>
    apiClient.get<{ data: import('@/types').VideoJobDetail }>(`/projects/${projectId}/video-jobs/${jobId}`),

  selectCandidate: (projectId: number, jobId: string, data: { video_id: string; reason?: string }) =>
    apiClient.post(`/projects/${projectId}/video-jobs/${jobId}/select`, data),

  autoPublishWithGate: (projectId: number, jobId: string, data?: { platform?: string }) =>
    apiClient.post(`/projects/${projectId}/video-jobs/${jobId}/auto-publish`, data || {}),
    
  getStatus: (projectId: number, data: {
    video_id: string;
    provider: import('@/types').VideoProviderKind;
  }) =>
    apiClient.post(`/projects/${projectId}/video-status`, data),
    
  list: (projectId: number, params?: {
    status?: string;
    limit?: number;
    offset?: number;
    include_scores?: boolean;
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
  getTimeline: (projectId: number) =>
    apiClient.get<import('@/types').StoryboardTimeline>(`/projects/${projectId}/storyboard-timeline`),
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
  updateShot: (projectId: number, shotId: number, data: import('@/types').StoryboardShotUpdatePayload) =>
    apiClient.patch<import('@/types').StoryboardShot>(`/projects/${projectId}/storyboard-shots/${shotId}`, data),
  generateShot: (projectId: number, shotId: number, data: import('@/types').StoryboardShotGeneratePayload) =>
    apiClient.post<import('@/types').StoryboardShot>(`/projects/${projectId}/storyboard-shots/${shotId}/generate`, data),
  regenerateShot: (projectId: number, shotId: number, data: import('@/types').StoryboardShotGeneratePayload) =>
    apiClient.post<import('@/types').StoryboardShot>(`/projects/${projectId}/storyboard-shots/${shotId}/regenerate`, data),
  renderTimeline: (projectId: number) =>
    apiClient.post<import('@/types').VideoTask>(`/projects/${projectId}/storyboard-timeline/render`),
  autoPublishDirectorCut: (projectId: number, exportId: string) =>
    apiClient.post<{ status: 'passed' | 'blocked'; reason: string; score: number; threshold: number }>(`/projects/${projectId}/storyboard-timeline/exports/${exportId}/auto-publish`, {}),
  publishDirectorCut: (projectId: number, exportId: string, data: { platform: import('@/types').PlatformKind }) =>
    apiClient.post<{ status: 'published' | 'blocked'; reason?: string; platform?: import('@/types').PlatformKind; video_id?: string; task_id?: number }>(`/projects/${projectId}/storyboard-timeline/exports/${exportId}/publish`, data),
  listDirectorPublishHistory: (projectId: number, params?: { export_id?: string }) =>
    apiClient.get<{ total: number; data: import('@/types').DirectorPublishRecord[] }>(`/projects/${projectId}/storyboard-timeline/publish-history`, { params }),
  retryDirectorPublish: (projectId: number, recordId: number, data?: { reason?: string }) =>
    apiClient.post<{ status: 'retried'; video_id?: string; task_id?: number; publish_record?: import('@/types').DirectorPublishRecord }>(`/projects/${projectId}/storyboard-timeline/publish-history/${recordId}/retry`, data || {}),
  fetchDirectorCutBlob: (projectId: number, exportId: string) =>
    apiClient.get<Blob>(`/projects/${projectId}/storyboard-timeline/exports/${exportId}`, { responseType: 'blob' }),
  listDirectorTemplates: (projectId: number) =>
    apiClient.get<{ total: number; data: import('@/types').DirectorTemplate[] }>(`/projects/${projectId}/director-templates`),
  createDirectorTemplate: (projectId: number, data: {
    name: string;
    slug?: string;
    sample_frame_url?: string;
    sample_video_url?: string;
    prompt_prefix?: string;
    camera_language?: string;
    emotion_tone?: string;
    transition_type?: 'cut' | 'fade' | 'wipe' | 'match';
    transition_duration_ms?: number;
    genre_keywords?: string;
    weight_narrative?: number;
    weight_visual?: number;
    weight_emotion?: number;
    weight_rhythm?: number;
    weight_continuity?: number;
  }) => apiClient.post<import('@/types').DirectorTemplate>(`/projects/${projectId}/director-templates`, data),
  updateDirectorTemplate: (projectId: number, templateId: number, data: {
    name?: string;
    slug?: string;
    sample_frame_url?: string;
    sample_video_url?: string;
    prompt_prefix?: string;
    camera_language?: string;
    emotion_tone?: string;
    transition_type?: 'cut' | 'fade' | 'wipe' | 'match';
    transition_duration_ms?: number;
    genre_keywords?: string;
    weight_narrative?: number;
    weight_visual?: number;
    weight_emotion?: number;
    weight_rhythm?: number;
    weight_continuity?: number;
  }) => apiClient.put<import('@/types').DirectorTemplate>(`/projects/${projectId}/director-templates/${templateId}`, data),
  deleteDirectorTemplate: (projectId: number, templateId: number) =>
    apiClient.delete<{ deleted: boolean }>(`/projects/${projectId}/director-templates/${templateId}`),
  autoDirectorStrategy: (projectId: number, data: { genre?: string; template_id?: number; apply?: boolean; tune_percent?: number }) =>
    apiClient.post<import('@/types').DirectorAutoStrategyResult>(`/projects/${projectId}/director-agent/auto-strategy`, data),
  compareDirectorAB: (projectId: number, data: { template_a_id: number; template_b_id: number; genre?: string; apply_best?: boolean; tune_percent?: number; render_best_cut?: boolean }) =>
    apiClient.post<import('@/types').DirectorABCompareResult>(`/projects/${projectId}/director-agent/ab-compare`, data),
};
