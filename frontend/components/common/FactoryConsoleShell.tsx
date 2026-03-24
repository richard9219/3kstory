"use client";

import Link from 'next/link';
import type { ReactNode } from 'react';
import { usePathname } from 'next/navigation';
import { ArrowLeft, Boxes, Clapperboard, Cpu, Layers3, Rocket, UserRoundCog, Wand2 } from 'lucide-react';

const HUB_LINKS = [
  {
    href: '/factory/models',
    title: '模型中心',
    description: '管理文生文、图像、视频、TTS 与审校模型供应商。',
    icon: Cpu,
    accent: 'from-black/10 to-black/0',
  },
  {
    href: '/factory/assets',
    title: '资产中心',
    description: '沉淀角色、场景、道具、品牌模板与提示词资产。',
    icon: Boxes,
    accent: 'from-black/10 to-black/0',
  },
  {
    href: '/factory/storyboards',
    title: '分镜中心',
    description: '承接章节拆分、镜头队列、版本审核与镜头生产。',
    icon: Clapperboard,
    accent: 'from-black/10 to-black/0',
  },
  {
    href: '/factory/releases',
    title: '发布历史',
    description: '查看导演版各平台上传回执，失败后可一键重试。',
    icon: Rocket,
    accent: 'from-black/10 to-black/0',
  },
  {
    href: '/factory/director-agents',
    title: '导演 Agent',
    description: '管理导演模板、自动策略与 A/B 双导演选优。',
    icon: UserRoundCog,
    accent: 'from-black/10 to-black/0',
  },
];

export function FactoryHubCards() {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
      {HUB_LINKS.map((item) => (
        <Link key={item.href} href={item.href} className="group rounded-2xl site-card p-5 transition hover:-translate-y-1 hover:border-black/20">
          <div className={`rounded-2xl bg-gradient-to-br ${item.accent} p-4 mb-4`}>
            <item.icon className="w-6 h-6 text-black" />
          </div>
          <div className="text-lg font-semibold text-black mb-2">{item.title}</div>
          <p className="text-sm text-black/65 leading-6">{item.description}</p>
          <div className="mt-4 text-sm text-black/70 group-hover:text-black transition">进入中心</div>
        </Link>
      ))}
    </div>
  );
}

export function FactoryConsoleShell({
  title,
  subtitle,
  eyebrow,
  children,
}: {
  title: string;
  subtitle: string;
  eyebrow: string;
  children: ReactNode;
}) {
  const pathname = usePathname();

  return (
    <main className="min-h-screen px-4 pb-10 bg-[#f7f7f7]">
      <div className="max-w-7xl mx-auto">
        <section className="min-h-[calc(100vh-96px)] flex items-center justify-center py-10">
          <div className="w-full max-w-4xl text-center">
            <Link href="/dashboard" className="inline-flex items-center gap-2 text-black/55 hover:text-black transition">
              <ArrowLeft className="w-4 h-4" />
              返回工厂控制台
            </Link>
            <div className="mx-auto ui-mt-16 inline-flex items-center gap-2 px-3 py-1.5 text-xs uppercase tracking-[0.24em] text-black/55 border border-black/10 rounded-full bg-white">
              <Wand2 className="w-3.5 h-3.5" />
              {eyebrow}
            </div>
            <h1 className="site-h1 ui-mt-12">{title}</h1>
            <p className="site-lead ui-mt-12 mx-auto max-w-3xl text-black/58">{subtitle}</p>

            <div className="ui-mt-24 rounded-2xl border border-black/10 bg-white p-2 overflow-x-auto">
              <div className="inline-flex gap-2 min-w-max">
                {HUB_LINKS.map((item) => {
                  const active = pathname.startsWith(item.href);
                  return (
                    <Link
                      key={item.href}
                      href={item.href}
                      className={`px-4 py-2 rounded-xl text-sm whitespace-nowrap transition ${
                        active
                          ? 'text-white bg-black border border-black'
                          : 'text-black/65 hover:text-black hover:bg-black/[0.06] border border-transparent'
                      }`}
                    >
                      {item.title}
                    </Link>
                  );
                })}
              </div>
            </div>
          </div>
        </section>

        <div className="grid grid-cols-1 xl:grid-cols-[260px,1fr] gap-6">
          <aside className="rounded-[28px] site-card p-5 h-fit sticky top-6">
            <div className="text-sm uppercase tracking-[0.24em] text-black/45 mb-4">工厂导航</div>
            <div className="space-y-3">
              {HUB_LINKS.map((item) => (
                <Link key={item.href} href={item.href} className="block rounded-2xl site-card p-4 hover:border-black/20 transition">
                  <div className="flex items-center gap-2 text-black font-medium mb-1">
                    <item.icon className="w-4 h-4 text-black/70" />
                    {item.title}
                  </div>
                  <div className="text-xs text-black/55 leading-5">{item.description}</div>
                </Link>
              ))}
            </div>
            <div className="mt-5 rounded-2xl bg-black/[0.02] border border-black/10 p-4">
              <div className="flex items-center gap-2 text-sm text-black font-medium mb-2">
                <Layers3 className="w-4 h-4 text-black/70" />
                下一阶段预留
              </div>
              <div className="text-xs text-black/55 leading-5">后续可继续扩展审片中心、投放中心、成本中心与工作流编排器。</div>
            </div>
          </aside>

          <div className="space-y-6">{children}</div>
        </div>
      </div>
    </main>
  );
}

export function ConsoleSection({ title, description, children }: { title: string; description: string; children: ReactNode }) {
  return (
    <section className="rounded-[28px] site-card p-6">
      <div className="mb-5">
        <h2 className="text-2xl font-semibold text-black mb-2">{title}</h2>
        <p className="text-sm text-black/55 leading-6 max-w-3xl">{description}</p>
      </div>
      {children}
    </section>
  );
}

export function MetricStrip({ items }: { items: { label: string; value: string; hint: string }[] }) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {items.map((item) => (
        <div key={item.label} className="rounded-2xl site-card p-4">
          <div className="text-xs uppercase tracking-[0.18em] text-black/45 mb-2">{item.label}</div>
          <div className="text-2xl font-bold text-black mb-2">{item.value}</div>
          <div className="text-sm text-black/55 leading-6">{item.hint}</div>
        </div>
      ))}
    </div>
  );
}