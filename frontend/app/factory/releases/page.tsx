"use client";

import { useEffect, useMemo, useState } from 'react';
import { AlertTriangle, CheckCircle2, Loader2, RefreshCcw, Rocket } from 'lucide-react';
import { ConsoleSection, FactoryConsoleShell, MetricStrip } from '@/components/common/FactoryConsoleShell';
import { EmptyBlock, LoadingBlock } from '@/components/common/StateBlocks';
import { projectAPI, storyboardAPI } from '@/lib/api/client';
import type { DirectorPublishRecord, PlatformKind, Project } from '@/types';

const PLATFORM_LABEL: Record<PlatformKind, string> = {
  douyin: '抖音',
  xiaohongshu: '小红书',
  bilibili: 'B站',
  weibo: '微博',
};

function fmtTime(value?: string) {
  if (!value) return '-';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toLocaleString('zh-CN', { hour12: false });
}

export default function ReleaseHistoryPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectId, setProjectId] = useState<number>(0);
  const [records, setRecords] = useState<DirectorPublishRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [retryingId, setRetryingId] = useState<number | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const loadData = async (nextProjectId: number) => {
    if (!nextProjectId) {
      setRecords([]);
      return;
    }
    const historyRes = await storyboardAPI.listDirectorPublishHistory(nextProjectId);
    setRecords(historyRes.data.data || []);
  };

  useEffect(() => {
    projectAPI.list()
      .then(async (res) => {
        const list = (res.data || []) as Project[];
        setProjects(list);
        if (list[0]?.id) {
          setProjectId(list[0].id);
          await loadData(list[0].id);
        }
      })
      .catch(() => setMessage('加载发布历史失败'))
      .finally(() => setLoading(false));
  }, []);

  const successCount = useMemo(() => records.filter((item) => item.status === 'success').length, [records]);
  const failedCount = useMemo(() => records.filter((item) => item.status === 'failed').length, [records]);

  const handleRetry = async (record: DirectorPublishRecord) => {
    if (!projectId || retryingId) return;
    setRetryingId(record.id);
    setMessage(null);
    try {
      await storyboardAPI.retryDirectorPublish(projectId, record.id, {
        reason: 'manual_retry_from_release_history',
      });
      await loadData(projectId);
      setMessage(`已重试 ${PLATFORM_LABEL[record.platform]} 发布`);
    } catch (error: any) {
      setMessage(error?.response?.data?.error || '重试失败');
    } finally {
      setRetryingId(null);
    }
  };

  return (
    <FactoryConsoleShell
      eyebrow="Publish Receipts"
      title="导演版发布历史"
      subtitle="每次平台上传都记录真实回执（请求号/HTTP状态/远端视频ID），失败记录可直接重试。"
    >
      <MetricStrip
        items={[
          { label: '发布总尝试', value: String(records.length), hint: '含首次发布与失败重试。' },
          { label: '成功回执', value: String(successCount), hint: '收到平台成功上传反馈。' },
          { label: '失败待处理', value: String(failedCount), hint: '可直接点击重试。' },
        ]}
      />

      <ConsoleSection title="项目筛选" description="可切换项目查看导演版导出的平台发布回执。">
        <div className="rounded-2xl site-card p-4 space-y-3">
          <label className="text-sm text-black/65 block">
            项目
            <select
              value={projectId}
              onChange={async (e) => {
                const nextId = Number(e.target.value) || 0;
                setProjectId(nextId);
                setLoading(true);
                try {
                  await loadData(nextId);
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
          {message ? <div className="text-sm text-black/70">{message}</div> : null}
        </div>
      </ConsoleSection>

      <ConsoleSection title="发布记录" description="按时间倒序展示每次上传尝试；失败项提供重试入口。">
        {loading ? (
          <LoadingBlock text="正在加载发布记录..." />
        ) : records.length === 0 ? (
          <EmptyBlock title="暂无发布记录" hint="先在分镜中心生成导演版并执行平台发布。" />
        ) : (
          <div className="space-y-3">
            {records.map((record) => (
              <div key={record.id} className="rounded-2xl border border-black/10 bg-white p-4">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div className="flex items-center gap-2 text-sm font-medium text-black">
                    <Rocket className="w-4 h-4 text-black/70" />
                    {PLATFORM_LABEL[record.platform]} · 导出 {record.export_id}
                  </div>
                  <span className={`inline-flex items-center gap-1 rounded-full border px-2 py-1 text-xs ${record.status === 'success' ? 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700' : record.status === 'failed' ? 'border-red-500/25 bg-red-500/10 text-red-700' : 'border-black/10 bg-black/[0.03] text-black/65'}`}>
                    {record.status === 'success' ? <CheckCircle2 className="w-3.5 h-3.5" /> : record.status === 'failed' ? <AlertTriangle className="w-3.5 h-3.5" /> : null}
                    {record.status}
                  </span>
                </div>

                <div className="mt-3 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-3 text-xs text-black/65">
                  <div>尝试次数: {record.attempt_no}</div>
                  <div>回执ID: {record.receipt_id || '-'}</div>
                  <div>远端视频ID: {record.remote_video_id || '-'}</div>
                  <div>完成时间: {fmtTime(record.completed_at)}</div>
                </div>

                <div className="mt-2 grid grid-cols-1 md:grid-cols-2 gap-3 text-xs text-black/60">
                  <div>请求号: {String(record.response_payload?.request_id || '-')}</div>
                  <div>HTTP状态: {String(record.response_payload?.http_status || '-')}</div>
                  {record.error_msg ? <div className="md:col-span-2 text-red-700">失败原因: {record.error_msg}</div> : null}
                </div>

                <div className="mt-3 flex items-center justify-between gap-3">
                  <div className="text-xs text-black/50">创建时间: {fmtTime(record.created_at)}</div>
                  {record.status === 'failed' ? (
                    <button
                      onClick={() => handleRetry(record)}
                      disabled={retryingId === record.id}
                      className="btn-base btn-dark btn-m disabled:opacity-60"
                    >
                      {retryingId === record.id ? <Loader2 className="w-4 h-4 animate-spin" /> : <RefreshCcw className="w-4 h-4" />}
                      重试发布
                    </button>
                  ) : null}
                </div>
              </div>
            ))}
          </div>
        )}
      </ConsoleSection>
    </FactoryConsoleShell>
  );
}
