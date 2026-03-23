import { CheckCircle2, Clock3, DollarSign, TrendingUp } from 'lucide-react';

import type { VideoProviderKind } from '@/types';

type ChartPoint = {
  label: string;
  value: number;
};

const VIDEO_PROVIDER_OPTIONS: Array<{ value: 'all' | VideoProviderKind; label: string }> = [
  { value: 'all', label: '全部引擎' },
  { value: 'runway', label: 'Runway' },
  { value: 'pika', label: 'Pika' },
  { value: 'local', label: '本地调试' },
  { value: 'minimax', label: 'MiniMax' },
  { value: 'seedance', label: 'Seedance' },
  { value: 'comfy', label: 'ComfyUI' },
];

type InsightCardProps = {
  title: string;
  value: string;
  hint: string;
  trendText: string;
  tone: 'cyan' | 'gold' | 'lime' | 'rose';
  points: ChartPoint[];
  area?: boolean;
};

const toneClass: Record<InsightCardProps['tone'], string> = {
  cyan: 'text-black/70',
  gold: 'text-black/70',
  lime: 'text-black/70',
  rose: 'text-black/70',
};

const toneStroke: Record<InsightCardProps['tone'], string> = {
  cyan: '#222222',
  gold: '#222222',
  lime: '#222222',
  rose: '#222222',
};

function buildLinePath(points: ChartPoint[], width: number, height: number, padding: number) {
  if (!points.length) return '';
  const values = points.map((p) => p.value);
  const max = Math.max(...values, 1);
  const min = Math.min(...values, 0);
  const range = Math.max(max - min, 1);

  return points
    .map((point, index) => {
      const x = padding + (index * (width - padding * 2)) / Math.max(points.length - 1, 1);
      const y = padding + ((max - point.value) / range) * (height - padding * 2);
      return `${index === 0 ? 'M' : 'L'}${x},${y}`;
    })
    .join(' ');
}

function buildAreaPath(linePath: string, points: ChartPoint[], width: number, height: number, padding: number) {
  if (!linePath || !points.length) return '';
  const firstX = padding;
  const lastX = padding + ((points.length - 1) * (width - padding * 2)) / Math.max(points.length - 1, 1);
  const floorY = height - padding;
  return `${linePath} L${lastX},${floorY} L${firstX},${floorY} Z`;
}

