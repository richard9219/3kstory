'use client';

import type { ReactNode } from 'react';
import { useState } from 'react';

export default function MagneticButton({
  children,
  className = '',
  onClick,
}: {
  children: ReactNode;
  className?: string;
  onClick?: () => void;
}) {
  const [offset, setOffset] = useState({ x: 0, y: 0 });

  const handleMove = (e: React.MouseEvent<HTMLButtonElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX - (rect.left + rect.width / 2);
    const y = e.clientY - (rect.top + rect.height / 2);
    setOffset({
      x: Math.max(Math.min(x * 0.12, 8), -8),
      y: Math.max(Math.min(y * 0.12, 8), -8),
    });
  };

  const reset = () => setOffset({ x: 0, y: 0 });

  return (
    <button
      type="button"
      onClick={onClick}
      onMouseMove={handleMove}
      onMouseLeave={reset}
      onBlur={reset}
      style={{ transform: `translate3d(${offset.x}px, ${offset.y}px, 0)` }}
      className={`transition-transform duration-200 ease-out ${className}`}
    >
      {children}
    </button>
  );
}

