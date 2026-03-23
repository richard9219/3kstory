import { AlertTriangle, Inbox, Loader2 } from 'lucide-react';

export function LoadingBlock({ text = '加载中...' }: { text?: string }) {
  return (
    <div className="rounded-2xl site-card p-5 text-black/65 inline-flex items-center gap-2">
      <Loader2 className="w-4 h-4 animate-spin" />
      {text}
    </div>
  );
}

export function EmptyBlock({ title, hint }: { title: string; hint?: string }) {
  return (
    <div className="rounded-2xl site-card p-6 text-center">
      <Inbox className="w-10 h-10 text-black/35 mx-auto mb-3" />
      <div className="text-black font-medium mb-1">{title}</div>
      {hint ? <div className="text-sm text-black/55">{hint}</div> : null}
    </div>
  );
}

export function ErrorBlock({ text }: { text: string }) {
  return (
    <div className="rounded-2xl site-card p-4 border border-red-300 text-red-700 inline-flex items-center gap-2">
      <AlertTriangle className="w-4 h-4" />
      {text}
    </div>
  );
}
