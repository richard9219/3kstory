"use client";

import { DragEvent, FormEvent, useEffect, useMemo, useState } from 'react';
import { Clapperboard, Eye, Film, GitBranchPlus, Loader2, Lock, PlayCircle, RefreshCcw, Sparkles, Unlock } from 'lucide-react';
import { ConsoleSection, FactoryConsoleShell, MetricStrip } from '@/components/common/FactoryConsoleShell';
import { EmptyBlock, LoadingBlock } from '@/components/common/StateBlocks';
import MiniButton from '@/components/common/MiniButton';
import { projectAPI, storyboardAPI } from '@/lib/api/client';
import type { PlatformKind, Project, StoryboardShot, StoryboardTimeline, StoryboardVersionNode } from '@/types';

const PLATFORM_LABEL: Record<PlatformKind, string> = {
  douyin: '抖音',
  xiaohongshu: '小红书',
  bilibili: 'B站',
  weibo: '微博',
};

const DIRECTOR_STYLE_TEMPLATES: Array<{
  key: string;
  label: string;
  camera_language: string;
  emotion_tone: string;
  transition_type: 'cut' | 'fade' | 'wipe' | 'match';
  prompt_prefix: string;
}> = [
  {
    key: 'kurosawa',
    label: '黑泽明',
    camera_language: '长焦压缩 + 低机位横移 + 风雨动势',
    emotion_tone: '肃杀、宿命、张力',
    transition_type: 'wipe',
    prompt_prefix: 'high-contrast monochrome mood, dynamic weather, samurai-like blocking',
  },
  {
    key: 'zhangyimou',
    label: '张艺谋',
    camera_language: '色块构图 + 仪式化调度 + 大景别对称',
    emotion_tone: '浓烈、仪式、情绪外放',
    transition_type: 'fade',
    prompt_prefix: 'bold color choreography, ceremonial composition, poetic epic tableau',
  },
  {
    key: 'chenkaige',
    label: '陈凯歌',
    camera_language: '戏剧性推轨 + 人物群像层次',
    emotion_tone: '抒情、史诗、反思',
    transition_type: 'match',
    prompt_prefix: 'operatic visual language, historical reflection, lyrical dramatic pacing',
  },
  {
    key: 'fengxiaogang',
    label: '冯小刚',
    camera_language: '生活流手持 + 快节奏切换',
    emotion_tone: '写实、讽刺、群体烟火气',
    transition_type: 'cut',
    prompt_prefix: 'urban realism, brisk rhythm, satirical social texture, handheld immediacy',
  },
];

const TIMELINE_GRID_MS = 500;

function formatMs(ms?: number) {
  const total = Math.max(0, Math.round((ms || 0) / 1000));
  const minutes = Math.floor(total / 60);
  const seconds = total % 60;
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
}

function statusTone(status?: string) {
  switch (status) {
    case 'completed':
      return 'bg-emerald-500/12 text-emerald-700 border-emerald-500/20';
    case 'processing':
      return 'bg-amber-500/12 text-amber-700 border-amber-500/20';
    case 'failed':
      return 'bg-red-500/12 text-red-700 border-red-500/20';
    default:
      return 'bg-black/[0.04] text-black/65 border-black/10';
  }
}

function timelineDurationMs(shot: StoryboardShot) {
  return Math.max(shot.timeline_duration_ms || 0, (shot.duration || 1) * 1000);
}

function snapToGrid(ms: number, gridMs: number) {
  if (gridMs <= 0) return Math.max(0, Math.round(ms));
  return Math.max(0, Math.round(ms / gridMs) * gridMs);
}

function resolveCollisions(startMs: number, durationMs: number, others: StoryboardShot[]) {
  const sorted = [...others].sort((a, b) => (a.timeline_start_ms || 0) - (b.timeline_start_ms || 0));
  let cursor = Math.max(0, startMs);
  for (const item of sorted) {
    const itemStart = item.timeline_start_ms || 0;
    const itemEnd = itemStart + timelineDurationMs(item);
    const candidateEnd = cursor + durationMs;
    if (candidateEnd <= itemStart || cursor >= itemEnd) {
      continue;
    }
    cursor = snapToGrid(itemEnd, TIMELINE_GRID_MS);
  }
  return cursor;
}

