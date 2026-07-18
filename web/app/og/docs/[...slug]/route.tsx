import { getPageImage, source } from '@/lib/source';
import { notFound } from 'next/navigation';
import { ImageResponse } from 'next/og';
import { site } from '@/lib/site';

import { readFile } from 'node:fs/promises';
import { join } from 'node:path';

const fontBuffer = async (...fontPath: string[]) => {
  const data = await readFile(join(process.cwd(), 'node_modules', ...fontPath));
  return data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength) as ArrayBuffer;
};

export const revalidate = false;

const inter = fontBuffer('@fontsource', 'inter', 'files', 'inter-latin-400-normal.woff');
const jetbrainsMono = fontBuffer('@fontsource', 'jetbrains-mono', 'files', 'jetbrains-mono-latin-500-normal.woff');

export async function GET(_req: Request, { params }: RouteContext<'/og/docs/[...slug]'>) {
  const { slug } = await params;
  const page = source.getPage(slug.slice(0, -1));
  if (!page) notFound();
  const [interData, jetbrainsMonoData] = await Promise.all([inter, jetbrainsMono]);

  return new ImageResponse(
    <div
      style={{
        display: 'flex',
        width: '100%',
        height: '100%',
        flexDirection: 'column',
        justifyContent: 'space-between',
        padding: '68px 76px',
        color: '#f3f4f1',
        background: '#131412',
        fontFamily: 'Inter',
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          color: site.accent,
          fontFamily: 'JetBrains Mono',
          fontSize: 24,
          letterSpacing: '0.08em',
        }}
      >
        &gt;_ {site.binary} docs
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', maxWidth: 1000 }}>
        <div style={{ fontSize: 76, lineHeight: 0.96, letterSpacing: '-0.045em' }}>
          {page.data.title}
        </div>
        <div
          style={{
            marginTop: 28,
            color: '#b6b8b3',
            fontSize: 28,
            lineHeight: 1.35,
          }}
        >
          {page.data.description}
        </div>
      </div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          color: '#7f827b',
          fontFamily: 'JetBrains Mono',
          fontSize: 20,
        }}
      >
        <span>{site.name}</span>
        <span style={{ color: site.accent }}>command your Jenkins instance</span>
      </div>
    </div>,
    {
      width: 1200,
      height: 630,
      fonts: [
        { name: 'Inter', data: interData, weight: 400 },
        { name: 'JetBrains Mono', data: jetbrainsMonoData, weight: 500 },
      ],
    },
  );
}

export function generateStaticParams() {
  return source.getPages().map((page) => ({
    lang: page.locale,
    slug: getPageImage(page).segments,
  }));
}
