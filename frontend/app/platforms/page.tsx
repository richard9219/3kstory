'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useAuthStore } from '@/lib/store/authStore';
import { platformAPI } from '@/lib/api/client';
import type { PlatformAccount, PlatformKind } from '@/types';
import { Video, Link2, Unlink, Loader2, ArrowLeft, Lock } from 'lucide-react';

const PLATFORMS: { id: PlatformKind; name: string; desc: string }[] = [
  { id: 'douyin', name: '抖音', desc: '绑定后可将作品发布到抖音' },
  { id: 'xiaohongshu', name: '小红书', desc: '绑定后可将作品发布到小红书' },
  { id: 'bilibili', name: 'B站', desc: '绑定后可将作品发布到 Bilibili' },
  { id: 'weibo', name: '微博', desc: '绑定后可将作品发布到微博' },
];

export default function PlatformsPage() {
  const { isAuthenticated } = useAuthStore();
  const [offlineMode, setOfflineMode] = useState(false);
  const [list, setList] = useState<PlatformAccount[]>([]);
  const [configured, setConfigured] = useState<PlatformKind[]>([]);
  const [loading, setLoading] = useState(true);
  const [binding, setBinding] = useState<PlatformKind | null>(null);
  const [unbinding, setUnbinding] = useState<PlatformKind | null>(null);
  const [bindError, setBindError] = useState<string | null>(null);

  useEffect(() => {
    platformAPI.configured().then((res) => setConfigured(res.data.data || [])).catch(() => setConfigured([]));
  }, []);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    setOfflineMode(sessionStorage.getItem('frontend-offline-fallback') === '1');
  }, []);

  useEffect(() => {
    if (!isAuthenticated) {
      setLoading(false);
      return;
    }
    platformAPI
      .list()
      .then((res) => setList(res.data.data || []))
      .catch(() => setList([]))
      .finally(() => setLoading(false));
  }, [isAuthenticated]);

  const connectedSet = new Set(list.map((a) => a.platform));
  const configuredSet = new Set(configured);

  const handleBind = (platform: PlatformKind) => {
    if (!configuredSet.has(platform)) return;
    setBindError(null);
    setBinding(platform);
    platformAPI
      .getAuthorizeUrl(platform)
      .then((res) => {
        window.location.href = res.data.authorize_url;
      })
      .catch((e: { response?: { data?: { error?: string } }; message?: string }) => {
        setBinding(null);
        setBindError(e.response?.data?.error || e.message || '获取授权链接失败，请稍后重试');
      });
  };

  const handleDisconnect = (platform: PlatformKind) => {
    if (!confirm(`确定要解绑${PLATFORMS.find((p) => p.id === platform)?.name}吗？`)) return;
    setUnbinding(platform);
    platformAPI
      .disconnect(platform)
      .then(() => setList((prev) => prev.filter((a) => a.platform !== platform)))
      .finally(() => setUnbinding(null));
  };

  if (!isAuthenticated) {
    return (
      <main className="min-h-screen bg-[#f7f7f7] flex items-center justify-center">
        <div className="text-center max-w-md mx-auto px-4">
          <Video className="w-16 h-16 text-black/70 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-black mb-2">平台账号</h1>
          <p className="text-black/55 mb-6">请先登录后再绑定抖音、小红书、B站、微博等平台账号。</p>
          <Link
            href="/"
            className="btn-base btn-dark btn-m"
          >
            返回首页
          </Link>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#f7f7f7] px-4 pb-12">
      <div className="max-w-2xl mx-auto">
        <section className="min-h-[calc(100vh-96px)] flex items-center justify-center py-10">
          <div className="w-full max-w-2xl text-center">
            <Link
              href="/"
              className="inline-flex items-center gap-2 text-black/55 hover:text-black transition"
            >
              <ArrowLeft className="w-4 h-4" />
              返回首页
            </Link>
            <div className="ui-mt-16 inline-flex h-14 w-14 items-center justify-center rounded-2xl border border-black/10 bg-white text-black/75">
              <Link2 className="w-6 h-6" />
            </div>
            <h1 className="site-h1 ui-mt-12">平台账号</h1>
            <p className="site-lead ui-mt-12 mx-auto max-w-xl text-black/58">
              绑定后可在发布作品时选择对应平台一键发布。授权信息仅用于发布，我们不会用于其他用途。
            </p>
          </div>
        </section>
        {offlineMode ? (
          <div className="mb-4 rounded-xl border border-amber-300/30 bg-amber-300/10 px-4 py-3 text-sm text-amber-800">
            当前处于演示模式：平台绑定数据可能来自本地兜底数据。
          </div>
        ) : null}
        {bindError && (
          <div className="mb-6 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
            {bindError}
          </div>
        )}

        {loading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="w-8 h-8 text-black/70 animate-spin" />
          </div>
        ) : (
          <ul className="space-y-4">
            {PLATFORMS.map((p) => {
              const connected = connectedSet.has(p.id);
              const acc = list.find((a) => a.platform === p.id);
              const isConfigured = configuredSet.has(p.id);
              return (
                <li
                  key={p.id}
                  className={`flex items-center justify-between p-4 rounded-xl border transition ${
                    isConfigured ? 'site-card border-black/10' : 'site-card border-black/5 opacity-75'
                  }`}
                >
                  <div className="flex items-center gap-4">
                    {acc?.avatar_url ? (
                      <div className="relative w-12 h-12 rounded-lg overflow-hidden bg-black/[0.04] flex-shrink-0">
                        {/* eslint-disable-next-line @next/next/no-img-element */}
                        <img src={acc.avatar_url} alt="" className="w-full h-full object-cover" referrerPolicy="no-referrer" />
                      </div>
                    ) : (
                      <div className="w-12 h-12 rounded-lg bg-black/[0.04] flex items-center justify-center flex-shrink-0">
                        <Video className="w-6 h-6 text-black/70" />
                      </div>
                    )}
                    <div>
                      <div className="font-medium text-black flex items-center gap-2">
                        {p.name}
                        {!isConfigured && (
                          <span className="text-xs text-gray-500 font-normal flex items-center gap-1">
                            <Lock className="w-3 h-3" /> 暂未开放
                          </span>
                        )}
                      </div>
                      <div className="text-sm text-black/55">{p.desc}</div>
                      {acc?.nickname && (
                        <div className="text-sm text-black/60 mt-1">已绑定：{acc.nickname}</div>
                      )}
                    </div>
                  </div>
                  <div>
                    {connected ? (
                      <button
                        onClick={() => handleDisconnect(p.id)}
                        disabled={!!unbinding}
                        className="btn-base btn-m border border-red-500/50 text-red-500 hover:bg-red-500/10 disabled:opacity-50"
                      >
                        {unbinding === p.id ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <Unlink className="w-4 h-4" />
                        )}
                        解绑
                      </button>
                    ) : (
                      <button
                        onClick={() => handleBind(p.id)}
                        disabled={!!binding || !isConfigured}
                        className="btn-base btn-dark btn-m disabled:opacity-50"
                      >
                        {binding === p.id ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <Link2 className="w-4 h-4" />
                        )}
                        {isConfigured ? '绑定' : '暂未开放'}
                      </button>
                    )}
                  </div>
                </li>
              );
            })}
          </ul>
        )}
      </div>
    </main>
  );
}