function Sparkline({ points, tone, area = false }: { points: ChartPoint[]; tone: InsightCardProps['tone']; area?: boolean }) {
  const width = 260;
  const height = 90;
  const padding = 10;
  const linePath = buildLinePath(points, width, height, padding);
  const areaPath = buildAreaPath(linePath, points, width, height, padding);

  return (
    <svg viewBox={`0 0 ${width} ${height}`} className="w-full h-24" preserveAspectRatio="none" role="img" aria-label="趋势图">
      {area && areaPath ? <path d={areaPath} fill={`${toneStroke[tone]}30`} /> : null}
      <path d={linePath} fill="none" stroke={toneStroke[tone]} strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export function OperationalInsightCard({ title, value, hint, trendText, tone, points, area }: InsightCardProps) {
  return (
    <div className="rounded-2xl site-card p-4">
      <div className="flex items-start justify-between gap-3 mb-2">
        <div className="text-sm text-black/60">{title}</div>
        <span className={`text-xs ${toneClass[tone]}`}>{trendText}</span>
      </div>
      <div className="text-2xl font-semibold text-black mb-2">{value}</div>
      <Sparkline points={points} tone={tone} area={area} />
      <div className="text-xs text-black/55 mt-2">{hint}</div>
    </div>
  );
}

export function OperationalInsightPanel({
  trendPoints,
  durationPoints,
  successPoints,
  costPoints,
  avgDurationMinutes,
  successRate,
  monthlyCost,
  range,
  onRangeChange,
  provider,
  onProviderChange,
  pipeline,
  onPipelineChange,
}: {
  trendPoints: ChartPoint[];
  durationPoints: ChartPoint[];
  successPoints: ChartPoint[];
  costPoints: ChartPoint[];
  avgDurationMinutes: number;
  successRate: number;
  monthlyCost: number;
  range: '7d' | '30d';
  onRangeChange: (v: '7d' | '30d') => void;
  provider: 'all' | VideoProviderKind;
  onProviderChange: (v: 'all' | VideoProviderKind) => void;
  pipeline: 'all' | 'narration' | 'scene';
  onPipelineChange: (v: 'all' | 'narration' | 'scene') => void;
}) {
  return (
    <section className="rounded-2xl site-card p-5 mb-8">
      <div className="flex flex-col xl:flex-row xl:items-center xl:justify-between gap-3 mb-4">
        <div>
          <h2 className="text-xl font-semibold text-black">运营洞察面板</h2>
          <div className="text-xs text-black/50">可复用卡片规范：趋势图 / 时长 / 成功率 / 成本曲线</div>
        </div>

        <div className="flex flex-wrap gap-2">
          <select
            value={range}
            onChange={(e) => onRangeChange(e.target.value as '7d' | '30d')}
            className="factory-input rounded-lg px-3 py-1.5 text-xs bg-white border-black/15 text-black"
          >
            <option value="7d">近 7 天</option>
            <option value="30d">近 30 天</option>
          </select>

          <select
            value={provider}
            onChange={(e) => onProviderChange(e.target.value as 'all' | VideoProviderKind)}
            className="factory-input rounded-lg px-3 py-1.5 text-xs bg-white border-black/15 text-black"
          >
            {VIDEO_PROVIDER_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>{option.label}</option>
            ))}
          </select>

          <select
            value={pipeline}
            onChange={(e) => onPipelineChange(e.target.value as 'all' | 'narration' | 'scene')}
            className="factory-input rounded-lg px-3 py-1.5 text-xs bg-white border-black/15 text-black"
          >
            <option value="all">全部业务线</option>
            <option value="narration">影视解说线</option>
            <option value="scene">镜头生成线</option>
          </select>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4">
        <OperationalInsightCard
          title="任务趋势"
          value={`${trendPoints.reduce((sum, point) => sum + point.value, 0)} 条`}
          hint="近 7 天任务提交数"
          trendText="趋势"
          tone="cyan"
          points={trendPoints}
        />

        <OperationalInsightCard
          title="任务耗时"
          value={`${avgDurationMinutes.toFixed(1)} 分钟`}
          hint="已完成任务平均耗时"
          trendText="耗时"
          tone="gold"
          points={durationPoints}
        />

        <OperationalInsightCard
          title="模型成功率"
          value={`${(successRate * 100).toFixed(1)}%`}
          hint="completed / (completed + failed)"
          trendText="质量"
          tone="lime"
          points={successPoints}
        />

        <OperationalInsightCard
          title="成本曲线"
          value={`¥${monthlyCost.toFixed(1)}`}
          hint="按 provider 估算的本月渲染成本"
          trendText="预算"
          tone="rose"
          points={costPoints}
          area
        />
      </div>

      <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-3 text-xs">
        <div className="rounded-xl site-card p-3 text-black/65 inline-flex items-center gap-2"><TrendingUp className="w-4 h-4 text-black/70" /> 趋势图</div>
        <div className="rounded-xl site-card p-3 text-black/65 inline-flex items-center gap-2"><Clock3 className="w-4 h-4 text-black/70" /> 任务耗时</div>
        <div className="rounded-xl site-card p-3 text-black/65 inline-flex items-center gap-2"><CheckCircle2 className="w-4 h-4 text-black/70" /> 成功率</div>
        <div className="rounded-xl site-card p-3 text-black/65 inline-flex items-center gap-2"><DollarSign className="w-4 h-4 text-black/70" /> 成本曲线</div>
      </div>
    </section>
  );
}
