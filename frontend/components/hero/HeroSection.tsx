'use client';

import Link from 'next/link';
import { ArrowRight, Sparkles, Zap } from 'lucide-react';
import MagneticButton from '@/components/common/MagneticButton';

export default function HeroSection({
  onRegister,
}: {
  onRegister: () => void;
}) {
  return (
    <section className="relative min-h-[calc(100vh-64px)] snap-start overflow-hidden border-b border-black/8 bg-[#f7f7f7]">
      <div className="mx-auto flex min-h-[calc(100vh-64px)] w-full max-w-[1120px] items-center justify-center px-5 pt-6 text-center sm:px-8">
        <div className="w-full max-w-3xl">
          <div className="mx-auto inline-flex h-14 w-14 items-center justify-center rounded-2xl border border-black/10 bg-white text-black/80">
            <Sparkles className="h-7 w-7" strokeWidth={1.35} />
          </div>
          <p className="site-kicker ui-mt-16">AI 视频工厂</p>
          <h1 className="site-h1 ui-mt-12 text-balance text-[clamp(2.6rem,7vw,4.8rem)]">
            把内容生产
            <span className="block">做成可复制的商业流水线</span>
          </h1>
          <p className="site-lead ui-mt-16 mx-auto max-w-2xl text-black/60">
            面向内容团队，把脚本、镜头、视频与分发统一到一条产线，降低试错成本，让交付更稳定。
          </p>

          <div className="ui-mt-24 flex flex-col items-center justify-center ui-gap-12">
            <MagneticButton
              onClick={onRegister}
              className="btn-base btn-dark btn-l w-full max-w-[280px]"
            >
              <Zap className="h-4 w-4" strokeWidth={1.5} />
              立即申请试用
            </MagneticButton>
            <div className="flex flex-wrap items-center justify-center gap-x-5 gap-y-2 text-sm text-black/45">
              <Link href="/dashboard" className="inline-flex items-center gap-1 text-black/60 transition hover:text-black">
                查看运营控制台
                <ArrowRight className="h-3.5 w-3.5" strokeWidth={1.5} />
              </Link>
              <span className="hidden text-black/20 sm:inline">|</span>
              <Link href="/platforms" className="text-black/60 transition hover:text-black">
                渠道账号绑定中心
              </Link>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
