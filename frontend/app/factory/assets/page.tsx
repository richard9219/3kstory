"use client";

import { FormEvent, useEffect, useMemo, useState } from 'react';
import { Boxes, Brush, Building2, Film, Loader2, Package, Pencil, PlusCircle, Trash2, Users2 } from 'lucide-react';
import { ConsoleSection, FactoryConsoleShell, MetricStrip } from '@/components/common/FactoryConsoleShell';
import { EmptyBlock, LoadingBlock } from '@/components/common/StateBlocks';
import MiniButton from '@/components/common/MiniButton';
import { assetAPI, projectAPI } from '@/lib/api/client';
import type { Project, PromptTemplate, RoleAsset } from '@/types';

const assetGroups = [
  {
    title: '角色资产库',
    description: '沉淀人物设定、视觉参考、口型偏好、声音绑定和角色提示词模板。',
    icon: Users2,
  },
  {
    title: '场景资产库',
    description: '管理固定场景、镜头风格、镜头运动预设和背景视觉规范。',
    icon: Building2,
  },
  {
    title: '道具与服化道',
    description: '用于短剧和电影的道具、服装、妆容和品牌素材统一复用。',
    icon: Package,
  },
  {
    title: '品牌与广告模板',
    description: '沉淀品牌语气、产品卖点、Logo 安全区和广告镜头模板。',
    icon: Brush,
  },
];

