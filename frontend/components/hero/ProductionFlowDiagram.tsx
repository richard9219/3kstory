'use client';

import { motion } from 'framer-motion';
import { BarChart3, Boxes, Clapperboard, ScanText } from 'lucide-react';

const FLOW_NODES = [
  {
    id: 'input',
    title: '内容输入',
    desc: '脚本 / 解说稿 / 选题',
    step: '01',
    icon: ScanText,
    glow: 'from-cyan-400/40 to-blue-500/10',
  },
  {
    id: 'split',
    title: '镜头拆解',
    desc: '章节 / 场景 / 节奏',
    step: '02',
    icon: Boxes,
    glow: 'from-amber-300/40 to-orange-500/10',
  },
  {
    id: 'render',
    title: '视频生成',
    desc: '多模型并行渲染',
    step: '03',
    icon: Clapperboard,
    glow: 'from-lime-300/36 to-emerald-500/10',
  },
  {
    id: 'review',
    title: '运营复盘',
    desc: '分发与数据回流',
    step: '04',
    icon: BarChart3,
    glow: 'from-cyan-300/36 to-indigo-500/10',
  },
];

export default function ProductionFlowDiagram() {
  return (
    <div className="rounded-[30px] border border-white/10 bg-black/30 p-5 backdrop-blur-xl sm:p-6">
      <div className="mb-4 text-[11px] uppercase tracking-[0.24em] text-cyan-200/72">Production Flow</div>
      <div className="overflow-hidden rounded-[24px] border border-white/10 bg-[linear-gradient(180deg,rgba(5,8,13,0.9),rgba(4,5,8,0.96))] p-4 sm:p-5">
        <div className="mb-4 hidden h-[2px] w-full bg-[linear-gradient(90deg,rgba(81,200,255,0.1),rgba(216,255,61,0.6),rgba(81,200,255,0.1))] lg:block" />

        <div className="grid gap-3 lg:grid-cols-4">
          {FLOW_NODES.map((node, index) => (
            <motion.div
              key={node.id}
              initial={{ opacity: 0, y: 8 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, amount: 0.3 }}
              transition={{ duration: 0.35, delay: index * 0.07 }}
              className="relative overflow-hidden rounded-2xl border border-white/12 bg-[#090c12]/95 p-3.5"
            >
              <div className={`absolute inset-0 bg-gradient-to-br ${node.glow}`} />
              <div className="relative z-10">
                <div className="flex items-center justify-between">
                  <div className="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-white/12 bg-white/[0.04] text-cyan-200">
                    <node.icon className="h-4.5 w-4.5" />
                  </div>
                  <div className="text-[11px] tracking-[0.2em] text-white/40">{node.step}</div>
                </div>
                <div className="mt-2.5 text-sm font-semibold text-white">{node.title}</div>
                <div className="mt-1 text-xs leading-5 text-white/62">{node.desc}</div>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </div>
  );
}
