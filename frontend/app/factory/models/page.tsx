"use client";

import { useEffect, useMemo, useState } from 'react';
import { AlertTriangle, CheckCircle2, Cpu, FileText, ImageIcon, Loader2, RadioTower, ShieldCheck, Video, Volume2 } from 'lucide-react';
import { ConsoleSection, FactoryConsoleShell, MetricStrip } from '@/components/common/FactoryConsoleShell';
import { EmptyBlock, ErrorBlock, LoadingBlock } from '@/components/common/StateBlocks';
import { modelCenterAPI } from '@/lib/api/client';
import type { ModelCenterOverview } from '@/types';

const modelStacks = [
  {
    name: '脚本生成',
    primary: 'Qwen / Claude / GPT',
    backup: '本地 vLLM / Ollama',
    icon: FileText,
  },
  {
    name: '图像生成',
    primary: 'Midjourney / Flux / SDXL',
    backup: 'ComfyUI 工作流',
    icon: ImageIcon,
  },
  {
    name: '视频生成',
    primary: 'Runway',
    backup: 'Pika / 本地调试链路',
    icon: Video,
  },
  {
    name: '旁白与配音',
    primary: 'Edge TTS / 云端语音',
    backup: '本地 TTS',
    icon: Volume2,
  },
];

const queues = [
  '供应商优先级：为每类模型维护主引擎、备选引擎、兜底引擎。',
  '配置健康度：校验 API Key、Base URL、模型名、超时和可达性。',
  '错误分类：区分配置错误、鉴权错误、额度限制、参数不合法、轮询失败。',
  '路由策略：按任务类型和项目阶段分配到影视解说、短剧或电影流水线。',
];