export default function StoryboardsCenterPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [activeProjectId, setActiveProjectId] = useState<number>(0);
  const [timeline, setTimeline] = useState<StoryboardTimeline | null>(null);
  const [versionTree, setVersionTree] = useState<StoryboardVersionNode[]>([]);
  const [selectedShotId, setSelectedShotId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [rendering, setRendering] = useState(false);
  const [publishChecking, setPublishChecking] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [publishPlatform, setPublishPlatform] = useState<PlatformKind>('douyin');
  const [directorTemplateKey, setDirectorTemplateKey] = useState<string>('');
  const [draggingShotId, setDraggingShotId] = useState<number | null>(null);
  const [previewURL, setPreviewURL] = useState<string | null>(null);
  const [msg, setMsg] = useState<string | null>(null);
  const [form, setForm] = useState({
    title: '',
    chapter: '',
    prompt: '',
    camera_language: '',
    emotion_tone: '',
    duration: 6,
    aspect_ratio: '16:9' as '16:9' | '9:16',
    transition_type: 'cut' as 'cut' | 'fade' | 'wipe' | 'match',
    transition_duration_ms: 0,
  });

  const shots = timeline?.shots || [];

  const selectedShot = useMemo(
    () => shots.find((item) => item.id === selectedShotId) || null,
    [shots, selectedShotId]
  );

  const groupedTracks = useMemo(() => {
    return shots.reduce<Record<number, StoryboardShot[]>>((acc, shot) => {
      const track = shot.track_index || 1;
      if (!acc[track]) acc[track] = [];
      acc[track].push(shot);
      return acc;
    }, {});
  }, [shots]);

  const timelineHorizonMs = useMemo(() => {
    const maxEnd = shots.reduce((acc, shot) => {
      const end = (shot.timeline_start_ms || 0) + timelineDurationMs(shot);
      return Math.max(acc, end);
    }, 0);
    return Math.max(timeline?.total_duration_ms || 0, maxEnd, 10000);
  }, [shots, timeline?.total_duration_ms]);

  const timelineRulerTicks = useMemo(() => {
    const ticks: number[] = [];
    for (let t = 0; t <= timelineHorizonMs; t += 1000) {
      ticks.push(t);
    }
    return ticks;
  }, [timelineHorizonMs]);

  const activeVersionNode = useMemo(() => {
    if (!selectedShot) return null;
    return versionTree.find((node) => node.root.id === selectedShot.root_shot_id || node.root.id === selectedShot.id || node.versions.some((version) => version.id === selectedShot.id)) || null;
  }, [selectedShot, versionTree]);

  const reloadProjectData = async (projectId: number) => {
    if (!projectId) {
      setTimeline(null);
      setVersionTree([]);
      setSelectedShotId(null);
      return;
    }
    const [timelineRes, treeRes] = await Promise.all([
      storyboardAPI.getTimeline(projectId),
      storyboardAPI.getVersionTree(projectId),
    ]);
    const nextTimeline = timelineRes.data || null;
    setTimeline(nextTimeline);
    setVersionTree(treeRes.data.data || []);
    setSelectedShotId((prev) => {
      if (prev && nextTimeline?.shots.some((item) => item.id === prev)) return prev;
      return nextTimeline?.shots[0]?.id || null;
    });
  };

  useEffect(() => {
    projectAPI.list()
      .then(async (res) => {
        const list = (res.data || []) as Project[];
        setProjects(list);
        if (list[0]?.id) {
          setActiveProjectId(list[0].id);
          await reloadProjectData(list[0].id);
        }
      })
      .catch(() => setMsg('加载项目或时间轴失败'))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (!selectedShot) return;
    setForm({
      title: selectedShot.title || '',
      chapter: selectedShot.chapter || '',
      prompt: selectedShot.prompt || '',
      camera_language: selectedShot.camera_language || '',
      emotion_tone: selectedShot.emotion_tone || '',
      duration: selectedShot.duration || 6,
      aspect_ratio: selectedShot.aspect_ratio || '16:9',
      transition_type: selectedShot.transition_type || 'cut',
      transition_duration_ms: selectedShot.transition_duration_ms || 0,
    });
  }, [selectedShot]);

  useEffect(() => {
    return () => {
      if (previewURL?.startsWith('blob:')) {
        URL.revokeObjectURL(previewURL);
      }
    };
  }, [previewURL]);

  const handleBootstrap = async () => {
    if (!activeProjectId) return;
    setSaving(true);
    setMsg(null);
    try {
      const resp = await storyboardAPI.bootstrapFromScenes(activeProjectId);
      await reloadProjectData(activeProjectId);
      setMsg(`已从场景初始化 ${resp.data.bootstrapped || 0} 个镜头`);
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '初始化失败');
    } finally {
      setSaving(false);
    }
  };

  const handleSaveShot = async (e: FormEvent) => {
    e.preventDefault();
    if (!activeProjectId || !selectedShot) return;
    setSaving(true);
    setMsg(null);
    try {
      await storyboardAPI.updateShot(activeProjectId, selectedShot.id, {
        title: form.title.trim(),
        chapter: form.chapter.trim(),
        prompt: form.prompt.trim(),
        camera_language: form.camera_language.trim(),
        emotion_tone: form.emotion_tone.trim(),
        duration: form.duration,
        timeline_duration_ms: form.duration * 1000,
        aspect_ratio: form.aspect_ratio,
        transition_type: form.transition_type,
        transition_duration_ms: form.transition_duration_ms,
      });
      await reloadProjectData(activeProjectId);
      setMsg('镜头参数已保存到执行时间轴');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '保存镜头失败');
    } finally {
      setSaving(false);
    }
  };

  const handleGenerate = async (mode: 'generate' | 'regenerate') => {
    if (!activeProjectId || !selectedShot) return;
    setSaving(true);
    setMsg(null);
    try {
      if (mode === 'generate') {
        await storyboardAPI.generateShot(activeProjectId, selectedShot.id, {
          prompt: form.prompt.trim(),
          duration: form.duration,
          aspect_ratio: form.aspect_ratio,
        });
      } else {
        const note = window.prompt('局部重生版本备注', '优化镜头节奏与动作连续性') || '';
        await storyboardAPI.regenerateShot(activeProjectId, selectedShot.id, {
          prompt: form.prompt.trim(),
          duration: form.duration,
          aspect_ratio: form.aspect_ratio,
          version_note: note,
        });
      }
      await reloadProjectData(activeProjectId);
      setMsg(mode === 'generate' ? '已提交镜头生成' : '已创建局部重生版本');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '镜头执行失败');
    } finally {
      setSaving(false);
    }
  };

  const handleToggleLock = async () => {
    if (!activeProjectId || !selectedShot) return;
    setSaving(true);
    try {
      await storyboardAPI.updateShot(activeProjectId, selectedShot.id, { locked: !selectedShot.locked });
      await reloadProjectData(activeProjectId);
      setMsg(selectedShot.locked ? '镜头已解锁，可继续重生' : '镜头已锁定，防止误改');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '切换锁定状态失败');
    } finally {
      setSaving(false);
    }
  };

  const handleRenderTimeline = async () => {
    if (!activeProjectId) return;
    setRendering(true);
    setMsg(null);
    try {
      const res = await storyboardAPI.renderTimeline(activeProjectId);
      const task = res.data;
      let nextPreviewURL = task.video_url || null;
      if ((!nextPreviewURL || nextPreviewURL.startsWith('/api/')) && task.video_id) {
        const blobRes = await storyboardAPI.fetchDirectorCutBlob(activeProjectId, task.video_id);
        nextPreviewURL = URL.createObjectURL(blobRes.data);
      }
      if (previewURL?.startsWith('blob:')) {
        URL.revokeObjectURL(previewURL);
      }
      setPreviewURL(nextPreviewURL);
      await reloadProjectData(activeProjectId);
      setMsg('导演合成片已生成，可直接预览');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '时间轴出片失败');
    } finally {
      setRendering(false);
    }
  };

  const handleShotDragStart = (shotId: number) => {
    setDraggingShotId(shotId);
  };

  const handleShotDropToTrack = async (event: DragEvent<HTMLDivElement>, trackIndex: number) => {
    event.preventDefault();
    if (!activeProjectId || !draggingShotId || !timeline) return;
    const dragged = timeline.shots.find((item) => item.id === draggingShotId);
    if (!dragged || dragged.locked) {
      setDraggingShotId(null);
      return;
    }
    const rect = event.currentTarget.getBoundingClientRect();
    const px = Math.min(Math.max(event.clientX - rect.left, 0), rect.width);
    const ratio = rect.width > 0 ? px / rect.width : 0;
    const desiredStart = snapToGrid(ratio * timelineHorizonMs, TIMELINE_GRID_MS);
    const currentDuration = timelineDurationMs(dragged);
    const othersOnTrack = timeline.shots.filter((item) => item.id !== draggingShotId && (item.track_index || 1) === trackIndex);
    const nextStartMs = resolveCollisions(desiredStart, currentDuration, othersOnTrack);

    setSaving(true);
    setMsg(null);
    try {
      await storyboardAPI.updateShot(activeProjectId, draggingShotId, {
        track_index: trackIndex,
        timeline_start_ms: nextStartMs,
      });
      await reloadProjectData(activeProjectId);
      setMsg(`已移动镜头到轨道 ${trackIndex}，起点 ${formatMs(nextStartMs)}（${TIMELINE_GRID_MS}ms 吸附）`);
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '拖拽调整失败');
    } finally {
      setDraggingShotId(null);
      setSaving(false);
    }
  };

  const handleApplyDirectorTemplate = async () => {
    if (!activeProjectId || !selectedShot || !directorTemplateKey) return;
    const template = DIRECTOR_STYLE_TEMPLATES.find((item) => item.key === directorTemplateKey);
    if (!template) return;
    setSaving(true);
    setMsg(null);
    try {
      const nextPrompt = `${template.prompt_prefix}. ${form.prompt || ''}`.trim();
      await storyboardAPI.updateShot(activeProjectId, selectedShot.id, {
        camera_language: template.camera_language,
        emotion_tone: template.emotion_tone,
        transition_type: template.transition_type,
        prompt: nextPrompt,
      });
      await reloadProjectData(activeProjectId);
      setMsg(`已应用 ${template.label} 风格模板`);
      setDirectorTemplateKey('');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '应用导演模板失败');
    } finally {
      setSaving(false);
    }
  };

  const handleDirectorGateCheck = async () => {
    if (!activeProjectId || !timeline?.latest_export?.video_id) return;
    setPublishChecking(true);
    setMsg(null);
    try {
      const resp = await storyboardAPI.autoPublishDirectorCut(activeProjectId, timeline.latest_export.video_id);
      setMsg(resp.data.status === 'passed' ? `导演版通过门禁（score ${resp.data.score.toFixed(3)} / 阈值 ${resp.data.threshold.toFixed(2)}）` : `导演版被门禁拦截：${resp.data.reason}`);
    } catch (error: any) {
      const reason = error?.response?.data?.reason;
      setMsg(reason || error?.response?.data?.error || '导演版门禁校验失败');
    } finally {
      setPublishChecking(false);
    }
  };

  const handleDirectorPublish = async () => {
    if (!activeProjectId || !timeline?.latest_export?.video_id) return;
    setPublishing(true);
    setMsg(null);
    try {
      const resp = await storyboardAPI.publishDirectorCut(activeProjectId, timeline.latest_export.video_id, { platform: publishPlatform });
      setMsg(resp.data.status === 'published' ? `导演版已提交到${PLATFORM_LABEL[publishPlatform]}发布流程` : (resp.data.reason || '发布未通过'));
      await reloadProjectData(activeProjectId);
    } catch (error: any) {
      const reason = error?.response?.data?.reason;
      setMsg(reason || error?.response?.data?.error || '导演版发布失败');
    } finally {
      setPublishing(false);
    }
  };

  return (
    <FactoryConsoleShell
      eyebrow="Director Workbench"
      title="分镜中心 / 导演工作台"
      subtitle="时间轴现在是执行源，不再只是镜头台账。镜头可单独生成、局部重生、锁定入轨，再由导演台一键合成成片。"
    >
      <MetricStrip
        items={[
          { label: '时间轴总时长', value: formatMs(timeline?.total_duration_ms), hint: '当前分镜轨道总长度。' },
          { label: '已就绪镜头', value: String(timeline?.ready_shot_count || 0), hint: '已完成并可进导演合成的镜头数。' },
          { label: '镜头总数', value: String(timeline?.total_shot_count || 0), hint: '当前项目时间轴上的镜头单元。' },
        ]}
      />

      <ConsoleSection title="项目与导演合成" description="先选项目，再从场景初始化镜头骨架，最后在就绪镜头上合成导演版成片。">
        <div className="grid grid-cols-1 xl:grid-cols-[1.4fr_1fr] gap-4 items-start">
          <div className="rounded-2xl site-card p-4 space-y-4">
            <label className="text-sm text-black/65 block">
              目标项目
              <select
                value={activeProjectId}
                onChange={async (e) => {
                  const nextId = Number(e.target.value) || 0;
                  setActiveProjectId(nextId);
                  setLoading(true);
                  try {
                    await reloadProjectData(nextId);
                  } catch {
                    setMsg('加载项目时间轴失败');
                  } finally {
                    setLoading(false);
                  }
                }}
                className="factory-input mt-1 w-full rounded-lg px-3 py-2"
              >
                {projects.length === 0 ? <option value={0}>暂无项目</option> : projects.map((project) => (
                  <option key={project.id} value={project.id}>{project.title || `项目 #${project.id}`}</option>
                ))}
              </select>
            </label>

            <div className="flex flex-wrap gap-2">
              <button onClick={handleBootstrap} disabled={saving || !activeProjectId} className="btn-base btn-light btn-m disabled:opacity-60">
                {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Sparkles className="w-4 h-4" />}
                从场景生成时间轴骨架
              </button>
              <button onClick={handleRenderTimeline} disabled={rendering || !activeProjectId} className="btn-base btn-dark btn-m disabled:opacity-60">
                {rendering ? <Loader2 className="w-4 h-4 animate-spin" /> : <Clapperboard className="w-4 h-4" />}
                导演合成出片
              </button>
            </div>

            {msg ? <div className="text-sm text-black/70">{msg}</div> : null}
          </div>

          <div className="rounded-2xl site-card p-4 space-y-3">
            <div className="text-sm font-medium text-black">最新导演版</div>
            {previewURL || timeline?.latest_export?.video_url ? (
              <div className="space-y-3">
                <video
                  controls
                  className="w-full rounded-xl border border-black/10 bg-black"
                  src={previewURL || timeline?.latest_export?.video_url}
                />
                <div className="text-xs text-black/55">
                  {timeline?.latest_export?.title || 'Director Cut'} · {timeline?.latest_export?.provider || 'local'}
                </div>
                <div className="grid grid-cols-1 md:grid-cols-[1fr_auto_auto] gap-2 items-center">
                  <select
                    value={publishPlatform}
                    onChange={(e) => setPublishPlatform(e.target.value as PlatformKind)}
                    className="factory-input w-full rounded-lg px-3 py-2 text-sm"
                  >
                    {(Object.keys(PLATFORM_LABEL) as PlatformKind[]).map((platform) => (
                      <option key={platform} value={platform}>{PLATFORM_LABEL[platform]}</option>
                    ))}
                  </select>
                  <button onClick={handleDirectorGateCheck} disabled={publishChecking || publishing} className="btn-base btn-light btn-m disabled:opacity-60">
                    {publishChecking ? <Loader2 className="w-4 h-4 animate-spin" /> : <Sparkles className="w-4 h-4" />}
                    门禁校验
                  </button>
                  <button onClick={handleDirectorPublish} disabled={publishing || publishChecking} className="btn-base btn-dark btn-m disabled:opacity-60">
                    {publishing ? <Loader2 className="w-4 h-4 animate-spin" /> : <Clapperboard className="w-4 h-4" />}
                    发布
                  </button>
                </div>
              </div>
            ) : (
              <EmptyBlock title="尚无导演版成片" hint="至少生成一个可用镜头后，再执行导演合成。" />
            )}
          </div>
        </div>
      </ConsoleSection>

      <ConsoleSection title="可执行时间轴" description="拖拽镜头卡可直接改轨道和起始时间，不再通过手工输入毫秒数。">
        {loading ? (
          <LoadingBlock text="正在加载时间轴..." />
        ) : shots.length === 0 ? (
          <EmptyBlock title="暂无镜头" hint="先从场景初始化，或在后续补一个镜头创建入口。" />
        ) : (
          <div className="space-y-4">
            <div className="rounded-2xl site-card p-4 overflow-x-auto">
              <div className="relative min-w-[720px] h-8 border-b border-black/10">
                {timelineRulerTicks.map((tick) => (
                  <div
                    key={tick}
                    className="absolute top-0 h-full"
                    style={{ left: `${(tick / timelineHorizonMs) * 100}%` }}
                  >
                    <div className="w-px h-3 bg-black/20" />
                    <div className="text-[10px] text-black/50 mt-1 -translate-x-1/2">{formatMs(tick)}</div>
                  </div>
                ))}
              </div>
            </div>
            {Object.entries(groupedTracks)
              .sort((a, b) => Number(a[0]) - Number(b[0]))
              .map(([track, items]) => (
                <div
                  key={track}
                  className={`rounded-2xl site-card p-4 transition ${draggingShotId ? 'border-black/20' : ''}`}
                  onDragOver={(event) => event.preventDefault()}
                  onDrop={(event) => handleShotDropToTrack(event, Number(track))}
                >
                  <div className="text-sm font-medium text-black mb-3">轨道 {track}</div>
                  <div className="relative min-h-44 rounded-xl border border-black/10 bg-[linear-gradient(to_right,rgba(0,0,0,0.06)_1px,transparent_1px)] bg-[size:24px_100%] overflow-x-auto">
                    <div className="relative min-w-[720px] h-44">
                    {items
                      .sort((a, b) => (a.timeline_start_ms || 0) - (b.timeline_start_ms || 0))
                      .map((shot) => (
                        <button
                          key={shot.id}
                          type="button"
                          draggable={!shot.locked}
                          onDragStart={() => handleShotDragStart(shot.id)}
                          onClick={() => setSelectedShotId(shot.id)}
                          className={`absolute top-4 text-left rounded-2xl border p-3 transition h-32 overflow-hidden ${selectedShotId === shot.id ? 'border-black bg-black/[0.03]' : 'border-black/10 bg-white hover:border-black/20'}`}
                          style={{
                            left: `${((shot.timeline_start_ms || 0) / timelineHorizonMs) * 100}%`,
                            width: `${Math.max((timelineDurationMs(shot) / timelineHorizonMs) * 100, 8)}%`,
                          }}
                        >
                          <div className="flex items-start justify-between gap-3">
                            <div>
                              <div className="text-sm font-medium text-black">#{shot.shot_number} {shot.title}</div>
                              <div className="text-xs text-black/55 mt-1">{formatMs(shot.timeline_start_ms)} 开始 · {formatMs(shot.timeline_duration_ms)} 时长</div>
                            </div>
                            <span className={`inline-flex rounded-full border px-2 py-1 text-[11px] ${statusTone(shot.clip_status || shot.status)}`}>
                              {shot.clip_status || shot.status}
                            </span>
                          </div>
                          <div className="mt-3 text-sm text-black/65 line-clamp-2">{shot.description || shot.prompt || '暂无镜头说明'}</div>
                          <div className="mt-3 flex flex-wrap gap-2 text-xs text-black/55">
                            <span>{shot.camera_language || '未设运镜'}</span>
                            <span>{shot.emotion_tone || '未设情绪'}</span>
                            <span>v{shot.version}</span>
                            {shot.locked ? <span>已锁定</span> : null}
                          </div>
                        </button>
                      ))}
                    </div>
                  </div>
                </div>
              ))}
          </div>
        )}
      </ConsoleSection>

      <ConsoleSection title="镜头导演面板" description="在这里微调单镜头，然后直接生成或局部重生，不必重跑整条视频。">
        {!selectedShot ? (
          <EmptyBlock title="未选中镜头" hint="从上方时间轴选一个镜头后，这里会切换到对应导演控制面板。" />
        ) : (
          <div className="grid grid-cols-1 xl:grid-cols-[1.2fr_0.8fr] gap-4">
            <form onSubmit={handleSaveShot} className="rounded-2xl site-card p-4 grid grid-cols-1 md:grid-cols-2 gap-4">
              <label className="text-sm text-black/65">
                镜头标题
                <input value={form.title} onChange={(e) => setForm((prev) => ({ ...prev, title: e.target.value }))} className="factory-input mt-1 w-full rounded-lg px-3 py-2" />
              </label>
              <label className="text-sm text-black/65">
                章节
                <input value={form.chapter} onChange={(e) => setForm((prev) => ({ ...prev, chapter: e.target.value }))} className="factory-input mt-1 w-full rounded-lg px-3 py-2" />
              </label>
              <label className="text-sm text-black/65">
                运镜语言
                <input value={form.camera_language} onChange={(e) => setForm((prev) => ({ ...prev, camera_language: e.target.value }))} className="factory-input mt-1 w-full rounded-lg px-3 py-2" />
              </label>
              <label className="text-sm text-black/65">
                情绪基调
                <input value={form.emotion_tone} onChange={(e) => setForm((prev) => ({ ...prev, emotion_tone: e.target.value }))} className="factory-input mt-1 w-full rounded-lg px-3 py-2" />
              </label>
              <label className="text-sm text-black/65">
                起始位置
                <div className="factory-input mt-1 w-full rounded-lg px-3 py-2 text-black/60 bg-black/[0.02]">
                  通过上方时间轴拖拽镜头卡调整（当前：{formatMs(selectedShot.timeline_start_ms)})
                </div>
              </label>
              <label className="text-sm text-black/65">
                时长（秒）
                <input type="number" min={1} max={60} value={form.duration} onChange={(e) => setForm((prev) => ({ ...prev, duration: Number(e.target.value) || 6 }))} className="factory-input mt-1 w-full rounded-lg px-3 py-2" />
              </label>
              <label className="text-sm text-black/65">
                画幅
                <select value={form.aspect_ratio} onChange={(e) => setForm((prev) => ({ ...prev, aspect_ratio: e.target.value as '16:9' | '9:16' }))} className="factory-input mt-1 w-full rounded-lg px-3 py-2">
                  <option value="16:9">16:9</option>
                  <option value="9:16">9:16</option>
                </select>
              </label>
              <label className="text-sm text-black/65">
                过渡方式
                <select value={form.transition_type} onChange={(e) => setForm((prev) => ({ ...prev, transition_type: e.target.value as 'cut' | 'fade' | 'wipe' | 'match' }))} className="factory-input mt-1 w-full rounded-lg px-3 py-2">
                  <option value="cut">cut</option>
                  <option value="fade">fade</option>
                  <option value="wipe">wipe</option>
                  <option value="match">match</option>
                </select>
              </label>
              <label className="text-sm text-black/65">
                导演风格模板
                <div className="mt-1 flex gap-2">
                  <select value={directorTemplateKey} onChange={(e) => setDirectorTemplateKey(e.target.value)} className="factory-input w-full rounded-lg px-3 py-2">
                    <option value="">选择导演模板</option>
                    {DIRECTOR_STYLE_TEMPLATES.map((item) => (
                      <option key={item.key} value={item.key}>{item.label}</option>
                    ))}
                  </select>
                  <button type="button" onClick={handleApplyDirectorTemplate} disabled={!directorTemplateKey || saving} className="btn-base btn-light btn-m disabled:opacity-60 whitespace-nowrap">
                    应用
                  </button>
                </div>
              </label>
              <label className="text-sm text-black/65 md:col-span-2">
                提示词
                <textarea value={form.prompt} onChange={(e) => setForm((prev) => ({ ...prev, prompt: e.target.value }))} className="factory-input mt-1 w-full rounded-lg px-3 py-2 min-h-28" />
              </label>

              <div className="md:col-span-2 flex flex-wrap gap-2">
                <button type="submit" disabled={saving} className="btn-base btn-light btn-m disabled:opacity-60">
                  {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Film className="w-4 h-4" />}
                  保存镜头参数
                </button>
                <MiniButton type="button" size="s" tone="neutral" icon={PlayCircle} iconSize={16} disabled={saving || selectedShot.locked} onClick={() => handleGenerate('generate')}>
                  生成镜头
                </MiniButton>
                <MiniButton type="button" size="s" tone="neutral" icon={RefreshCcw} iconSize={16} disabled={saving || selectedShot.locked} onClick={() => handleGenerate('regenerate')}>
                  局部重生
                </MiniButton>
                <MiniButton type="button" size="s" tone={selectedShot.locked ? 'danger' : 'light'} icon={selectedShot.locked ? Unlock : Lock} iconSize={16} disabled={saving} onClick={handleToggleLock}>
                  {selectedShot.locked ? '解除锁定' : '锁定镜头'}
                </MiniButton>
              </div>
            </form>

            <div className="space-y-4">
              <div className="rounded-2xl site-card p-4 space-y-3">
                <div className="text-sm font-medium text-black">镜头执行状态</div>
                <div className="flex flex-wrap gap-2 text-xs">
                  <span className={`inline-flex rounded-full border px-2 py-1 ${statusTone(selectedShot.clip_status || selectedShot.status)}`}>
                    {(selectedShot.clip_status || selectedShot.status).toUpperCase()}
                  </span>
                  <span className="inline-flex rounded-full border border-black/10 bg-black/[0.03] px-2 py-1 text-black/60">{selectedShot.clip_provider || '未指定 provider'}</span>
                  <span className="inline-flex rounded-full border border-black/10 bg-black/[0.03] px-2 py-1 text-black/60">评分 {selectedShot.clip_score?.toFixed(2) || '0.00'}</span>
                </div>
                <div className="text-sm text-black/65">{selectedShot.clip_notes || '暂无执行说明'}</div>
                {selectedShot.clip_video_url ? (
                  <video controls className="w-full rounded-xl border border-black/10 bg-black" src={selectedShot.clip_video_url} />
                ) : (
                  <EmptyBlock title="暂无镜头预览" hint="执行镜头生成后，这里会显示当前入轨版本。" />
                )}
              </div>

              <div className="rounded-2xl site-card p-4 space-y-3">
                <div className="text-sm font-medium text-black">版本分支</div>
                {activeVersionNode ? (
                  <div className="space-y-2">
                    {activeVersionNode.versions.map((version) => (
                      <div key={version.id} className="rounded-xl border border-black/10 bg-black/[0.02] px-3 py-3 text-sm text-black/65">
                        <div className="flex items-center justify-between gap-2">
                          <span>v{version.version} · {version.title}</span>
                          <MiniButton type="button" size="xs" tone="neutral" icon={GitBranchPlus} iconSize={14} onClick={() => setSelectedShotId(version.id)}>
                            打开
                          </MiniButton>
                        </div>
                        {version.version_note ? <div className="text-xs text-black/55 mt-1">{version.version_note}</div> : null}
                      </div>
                    ))}
                    {activeVersionNode.versions.length === 0 ? <div className="text-sm text-black/55">当前根镜头还没有衍生版本。</div> : null}
                  </div>
                ) : (
                  <EmptyBlock title="暂无衍生版本" hint="对当前镜头执行局部重生后，这里会形成版本分支。" />
                )}
              </div>
            </div>
          </div>
        )}
      </ConsoleSection>

      <ConsoleSection title="导演流程说明" description="这三个动作对应产品升级后的核心链路。">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
          <div className="rounded-2xl site-card p-4 text-black/65">
            <Sparkles className="w-5 h-5 text-black/70 mb-3" />
            脚本先被拆成镜头并入轨，时间轴字段决定真实执行顺序与合成位置。
          </div>
          <div className="rounded-2xl site-card p-4 text-black/65">
            <RefreshCcw className="w-5 h-5 text-black/70 mb-3" />
            问题镜头只做局部重生，不再重跑整条片，导演能锁定已满意镜头防止回退。
          </div>
          <div className="rounded-2xl site-card p-4 text-black/65">
            <Eye className="w-5 h-5 text-black/70 mb-3" />
            就绪镜头自动汇入导演版，形成最终预览与后续发布入口。
          </div>
        </div>
      </ConsoleSection>
    </FactoryConsoleShell>
  );
}