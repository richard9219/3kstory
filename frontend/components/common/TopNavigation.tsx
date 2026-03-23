"use client";

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { ArrowRight } from 'lucide-react';
import BrandLogo from '@/components/common/BrandLogo';

type NavItem = {
  href: string;
  label: string;
  match: (pathname: string) => boolean;
};

const NAV_ITEMS: NavItem[] = [
  { href: '/', label: '官网', match: (p) => p === '/' },
  { href: '/dashboard', label: '控制台', match: (p) => p.startsWith('/dashboard') },
  { href: '/factory/models', label: '产品', match: (p) => p.startsWith('/factory') },
  { href: '/platforms', label: '渠道', match: (p) => p.startsWith('/platforms') || p.startsWith('/platform-bound') },
];

export default function TopNavigation() {
  const pathname = usePathname();
  const isHome = pathname === '/';

  return (
    <header className="sticky top-0 z-40 border-b border-black/10 bg-[#f7f7f7]">
      <div className="mx-auto max-w-[1120px] px-4 sm:px-6">
        <div className="flex h-16 items-center gap-3 overflow-x-auto whitespace-nowrap [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          <div className="flex shrink-0 justify-start">
            <BrandLogo />
          </div>

          <nav className="flex shrink-0 items-center gap-1 pr-1">
            {NAV_ITEMS.map((item) => {
              const active = item.match(pathname);
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={`shrink-0 rounded-lg px-3.5 py-2 text-sm font-medium transition ${
                    active
                      ? 'bg-black text-white'
                      : 'text-black/65 hover:bg-black/[0.04] hover:text-black'
                  }`}
                >
                  {item.label}
                </Link>
              );
            })}
          </nav>

          <div className="ml-auto hidden shrink-0 items-center gap-2 md:flex">
            {isHome ? (
              <>
                <Link
                  href="/factory/models"
                  className="rounded-lg px-3 py-2 text-sm text-black/65 transition hover:bg-black/[0.04] hover:text-black"
                >
                  能力详情
                </Link>
                <Link
                  href="/dashboard"
                  className="btn-base btn-dark btn-m"
                >
                  进入工厂
                  <ArrowRight className="h-3.5 w-3.5" strokeWidth={1.5} />
                </Link>
              </>
            ) : (
              <Link
                href="/dashboard"
                className="btn-base btn-light btn-m"
              >
                控制台
                <ArrowRight className="h-3.5 w-3.5" strokeWidth={1.5} />
              </Link>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}
