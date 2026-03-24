"use client";

import { FormEvent, useEffect, useMemo, useState } from 'react';
import { Bot, Copy, Loader2, Sparkles, Swords, Trash2, X } from 'lucide-react';
import { ConsoleSection, FactoryConsoleShell, MetricStrip } from '@/components/common/FactoryConsoleShell';
import { EmptyBlock, LoadingBlock } from '@/components/common/StateBlocks';
import { projectAPI, storyboardAPI } from '@/lib/api/client';
import type { DirectorABCompareResult, DirectorAutoStrategyResult, DirectorTemplate, Project } from '@/types';

function toNum(v: string, fallback = 0.2) {
  const n = Number(v);
  if (!Number.isFinite(n)) return fallback;
  return Math.max(0, Math.min(1, n));
}

export default function DirectorAgentsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectId, setProjectId] = useState(0);
  const [templates, setTemplates] = useState<DirectorTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [previewTemplate, setPreviewTemplate] = useState<DirectorTemplate | null>(null);
  const [nextTuneHint, setNextTuneHint] = useState<string | null>(null);

  const [newForm, setNewForm] = useState({
    name: '',
    slug: '',
    sample_frame_url: '',
    sample_video_url: '',
    prompt_prefix: '',
    camera_language: '',
    emotion_tone: '',
    transition_type: 'cut' as 'cut' | 'fade' | 'wipe' | 'match',
    transition_duration_ms: 220,
    genre_keywords: '',
    weight_narrative: '0.2',
    weight_visual: '0.2',
    weight_emotion: '0.2',
    weight_rhythm: '0.2',
    weight_continuity: '0.2',
  });

  const [strategyGenre, setStrategyGenre] = useState('');
  const [strategyTemplateID, setStrategyTemplateID] = useState<number>(0);
  const [strategyApply, setStrategyApply] = useState(true);
  const [strategyTune, setStrategyTune] = useState(70);
  const [strategyResult, setStrategyResult] = useState<DirectorAutoStrategyResult | null>(null);

  const [abA, setAbA] = useState<number>(0);
  const [abB, setAbB] = useState<number>(0);
  const [abGenre, setAbGenre] = useState('');
  const [abApplyBest, setAbApplyBest] = useState(false);
  const [abRenderBestCut, setAbRenderBestCut] = useState(false);
  const [abResult, setAbResult] = useState<DirectorABCompareResult | null>(null);
  const [editingTemplateID, setEditingTemplateID] = useState<number | null>(null);
  const [editForm, setEditForm] = useState({
    name: '',
    slug: '',
    sample_frame_url: '',
    sample_video_url: '',
    prompt_prefix: '',
    camera_language: '',
    emotion_tone: '',
    transition_type: 'cut' as 'cut' | 'fade' | 'wipe' | 'match',
    transition_duration_ms: 220,
    genre_keywords: '',
    weight_narrative: '0.2',
    weight_visual: '0.2',
    weight_emotion: '0.2',
    weight_rhythm: '0.2',
    weight_continuity: '0.2',
  });

  const buildTemplateForm = (tpl: DirectorTemplate) => ({
    name: tpl.name || '',
    slug: tpl.slug || '',
    sample_frame_url: tpl.sample_frame_url || '',
    sample_video_url: tpl.sample_video_url || '',
    prompt_prefix: tpl.prompt_prefix || '',
    camera_language: tpl.camera_language || '',
    emotion_tone: tpl.emotion_tone || '',
    transition_type: (tpl.transition_type || 'cut') as 'cut' | 'fade' | 'wipe' | 'match',
    transition_duration_ms: tpl.transition_duration_ms || 220,
    genre_keywords: tpl.genre_keywords || '',
    weight_narrative: String(tpl.weight_narrative ?? 0.2),
    weight_visual: String(tpl.weight_visual ?? 0.2),
    weight_emotion: String(tpl.weight_emotion ?? 0.2),
    weight_rhythm: String(tpl.weight_rhythm ?? 0.2),
    weight_continuity: String(tpl.weight_continuity ?? 0.2),
  });

  const loadTemplates = async (pid: number) => {
    if (!pid) {
      setTemplates([]);
      return;
    }
    const res = await storyboardAPI.listDirectorTemplates(pid);
    const list = res.data.data || [];
    setTemplates(list);
    if (list.length >= 2) {
      if (!abA) setAbA(list[0].id);
      if (!abB) setAbB(list[1].id);
    }
  };

  useEffect(() => {
    projectAPI.list()
      .then(async (res) => {
        const list = (res.data || []) as Project[];
        setProjects(list);
        if (list[0]?.id) {
          setProjectId(list[0].id);
          await loadTemplates(list[0].id);
        }
      })
      .catch(() => setMessage('加载导演模板失败'))
      .finally(() => setLoading(false));
  }, []);

  const builtInCount = useMemo(() => templates.filter((item) => item.is_builtin).length, [templates]);
  const customCount = Math.max(0, templates.length - builtInCount);

  useEffect(() => {
    if (!editingTemplateID) return;
    if (!templates.some((item) => item.id === editingTemplateID)) {
      setEditingTemplateID(null);
    }
  }, [templates, editingTemplateID]);

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    if (!projectId || !newForm.name.trim()) return;
    setSubmitting(true);
    setMessage(null);
    try {
      await storyboardAPI.createDirectorTemplate(projectId, {
        name: newForm.name.trim(),
        slug: newForm.slug.trim(),
        sample_frame_url: newForm.sample_frame_url.trim(),
        sample_video_url: newForm.sample_video_url.trim(),
        prompt_prefix: newForm.prompt_prefix.trim(),
        camera_language: newForm.camera_language.trim(),
        emotion_tone: newForm.emotion_tone.trim(),
        transition_type: newForm.transition_type,
        transition_duration_ms: newForm.transition_duration_ms,
        genre_keywords: newForm.genre_keywords.trim(),
        weight_narrative: toNum(newForm.weight_narrative),
        weight_visual: toNum(newForm.weight_visual),
        weight_emotion: toNum(newForm.weight_emotion),
        weight_rhythm: toNum(newForm.weight_rhythm),
        weight_continuity: toNum(newForm.weight_continuity),
      });
      await loadTemplates(projectId);
      setMessage('导演模板已创建');
      setNewForm((prev) => ({ ...prev, name: '', slug: '', sample_frame_url: '', sample_video_url: '', prompt_prefix: '', camera_language: '', emotion_tone: '', genre_keywords: '' }));
    } catch (error: any) {
      setMessage(error?.response?.data?.error || '创建模板失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (template: DirectorTemplate) => {
    if (!projectId) return;
    if (!window.confirm(`确认删除模板 ${template.name} ?`)) return;
    setSubmitting(true);
    setMessage(null);
    try {
      await storyboardAPI.deleteDirectorTemplate(projectId, template.id);
      await loadTemplates(projectId);
      setMessage('模板已删除');
    } catch (error: any) {
      setMessage(error?.response?.data?.error || '删除模板失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDuplicate = (template: DirectorTemplate) => {
    setNewForm({
      name: `${template.name} 副本`,
      slug: `${template.slug}-copy`,
      sample_frame_url: template.sample_frame_url || '',
      sample_video_url: template.sample_video_url || '',
      prompt_prefix: template.prompt_prefix || '',
      camera_language: template.camera_language || '',
      emotion_tone: template.emotion_tone || '',
      transition_type: (template.transition_type || 'cut') as 'cut' | 'fade' | 'wipe' | 'match',
      transition_duration_ms: template.transition_duration_ms || 220,
      genre_keywords: template.genre_keywords || '',
      weight_narrative: String(template.weight_narrative ?? 0.2),
      weight_visual: String(template.weight_visual ?? 0.2),
      weight_emotion: String(template.weight_emotion ?? 0.2),
      weight_rhythm: String(template.weight_rhythm ?? 0.2),
      weight_continuity: String(template.weight_continuity ?? 0.2),
    });
    setMessage(`已将模板 ${template.name} 复制到“新增模板”表单，可直接修改后创建`);
  };

  const handleStartEdit = (template: DirectorTemplate) => {
    setEditingTemplateID(template.id);
    setEditForm(buildTemplateForm(template));
    setMessage(null);
  };

  const handleCancelEdit = () => {
    setEditingTemplateID(null);
    setMessage(null);
  };

  const handleSaveEdit = async (e: FormEvent) => {
    e.preventDefault();
    if (!projectId || !editingTemplateID || !editForm.name.trim()) return;
    setSubmitting(true);
    setMessage(null);
    try {
      await storyboardAPI.updateDirectorTemplate(projectId, editingTemplateID, {
        name: editForm.name.trim(),
        slug: editForm.slug.trim(),
        sample_frame_url: editForm.sample_frame_url.trim(),
        sample_video_url: editForm.sample_video_url.trim(),
        prompt_prefix: editForm.prompt_prefix.trim(),
        camera_language: editForm.camera_language.trim(),
        emotion_tone: editForm.emotion_tone.trim(),
        transition_type: editForm.transition_type,
        transition_duration_ms: editForm.transition_duration_ms,
        genre_keywords: editForm.genre_keywords.trim(),
        weight_narrative: toNum(editForm.weight_narrative),
        weight_visual: toNum(editForm.weight_visual),
        weight_emotion: toNum(editForm.weight_emotion),
        weight_rhythm: toNum(editForm.weight_rhythm),
        weight_continuity: toNum(editForm.weight_continuity),
      });
      await loadTemplates(projectId);
      setEditingTemplateID(null);
      setMessage('模板已保存');
    } catch (error: any) {
      setMessage(error?.response?.data?.error || '保存模板失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleSuggestNextRound = () => {
    if (!abResult) return;
    const diff = Math.abs(abResult.score_a - abResult.score_b);
    const winnerScore = Math.max(abResult.score_a, abResult.score_b);
    let nextTune = strategyTune;
    if (diff < 0.02) {
      nextTune += 12;
    } else if (diff < 0.05) {
      nextTune += 8;
    } else {
      nextTune += 4;
    }
    if (winnerScore < 0.78) {
      nextTune += 6;
    }
    nextTune = Math.max(10, Math.min(100, nextTune));

    setStrategyTemplateID(abResult.winner_template_id);
    setStrategyTune(nextTune);
    if (!strategyGenre && abResult.genre) {
      setStrategyGenre(abResult.genre);
    }
    setStrategyApply(false);
    setNextTuneHint(`建议下一轮以“${abResult.winner_template}”为基线，tune_percent 调整为 ${nextTune}，先预估再决定是否应用。`);
    setMessage('已生成下一轮自动调参建议，并同步到自动策略器参数。');
  };

  const handleAutoStrategy = async () => {
    if (!projectId) return;
    setSubmitting(true);
    setMessage(null);
    try {
      const resp = await storyboardAPI.autoDirectorStrategy(projectId, {
        genre: strategyGenre.trim(),
        template_id: strategyTemplateID || undefined,
        apply: strategyApply,
        tune_percent: strategyTune,
      });
      setStrategyResult(resp.data);
      setMessage(`自动策略完成：${resp.data.selected.name}`);
    } catch (error: any) {
      setMessage(error?.response?.data?.error || '自动策略失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleABCompare = async () => {
    if (!projectId || !abA || !abB || abA === abB) {
      setMessage('请先选择两个不同模板');
      return;
    }
    setSubmitting(true);
    setMessage(null);
    try {
      const resp = await storyboardAPI.compareDirectorAB(projectId, {
        template_a_id: abA,
        template_b_id: abB,
        genre: abGenre.trim(),
        apply_best: abApplyBest,
        tune_percent: strategyTune,
        render_best_cut: abRenderBestCut,
      });
      setAbResult(resp.data);
      setMessage(`A/B 完成，胜出：${resp.data.winner_template}`);
    } catch (error: any) {
      setMessage(error?.response?.data?.error || 'A/B 对比失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <FactoryConsoleShell
      eyebrow="Director Agent"
      title="导演 Agent 中心"
      subtitle="把导演风格从固定枚举升级为可配置模板，再交给自动策略器和 A/B 双导演并行选优。"
    >
      <MetricStrip
        items={[
          { label: '模板总数', value: String(templates.length), hint: '当前项目导演风格模板池。' },
          { label: '内置模板', value: String(builtInCount), hint: '系统预置导演模板。' },
          { label: '自定义模板', value: String(customCount), hint: '你新增并可持续调优的模板。' },
        ]}
      />

      <ConsoleSection title="项目与模板池" description="先切项目，再管理导演模板参数权重。">
        <div className="rounded-2xl site-card p-4 space-y-4">
          <label className="text-sm text-black/65 block">
            项目
            <select
              value={projectId}
              onChange={async (e) => {
                const next = Number(e.target.value) || 0;
                setProjectId(next);
                setLoading(true);
                try {
                  await loadTemplates(next);
                } finally {
                  setLoading(false);
                }
              }}
              className="factory-input mt-1 w-full rounded-lg px-3 py-2"
            >
              {projects.length === 0 ? <option value={0}>暂无项目</option> : projects.map((item) => (
                <option key={item.id} value={item.id}>{item.title || `项目 #${item.id}`}</option>
              ))}
            </select>
          </label>

          {loading ? <LoadingBlock text="正在加载模板..." /> : templates.length === 0 ? <EmptyBlock title="暂无模板" hint="创建一个导演模板后可启用自动策略与 A/B。" /> : (
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
              {templates.map((item) => (
                <div key={item.id} className="rounded-xl border border-black/10 bg-white p-4">
                  {item.sample_frame_url ? (
                    <button
                      type="button"
                      onClick={() => setPreviewTemplate(item)}
                      className="block w-full mb-3 rounded-lg overflow-hidden border border-black/10 bg-black/[0.03]"
                    >
                      <img src={item.sample_frame_url} alt={`${item.name} 示例帧`} className="w-full h-32 object-cover" />
                    </button>
                  ) : (
                    <button
                      type="button"
                      onClick={() => setPreviewTemplate(item)}
                      className="block w-full mb-3 rounded-lg h-32 border border-dashed border-black/15 bg-black/[0.02] text-xs text-black/45"
                    >
                      暂无示例帧，点击预览模板详情
                    </button>
                  )}
                  <div className="flex items-center justify-between gap-2 mb-2">
                    <div className="font-medium text-black">{item.name}</div>
                    {item.is_builtin ? <span className="text-[11px] text-black/50">内置</span> : null}
                  </div>
                  <div className="text-xs text-black/60">{item.slug}</div>
                  <div className="text-xs text-black/60 mt-1">{item.camera_language || '未配置运镜'}</div>
                  <div className="text-xs text-black/60">{item.emotion_tone || '未配置情绪'}</div>
                  <div className="text-xs text-black/60 mt-2">N:{item.weight_narrative.toFixed(2)} V:{item.weight_visual.toFixed(2)} E:{item.weight_emotion.toFixed(2)}</div>
                  <div className="text-xs text-black/60">R:{item.weight_rhythm.toFixed(2)} C:{item.weight_continuity.toFixed(2)}</div>
                  <div className="mt-3 flex flex-wrap gap-2">
                    <button onClick={() => handleStartEdit(item)} disabled={submitting} className="btn-base btn-light btn-m disabled:opacity-60">
                      编辑
                    </button>
                    <button onClick={() => setPreviewTemplate(item)} disabled={submitting} className="btn-base btn-light btn-m disabled:opacity-60">
                      一键预览
                    </button>
                    <button onClick={() => handleDuplicate(item)} disabled={submitting} className="btn-base btn-light btn-m disabled:opacity-60">
                      <Copy className="w-4 h-4" /> 复制
                    </button>
                    {!item.is_builtin ? (
                      <button onClick={() => handleDelete(item)} disabled={submitting} className="btn-base btn-light btn-m disabled:opacity-60">
                        <Trash2 className="w-4 h-4" /> 删除
                      </button>
                    ) : null}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </ConsoleSection>

      <ConsoleSection title="新增模板" description="可配置风格参数与五维权重（叙事/视觉/情绪/节奏/连贯）。">
        <form onSubmit={handleCreate} className="rounded-2xl site-card p-4 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
          <input value={newForm.name} onChange={(e) => setNewForm((p) => ({ ...p, name: e.target.value }))} placeholder="模板名称" className="factory-input rounded-lg px-3 py-2" />
          <input value={newForm.slug} onChange={(e) => setNewForm((p) => ({ ...p, slug: e.target.value }))} placeholder="slug" className="factory-input rounded-lg px-3 py-2" />
          <input value={newForm.sample_frame_url} onChange={(e) => setNewForm((p) => ({ ...p, sample_frame_url: e.target.value }))} placeholder="示例帧 URL" className="factory-input rounded-lg px-3 py-2" />
          <input value={newForm.sample_video_url} onChange={(e) => setNewForm((p) => ({ ...p, sample_video_url: e.target.value }))} placeholder="示例视频 URL" className="factory-input rounded-lg px-3 py-2" />
          <input value={newForm.genre_keywords} onChange={(e) => setNewForm((p) => ({ ...p, genre_keywords: e.target.value }))} placeholder="题材关键词（逗号分隔）" className="factory-input rounded-lg px-3 py-2" />
          <input value={newForm.camera_language} onChange={(e) => setNewForm((p) => ({ ...p, camera_language: e.target.value }))} placeholder="运镜语言" className="factory-input rounded-lg px-3 py-2" />
          <input value={newForm.emotion_tone} onChange={(e) => setNewForm((p) => ({ ...p, emotion_tone: e.target.value }))} placeholder="情绪基调" className="factory-input rounded-lg px-3 py-2" />
          <select value={newForm.transition_type} onChange={(e) => setNewForm((p) => ({ ...p, transition_type: e.target.value as 'cut' | 'fade' | 'wipe' | 'match' }))} className="factory-input rounded-lg px-3 py-2">
            <option value="cut">cut</option>
            <option value="fade">fade</option>
            <option value="wipe">wipe</option>
            <option value="match">match</option>
          </select>
          <input value={newForm.prompt_prefix} onChange={(e) => setNewForm((p) => ({ ...p, prompt_prefix: e.target.value }))} placeholder="提示词前缀" className="factory-input rounded-lg px-3 py-2 md:col-span-2 xl:col-span-3" />
          <input value={newForm.weight_narrative} onChange={(e) => setNewForm((p) => ({ ...p, weight_narrative: e.target.value }))} placeholder="叙事权重" className="factory-input rounded-lg px-3 py-2" />
          <input value={newForm.weight_visual} onChange={(e) => setNewForm((p) => ({ ...p, weight_visual: e.target.value }))} placeholder="视觉权重" className="factory-input rounded-lg px-3 py-2" />
          <input value={newForm.weight_emotion} onChange={(e) => setNewForm((p) => ({ ...p, weight_emotion: e.target.value }))} placeholder="情绪权重" className="factory-input rounded-lg px-3 py-2" />
          <input value={newForm.weight_rhythm} onChange={(e) => setNewForm((p) => ({ ...p, weight_rhythm: e.target.value }))} placeholder="节奏权重" className="factory-input rounded-lg px-3 py-2" />
          <input value={newForm.weight_continuity} onChange={(e) => setNewForm((p) => ({ ...p, weight_continuity: e.target.value }))} placeholder="连贯权重" className="factory-input rounded-lg px-3 py-2" />
          <button type="submit" disabled={submitting || !projectId} className="btn-base btn-dark btn-m disabled:opacity-60">
            {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Sparkles className="w-4 h-4" />}
            创建模板
          </button>
        </form>
      </ConsoleSection>

      <ConsoleSection title="自动导演策略器" description="按题材自动选导演模板，并可直接批量应用到时间轴。">
        <div className="rounded-2xl site-card p-4 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-3 items-end">
          <input value={strategyGenre} onChange={(e) => setStrategyGenre(e.target.value)} placeholder="题材（如 历史/都市/喜剧）" className="factory-input rounded-lg px-3 py-2" />
          <select value={strategyTemplateID} onChange={(e) => setStrategyTemplateID(Number(e.target.value) || 0)} className="factory-input rounded-lg px-3 py-2">
            <option value={0}>自动选择模板</option>
            {templates.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </select>
          <input type="number" min={10} max={100} value={strategyTune} onChange={(e) => setStrategyTune(Number(e.target.value) || 70)} className="factory-input rounded-lg px-3 py-2" placeholder="微调强度" />
          <label className="inline-flex items-center gap-2 text-sm text-black/70"><input type="checkbox" checked={strategyApply} onChange={(e) => setStrategyApply(e.target.checked)} />应用到镜头</label>
          <button onClick={handleAutoStrategy} disabled={submitting || !projectId} className="btn-base btn-light btn-m disabled:opacity-60 md:col-span-2 xl:col-span-1">
            {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Bot className="w-4 h-4" />}
            自动策略
          </button>
        </div>
        {strategyResult ? (
          <div className="mt-3 rounded-xl border border-black/10 bg-white p-4 text-sm text-black/70">
            选择模板：{strategyResult.selected.name} · 预测分：{strategyResult.predicted_score.toFixed(3)} · 更新镜头：{strategyResult.shot_updates}
          </div>
        ) : null}
      </ConsoleSection>

      <ConsoleSection title="A/B 双导演并行对比" description="同剧本并行打分两种导演风格，自动选优并可一键应用最佳方案。">
        <div className="rounded-2xl site-card p-4 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-3 items-end">
          <select value={abA} onChange={(e) => setAbA(Number(e.target.value) || 0)} className="factory-input rounded-lg px-3 py-2">
            <option value={0}>选择 A 导演模板</option>
            {templates.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </select>
          <select value={abB} onChange={(e) => setAbB(Number(e.target.value) || 0)} className="factory-input rounded-lg px-3 py-2">
            <option value={0}>选择 B 导演模板</option>
            {templates.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </select>
          <input value={abGenre} onChange={(e) => setAbGenre(e.target.value)} placeholder="题材（可选）" className="factory-input rounded-lg px-3 py-2" />
          <label className="inline-flex items-center gap-2 text-sm text-black/70"><input type="checkbox" checked={abApplyBest} onChange={(e) => setAbApplyBest(e.target.checked)} />应用最佳模板</label>
          <label className="inline-flex items-center gap-2 text-sm text-black/70 md:col-span-2"><input type="checkbox" checked={abRenderBestCut} onChange={(e) => setAbRenderBestCut(e.target.checked)} />应用后自动渲染最佳导演版</label>
          <button onClick={handleABCompare} disabled={submitting || !projectId} className="btn-base btn-dark btn-m disabled:opacity-60 xl:col-span-1">
            {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Swords className="w-4 h-4" />}
            开始 A/B 对比
          </button>
        </div>
        {abResult ? (
          <div className="mt-3 rounded-xl border border-black/10 bg-white p-4 text-sm text-black/70">
            A: {abResult.template_a.name} ({abResult.score_a.toFixed(3)}) · B: {abResult.template_b.name} ({abResult.score_b.toFixed(3)}) · 胜出：{abResult.winner_template}
            {abResult.rendered_export_id ? ` · 导出ID: ${abResult.rendered_export_id}` : ''}
            <div className="mt-3 flex flex-wrap gap-2">
              <button type="button" onClick={handleSuggestNextRound} className="btn-base btn-light btn-m">
                下一轮自动调参建议
              </button>
              {nextTuneHint ? <span className="text-xs text-black/60">{nextTuneHint}</span> : null}
            </div>
          </div>
        ) : null}
      </ConsoleSection>

      {message ? <div className="text-sm text-black/70">{message}</div> : null}

      {editingTemplateID ? (
        <div className="fixed inset-0 z-50 bg-black/45 backdrop-blur-[1px] px-3 py-6 overflow-y-auto" onClick={handleCancelEdit}>
          <div className="max-w-5xl mx-auto" onClick={(e) => e.stopPropagation()}>
            <form onSubmit={handleSaveEdit} className="rounded-2xl site-card p-4 md:p-5 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
              <div className="md:col-span-2 xl:col-span-3 flex items-center justify-between gap-3 mb-1">
                <div>
                  <div className="text-sm text-black/50">模板编辑</div>
                  <div className="text-lg font-semibold text-black">编辑导演模板并保存</div>
                </div>
                <button type="button" onClick={handleCancelEdit} className="btn-base btn-light btn-m">
                  <X className="w-4 h-4" /> 关闭
                </button>
              </div>

              <input value={editForm.name} onChange={(e) => setEditForm((p) => ({ ...p, name: e.target.value }))} placeholder="模板名称" className="factory-input rounded-lg px-3 py-2" />
              <input value={editForm.slug} onChange={(e) => setEditForm((p) => ({ ...p, slug: e.target.value }))} placeholder="slug" className="factory-input rounded-lg px-3 py-2" />
              <input value={editForm.sample_frame_url} onChange={(e) => setEditForm((p) => ({ ...p, sample_frame_url: e.target.value }))} placeholder="示例帧 URL" className="factory-input rounded-lg px-3 py-2" />
              <input value={editForm.sample_video_url} onChange={(e) => setEditForm((p) => ({ ...p, sample_video_url: e.target.value }))} placeholder="示例视频 URL" className="factory-input rounded-lg px-3 py-2" />
              <input value={editForm.genre_keywords} onChange={(e) => setEditForm((p) => ({ ...p, genre_keywords: e.target.value }))} placeholder="题材关键词（逗号分隔）" className="factory-input rounded-lg px-3 py-2" />
              <input value={editForm.camera_language} onChange={(e) => setEditForm((p) => ({ ...p, camera_language: e.target.value }))} placeholder="运镜语言" className="factory-input rounded-lg px-3 py-2" />
              <input value={editForm.emotion_tone} onChange={(e) => setEditForm((p) => ({ ...p, emotion_tone: e.target.value }))} placeholder="情绪基调" className="factory-input rounded-lg px-3 py-2" />
              <select value={editForm.transition_type} onChange={(e) => setEditForm((p) => ({ ...p, transition_type: e.target.value as 'cut' | 'fade' | 'wipe' | 'match' }))} className="factory-input rounded-lg px-3 py-2">
                <option value="cut">cut</option>
                <option value="fade">fade</option>
                <option value="wipe">wipe</option>
                <option value="match">match</option>
              </select>
              <input value={editForm.prompt_prefix} onChange={(e) => setEditForm((p) => ({ ...p, prompt_prefix: e.target.value }))} placeholder="提示词前缀" className="factory-input rounded-lg px-3 py-2 md:col-span-2 xl:col-span-3" />
              <input value={editForm.weight_narrative} onChange={(e) => setEditForm((p) => ({ ...p, weight_narrative: e.target.value }))} placeholder="叙事权重" className="factory-input rounded-lg px-3 py-2" />
              <input value={editForm.weight_visual} onChange={(e) => setEditForm((p) => ({ ...p, weight_visual: e.target.value }))} placeholder="视觉权重" className="factory-input rounded-lg px-3 py-2" />
              <input value={editForm.weight_emotion} onChange={(e) => setEditForm((p) => ({ ...p, weight_emotion: e.target.value }))} placeholder="情绪权重" className="factory-input rounded-lg px-3 py-2" />
              <input value={editForm.weight_rhythm} onChange={(e) => setEditForm((p) => ({ ...p, weight_rhythm: e.target.value }))} placeholder="节奏权重" className="factory-input rounded-lg px-3 py-2" />
              <input value={editForm.weight_continuity} onChange={(e) => setEditForm((p) => ({ ...p, weight_continuity: e.target.value }))} placeholder="连贯权重" className="factory-input rounded-lg px-3 py-2" />

              <div className="flex flex-wrap gap-2 md:col-span-2 xl:col-span-3">
                <button type="submit" disabled={submitting || !projectId} className="btn-base btn-dark btn-m disabled:opacity-60">
                  {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Sparkles className="w-4 h-4" />}
                  保存模板
                </button>
                <button type="button" onClick={handleCancelEdit} disabled={submitting} className="btn-base btn-light btn-m disabled:opacity-60">
                  取消
                </button>
              </div>
            </form>
          </div>
        </div>
      ) : null}

      {previewTemplate ? (
        <div className="fixed inset-0 z-50 bg-black/60 px-3 py-6 overflow-y-auto" onClick={() => setPreviewTemplate(null)}>
          <div className="max-w-3xl mx-auto" onClick={(e) => e.stopPropagation()}>
            <div className="rounded-2xl site-card p-4 md:p-5">
              <div className="flex items-center justify-between gap-3 mb-3">
                <div>
                  <div className="text-sm text-black/50">模板预览</div>
                  <div className="text-lg font-semibold text-black">{previewTemplate.name}</div>
                </div>
                <button type="button" onClick={() => setPreviewTemplate(null)} className="btn-base btn-light btn-m">
                  <X className="w-4 h-4" /> 关闭
                </button>
              </div>

              {previewTemplate.sample_video_url ? (
                <video controls className="w-full rounded-xl border border-black/10 bg-black" src={previewTemplate.sample_video_url} />
              ) : previewTemplate.sample_frame_url ? (
                <img src={previewTemplate.sample_frame_url} alt={`${previewTemplate.name} 示例帧预览`} className="w-full rounded-xl border border-black/10 object-cover" />
              ) : (
                <EmptyBlock title="暂无示例媒体" hint="请在模板中填写示例帧或示例视频 URL。" />
              )}

              <div className="mt-3 text-xs text-black/60">{previewTemplate.camera_language || '未配置运镜'} · {previewTemplate.emotion_tone || '未配置情绪'} · {previewTemplate.genre_keywords || '未配置题材'}</div>
            </div>
          </div>
        </div>
      ) : null}
    </FactoryConsoleShell>
  );
}
