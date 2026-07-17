import type { Metadata } from 'next';
import { site } from '@/lib/site';
import { siteUrl } from '@/lib/shared';

interface PageMetadataOptions {
  title: string;
  description: string;
  socialDescription?: string;
  path: string;
  image?: string;
  type?: 'article' | 'website';
}

export const siteMetadataDescription =
  'Independent, unofficial open-source Jenkins CLI, agent-ready and harness-agnostic, with JSON/YAML output, read-only safety, and no-input automation.';

export function createPageMetadata({
  title,
  description,
  socialDescription = description,
  path,
  image = '/og/docs/image.png',
  type = 'website',
}: PageMetadataOptions): Metadata {
  const socialTitle = `${title} · ${site.name}`;
  const canonicalUrl = `${siteUrl}${path}`;
  const socialImage = {
    url: `${siteUrl}${image}`,
    width: 1200,
    height: 630,
    alt: `${title} for ${site.name}`,
  };

  return {
    title,
    description,
    alternates: { canonical: canonicalUrl },
    openGraph: {
      type,
      url: canonicalUrl,
      siteName: site.name,
      title: socialTitle,
      description: socialDescription,
      images: [socialImage],
    },
    twitter: {
      card: 'summary_large_image',
      title: socialTitle,
      description: socialDescription,
      images: [socialImage],
    },
  };
}
