import { source } from '@/lib/source';
import { llms } from 'fumadocs-core/source';
import { site } from '@/lib/site';
import { getOtherSuiteProjects } from '@/lib/suite';

export const revalidate = false;

export function GET() {
  const preamble =
    'Jenkins CLI is agent-ready and harness-agnostic: it works with Claude Code, OpenAI Codex, Cursor, or any agent harness that can run shell commands. Structured JSON/YAML output, read-only safety mode, and no-input flags let coding agents manage Jenkins jobs, builds, and pipelines from the terminal.';
  const related = getOtherSuiteProjects(site.repo)
    .map(({ name, href }) => `- ${name}: ${href}`)
    .join('\n');

  return new Response(
    `${preamble}\n\n${llms(source).index()}\n\n## Related CLI projects\n\n${related}\n`,
  );
}
