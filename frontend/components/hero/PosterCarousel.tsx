'use client';

import { AnimatePresence, motion } from 'framer-motion';
import { ChevronLeft, ChevronRight, Clapperboard, Film, Sparkles } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';

type PosterItem = {
  title: string;
  badge: string;
  tags: string[];
};

const POSTERS: PosterItem[] = [
  {
    title: '药命效应 · 电影解说',
    badge: 'Commentary',
    tags: ['剧情拆解', '旁白生成', '镜头拼接'],
  },
  {
    title: '72小时追凶 · AI短剧',
    badge: 'Drama',
    tags: ['角色资产', '分镜版本', '连续更新'],
  },
  {
    title: '品牌首发片 · 广告内容',
    badge: 'Campaign',
    tags: ['品牌语调', '高频A/B', '跨平台分发'],
  },
];

export default function PosterCarousel() {
  const [index, setIndex] = useState(0);

  useEffect(() => {
    const timer = window.setInterval(() => {
      setIndex((prev) => (prev + 1) % POSTERS.length);
    }, 4200);

    return () => window.clearInterval(timer);
  }, []);

  const activePoster = useMemo(() => POSTERS[index], [index]);

  const goPrev = () => {
    setIndex((prev) => (prev - 1 + POSTERS.length) % POSTERS.length);
  };

  const goNext = () => {
    setIndex((prev) => (prev + 1) % POSTERS.length);
  };

  return (
    <div className="rounded-[30px] border border-black/10 bg-white p-4 shadow-[0_24px_70px_rgba(0,0,0,0.08)] sm:p-5">
      <div className="mb-4 flex items-center justify-between">
        <div className="inline-flex items-center ui-gap-8 text-xs font-medium uppercase tracking-[0.16em] text-black/45">
          <Film className="h-4 w-4" strokeWidth={1.35} />
          案例轮播
        </div>
        <div className="flex items-center ui-gap-8">
          <button
            type="button"
            onClick={goPrev}
            className="btn-base btn-light btn-s w-8 !px-0"
            aria-label="上一张"
          >
            <ChevronLeft className="h-4 w-4" strokeWidth={1.35} />
          </button>
          <button
            type="button"
            onClick={goNext}
            className="btn-base btn-light btn-s w-8 !px-0"
            aria-label="下一张"
          >
            <ChevronRight className="h-4 w-4" strokeWidth={1.35} />
          </button>
        </div>
      </div>

      <div className="relative h-[320px] overflow-hidden rounded-[24px] border border-black/10 bg-[linear-gradient(160deg,#f4f4f5,#ececef)] sm:h-[360px]">
        <AnimatePresence mode="wait">
          <motion.div
            key={activePoster.title}
            initial={{ opacity: 0, x: 24 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -24 }}
            transition={{ duration: 0.45 }}
            className="absolute inset-0"
          >
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_25%_28%,rgba(255,255,255,0.95),transparent_35%),radial-gradient(circle_at_74%_74%,rgba(0,0,0,0.06),transparent_36%)]" />
            <div className="absolute bottom-0 left-0 right-0 p-5 sm:p-6">
              <div className="inline-flex items-center ui-gap-8 rounded-full border border-black/12 bg-white/85 px-3 py-1 text-[11px] uppercase tracking-[0.16em] text-black/78">
                <Sparkles className="h-3.5 w-3.5 text-black/55" strokeWidth={1.35} />
                {activePoster.badge}
              </div>

              <h3 className="factory-title mt-3 max-w-xl text-2xl leading-tight text-black sm:text-[32px]">
                {activePoster.title}
              </h3>

              <div className="mt-4 flex flex-wrap gap-2">
                {activePoster.tags.map((tag) => (
                  <span key={tag} className="rounded-full border border-black/12 bg-white/70 px-3 py-1 text-xs text-black/72">
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          </motion.div>
        </AnimatePresence>
      </div>

      <div className="mt-4 grid grid-cols-3 gap-2">
        {POSTERS.map((poster, idx) => (
          <button
            type="button"
            key={poster.title}
            onClick={() => setIndex(idx)}
            className={`rounded-xl border px-3 py-2 text-left text-xs transition ${
              idx === index
                ? 'border-black/25 bg-black text-white'
                : 'border-black/10 bg-black/[0.02] text-black/58 hover:text-black/85'
            }`}
          >
            <div className="inline-flex items-center gap-1.5">
              <Clapperboard className="h-3.5 w-3.5" strokeWidth={1.35} />
              案例 {idx + 1}
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
