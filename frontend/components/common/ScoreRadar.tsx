'use client';

import { useMemo } from 'react';

type RadarMetric = {
  key: string;
  label: string;
  value: number;
};

type ScoreRadarProps = {
  detail?: Record<string, any>;
  size?: number;
};

const METRICS: Array<{ key: string; label: string }> = [
  { key: 'sync', label: '时长同步' },
  { key: 'readability', label: '字幕可读' },
  { key: 'visual', label: '画面质量' },
  { key: 'audio', label: '音频质量' },
  { key: 'stability', label: '稳定性' },
  { key: 'cost_efficiency', label: '成本效率' },
];

function clamp(v: number) {
  if (!Number.isFinite(v)) return 0;
  if (v < 0) return 0;
  if (v > 1) return 1;
  return v;
}

export default function ScoreRadar({ detail, size = 220 }: ScoreRadarProps) {
  const center = size / 2;
  const maxRadius = size * 0.34;

  const metrics = useMemo<RadarMetric[]>(() => {
    return METRICS.map((item) => ({
      key: item.key,
      label: item.label,
      value: clamp(Number(detail?.[item.key] ?? 0)),
    }));
  }, [detail]);

  const levelPolygons = [0.2, 0.4, 0.6, 0.8, 1].map((level) => {
    const points = metrics.map((_, i) => {
      const angle = (Math.PI * 2 * i) / metrics.length - Math.PI / 2;
      const radius = maxRadius * level;
      const x = center + Math.cos(angle) * radius;
      const y = center + Math.sin(angle) * radius;
      return `${x},${y}`;
    });
    return points.join(' ');
  });

  const axisLines = metrics.map((_, i) => {
    const angle = (Math.PI * 2 * i) / metrics.length - Math.PI / 2;
    const x = center + Math.cos(angle) * maxRadius;
    const y = center + Math.sin(angle) * maxRadius;
    return { x, y };
  });

  const dataPoints = metrics.map((m, i) => {
    const angle = (Math.PI * 2 * i) / metrics.length - Math.PI / 2;
    const radius = maxRadius * m.value;
    const x = center + Math.cos(angle) * radius;
    const y = center + Math.sin(angle) * radius;
    return { x, y, label: m.label, value: m.value };
  });

  const dataPolygon = dataPoints.map((p) => `${p.x},${p.y}`).join(' ');

  return (
    <div className="rounded-lg border border-black/10 bg-white p-3">
      <div className="text-xs text-black/55 mb-2">评分雷达图</div>
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} className="mx-auto block">
        {levelPolygons.map((poly, idx) => (
          <polygon key={idx} points={poly} fill="none" stroke="rgba(0,0,0,0.12)" strokeWidth="1" />
        ))}

        {axisLines.map((line, idx) => (
          <line
            key={idx}
            x1={center}
            y1={center}
            x2={line.x}
            y2={line.y}
            stroke="rgba(0,0,0,0.12)"
            strokeWidth="1"
          />
        ))}

        <polygon points={dataPolygon} fill="rgba(0,0,0,0.12)" stroke="rgba(0,0,0,0.55)" strokeWidth="2" />

        {dataPoints.map((p, idx) => (
          <circle key={idx} cx={p.x} cy={p.y} r="3" fill="rgba(0,0,0,0.75)" />
        ))}

        {metrics.map((m, i) => {
          const angle = (Math.PI * 2 * i) / metrics.length - Math.PI / 2;
          const labelRadius = maxRadius + 20;
          const x = center + Math.cos(angle) * labelRadius;
          const y = center + Math.sin(angle) * labelRadius;
          return (
            <text
              key={m.key}
              x={x}
              y={y}
              fontSize="11"
              textAnchor="middle"
              dominantBaseline="middle"
              fill="rgba(0,0,0,0.72)"
            >
              {m.label}
            </text>
          );
        })}
      </svg>
      <div className="grid grid-cols-2 gap-1 mt-1">
        {metrics.map((m) => (
          <div key={m.key} className="text-[11px] text-black/65">{m.label}: {m.value.toFixed(3)}</div>
        ))}
      </div>
    </div>
  );
}
