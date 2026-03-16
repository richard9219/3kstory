'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { analyticsAPI, projectAPI, videoAPI } from '@/lib/api/client';
import { useAuthStore } from '@/lib/store/authStore';
import type { AnalyticsSummary, PlatformKind, Project, VideoTask } from '@/types';
import { ArrowLeft, BarChart3, CheckCircle2, Clock3, Link2, Video, XCircle } from 'lucide-react';

const PLATFORM_LABEL: Record<PlatformKind, string> = {
  douyin: '抖音',
  xiaohongshu: '小红书',
  bilibili: 'B站',
  weibo: '微博',
};

function statusBadge(status: VideoTask['status']) {
  switch (status) {
    case 'completed':
      return 'text-green-300 bg-green-500/10 border-green-500/30';
    case 'failed':
      return 'text-red-300 bg-red-500/10 border-red-500/30';
    case 'processing':
      return 'text-yellow-300 bg-yellow-500/10 border-yellow-500/30';
    default:
      return 'text-gray-300 bg-gray-500/10 border-gray-500/30';
  }
}

export default function DashboardPage() {
  const { isAuthenticated } = useAuthStore();
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null);
  const [videos, setVideos] = useState<VideoTask[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitLoading, setSubmitLoading] = useState(false);
  const [submitMsg, setSubmitMsg] = useState<string | null>(null);
  const [form, setForm] = useState({
    projectId: 0,
    movieTitle: '',
    synopsis: '',
    style: '深度分析',
    targetDuration: 90,
    provider: 'local' as 'runway' | 'pika' | 'local',
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

  const boundSet = useMemo(() => {
    return new Set((summary?.bound_platforms || []).map((p) => p.platform));
  }, [summary?.bound_platforms]);

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
        provider: form.provider,
        aspect_ratio: form.aspectRatio,
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
      <main className="min-h-screen bg-gradient-to-br from-purple-900/20 via-blue-900/20 to-black flex items-center justify-center px-4">
        <div className="max-w-md w-full text-center p-8 rounded-2xl border border-white/10 bg-white/5">
          <BarChart3 className="w-12 h-12 text-purple-300 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-white mb-2">运营看板</h1>
          <p className="text-gray-400 mb-6">请先登录后查看平台绑定状态和视频运营数据。</p>
          <Link href="/" className="inline-flex items-center justify-center px-5 py-2.5 rounded-lg bg-purple-600 text-white hover:bg-purple-700 transition">
            返回首页
          </Link>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-gradient-to-br from-purple-900/20 via-blue-900/20 to-black py-10 px-4">
      <div className="max-w-6xl mx-auto">
        <Link href="/" className="inline-flex items-center gap-2 text-gray-400 hover:text-white mb-6 transition">
          <ArrowLeft className="w-4 h-4" />
          返回首页
        </Link>

        <h1 className="text-3xl font-bold text-white mb-2">运营看板</h1>
        <p className="text-gray-400 mb-8">聚合平台绑定状态与内部生成数据（里程碑 1 版本）。</p>

        {loading ? (
          <div className="text-gray-300">加载中...</div>
        ) : (
          <>
            <section className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
              <div className="rounded-xl border border-white/10 bg-white/5 p-4">
                <div className="text-gray-400 text-sm mb-1">总作品数</div>
                <div className="text-2xl text-white font-semibold">{summary?.total_projects ?? 0}</div>
              </div>
              <div className="rounded-xl border border-white/10 bg-white/5 p-4">
                <div className="text-gray-400 text-sm mb-1">生成任务总数</div>
                <div className="text-2xl text-white font-semibold">{summary?.total_generated_videos ?? 0}</div>
              </div>
              <div className="rounded-xl border border-white/10 bg-white/5 p-4">
                <div className="text-gray-400 text-sm mb-1">完成视频数</div>
                <div className="text-2xl text-white font-semibold">{summary?.total_completed_videos ?? 0}</div>
              </div>
              <div className="rounded-xl border border-white/10 bg-white/5 p-4">
                <div className="text-gray-400 text-sm mb-1">总互动（内部）</div>
                <div className="text-2xl text-white font-semibold">{summary?.total_interactions ?? 0}</div>
              </div>
            </section>

            <section className="mb-8">
              <h2 className="text-xl font-semibold text-white mb-3">创建解说视频</h2>
              <div className="rounded-xl border border-white/10 bg-white/5 p-4 mb-8">
                <form onSubmit={handleGenerateNarration} className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <label className="text-sm text-gray-300">
                    项目
                    <select
                      value={form.projectId}
                      onChange={(e) => setForm((prev) => ({ ...prev, projectId: Number(e.target.value) }))}
                      className="mt-1 w-full rounded-lg bg-black/30 border border-white/15 text-white px-3 py-2"
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

                  <label className="text-sm text-gray-300">
                    影片/剧名
                    <input
                      value={form.movieTitle}
                      onChange={(e) => setForm((prev) => ({ ...prev, movieTitle: e.target.value }))}
                      className="mt-1 w-full rounded-lg bg-black/30 border border-white/15 text-white px-3 py-2"
                      placeholder="例如：狂飙"
                    />
                  </label>

                  <label className="text-sm text-gray-300 md:col-span-2">
                    剧情简介（可选）
                    <textarea
                      value={form.synopsis}
                      onChange={(e) => setForm((prev) => ({ ...prev, synopsis: e.target.value }))}
                      className="mt-1 w-full rounded-lg bg-black/30 border border-white/15 text-white px-3 py-2 min-h-24"
                      placeholder="填写简要剧情，帮助生成更准确解说文案"
                    />
                  </label>

                  <label className="text-sm text-gray-300">
                    风格
                    <select
                      value={form.style}
                      onChange={(e) => setForm((prev) => ({ ...prev, style: e.target.value }))}
                      className="mt-1 w-full rounded-lg bg-black/30 border border-white/15 text-white px-3 py-2"
                    >
                      <option value="深度分析">深度分析</option>
                      <option value="搞笑">搞笑</option>
                      <option value="情感向">情感向</option>
                    </select>
                  </label>

                  <label className="text-sm text-gray-300">
                    目标时长（秒）
                    <input
                      type="number"
                      min={10}
                      max={600}
                      value={form.targetDuration}
                      onChange={(e) => setForm((prev) => ({ ...prev, targetDuration: Number(e.target.value) || 90 }))}
                      className="mt-1 w-full rounded-lg bg-black/30 border border-white/15 text-white px-3 py-2"
                    />
                  </label>

                  <label className="text-sm text-gray-300">
                    视频 Provider
                    <select
                      value={form.provider}
                      onChange={(e) => setForm((prev) => ({ ...prev, provider: e.target.value as 'runway' | 'pika' | 'local' }))}
                      className="mt-1 w-full rounded-lg bg-black/30 border border-white/15 text-white px-3 py-2"
                    >
                      <option value="local">local（推荐）</option>
                      <option value="runway">runway</option>
                      <option value="pika">pika</option>
                    </select>
                  </label>

                  <label className="text-sm text-gray-300">
                    画幅
                    <select
                      value={form.aspectRatio}
                      onChange={(e) => setForm((prev) => ({ ...prev, aspectRatio: e.target.value as '16:9' | '9:16' }))}
                      className="mt-1 w-full rounded-lg bg-black/30 border border-white/15 text-white px-3 py-2"
                    >
                      <option value="16:9">16:9</option>
                      <option value="9:16">9:16</option>
                    </select>
                  </label>

                  <div className="md:col-span-2 flex items-center gap-3">
                    <button
                      type="submit"
                      disabled={submitLoading || projects.length === 0}
                      className="px-4 py-2 rounded-lg bg-purple-600 text-white hover:bg-purple-700 disabled:opacity-60"
                    >
                      {submitLoading ? '提交中...' : '生成解说视频'}
                    </button>
                    {submitMsg && <span className="text-sm text-gray-300">{submitMsg}</span>}
                  </div>
                </form>
              </div>

              <div className="flex items-center justify-between mb-3">
                <h2 className="text-xl font-semibold text-white">平台状态</h2>
                <Link href="/platforms" className="inline-flex items-center gap-1 text-purple-300 hover:text-purple-200 text-sm">
                  <Link2 className="w-4 h-4" />
                  管理平台绑定
                </Link>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {(summary?.configured_platforms || ['douyin', 'xiaohongshu', 'bilibili', 'weibo']).map((platform) => {
                  const isBound = boundSet.has(platform as PlatformKind);
                  const account = summary?.bound_platforms.find((p) => p.platform === platform);
                  return (
                    <div key={platform} className="rounded-xl border border-white/10 bg-white/5 p-4 flex items-center justify-between">
                      <div>
                        <div className="text-white font-medium">{PLATFORM_LABEL[platform as PlatformKind] || platform}</div>
                        <div className="text-gray-400 text-sm mt-1">
                          {isBound ? `已绑定：${account?.nickname || '已连接'}` : '未绑定'}
                        </div>
                      </div>
                      <div>
                        {isBound ? <CheckCircle2 className="w-5 h-5 text-green-400" /> : <XCircle className="w-5 h-5 text-gray-500" />}
                      </div>
                    </div>
                  );
                })}
              </div>
            </section>

            <section>
              <h2 className="text-xl font-semibold text-white mb-3">最近记录</h2>
              <div className="rounded-xl border border-white/10 bg-white/5 overflow-hidden">
                <div className="grid grid-cols-12 gap-3 px-4 py-3 text-xs text-gray-400 border-b border-white/10">
                  <div className="col-span-4">标题</div>
                  <div className="col-span-2">类型</div>
                  <div className="col-span-2">状态</div>
                  <div className="col-span-2">来源</div>
                  <div className="col-span-1">时间</div>
                  <div className="col-span-1">操作</div>
                </div>
                {videos.length === 0 ? (
                  <div className="px-4 py-6 text-gray-500 text-sm">暂无记录</div>
                ) : (
                  videos.map((item) => (
                    <div key={item.id} className="grid grid-cols-12 gap-3 px-4 py-3 border-b border-white/5 last:border-b-0 items-center">
                      <div className="col-span-4 text-white text-sm truncate flex items-center gap-2">
                        <Video className="w-4 h-4 text-purple-300" />
                        {item.title || item.video_id}
                      </div>
                      <div className="col-span-2 text-gray-300 text-sm">{item.task_type || 'generate_video'}</div>
                      <div className="col-span-2 text-sm">
                        <span className={`inline-flex px-2 py-1 rounded-md border ${statusBadge(item.status)}`}>{item.status}</span>
                      </div>
                      <div className="col-span-2 text-gray-300 text-sm">{item.provider}</div>
                      <div className="col-span-1 text-gray-400 text-xs inline-flex items-center gap-1">
                        <Clock3 className="w-3.5 h-3.5" />
                        {new Date(item.created_at).toLocaleDateString()}
                      </div>
                      <div className="col-span-1 text-xs">
                        {item.video_url ? (
                          <a
                            href={item.video_url}
                            target="_blank"
                            rel="noreferrer"
                            className="text-purple-300 hover:text-purple-200"
                          >
                            查看
                          </a>
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
