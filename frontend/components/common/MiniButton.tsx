'use client';

import Link from 'next/link';
import type { LucideIcon } from 'lucide-react';
import type { ButtonHTMLAttributes } from 'react';
import type { ReactNode } from 'react';

type MiniButtonSize = 'xs' | 's';
type MiniButtonTone = 'neutral' | 'light' | 'danger';
type IconPixel = 14 | 16 | 20;

type MiniButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  size?: MiniButtonSize;
  tone?: MiniButtonTone;
  icon?: LucideIcon;
  iconSize?: IconPixel;
  iconStroke?: number;
};

type MiniLinkButtonProps = {
  href: string;
  size?: MiniButtonSize;
  tone?: MiniButtonTone;
  icon?: LucideIcon;
  iconSize?: IconPixel;
  iconStroke?: number;
  className?: string;
  children: ReactNode;
};

const sizeClasses: Record<MiniButtonSize, string> = {
  xs: 'h-7 px-2.5 text-xs',
  s: 'h-8 px-3 text-xs',
};

const toneClasses: Record<MiniButtonTone, string> = {
  neutral: 'border-black/12 bg-black/[0.04] text-black/78 hover:bg-black/[0.08] hover:border-black/20',
  light: 'border-black/14 bg-white text-black/75 hover:bg-black/[0.03] hover:border-black/22',
  danger: 'border-red-500/35 bg-red-500/12 text-red-700 hover:bg-red-500/18 hover:border-red-500/45',
};

const iconSizeClasses: Record<IconPixel, string> = {
  14: 'h-[14px] w-[14px]',
  16: 'h-4 w-4',
  20: 'h-5 w-5',
};

function baseClass(size: MiniButtonSize, tone: MiniButtonTone, className: string) {
  return `inline-flex items-center justify-center gap-1.5 rounded-md border font-medium transition ${sizeClasses[size]} ${toneClasses[tone]} ${className}`;
}

export default function MiniButton({
  size = 'xs',
  tone = 'neutral',
  icon: Icon,
  iconSize = 14,
  iconStroke = 1.5,
  className = '',
  children,
  ...props
}: MiniButtonProps) {
  return (
    <button
      className={baseClass(size, tone, className)}
      {...props}
    >
      {Icon ? <Icon className={iconSizeClasses[iconSize]} strokeWidth={iconStroke} /> : null}
      {children}
    </button>
  );
}

export function MiniLinkButton({
  href,
  size = 'xs',
  tone = 'neutral',
  icon: Icon,
  iconSize = 14,
  iconStroke = 1.5,
  className = '',
  children,
}: MiniLinkButtonProps) {
  return (
    <Link href={href} className={baseClass(size, tone, className)}>
      {Icon ? <Icon className={iconSizeClasses[iconSize]} strokeWidth={iconStroke} /> : null}
      {children}
    </Link>
  );
}
