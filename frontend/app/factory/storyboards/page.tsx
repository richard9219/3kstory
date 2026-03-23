"use client";

import { FormEvent, useEffect, useMemo, useState } from 'react';
import { CheckSquare, Clock3, Columns3, Eye, GitBranchPlus, ListTodo, Loader2, PlusCircle, SplitSquareVertical } from 'lucide-react';
import { ConsoleSection, FactoryConsoleShell, MetricStrip } from '@/components/common/FactoryConsoleShell';
import { EmptyBlock, LoadingBlock } from '@/components/common/StateBlocks';
import MiniButton from '@/components/common/MiniButton';
import { projectAPI, storyboardAPI } from '@/lib/api/client';
import type { Project, StoryboardShot, StoryboardVersionNode } from '@/types';

const boardColumns = [
  {
    title: '待拆分章节',
    items: ['第一章：隆中', '第二章：三分天下', '第三章：赤壁与入蜀'],
  },
  {
    title: '待生成镜头',
    items: ['镜头 01：草庐夜读', '镜头 02：扇柄开机', '镜头 03：荆州布防推演'],
  },
  {
    title: '待审核版本',
    items: ['V2：羽扇近景', 'V3：江面火光', 'V1：五丈原静物收尾'],
  },
  {
    title: '已完成镜头',
    items: ['地图沙盘', '三顾门前', '军帐灯影'],
  },
];

