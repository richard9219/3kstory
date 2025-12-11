'use client';

import { useState } from 'react';
import HeroSection from '@/components/hero/HeroSection';
import AuthModal from '@/components/auth/AuthModal';

export default function HomePage() {
  const [isAuthModalOpen, setIsAuthModalOpen] = useState(false);
  const [authMode, setAuthMode] = useState<'login' | 'register'>('login');

  const openAuthModal = (mode: 'login' | 'register') => {
    setAuthMode(mode);
    setIsAuthModalOpen(true);
  };

  return (
    <main>
      <HeroSection />
      <AuthModal 
        isOpen={isAuthModalOpen} 
        onClose={() => setIsAuthModalOpen(false)} 
        initialMode={authMode}
      />
    </main>
  );
}