export default function AssetsCenterPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [activeProjectId, setActiveProjectId] = useState<number | undefined>(undefined);
  const [roleAssets, setRoleAssets] = useState<RoleAsset[]>([]);
  const [templates, setTemplates] = useState<PromptTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);
  const [keyword, setKeyword] = useState('');
  const [tags, setTags] = useState('');

  const [roleForm, setRoleForm] = useState({ name: '', roleType: '', description: '', stylePrompt: '' });
  const [tplForm, setTplForm] = useState({ name: '', templateType: 'narration', content: '' });

  const loadAssets = (projectId?: number, q?: string, tagsValue?: string) => {
    return Promise.all([
      assetAPI.listRoleAssets({ project_id: projectId, q, tags: tagsValue }),
      assetAPI.listPromptTemplates({ project_id: projectId, q, tags: tagsValue }),
    ]).then(([rolesRes, templatesRes]) => {
      setRoleAssets(rolesRes.data.data || []);
      setTemplates(templatesRes.data.data || []);
    });
  };

  useEffect(() => {
    Promise.all([projectAPI.list(), loadAssets(undefined, '', '')])
      .then(([projectRes]) => {
        const list = (projectRes.data || []) as Project[];
        setProjects(list);
      })
      .catch(() => setMsg('加载资产数据失败'))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (!loading) {
      loadAssets(activeProjectId, keyword, tags).catch(() => setMsg('加载资产数据失败'));
    }
  }, [activeProjectId]);

  const stats = useMemo(() => ({ roles: roleAssets.length, templates: templates.length }), [roleAssets, templates]);

  const submitRole = async (e: FormEvent) => {
    e.preventDefault();
    if (!roleForm.name.trim()) return;
    setSaving(true);
    setMsg(null);
    try {
      await assetAPI.createRoleAsset({
        project_id: activeProjectId,
        name: roleForm.name.trim(),
        role_type: roleForm.roleType.trim(),
        description: roleForm.description.trim(),
        style_prompt: roleForm.stylePrompt.trim(),
        tags: tags.trim(),
      });
      setRoleForm({ name: '', roleType: '', description: '', stylePrompt: '' });
      await loadAssets(activeProjectId, keyword, tags);
      setMsg('角色资产已创建');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '创建角色资产失败');
    } finally {
      setSaving(false);
    }
  };

  const submitTemplate = async (e: FormEvent) => {
    e.preventDefault();
    if (!tplForm.name.trim() || !tplForm.content.trim()) return;
    setSaving(true);
    setMsg(null);
    try {
      await assetAPI.createPromptTemplate({
        project_id: activeProjectId,
        name: tplForm.name.trim(),
        template_type: tplForm.templateType,
        content: tplForm.content.trim(),
        tags: tags.trim(),
      });
      setTplForm({ name: '', templateType: 'narration', content: '' });
      await loadAssets(activeProjectId, keyword, tags);
      setMsg('提示词模板已创建');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '创建模板失败');
    } finally {
      setSaving(false);
    }
  };

  const handleSearch = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMsg(null);
    try {
      await loadAssets(activeProjectId, keyword.trim(), tags.trim());
    } catch {
      setMsg('检索资产失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteRole = async (id: number) => {
    setSaving(true);
    try {
      await assetAPI.deleteRoleAsset(id);
      await loadAssets(activeProjectId, keyword, tags);
      setMsg('角色资产已删除');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '删除角色资产失败');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteTemplate = async (id: number) => {
    setSaving(true);
    try {
      await assetAPI.deletePromptTemplate(id);
      await loadAssets(activeProjectId, keyword, tags);
      setMsg('模板已删除');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '删除模板失败');
    } finally {
      setSaving(false);
    }
  };

  const handleQuickEditRole = async (asset: RoleAsset) => {
    const name = window.prompt('更新角色名称', asset.name);
    if (!name) return;
    setSaving(true);
    try {
      await assetAPI.updateRoleAsset(asset.id, { name: name.trim(), tags: asset.tags || '' });
      await loadAssets(activeProjectId, keyword, tags);
      setMsg('角色资产已更新');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '更新角色资产失败');
    } finally {
      setSaving(false);
    }
  };

  const handleQuickEditTemplate = async (tpl: PromptTemplate) => {
    const name = window.prompt('更新模板名称', tpl.name);
    if (!name) return;
    setSaving(true);
    try {
      await assetAPI.updatePromptTemplate(tpl.id, { name: name.trim(), tags: tpl.tags || '' });
      await loadAssets(activeProjectId, keyword, tags);
      setMsg('模板已更新');
    } catch (error: any) {
      setMsg(error?.response?.data?.error || '更新模板失败');
    } finally {
      setSaving(false);
    }
  };

  return (
    <FactoryConsoleShell
      eyebrow="Asset Center"
      title="资产中心"
      subtitle="资产中心决定三千影视是否能从单条视频工具，走到可规模化复用的工厂系统。角色、场景、道具、品牌模板都应该在这里沉淀。"
    >
      <MetricStrip
        items={[
          { label: '角色资产', value: String(stats.roles), hint: '已写入数据库的角色资产数量。' },
          { label: '提示词模板', value: String(stats.templates), hint: '可复用提示词模板数量。' },
          { label: '当前作用域', value: activeProjectId ? `项目 #${activeProjectId}` : '全局资产', hint: '支持项目级和全局级资产并存。' },
        ]}
      />

      <ConsoleSection title="资产管理范围" description="可以切换项目视图，查看全局资产或项目资产。">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 items-end mb-4">
          <label className="text-sm text-black/65">
            项目范围
            <select
              value={activeProjectId ?? 0}
              onChange={(e) => setActiveProjectId(Number(e.target.value) || undefined)}
              className="factory-input mt-1 w-full rounded-lg px-3 py-2"
            >
              <option value={0}>全局资产</option>
              {projects.map((project) => (
                <option key={project.id} value={project.id}>{project.title || `项目 #${project.id}`}</option>
              ))}
            </select>
          </label>
          {msg && <div className="text-sm text-black/70">{msg}</div>}
        </div>

        <form onSubmit={handleSearch} className="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">
          <label className="text-sm text-black/65">
            关键词
            <input
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className="factory-input mt-1 w-full rounded-lg px-3 py-2"
              placeholder="名称/描述/内容"
            />
          </label>
          <label className="text-sm text-black/65">
            标签检索
            <input
              value={tags}
              onChange={(e) => setTags(e.target.value)}
              className="factory-input mt-1 w-full rounded-lg px-3 py-2"
              placeholder="逗号分隔，如 历史,男主"
            />
          </label>
          <button type="submit" disabled={loading} className="btn-base btn-light btn-m disabled:opacity-60">
            {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : '检索资产'}
          </button>
        </form>
      </ConsoleSection>

      <ConsoleSection title="资产分层" description="先把资产中心按对象拆清楚，后续再逐步接数据库和上传能力。">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {assetGroups.map((item) => (
            <div key={item.title} className="rounded-2xl site-card p-5">
              <item.icon className="w-6 h-6 text-black/70 mb-4" />
              <div className="text-lg font-semibold text-black mb-2">{item.title}</div>
              <p className="text-sm text-black/55 leading-6">{item.description}</p>
            </div>
          ))}
        </div>
      </ConsoleSection>

      <ConsoleSection title="创建角色资产" description="最小可用版本已支持角色资产入库，后续可再补头像上传、标签体系和批量导入。">
        <form onSubmit={submitRole} className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <label className="text-sm text-black/65">
            角色名称
            <input
              value={roleForm.name}
              onChange={(e) => setRoleForm((prev) => ({ ...prev, name: e.target.value }))}
              className="factory-input mt-1 w-full rounded-lg px-3 py-2"
              placeholder="例如：诸葛亮"
            />
          </label>
          <label className="text-sm text-black/65">
            角色类型
            <input
              value={roleForm.roleType}
              onChange={(e) => setRoleForm((prev) => ({ ...prev, roleType: e.target.value }))}
              className="factory-input mt-1 w-full rounded-lg px-3 py-2"
              placeholder="例如：主角 / 配角 / 旁白"
            />
          </label>
          <label className="text-sm text-black/65 md:col-span-2">
            角色描述
            <textarea
              value={roleForm.description}
              onChange={(e) => setRoleForm((prev) => ({ ...prev, description: e.target.value }))}
              className="factory-input mt-1 w-full rounded-lg px-3 py-2 min-h-24"
            />
          </label>
          <label className="text-sm text-black/65 md:col-span-2">
            风格提示词
            <textarea
              value={roleForm.stylePrompt}
              onChange={(e) => setRoleForm((prev) => ({ ...prev, stylePrompt: e.target.value }))}
              className="factory-input mt-1 w-full rounded-lg px-3 py-2 min-h-20"
            />
          </label>
          <button type="submit" disabled={saving} className="btn-base btn-dark btn-m disabled:opacity-60">
            {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <PlusCircle className="w-4 h-4" />}
            创建角色资产
          </button>
        </form>
      </ConsoleSection>

      <ConsoleSection title="创建提示词模板" description="模板可复用于影视解说、短剧分镜、广告脚本等不同生产线。">
        <form onSubmit={submitTemplate} className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-5">
          <label className="text-sm text-black/65">
            模板名称
            <input
              value={tplForm.name}
              onChange={(e) => setTplForm((prev) => ({ ...prev, name: e.target.value }))}
              className="factory-input mt-1 w-full rounded-lg px-3 py-2"
              placeholder="例如：历史化第一人称解说模板"
            />
          </label>
          <label className="text-sm text-black/65">
            模板类型
            <select
              value={tplForm.templateType}
              onChange={(e) => setTplForm((prev) => ({ ...prev, templateType: e.target.value }))}
              className="factory-input mt-1 w-full rounded-lg px-3 py-2"
            >
              <option value="narration">narration</option>
              <option value="storyboard">storyboard</option>
              <option value="ad-copy">ad-copy</option>
              <option value="movie-script">movie-script</option>
            </select>
          </label>
          <label className="text-sm text-black/65 md:col-span-2">
            模板内容
            <textarea
              value={tplForm.content}
              onChange={(e) => setTplForm((prev) => ({ ...prev, content: e.target.value }))}
              className="factory-input mt-1 w-full rounded-lg px-3 py-2 min-h-28"
            />
          </label>
          <button type="submit" disabled={saving} className="btn-base btn-light btn-m disabled:opacity-60">
            {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <PlusCircle className="w-4 h-4" />}
            创建模板
          </button>
        </form>

        {loading ? (
          <LoadingBlock text="正在加载资产列表..." />
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <h3 className="text-black font-medium mb-3">角色资产列表</h3>
              <div className="space-y-2">
                {roleAssets.length === 0 ? <EmptyBlock title="暂无角色资产" hint="可先创建角色资产或切换项目范围。" /> : roleAssets.map((item) => (
                  <div key={item.id} className="rounded-xl site-card p-3">
                    <div className="text-sm text-black font-medium">{item.name}</div>
                    <div className="text-xs text-black/55 mt-1">{item.role_type || '未分类'}</div>
                    <div className="text-xs text-black/45 mt-1">标签：{item.tags || '无'}</div>
                    <div className="mt-2 flex gap-2">
                      <MiniButton type="button" size="xs" tone="neutral" icon={Pencil} iconSize={14} onClick={() => handleQuickEditRole(item)}>
                        编辑
                      </MiniButton>
                      <MiniButton type="button" size="xs" tone="danger" icon={Trash2} iconSize={14} onClick={() => handleDeleteRole(item.id)}>
                        删除
                      </MiniButton>
                    </div>
                  </div>
                ))}
              </div>
            </div>
            <div>
              <h3 className="text-black font-medium mb-3">提示词模板列表</h3>
              <div className="space-y-2">
                {templates.length === 0 ? <EmptyBlock title="暂无提示词模板" hint="可以先创建一条模板，沉淀可复用提示词。" /> : templates.map((item) => (
                  <div key={item.id} className="rounded-xl site-card p-3">
                    <div className="text-sm text-black font-medium">{item.name}</div>
                    <div className="text-xs text-black/55 mt-1">{item.template_type}</div>
                    <div className="text-xs text-black/45 mt-1">标签：{item.tags || '无'}</div>
                    <div className="mt-2 flex gap-2">
                      <MiniButton type="button" size="xs" tone="neutral" icon={Pencil} iconSize={14} onClick={() => handleQuickEditTemplate(item)}>
                        编辑
                      </MiniButton>
                      <MiniButton type="button" size="xs" tone="danger" icon={Trash2} iconSize={14} onClick={() => handleDeleteTemplate(item.id)}>
                        删除
                      </MiniButton>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </ConsoleSection>

      <ConsoleSection title="建议数据结构" description="当前页面先作为骨架，后续最先应补数据结构、上传入口和引用关系。">
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-4 text-sm">
          <div className="rounded-2xl site-card p-5">
            <div className="flex items-center gap-2 text-black font-medium mb-3">
              <Boxes className="w-5 h-5 text-black/70" />
              资产字段建议
            </div>
            <div className="space-y-2 text-black/65 leading-6">
              <div>角色：姓名、身份、视觉参考、声音、提示词、禁用词。</div>
              <div>场景：场景名称、时代背景、镜头风格、光线、调色、参考图。</div>
              <div>道具：道具名称、适用项目、镜头标签、绑定角色、引用次数。</div>
              <div>模板：适用业务线、时长规格、平台规格、语气风格、镜头模版。</div>
            </div>
          </div>
          <div className="rounded-2xl site-card p-5">
            <div className="flex items-center gap-2 text-black font-medium mb-3">
              <Film className="w-5 h-5 text-black/70" />
              资产使用关系
            </div>
            <div className="space-y-2 text-black/65 leading-6">
              <div>解说项目：主要复用主题模板、标题模板、配音模板。</div>
              <div>短剧项目：重点复用角色、场景、镜头和提示词模板。</div>
              <div>电影项目：按章节继承资产，控制角色和场景长程一致性。</div>
              <div>广告项目：品牌模板与商品素材是核心资产对象。</div>
            </div>
          </div>
        </div>
      </ConsoleSection>
    </FactoryConsoleShell>
  );
}