export default function ModelsCenterPage() {
  const [offlineMode, setOfflineMode] = useState(false);
  const [overview, setOverview] = useState<ModelCenterOverview | null>(null);
  const [loading, setLoading] = useState(true);
  const [probing, setProbing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    modelCenterAPI
      .overview()
      .then((res) => setOverview(res.data.data))
      .catch((err) => setError(err?.response?.data?.error || '加载模型中心状态失败'))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    setOfflineMode(sessionStorage.getItem('frontend-offline-fallback') === '1');
  }, []);

  const handleProbe = async () => {
    setProbing(true);
    setError(null);
    try {
      const res = await modelCenterAPI.triggerProbe();
      setOverview(res.data.data);
    } catch (err: any) {
      setError(err?.response?.data?.error || '触发探活失败');
    } finally {
      setProbing(false);
    }
  };

  const statusCount = useMemo(() => {
    const providers = overview?.video_providers || [];
    return {
      healthy: providers.filter((p) => p.healthy).length,
      total: providers.length,
      configured: providers.filter((p) => p.configured).length,
    };
  }, [overview]);

  return (
    <FactoryConsoleShell
      eyebrow="Model Center"
      title="模型中心"
      subtitle="统一维护脚本、图像、视频、配音与审校模型供应链，决定每条生产线的主引擎、备选引擎和兜底策略。"
    >
      <MetricStrip
        items={[
          { label: '视频主引擎', value: (overview?.preferred_video_provider || 'runway').toUpperCase(), hint: '解说、短剧和电影任务优先接入云端视频模型。' },
          { label: '已配置引擎', value: `${statusCount.configured}/${statusCount.total}`, hint: '视频 provider 配置完成度（Runway / Pika / Local）。' },
          { label: '健康状态', value: `${statusCount.healthy}/${statusCount.total}`, hint: '后台定时探活 + 请求时读取缓存结果。' },
          { label: '告警阈值', value: String(overview?.probe_task?.failure_threshold || 0), hint: '连续失败达到阈值后进入告警。' },
        ]}
      />

      <div className="rounded-2xl site-card p-4 flex flex-wrap gap-3 items-center justify-between">
        <div className="text-sm text-black/65">
          探活任务：{overview?.probe_task?.enabled ? '已启用' : '未启用'}
          {overview?.probe_task?.interval_seconds ? ` · 周期 ${overview.probe_task.interval_seconds}s` : ''}
          {overview?.probe_task?.last_probe_at ? ` · 最近探活 ${new Date(overview.probe_task.last_probe_at).toLocaleString()}` : ''}
        </div>
        <button
          onClick={handleProbe}
          disabled={loading || probing}
          className="btn-base btn-light btn-m disabled:opacity-60"
        >
          {(loading || probing) ? <Loader2 className="w-4 h-4 animate-spin" /> : <RadioTower className="w-4 h-4" />}
          立即探活
        </button>
      </div>

      {offlineMode ? (
        <div className="rounded-xl border border-amber-300/30 bg-amber-300/10 px-4 py-3 text-sm text-amber-800">
          当前处于演示模式：模型中心状态为前端兜底数据，仅用于联调与界面验证。
        </div>
      ) : null}

      {loading && (
        <LoadingBlock text="正在加载模型中心状态..." />
      )}

      {error && (
        <ErrorBlock text={error} />
      )}

      <ConsoleSection title="模型栈视图" description="按生产环节查看模型分层，避免系统只围绕单一模型设计。">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {modelStacks.map((item) => (
            <div key={item.name} className="rounded-2xl site-card p-5">
              <item.icon className="w-6 h-6 text-black/70 mb-4" />
              <div className="text-lg font-semibold text-black mb-2">{item.name}</div>
              <div className="text-sm text-black/65 mb-1">主引擎：{item.primary}</div>
              <div className="text-sm text-black/55">备选：{item.backup}</div>
            </div>
          ))}
        </div>
      </ConsoleSection>

      <ConsoleSection title="真实健康检查" description="当前版本已经接入后端真实接口，状态由配置校验与可达性检查共同决定。">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {(overview?.video_providers || []).map((provider) => (
            <div key={provider.provider} className="rounded-2xl site-card p-4">
              <div className="text-black font-medium mb-2 uppercase">{provider.provider}</div>
              <div className="text-sm text-black/65 mb-2">{provider.message}</div>
              <div className="text-xs text-black/55 mb-3">检查时间：{new Date(provider.checked_at).toLocaleString()}</div>
              <div className="inline-flex items-center gap-2 text-sm">
                {provider.healthy ? (
                  <span className="inline-flex items-center gap-1 text-black/70"><CheckCircle2 className="w-4 h-4" /> 健康</span>
                ) : (
                  <span className="inline-flex items-center gap-1 text-black/70"><AlertTriangle className="w-4 h-4" /> 异常</span>
                )}
                {!provider.configured && <span className="text-red-700">未配置</span>}
              </div>
            </div>
          ))}
        </div>
      </ConsoleSection>

      <ConsoleSection title="失败告警阈值" description="当同一模型连续失败达到阈值时触发告警，直到探活恢复。">
        {(overview?.alerts || []).length === 0 ? (
          <EmptyBlock title="暂无告警" hint="探活任务运行正常，或尚未触发失败阈值。" />
        ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {(overview?.alerts || []).map((alert) => (
            <div key={`${alert.category}-${alert.name}`} className="rounded-2xl site-card p-4">
              <div className="flex items-center justify-between mb-2">
                <div className="text-black font-medium">{alert.category} / {alert.name}</div>
                <div className={alert.alerting ? 'text-red-700 text-xs' : 'text-black/70 text-xs'}>
                  {alert.alerting ? '告警中' : '正常'}
                </div>
              </div>
              <div className="text-sm text-black/65 mb-2">连续失败：{alert.failure_streak} / {alert.failure_threshold}</div>
              <div className="text-xs text-black/55">最近状态：{alert.last_message || '无'}</div>
            </div>
          ))}
        </div>
        )}
      </ConsoleSection>

      <ConsoleSection title="治理待办" description="模型中心不是展示页，而是生产稳定性的控制面。下面这些项目后续都可以接真实数据。">
        <div className="grid grid-cols-1 xl:grid-cols-[1fr,0.9fr] gap-4">
          <div className="space-y-3">
            {queues.map((item) => (
              <div key={item} className="rounded-2xl site-card p-4 text-sm text-black/65 leading-6">{item}</div>
            ))}
          </div>
          <div className="rounded-2xl site-card p-5">
            <div className="flex items-center gap-2 text-black font-medium mb-4">
              <Cpu className="w-5 h-5 text-black/70" />
              配置状态样板
            </div>
            <div className="space-y-3 text-sm">
              <div className="flex items-center justify-between rounded-xl bg-black/[0.02] px-4 py-3">
                <span className="text-black/65">Runway API</span>
                <span className="inline-flex items-center gap-1 text-black/70"><CheckCircle2 className="w-4 h-4" /> 已配置</span>
              </div>
              <div className="flex items-center justify-between rounded-xl bg-black/[0.02] px-4 py-3">
                <span className="text-black/65">Pika API</span>
                <span className="inline-flex items-center gap-1 text-black/70"><AlertTriangle className="w-4 h-4" /> 作为备选</span>
              </div>
              <div className="flex items-center justify-between rounded-xl bg-black/[0.02] px-4 py-3">
                <span className="text-black/65">审核模型</span>
                <span className="inline-flex items-center gap-1 text-black/65"><ShieldCheck className="w-4 h-4" /> 待接入</span>
              </div>
              <div className="flex items-center justify-between rounded-xl bg-black/[0.02] px-4 py-3">
                <span className="text-black/65">状态轮询网关</span>
                <span className="inline-flex items-center gap-1 text-black/70"><RadioTower className="w-4 h-4" /> 已抽象</span>
              </div>
            </div>
          </div>
        </div>
      </ConsoleSection>
    </FactoryConsoleShell>
  );
}