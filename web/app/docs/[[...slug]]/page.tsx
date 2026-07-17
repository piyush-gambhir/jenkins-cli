import { getPageImage, getPageMarkdownUrl, source } from '@/lib/source';
import {
  DocsBody,
  DocsDescription,
  DocsPage,
  DocsTitle,
  MarkdownCopyButton,
  ViewOptionsPopover,
} from 'fumadocs-ui/layouts/docs/page';
import { notFound } from 'next/navigation';
import { getMDXComponents } from '@/components/mdx';
import type { Metadata } from 'next';
import { createRelativeLink } from 'fumadocs-ui/mdx';
import { gitConfig } from '@/lib/shared';
import { site } from '@/lib/site';
import { siteUrl } from '@/lib/shared';
import { createPageMetadata, siteMetadataDescription } from '@/lib/metadata';

function getMetadataDescription(description?: string) {
  const base =
    'Independent, unofficial Jenkins CLI for any shell-based coding agent harness.';
  return description ? `${description} ${base}` : base;
}

export default async function Page(props: PageProps<'/docs/[[...slug]]'>) {
  const params = await props.params;
  const page = source.getPage(params.slug);
  if (!page) notFound();

  const MDX = page.data.body;
  const markdownUrl = `${siteUrl}${getPageMarkdownUrl(page).url}`;
  const canonicalUrl = `${siteUrl}${page.url}`;
  const metadataDescription = getMetadataDescription(page.data.description);
  const breadcrumbItems = [
    { '@type': 'ListItem', position: 1, name: 'Home', item: siteUrl },
    {
      '@type': 'ListItem',
      position: 2,
      name: 'Documentation',
      item: `${siteUrl}/docs`,
    },
    ...(page.url === '/docs'
      ? []
      : [{ '@type': 'ListItem', position: 3, name: page.data.title, item: canonicalUrl }]),
  ];
  const structuredData = {
    '@context': 'https://schema.org',
    '@graph': [
      { '@type': 'BreadcrumbList', itemListElement: breadcrumbItems },
      {
        '@type': 'TechArticle',
        headline: page.data.title,
        description: metadataDescription,
        url: canonicalUrl,
        mainEntityOfPage: canonicalUrl,
        author: { '@type': 'Person', name: 'Piyush Gambhir' },
        publisher: { '@type': 'Organization', name: site.name, url: siteUrl },
      },
    ],
  };

  return (
    <DocsPage toc={page.data.toc} full={page.data.full}>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData) }}
      />
      <DocsTitle>{page.data.title}</DocsTitle>
      <DocsDescription className="mb-0">{page.data.description}</DocsDescription>
      <div className="flex flex-row gap-2 items-center pb-6">
        <MarkdownCopyButton markdownUrl={markdownUrl} />
        <ViewOptionsPopover
          markdownUrl={markdownUrl}
          githubUrl={`https://github.com/${gitConfig.user}/${gitConfig.repo}/blob/${gitConfig.branch}/content/docs/${page.path}`}
        />
      </div>
      <DocsBody>
        <MDX
          components={getMDXComponents({
            // this allows you to link to other pages with relative file paths
            a: createRelativeLink(source, page),
          })}
        />
      </DocsBody>
    </DocsPage>
  );
}

export async function generateStaticParams() {
  return source.generateParams();
}

export async function generateMetadata(props: PageProps<'/docs/[[...slug]]'>): Promise<Metadata> {
  const params = await props.params;
  const page = source.getPage(params.slug);
  if (!page) notFound();

  return createPageMetadata({
    title: page.data.title,
    description: getMetadataDescription(page.data.description),
    socialDescription: siteMetadataDescription,
    path: page.url,
    image: getPageImage(page).url,
    type: 'article',
  });
}
