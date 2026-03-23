'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { analyticsAPI, projectAPI, videoAPI } from '@/lib/api/client';
import { FactoryHubCards } from '@/components/common/FactoryConsoleShell';
import { OperationalInsightPanel } from '@/components/common/OperationalInsightCards';
import { EmptyBlock, LoadingBlock } from '@/components/common/StateBlocks';
import MiniButton, { MiniLinkButton } from '@/components/common/MiniButton';
import { useAuthStore } from '@/lib/store/authStore';
import type { AnalyticsSummary, PlatformKind, Project, VideoProviderKind, VideoTask } from '@/types';
import { ArrowLeft, BarChart3, CheckCircle2, Clapperboard, Clock3, ExternalLink, Film, Link2, Megaphone, Sparkles, Video, XCircle } from 'lucide-react';

const PLATFORM_LABEL: Record<PlatformKind, string> = {
  douyin: '抖音',
  xiaohongshu: '小红书',
  bilibili: 'B站',
  weibo: '微博',
};

const PROVIDER_LABEL: Record<VideoTask['provider'], string> = {
  runway: 'Runway',
  pika: 'Pika',
  local: '本地调试',
  minimax: 'MiniMax',
  seedance: 'Seedance',
  comfy: 'Comfy',
};

const STAGE_CARDS = [
  {
    icon: Video,
    title: '影视解说',
    keywords: ['脚本', '旁白', '素材', '分发'],
  },
  {
    icon: Clapperboard,
    title: 'AI短剧',
    keywords: ['角色', '分镜', '版本', '连载'],
  },
  {
    icon: Film,
    title: 'AI电影',
    keywords: ['章节', '镜头', '一致性', '审片'],
  },
  {
    icon: Megaphone,
    title: '广告内容',
    keywords: ['品牌', '商品', '模板', '投放'],
  },
];

const CONTROL_MODULES = [
  {
    icon: Sparkles,
    title: '任务中心',
    keywords: ['排产', '追踪', '回调', '看板'],
  },
  {
    icon: Clapperboard,
    title: '模型中心',
    keywords: ['路由', '配额', '探活', '告警'],
  },
  {
    icon: Film,
    title: '资产中心',
    keywords: ['角色', '场景', '镜头', '模板'],
  },
  {
    icon: Link2,
    title: '渠道中心',
    keywords: ['绑定', '发布', '回流', '复盘'],
  },
];

function statusBadge(status: VideoTask['status']) {
  switch (status) {
    case 'completed':
      return 'text-black bg-black/[0.08] border-black/20';
    case 'failed':
      return 'text-black bg-black/[0.12] border-black/25';
    case 'processing':
      return 'text-black bg-black/[0.06] border-black/20';
    default:
      return 'text-black/65 bg-black/[0.04] border-black/15';
  }
}

function toDayKey(value: string) {
  return new Date(value).toISOString().slice(0, 10);
}

function getRecentDays(count: number) {
  const days: string[] = [];
  const now = new Date();
  for (let i = count - 1; i >= 0; i -= 1) {
    const d = new Date(now);
    d.setDate(now.getDate() - i);
    days.push(d.toISOString().slice(0, 10));
  }
  return days;
}

