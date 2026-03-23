'use client';

import type { ReactNode } from 'react';

/** 全站区块：限制宽度 + 统一水平内边距 + 垂直节奏（参考 Apple/Tesla 大留白） */
export function SiteSection({
  id,
  children,
  className = '',
  tight = false,
  screen = false,
}: {
  id?: string;
  children: ReactNode;
  className?: string;
  tight?: boolean;
  screen?: boolean;
}) {
  const py = screen ? 'py-0' : tight ? 'py-16 md:py-20' : 'py-20 md:py-28 lg:py-32';
  return (
    <section id={id} className={`${py} ${screen ? 'min-h-[calc(100vh-64px)] snap-start flex items-center' : ''} ${className}`}>
      <div className="mx-auto w-full max-w-[1120px] px-5 sm:px-8">{children}</div>
    </section>
  );
}

/** 玻璃卡片：统一圆角、边框、背景（商务简洁） */
export function SiteCard({
  children,
  className = '',
  padding = 'p-6 md:p-8',
}: {
  children: ReactNode;
  className?: string;
  padding?: string;
}) {
  return (
    <div
      className={`site-card ui-lift rounded-2xl ${padding} ${className}`}
    >
      {children}
    </div>
  );
}

/** 图标容器：统一尺寸与描边，全站只用这一套 */
export function SiteIconBox({
  children,
  className = '',
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <span
      className={`inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-black/10 bg-black/[0.03] text-black/85 [&_svg]:h-[22px] [&_svg]:w-[22px] [&_svg]:stroke-[1.35] ${className}`}
    >
      {children}
    </span>
  );
}

/** 区块标题：可选居中 */
export function SiteSectionTitle({
  overline,
  title,
  subtitle,
  centered = false,
}: {
  overline?: string;
  title: ReactNode;
  subtitle?: string;
  centered?: boolean;
}) {
  return (
    <div className={centered ? 'mx-auto max-w-[42rem] text-center' : 'max-w-[42rem]'}>
      {overline ? (
        <p className="site-kicker mb-4">{overline}</p>
      ) : null}
      <h2 className="site-h2 text-balance">{title}</h2>
      {subtitle ? (
        <p className={`site-lead mt-4 text-balance ${centered ? 'mx-auto' : ''}`}>{subtitle}</p>
      ) : null}
    </div>
  );
}
