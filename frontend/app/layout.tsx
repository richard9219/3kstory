import type { Metadata } from 'next';
import './globals.css';
import TopNavigation from '@/components/common/TopNavigation';

export const metadata: Metadata = {
  title: '3K STORY | AI Video Factory',
  description: '3K STORY AI Video Factory',
  icons: {
    icon: '/images/brand-mark.svg',
    shortcut: '/images/brand-mark.svg',
    apple: '/images/brand-mark.png',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN">
      <body>
        <TopNavigation />
        {children}
      </body>
    </html>
  );
}