export default function DashboardPage() {
  const { isAuthenticated } = useAuthStore();
  const [offlineMode, setOfflineMode] = useState(false);
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null);
  const [videos, setVideos] = useState<VideoTask[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitLoading, setSubmitLoading] = useState(false);
  const [submitMsg, setSubmitMsg] = useState<string | null>(null);
  const [range, setRange] = useState<'7d' | '30d'>('7d');
  const [providerFilter, setProviderFilter] = useState<'all' | VideoProviderKind>('all');
  const [pipelineFilter, setPipelineFilter] = useState<'all' | 'narration' | 'scene'>('all');
  const [form, setForm] = useState({
    projectId: 0,
    movieTitle: '',
    synopsis: '',
    style: '深度分析',
    targetDuration: 90,
    voice: 'female_cn',
    sourceVideoPath: '',
    sourceVideoUrl: '',
    creativeBrief: '',
    provider: 'runway' as 'runway' | 'pika' | 'local',
    aspectRatio: '16:9' as '16:9' | '9:16',
  });

  const loadData = () => {
    return Promise.all([
      analyticsAPI.summary(),
      analyticsAPI.videos({ limit: 20 }),
      projectAPI.list(),
    ]).then(([s, v, p]) => {
      setSummary(s.data.data);
      setVideos(v.data.data || []);

      const list = (p.data || []) as Project[];
      setProjects(list);
      setForm((prev) => ({
        ...prev,
        projectId: prev.projectId || (list[0]?.id ?? 0),
      }));
    });
  };

  useEffect(() => {
    if (!isAuthenticated) {
      setLoading(false);
      return;
    }

    loadData()
      .finally(() => setLoading(false));
  }, [isAuthenticated]);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    setOfflineMode(sessionStorage.getItem('frontend-offline-fallback') === '1');
  }, []);

  const boundSet = useMemo(() => {
    return new Set((summary?.bound_platforms || []).map((p) => p.platform));
  }, [summary?.bound_platforms]);

  const filteredVideos = useMemo(() => {
    return videos.filter((video) => {
      if (providerFilter !== 'all' && video.provider !== providerFilter) {
        return false;
      }
      if (pipelineFilter === 'narration' && video.task_type !== 'narration_video') {
        return false;
      }
      if (pipelineFilter === 'scene' && video.task_type === 'narration_video') {
        return false;
      }
      return true;
    });
  }, [videos, providerFilter, pipelineFilter]);

  const insight = useMemo(() => {
    const recentDays = getRecentDays(range === '7d' ? 7 : 30);
    const taskByDay = new Map<string, number>();
    const completedByDay = new Map<string, number>();
    const failedByDay = new Map<string, number>();
    const durationByDay = new Map<string, { total: number; count: number }>();
    const costByDay = new Map<string, number>();

    for (const day of recentDays) {
      taskByDay.set(day, 0);
      completedByDay.set(day, 0);
      failedByDay.set(day, 0);
      durationByDay.set(day, { total: 0, count: 0 });
      costByDay.set(day, 0);
    }

    const providerCost: Record<VideoTask['provider'], number> = {
      runway: 2.8,
      pika: 1.9,
      local: 0.3,
      minimax: 1.2,
      seedance: 2.2,
      comfy: 0.5,
    };

    filteredVideos.forEach((video) => {
      const dayKey = toDayKey(video.created_at);
      if (!taskByDay.has(dayKey)) return;

      taskByDay.set(dayKey, (taskByDay.get(dayKey) || 0) + 1);
      costByDay.set(dayKey, (costByDay.get(dayKey) || 0) + (providerCost[video.provider] || 0));

      if (video.status === 'completed') {
        completedByDay.set(dayKey, (completedByDay.get(dayKey) || 0) + 1);
      }
      if (video.status === 'failed') {
        failedByDay.set(dayKey, (failedByDay.get(dayKey) || 0) + 1);
      }

      if (video.completed_at) {
        const start = new Date(video.created_at).getTime();
        const end = new Date(video.completed_at).getTime();
        const minutes = Math.max((end - start) / 60000, 0);
        const bucket = durationByDay.get(dayKey) || { total: 0, count: 0 };
        bucket.total += minutes;
        bucket.count += 1;
        durationByDay.set(dayKey, bucket);
      }
    });

    let cumulativeCost = 0;

    const trendPoints = recentDays.map((day) => ({ label: day.slice(5), value: taskByDay.get(day) || 0 }));
    const durationPoints = recentDays.map((day) => {
      const bucket = durationByDay.get(day) || { total: 0, count: 0 };
      const value = bucket.count > 0 ? bucket.total / bucket.count : 0;
      return { label: day.slice(5), value };
    });
    const successPoints = recentDays.map((day) => {
      const success = completedByDay.get(day) || 0;
      const failed = failedByDay.get(day) || 0;
      const denom = success + failed;
      return { label: day.slice(5), value: denom > 0 ? success / denom : 1 };
    });
    const costPoints = recentDays.map((day) => {
      cumulativeCost += costByDay.get(day) || 0;
      return { label: day.slice(5), value: cumulativeCost };
    });

    const avgDurationMinutes = durationPoints.reduce((sum, point) => sum + point.value, 0) / Math.max(durationPoints.length, 1);
    const totalCompleted = filteredVideos.filter((video) => video.status === 'completed').length;
    const totalFailed = filteredVideos.filter((video) => video.status === 'failed').length;
    const successRate = (totalCompleted + totalFailed) > 0 ? totalCompleted / (totalCompleted + totalFailed) : 1;
    const monthlyCost = costPoints[costPoints.length - 1]?.value || 0;

    return {
      trendPoints,
      durationPoints,
      successPoints,
      costPoints,
      avgDurationMinutes,
      successRate,
      monthlyCost,
    };
  }, [filteredVideos, range]);

  const handleGenerateNarration = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.projectId || !form.movieTitle.trim()) {
      setSubmitMsg('请选择项目并填写影片名称。');
      return;
    }

    setSubmitLoading(true);
    setSubmitMsg(null);
    try {
      const resp = await videoAPI.generateNarration(form.projectId, {
        movie_title: form.movieTitle.trim(),
        synopsis: form.synopsis.trim(),
        style: form.style,
        target_duration: form.targetDuration,
        voice: form.voice,
        provider: form.provider,
        aspect_ratio: form.aspectRatio,
        source_video_path: form.sourceVideoPath.trim(),
        source_video_url: form.sourceVideoUrl.trim(),
        creative_brief: form.creativeBrief.trim(),
      });
      const status = resp.data?.status || 'processing';
      const videoId = resp.data?.video_id || '';
      setSubmitMsg(`已提交解说视频任务，状态：${status}${videoId ? `，video_id: ${videoId}` : ''}`);
      await loadData();
    } catch (error: any) {
      setSubmitMsg(error?.response?.data?.error || '提交失败，请稍后重试。');
    } finally {
      setSubmitLoading(false);
    }
  };

  if (!isAuthenticated) {
    return (
      <main className="min-h-screen bg-[#f7f7f7] flex items-center justify-center px-4">
        <div className="max-w-md w-full text-center p-8 rounded-2xl site-card">
          <BarChart3 className="w-12 h-12 text-black/70 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-black mb-2">工厂控制台</h1>
          <p className="text-black/55 mb-6">请先登录后查看 AI 视频任务、渠道绑定状态与运营数据。</p>
          <Link href="/" className="inline-flex items-center justify-center px-5 py-2.5 rounded-lg bg-black text-white hover:bg-black/90 transition">
            返回首页
          </Link>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#f7f7f7] px-4 pb-10">
      <div className="max-w-6xl mx-auto">
        <section className="min-h-[calc(100vh-96px)] flex items-center justify-center py-10">
          <div className="w-full max-w-3xl text-center">
            <Link href="/" className="inline-flex items-center gap-2 text-black/55 hover:text-black transition">
              <ArrowLeft className="w-4 h-4" />
              返回首页
            </Link>
            <h1 className="site-h1 ui-mt-16 factory-title">AI 视频工厂控制台</h1>
            <p className="site-lead ui-mt-12 mx-auto max-w-2xl text-black/58">
              统一管理视频任务、模型引擎、渠道分发与运营复盘。当前以影视解说生产线为主，持续扩展到短剧、电影和广告内容。
            </p>
          </div>
        </section>

        {offlineMode ? (
          <div className="mb-6 rounded-xl border border-amber-300/30 bg-amber-300/10 px-4 py-3 text-sm text-amber-800">
            当前处于演示模式：后端接口不可达，页面展示的是前端兜底数据。
          </div>
        ) : null}

        {loading ? (
          <LoadingBlock text="正在加载控制台运营数据..." />
        ) : (
          <>
            <section className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
              <div className="rounded-xl site-card p-4">
                <div className="text-black/50 text-sm mb-1">总作品数</div>
                <div className="text-2xl text-black font-semibold">{summary?.total_projects ?? 0}</div>
              </div>
              <div className="rounded-xl site-card p-4">
                <div className="text-black/50 text-sm mb-1">生产任务总数</div>
                <div className="text-2xl text-black font-semibold">{summary?.total_generated_videos ?? 0}</div>
              </div>
              <div className="rounded-xl site-card p-4">
                <div className="text-black/50 text-sm mb-1">交付视频数</div>
                <div className="text-2xl text-black font-semibold">{summary?.total_completed_videos ?? 0}</div>
              </div>
              <div className="rounded-xl site-card p-4">
                <div className="text-black/50 text-sm mb-1">内部互动</div>
                <div className="text-2xl text-black font-semibold">{summary?.total_interactions ?? 0}</div>
              </div>
            </section>

            <OperationalInsightPanel
              trendPoints={insight.trendPoints}
              durationPoints={insight.durationPoints}
              successPoints={insight.successPoints}
              costPoints={insight.costPoints}
              avgDurationMinutes={insight.avgDurationMinutes}
              successRate={insight.successRate}
              monthlyCost={insight.monthlyCost}
                range={range}
                onRangeChange={setRange}
                provider={providerFilter}
                onProviderChange={setProviderFilter}
                pipeline={pipelineFilter}
                onPipelineChange={setPipelineFilter}
            />

            <section className="grid grid-cols-1 xl:grid-cols-2 gap-4 mb-8">
              <div className="rounded-xl site-card p-5">
                <div className="inline-flex items-center gap-2 text-sm text-black/65 mb-4">
                  <Sparkles className="w-4 h-4" />
                  业务阶段
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  {STAGE_CARDS.map((item) => (
                    <div key={item.title} className="group rounded-xl site-card p-4 transition-all duration-300 hover:border-black/25 cursor-default">
                      <div className="flex items-center gap-2 text-black font-medium mb-3">
                        <item.icon className="w-4 h-4 text-black/70 transition-all duration-300 group-hover:text-black" />
                        {item.title}
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {item.keywords.map((keyword) => (
                          <span key={keyword} className="px-2.5 py-1 rounded-full border border-black/10 bg-black/[0.03] text-xs text-black/70 transition-all duration-200 group-hover:border-black/30 group-hover:bg-black/[0.06] group-hover:text-black">
                            {keyword}
                          </span>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              <div className="rounded-xl site-card p-5">
                <div className="inline-flex items-center gap-2 text-sm text-black/65 mb-4">
                  <BarChart3 className="w-4 h-4" />
                  控制塔模块
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  {CONTROL_MODULES.map((item) => (
                    <div key={item.title} className="group rounded-xl site-card p-4 transition-all duration-300 hover:border-black/25 cursor-default">
                      <div className="flex items-center gap-2 text-black font-medium mb-3">
                        <item.icon className="w-4 h-4 text-black/70 transition-all duration-300 group-hover:text-black" />
                        {item.title}
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {item.keywords.map((keyword) => (
                          <span key={keyword} className="px-2.5 py-1 rounded-full border border-black/10 bg-black/[0.03] text-xs text-black/70 transition-all duration-200 group-hover:border-black/30 group-hover:bg-black/[0.06] group-hover:text-black">
                            {keyword}
                          </span>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </section>

            <section className="mb-8">
              <div className="flex items-center justify-between mb-3">
                <h2 className="text-xl font-semibold text-black">工厂核心中心</h2>
                <span className="text-sm text-black/50">先补齐模型、资产、分镜三层控制面板</span>
              </div>
              <FactoryHubCards />
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-semibold text-black mb-2">创建 AI 视频任务</h2>
              <p className="text-sm text-black/55 mb-3">核心必填在首层，高级输入可按需展开。</p>
              <div className="rounded-xl site-card p-4 mb-8">
                <form onSubmit={handleGenerateNarration} className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <label className="text-sm text-black/65">
                    项目
                    <select
                      value={form.projectId}
                      onChange={(e) => setForm((prev) => ({ ...prev, projectId: Number(e.target.value) }))}
                      className="factory-input mt-1 w-full rounded-lg px-3 py-2 bg-white border-black/15 text-black"
                    >
                      {projects.length === 0 ? (
                        <option value={0}>暂无项目</option>
                      ) : (
                        projects.map((p) => (
                          <option key={p.id} value={p.id}>{p.title || `项目 #${p.id}`}</option>
                        ))
                      )}
                    </select>
                  </label>

                  <label className="text-sm text-black/65">
                    视频主题 / 片名
                    <input
                      value={form.movieTitle}
                      onChange={(e) => setForm((prev) => ({ ...prev, movieTitle: e.target.value }))}
                      className="factory-input mt-1 w-full rounded-lg px-3 py-2 bg-white border-black/15 text-black"
                      placeholder="例如：药命效应：卧龙入局"
                    />
                  </label>

                  <label className="text-sm text-black/65 md:col-span-2">
                    剧情简介（可选）
                    <textarea
                      value={form.synopsis}
                      onChange={(e) => setForm((prev) => ({ ...prev, synopsis: e.target.value }))}
                      className="factory-input mt-1 w-full rounded-lg px-3 py-2 min-h-24 bg-white border-black/15 text-black"
                      placeholder="一句话概述剧情或主题"
                    />
                  </label>

                  <details className="md:col-span-2 rounded-xl site-card p-4">
                    <summary className="cursor-pointer text-sm text-black/70">高级输入（可选）：素材与创作要求</summary>
                    <div className="mt-4 grid grid-cols-1 gap-4">
                      <label className="text-sm text-black/65">
                        私有素材路径
                        <input
                          value={form.sourceVideoPath}
                          onChange={(e) => setForm((prev) => ({ ...prev, sourceVideoPath: e.target.value }))}
                          className="factory-input mt-1 w-full rounded-lg px-3 py-2 bg-white border-black/15 text-black"
                          placeholder="/mnt/shared/story-assets/episode-01.mp4"
                        />
                      </label>

                      <label className="text-sm text-black/65">
                        公网素材直链
                        <input
                          value={form.sourceVideoUrl}
                          onChange={(e) => setForm((prev) => ({ ...prev, sourceVideoUrl: e.target.value }))}
                          className="factory-input mt-1 w-full rounded-lg px-3 py-2 bg-white border-black/15 text-black"
                          placeholder="https://cdn.example.com/assets/episode-01.mp4"
                        />
                      </label>

                      <label className="text-sm text-black/65">
                        创作要求
                        <textarea
                          value={form.creativeBrief}
                          onChange={(e) => setForm((prev) => ({ ...prev, creativeBrief: e.target.value }))}
                          className="factory-input mt-1 w-full rounded-lg px-3 py-2 min-h-24 bg-white border-black/15 text-black"
                          placeholder="例如：历史化第一人称、偏策略叙事、结尾反转"
                        />
                      </label>
                    </div>
                  </details>

                  <label className="text-sm text-black/65">
                    风格
                    <select
                      value={form.style}
                      onChange={(e) => setForm((prev) => ({ ...prev, style: e.target.value }))}
                      className="factory-input mt-1 w-full rounded-lg px-3 py-2 bg-white border-black/15 text-black"
                    >
                      <option value="深度分析">深度分析</option>
                      <option value="搞笑">搞笑</option>
                      <option value="情感向">情感向</option>
                    </select>
                  </label>

                  <label className="text-sm text-black/65">
                    目标时长（秒）
                    <input
                      type="number"
                      min={10}
                      max={600}
                      value={form.targetDuration}
                      onChange={(e) => setForm((prev) => ({ ...prev, targetDuration: Number(e.target.value) || 90 }))}
                      className="factory-input mt-1 w-full rounded-lg px-3 py-2 bg-white border-black/15 text-black"
                    />
                  </label>

                  <label className="text-sm text-black/65">
                    旁白声音（解说阶段）
                    <select
                      value={form.voice}
                      onChange={(e) => setForm((prev) => ({ ...prev, voice: e.target.value }))}
                      className="factory-input mt-1 w-full rounded-lg px-3 py-2 bg-white border-black/15 text-black"
                    >
                      <option value="female_cn">女声（Ting-Ting）</option>
                      <option value="male_cn">男声（Sin-ji）</option>
                      <option value="Mei-Jia">女声（Mei-Jia）</option>
                    </select>
                  </label>

                  <label className="text-sm text-black/65">
                    AI 视频引擎
                    <select
                      value={form.provider}
                      onChange={(e) => setForm((prev) => ({ ...prev, provider: e.target.value as 'runway' | 'pika' | 'local' }))}
                      className="factory-input mt-1 w-full rounded-lg px-3 py-2 bg-white border-black/15 text-black"
                    >
                      <option value="runway">Runway（主引擎）</option>
                      <option value="pika">Pika（备选）</option>
                      <option value="local">本地调试链路</option>
                    </select>
                  </label>

                  <label className="text-sm text-black/65">
                    画幅
                    <select
                      value={form.aspectRatio}
                      onChange={(e) => setForm((prev) => ({ ...prev, aspectRatio: e.target.value as '16:9' | '9:16' }))}
                      className="factory-input mt-1 w-full rounded-lg px-3 py-2 bg-white border-black/15 text-black"
                    >
                      <option value="16:9">16:9</option>
                      <option value="9:16">9:16</option>
                    </select>
                  </label>

                  <div className="md:col-span-2 flex items-center gap-3">
                    <button
                      type="submit"
                      disabled={submitLoading || projects.length === 0}
                      className="btn-base btn-dark btn-m disabled:opacity-60"
                    >
                      {submitLoading ? '提交中...' : '创建视频任务'}
                    </button>
                    {submitMsg && <span className="text-sm text-black/65">{submitMsg}</span>}
                  </div>
                </form>
              </div>

              <div className="flex items-center justify-between mb-3">
                <h2 className="text-xl font-semibold text-black">平台状态</h2>
                <MiniLinkButton href="/platforms" size="s" tone="light" icon={Link2} iconSize={14}>
                  管理平台绑定
                </MiniLinkButton>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {(summary?.configured_platforms || ['douyin', 'xiaohongshu', 'bilibili', 'weibo']).map((platform) => {
                  const isBound = boundSet.has(platform as PlatformKind);
                  const account = summary?.bound_platforms.find((p) => p.platform === platform);
                  return (
                    <div key={platform} className="rounded-xl site-card p-4 flex items-center justify-between">
                      <div>
                        <div className="text-black font-medium">{PLATFORM_LABEL[platform as PlatformKind] || platform}</div>
                        <div className="text-black/55 text-sm mt-1">
                          {isBound ? `已绑定：${account?.nickname || '已连接'}` : '未绑定'}
                        </div>
                      </div>
                      <div>
                        {isBound ? <CheckCircle2 className="w-5 h-5 text-black/80" /> : <XCircle className="w-5 h-5 text-black/35" />}
                      </div>
                    </div>
                  );
                })}
              </div>
            </section>

            <section>
              <h2 className="text-xl font-semibold text-black mb-3">最近任务记录</h2>
              <div className="rounded-xl border border-black/10 bg-white overflow-hidden">
                <div className="grid grid-cols-12 gap-3 px-4 py-3 text-xs text-black/50 border-b border-black/10">
                  <div className="col-span-4">标题</div>
                  <div className="col-span-2">类型</div>
                  <div className="col-span-2">状态</div>
                  <div className="col-span-2">来源</div>
                  <div className="col-span-1">时间</div>
                  <div className="col-span-1">操作</div>
                </div>
                {videos.length === 0 ? (
                  <div className="p-4">
                    <EmptyBlock title="暂无任务记录" hint="先创建一个视频任务，系统会在这里显示执行状态与产出链接。" />
                  </div>
                ) : (
                  videos.map((item) => (
                    <div key={item.id} className="grid grid-cols-12 gap-3 px-4 py-3 border-b border-black/5 last:border-b-0 items-center">
                      <div className="col-span-4 text-black text-sm truncate flex items-center gap-2">
                        <Video className="w-4 h-4 text-black/70" />
                        {item.title || item.video_id}
                      </div>
                      <div className="col-span-2 text-black/65 text-sm">{item.task_type === 'narration_video' ? '影视解说' : '镜头任务'}</div>
                      <div className="col-span-2 text-sm">
                        <span className={`inline-flex px-2 py-1 rounded-md border ${statusBadge(item.status)}`}>{item.status}</span>
                      </div>
                      <div className="col-span-2 text-black/65 text-sm">{PROVIDER_LABEL[item.provider] || item.provider}</div>
                      <div className="col-span-1 text-black/50 text-xs inline-flex items-center gap-1">
                        <Clock3 className="w-3.5 h-3.5" />
                        {new Date(item.created_at).toLocaleDateString()}
                      </div>
                      <div className="col-span-1 text-xs">
                        {item.video_url ? (
                          <MiniButton
                            type="button"
                            size="xs"
                            tone="light"
                            icon={ExternalLink}
                            iconSize={14}
                            onClick={() => window.open(item.video_url!, '_blank', 'noopener,noreferrer')}
                          >
                            查看
                          </MiniButton>
                        ) : (
                          <span className="text-gray-500">-</span>
                        )}
                      </div>
                    </div>
                  ))
                )}
              </div>
            </section>
          </>
        )}
      </div>
    </main>
  );
}
