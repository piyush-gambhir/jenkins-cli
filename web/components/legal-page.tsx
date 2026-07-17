import type { ReactNode } from 'react';
import { FloatingHeader } from '@/components/floating-header';
import { SiteFooter } from '@/components/site-footer';

interface LegalPageProps {
  title: string;
  intro: ReactNode;
  children: ReactNode;
}

export function LegalPage({ title, intro, children }: LegalPageProps) {
  return (
    <div className="marketing-shell site-route-shell">
      <FloatingHeader />
      <main className="legal-page">
        <div className="legal-page__inner">
          <header className="legal-page__header">
            <h1>{title}</h1>
            <p className="legal-page__effective">Effective June 14, 2026</p>
          </header>
          <p className="legal-page__lede">{intro}</p>
          <div className="legal-page__content">{children}</div>
        </div>
      </main>
      <SiteFooter />
    </div>
  );
}
