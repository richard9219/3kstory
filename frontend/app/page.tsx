'use client';

import { useState } from 'react';
import HeroSection from '@/components/hero/HeroSection';
import LandingSections from '@/components/hero/LandingSections';
import AuthModal from '@/components/auth/AuthModal';

export default function HomePage() {
  const [isAuthModalOpen, setIsAuthModalOpen] = useState(false);
  const [authMode, setAuthMode] = useState<'login' | 'register'>('login');

  const openAuthModal = (mode: 'login' | 'register') => {
    setAuthMode(mode);
    setIsAuthModalOpen(true);
  };

  return (
    <main className="h-[calc(100vh-64px)] overflow-y-auto snap-y snap-mandatory">
      <HeroSection onRegister={() => openAuthModal('register')} />
      <LandingSections onRegister={() => openAuthModal('register')} />
      <AuthModal 
        isOpen={isAuthModalOpen} 
        onClose={() => setIsAuthModalOpen(false)} 
        initialMode={authMode}
      />
    </main>
  );
}
