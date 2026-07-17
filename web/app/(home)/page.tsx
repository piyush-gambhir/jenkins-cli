import { ArrowRight } from 'lucide-react';
import Link from 'next/link';
import { HomeHero } from '@/components/home-hero';
import { InstallCommand } from '@/components/install-command';
import { Reveal } from '@/components/reveal';
import { SiteFooter } from '@/components/site-footer';
import { OsmoButton } from '@/components/ui/osmo-button';
import { site } from '@/lib/site';
import { siteUrl } from '@/lib/shared';
import { getOtherSuiteProjects } from '@/lib/suite';

const revealDelays = ['0s', '0.075s', '0.15s'] as const;
const featureLinks = [
  '/docs/commands/jobs-builds',
  '/docs/authentication',
  '/docs/agents',
  '/docs/commands/pipelines-administration',
  '/docs/quickstart',
  '/docs/commands/nodes-views-queue',
] as const;

export default function HomePage() {
  const repoUrl = `https://github.com/${site.repo}`;
  const relatedLink = getOtherSuiteProjects(site.repo).map(({ href }) => href);
  const structuredData = {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'SoftwareApplication',
        name: site.name,
        applicationCategory: 'DeveloperApplication',
        operatingSystem: 'macOS, Linux',
        description: site.description,
        url: siteUrl,
        downloadUrl: `${repoUrl}/releases`,
        license: `${repoUrl}/blob/main/LICENSE`,
        sameAs: repoUrl,
        relatedLink,
        featureList: [
          'Structured JSON and YAML output for coding agents',
          'Read-only safety mode',
          'Non-interactive automation flags',
          'Works with any coding agent or agent harness that can run shell commands',
        ],
        keywords:
          'coding agent, AI agent CLI, agent harness, MCP-free shell integration, terminal automation, jenkins automation',
      },
      {
        '@type': 'WebSite',
        name: site.name,
        url: siteUrl,
        description: site.description,
        sameAs: repoUrl,
        relatedLink,
      },
    ],
  };

  return (
    <main className="osmo-home flex flex-1 flex-col">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData) }}
      />
      <HomeHero />

      {/* Stack strip */}
      {site.compatible && site.compatible.length > 0 ? (
        <section className="osmo-section">
          <div className="osmo-container">
            <Reveal className="compatible-marquee">
              <div className="compatible-marquee__track">
                {[false, true].map((hidden) => (
                  <span
                    className="compatible-marquee__list"
                    aria-hidden={hidden || undefined}
                    key={String(hidden)}
                  >
                    {site.compatible?.map((item) => (
                      <span className="compatible-marquee__item" key={item}>
                        {item}
                        <span aria-hidden>{' · '}</span>
                      </span>
                    ))}
                  </span>
                ))}
              </div>
            </Reveal>
          </div>
        </section>
      ) : null}

      {/* Features */}
      <section
        className="osmo-section osmo-section--features"
        data-theme-section="dark"
        aria-labelledby="features-heading"
      >
        <div className="osmo-container">
          <Reveal className="osmo-section__header">
            <h2 id="features-heading" className="osmo-section__title">
              {site.featuresTitle ?? 'Everything, from one binary'}
            </h2>
            <p className="osmo-home-hero__description">
              {site.featuresSubtitle ??
                'Built for humans at the keyboard and coding agents alike.'}
            </p>
          </Reveal>

          <div className="osmo-card-grid osmo-card-grid--features">
            {site.features.map(({ title, body }, index) => (
              <Reveal
                key={title}
                delay={revealDelays[index % revealDelays.length]}
                className="osmo-card osmo-feature-card"
              >
                <span className="osmo-eyebrow osmo-card__number">
                  {String(index + 1).padStart(2, '0')}
                </span>
                <h3 className="osmo-card__title">
                  <Link href={featureLinks[index] ?? '/docs'}>{title}</Link>
                </h3>
                <p className="osmo-card__body">{body}</p>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* CTA band */}
      <section className="osmo-section">
        <div className="osmo-container">
          <Reveal className="osmo-section__header mx-auto items-center text-center">
            <h2 className="osmo-section__title">Ready in one command</h2>
            <p className="osmo-home-hero__description">
              {site.ctaBody ?? (
                <>
                  <Link href="/docs/installation">Install the binary</Link>,{' '}
                  <Link href="/docs/authentication">authenticate</Link>, and start
                  querying. No runtime, no dependencies.
                </>
              )}
            </p>
            <div className="osmo-home-hero__install">
              <InstallCommand command={site.installCommand} />
            </div>
            <div className="osmo-home-hero__actions">
              <OsmoButton
                href="/docs"
                aria-label="Read the docs"
                icon={<ArrowRight />}
              >
                Read the docs
              </OsmoButton>
            </div>
          </Reveal>
        </div>
      </section>

      <SiteFooter />
    </main>
  );
}
