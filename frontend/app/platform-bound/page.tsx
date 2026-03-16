'use client';

import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { CheckCircle, XCircle } from 'lucide-react';

export default function PlatformBoundPage() {
  const searchParams = useSearchParams();
  const ok = searchParams.get('ok') === '1';
  const err = searchParams.get('err') || '';

  return (
    <main className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-900/20 via-blue-900/20 to-black">
      <div className="text-center max-w-md mx-auto px-4">
        {ok ? (
          <>
            <CheckCircle className="w-16 h-16 text-green-500 mx-auto mb-4" />
            <h1 className="text-2xl font-bold text-white mb-2">绑定成功</h1>
            <p className="text-gray-400 mb-6">平台账号已成功绑定到您的 3kstory 账号。</p>
          </>
        ) : (
          <>
            <XCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
            <h1 className="text-2xl font-bold text-white mb-2">绑定未完成</h1>
            <p className="text-gray-400 mb-2">
              {err ? decodeURIComponent(err) : '授权已取消或发生错误，请重试。'}
            </p>
            <p className="text-gray-500 text-sm mb-6">请确保已在对应开放平台完成应用配置与回调地址。</p>
          </>
        )}
        <Link
          href="/platforms"
          className="inline-block px-6 py-3 rounded-lg bg-purple-600 hover:bg-purple-700 text-white font-medium transition"
        >
          返回平台账号
        </Link>
      </div>
    </main>
  );
}