export default function StoryboardsCenterPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [activeProjectId, setActiveProjectId] = useState<number>(0);
  const [shots, setShots] = useState<StoryboardShot[]>([]);
  const [versionTree, setVersionTree] = useState<StoryboardVersionNode[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);
  const [draggingID, setDraggingID] = useState<number | null>(null);
  const [form, setForm] = useState({ title: '', chapter: '', duration: 6, cameraLanguage: '', prompt: '' });

  const reloadShots = async (projectId: number) => {
    if (!projectId) {
      setShots([]);
      setVersionTree([]);
      return;
    }
    const [res, treeRes] = await Promise.all([
      storyboardAPI.listProjectShots(projectId),
      storyboardAPI.getVersionTree(projectId),
    ]);
    setShots(res.data.data || []);
    setVersionTree(treeRes.data.data || []);
  };

  useEffect(() => {
    projectAPI.list()
      .then((res) => {
        const list = (res.data || []) as Project[];
        setProjects(list);
        if (list[0]?.id) {
          setActiveProjectId(list[0].id);
          return reloadShots(list[0].id);
        }
      })
      .catch(() => setMsg('加载项目或分镜失败'))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (activeProjectId) {
      reloadShots(activeProjectId).catch(() => setMsg('加载项目镜头失败'));
    }
  }, [activeProjectId]);

  const grouped = useMemo(() => {
    const map: Record<string, StoryboardShot[]> = {};
    shots.forEach((shot) => {
      const key = shot.status || 'draft';
      if (!map[key]) map[key] = [];
      map[key].push(shot);
    });
    return map;
  }, [shots]);

  const handleBootstrap = async () => {
    if (!activeProjectId) return;
    setSaving(true);
    setMsg(null);
    try {
      const resp = await storyboardAPI.bootstrapFromScenes(activeProjectId);
      setMsg(`已初始化 ${resp.data.bootstrapped || 0} 条镜头`);
      await reloadShots(activeProjectId);
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '初始化失败');
    } finally {
      setSaving(false);
    }
  };

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    if (!activeProjectId || !form.title.trim()) return;
    setSaving(true);
    setMsg(null);
    try {
      await storyboardAPI.createShot(activeProjectId, {
        title: form.title.trim(),
        chapter: form.chapter.trim(),
        duration: form.duration,
        camera_language: form.cameraLanguage.trim(),
        prompt: form.prompt.trim(),
      });
      setForm({ title: '', chapter: '', duration: 6, cameraLanguage: '', prompt: '' });
      await reloadShots(activeProjectId);
      setMsg('镜头创建成功');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '创建镜头失败');
    } finally {
      setSaving(false);
    }
  };

  const handleCreateVersion = async (sourceShotID: number) => {
    if (!activeProjectId) return;
    const note = window.prompt('版本备注（可选）', '细化运镜与角色表情');
    setSaving(true);
    setMsg(null);
    try {
      await storyboardAPI.createShotVersion(activeProjectId, {
        source_shot_id: sourceShotID,
        version_note: note || '',
      });
      await reloadShots(activeProjectId);
      setMsg('已创建新镜头版本');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '创建版本失败');
    } finally {
      setSaving(false);
    }
  };

  const handleDropTo = async (targetID: number) => {
    if (!activeProjectId || !draggingID || draggingID === targetID) return;
    const ordered = [...shots];
    const fromIndex = ordered.findIndex((item) => item.id === draggingID);
    const toIndex = ordered.findIndex((item) => item.id === targetID);
    if (fromIndex < 0 || toIndex < 0) return;

    const [moved] = ordered.splice(fromIndex, 1);
    ordered.splice(toIndex, 0, moved);

    setSaving(true);
    try {
      await storyboardAPI.reorderShots(activeProjectId, ordered.map((item) => item.id));
      await reloadShots(activeProjectId);
      setMsg('镜头顺序已更新');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '拖拽排序失败');
    } finally {
      setSaving(false);
      setDraggingID(null);
    }
  };

  return (
    <FactoryConsoleShell
      eyebrow="Storyboard Center"
      title="分镜中心"
      subtitle="短剧和电影阶段不能再只靠文本表单，需要一个把章节、镜头、版本和审核串起来的分镜中心。当前先搭建信息架构骨架。"
    >
      <MetricStrip
        items={[
          { label: '镜头总数', value: String(shots.length), hint: '当前项目下已入库镜头数量。' },
          { label: '草稿镜头', value: String((grouped.draft || []).length), hint: '可继续编辑和补充提示词。' },
          { label: '已完成镜头', value: String((grouped.completed || []).length), hint: '完成渲染并可进入剪辑的镜头数量。' },
        ]}
      />

      <ConsoleSection title="项目与镜头初始化" description="先选择项目，再从场景自动生成镜头骨架，避免手工重复录入。">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 items-end">
          <label className="text-sm text-black/65">
            目标项目
            <select
              value={activeProjectId}
              onChange={(e) => setActiveProjectId(Number(e.target.value) || 0)}
              className="factory-input mt-1 w-full rounded-lg px-3 py-2"
            >
              {projects.length === 0 ? <option value={0}>暂无项目</option> : projects.map((p) => (
                <option key={p.id} value={p.id}>{p.title || `项目 #${p.id}`}</option>
              ))}
            </select>
          </label>
          <button
            onClick={handleBootstrap}
            disabled={saving || !activeProjectId}
            className="btn-base btn-dark btn-m disabled:opacity-60"
          >
            {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <ListTodo className="w-4 h-4" />}
            从场景初始化镜头
          </button>
        </div>
        {msg && <div className="mt-3 text-sm text-black/70">{msg}</div>}
      </ConsoleSection>

      <ConsoleSection title="新增镜头" description="镜头数据结构已入库，可逐步替换为拖拽式分镜编辑器。">
        <form onSubmit={handleCreate} className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <label className="text-sm text-black/65">
            镜头标题
            <input value={form.title} onChange={(e) => setForm((prev) => ({ ...prev, title: e.target.value }))} className="factory-input mt-1 w-full rounded-lg px-3 py-2" />
          </label>
          <label className="text-sm text-black/65">
            所属章节
            <input value={form.chapter} onChange={(e) => setForm((prev) => ({ ...prev, chapter: e.target.value }))} className="factory-input mt-1 w-full rounded-lg px-3 py-2" placeholder="例如：第一章 隆中" />
          </label>
          <label className="text-sm text-black/65">
            镜头时长（秒）
            <input type="number" min={1} max={60} value={form.duration} onChange={(e) => setForm((prev) => ({ ...prev, duration: Number(e.target.value) || 6 }))} className="factory-input mt-1 w-full rounded-lg px-3 py-2" />
          </label>
          <label className="text-sm text-black/65">
            运镜语言
            <input value={form.cameraLanguage} onChange={(e) => setForm((prev) => ({ ...prev, cameraLanguage: e.target.value }))} className="factory-input mt-1 w-full rounded-lg px-3 py-2" placeholder="例如：慢推近景" />
          </label>
          <label className="text-sm text-black/65 md:col-span-2">
            镜头提示词
            <textarea value={form.prompt} onChange={(e) => setForm((prev) => ({ ...prev, prompt: e.target.value }))} className="factory-input mt-1 w-full rounded-lg px-3 py-2 min-h-24" />
          </label>
          <button type="submit" disabled={saving || !activeProjectId} className="btn-base btn-light btn-m disabled:opacity-60">
            {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <PlusCircle className="w-4 h-4" />}
            创建镜头
          </button>
        </form>
      </ConsoleSection>

      <ConsoleSection title="分镜看板" description="这个页面先给出未来工作区的基础布局，后续可以接真实镜头数据和拖拽排序。">
        {loading ? (
          <LoadingBlock text="正在加载项目镜头..." />
        ) : (
        <div className="grid grid-cols-1 xl:grid-cols-4 gap-4">
          {boardColumns.map((column) => {
            const mappedStatus = column.title.includes('待生成') ? 'pending' : column.title.includes('待审核') ? 'processing' : column.title.includes('已完成') ? 'completed' : 'draft';
            const items = (grouped[mappedStatus] || []).slice(0, 8);
            return (
            <div key={column.title} className="rounded-2xl site-card p-4">
              <div className="flex items-center gap-2 text-black font-medium mb-3">
                <Columns3 className="w-4 h-4 text-black/70" />
                {column.title}
              </div>
              <div className="space-y-2">
                {items.length === 0 ? <div className="text-xs text-black/45">暂无镜头</div> : items.map((item) => (
                  <div
                    key={item.id}
                    draggable
                    onDragStart={() => setDraggingID(item.id)}
                    onDragOver={(e) => e.preventDefault()}
                    onDrop={() => handleDropTo(item.id)}
                    className="rounded-xl bg-black/[0.02] border border-black/10 px-3 py-3 text-sm text-black/65"
                  >
                    <div className="text-black">#{item.shot_number} {item.title}</div>
                    <div className="text-xs text-black/55 mt-1">{item.chapter || '未分章'} · {item.duration}s · v{item.version}</div>
                    <div className="mt-2 flex gap-2">
                      <MiniButton
                        type="button"
                        size="xs"
                        tone="neutral"
                        icon={GitBranchPlus}
                        iconSize={14}
                        disabled={saving}
                        onClick={() => handleCreateVersion(item.id)}
                      >
                        新建版本
                      </MiniButton>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )})}
        </div>
        )}
      </ConsoleSection>

      <ConsoleSection title="镜头版本树" description="按根镜头聚合全部版本，支持短剧/电影反复打磨到可用版本。">
        <div className="space-y-3">
          {versionTree.length === 0 ? (
            <EmptyBlock title="暂无版本树数据" hint="先创建镜头，再为镜头创建版本。" />
          ) : versionTree.map((node) => (
            <div key={node.root.id} className="rounded-2xl site-card p-4">
              <div className="text-black font-medium">根镜头 #{node.root.shot_number} · {node.root.title} · v{node.root.version}</div>
              <div className="text-xs text-black/55 mt-1">{node.root.chapter || '未分章'} · {node.root.duration}s</div>
              {node.versions.length > 0 ? (
                <div className="mt-3 space-y-2">
                  {node.versions.map((version) => (
                    <div key={version.id} className="rounded-lg bg-black/[0.02] border border-black/10 px-3 py-2 text-sm text-black/65">
                      v{version.version} · {version.title}
                      {version.version_note ? <span className="text-xs text-black/55 ml-2">({version.version_note})</span> : null}
                    </div>
                  ))}
                </div>
              ) : (
                <div className="mt-2 text-xs text-black/45">暂无衍生版本</div>
              )}
            </div>
          ))}
        </div>
      </ConsoleSection>

      <ConsoleSection title="镜头卡片建议字段" description="这部分决定未来短剧和电影生产线是否好用，先把字段边界定义清楚。">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 text-sm">
          <div className="rounded-2xl site-card p-5 space-y-3 text-black/65 leading-6">
            <div className="flex items-center gap-2 text-black font-medium"><SplitSquareVertical className="w-5 h-5 text-black/70" /> 生成前字段</div>
            <div>章节编号、镜头编号、镜头标题、目标时长、画幅、景别、运镜、情绪。</div>
            <div>角色绑定、场景绑定、参考图、镜头提示词、负向提示词、音效说明。</div>
            <div>模型选择、主 provider、备选 provider、预算限制、优先级。</div>
          </div>
          <div className="rounded-2xl site-card p-5 space-y-3 text-black/65 leading-6">
            <div className="flex items-center gap-2 text-black font-medium"><Eye className="w-5 h-5 text-black/70" /> 生成后字段</div>
            <div>版本号、缩略图、预览地址、生成时间、失败原因、审核意见。</div>
            <div>角色一致性评分、口型匹配评分、镜头可用性标签、是否入库。</div>
            <div>最终采用版本、复用次数、发布去向、二次剪辑引用关系。</div>
          </div>
        </div>
      </ConsoleSection>

      <ConsoleSection title="工作流节点" description="下一步可把这里演进成可视化工作流，串联脚本拆分、镜头生成和审核。">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 text-sm">
          <div className="rounded-2xl site-card p-4 text-black/65"><ListTodo className="w-5 h-5 text-black/70 mb-3" />脚本拆分为章节与镜头</div>
          <div className="rounded-2xl site-card p-4 text-black/65"><Clock3 className="w-5 h-5 text-black/70 mb-3" />镜头排队并按优先级生成</div>
          <div className="rounded-2xl site-card p-4 text-black/65"><CheckSquare className="w-5 h-5 text-black/70 mb-3" />审核并标记可用版本</div>
          <div className="rounded-2xl site-card p-4 text-black/65"><Eye className="w-5 h-5 text-black/70 mb-3" />进入剪辑与发布阶段</div>
        </div>
      </ConsoleSection>
    </FactoryConsoleShell>
  );
}