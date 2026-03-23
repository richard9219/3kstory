'use client';

import { Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { CheckCircle, XCircle } from 'lucide-react';

export default function PlatformBoundPage() {
  return (
    <Suspense fallback={<PlatformBoundFallback />}>
      <PlatformBoundContent />
    </Suspense>
  );
}

function PlatformBoundContent() {
  const searchParams = useSearchParams();
  const ok = searchParams.get('ok') === '1';
  const err = searchParams.get('err') || '';

  return (
    <main className="min-h-screen bg-[#f7f7f7] flex items-center justify-center px-4">
      <div className="text-center max-w-md mx-auto px-4 py-8 rounded-2xl site-card">
        {ok ? (
          <>
            <CheckCircle className="w-16 h-16 text-black/70 mx-auto mb-4" />
            <h1 className="text-2xl font-bold text-black mb-2">绑定成功</h1>
            <p className="text-black/55 mb-6">平台账号已成功绑定到您的 3kstory 账号。</p>
          </>
        ) : (
          <>
            <XCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
            <h1 className="text-2xl font-bold text-black mb-2">绑定未完成</h1>
            <p className="text-black/55 mb-2">
              {err ? decodeURIComponent(err) : '授权已取消或发生错误，请重试。'}
            </p>
            <p className="text-black/45 text-sm mb-6">请确保已在对应开放平台完成应用配置与回调地址。</p>
          </>
        )}
        <Link
          href="/platforms"
          className="btn-base btn-dark btn-m"
        >
          返回平台账号
        </Link>
      </div>
    </main>
  );
}

function PlatformBoundFallback() {
  return (
    <main className="min-h-screen bg-[#f7f7f7] flex items-center justify-center px-4">
      <div className="text-center max-w-md mx-auto px-4 py-8 rounded-2xl site-card">
        <h1 className="text-2xl font-bold text-black mb-2">正在读取绑定结果</h1>
        <p className="text-black/55">请稍候，页面正在恢复授权状态。</p>
      </div>
    </main>
  );
}
