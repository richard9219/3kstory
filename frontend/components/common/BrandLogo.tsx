import Image from 'next/image';
import Link from 'next/link';

export default function BrandLogo({ compact = false }: { compact?: boolean }) {
  return (
    <Link href="/" className="inline-flex items-center gap-3 group">
      <span className="relative inline-flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl border border-black/12 bg-white">
        <Image src="/images/brand-mark.svg" alt="三千视界" width={34} height={34} priority />
      </span>
      {!compact && (
        <span className="hidden leading-none sm:block">
          <span className="block text-[15px] font-semibold tracking-[0.04em] text-black">3K STORY</span>
          <span className="mt-1 block text-[10px] uppercase tracking-[0.22em] text-black/45">AI VIDEO FACTORY</span>
        </span>
      )}
    </Link>
  );
}
