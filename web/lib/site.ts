import {
  Bot,
  GitBranch,
  KeyRound,
  ListChecks,
  Search,
  Zap,
  type LucideIcon,
} from 'lucide-react';

export interface Feature {
  icon: LucideIcon;
  title: string;
  body: string;
}

export interface SiteConfig {
  /** Display name, e.g. "Acme CLI" */
  name: string;
  /** The binary invoked in examples, e.g. "acme" */
  binary: string;
  /** GitHub "owner/repo" */
  repo: string;
  /** One-line hero heading */
  tagline: string;
  /** Hero sub-paragraph */
  description: string;
  /** Small pill above the heading */
  badge: string;
  /** One-line install command shown in the hero */
  installCommand: string;
  /** Feature cards */
  features: Feature[];
  /** Title above the code block */
  exampleTitle: string;
  /** Shell example rendered in the terminal card */
  example: string;
  /** Optional: tech / query languages this CLI speaks (logo strip) */
  compatible?: string[];
  /** Optional: features section heading (default: "Everything, from one binary") */
  featuresTitle?: string;
  /** Optional: features section subheading */
  featuresSubtitle?: string;
  /** Optional: CTA band body (default mentions installing the binary) */
  ctaBody?: string;
  /** Optional: per-site accent expressed as an OKLCH color */
  accent?: string;
  /** Optional: human-readable accent name */
  accentName?: string;
  /** Optional: hex twin of the accent, for surfaces without oklch() support (OG images) */
  accentHex?: string;
}

export const site: SiteConfig = {
  name: 'Jenkins CLI',
  binary: 'jenkins',
  repo: 'piyush-gambhir/jenkins-cli',
  tagline: 'Jenkins from your terminal',
  description:
    'An independent, unofficial open-source CLI for Jenkins. Manage jobs, builds, agents, queues, views, plugins, credentials, pipelines, and system operations, built for humans and coding agents alike.',
  badge: 'Open-source · Agent-friendly',
  accent: 'oklch(0.71 0.16 30)',
  accentName: 'coral',
  accentHex: '#f57663',
  installCommand:
    'curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/jenkins-cli/main/install.sh | sh',
  features: [
    {
      icon: Search,
      title: 'Jobs & builds',
      body: 'Discover nested jobs, trigger parameterized builds, follow console output, and inspect stages, tests, artifacts, and environments.',
    },
    {
      icon: KeyRound,
      title: 'Token authentication',
      body: 'Connect with a Jenkins username and API token. Keep multiple named profiles or configure credentials entirely through environment variables.',
    },
    {
      icon: Bot,
      title: 'Agent-friendly',
      body: '-o json|yaml for structured reads, --read-only safety mode, --no-input, idempotency flags, and structured JSON errors.',
    },
    {
      icon: GitBranch,
      title: 'Pipeline operations',
      body: 'Validate declarative Jenkinsfiles, trigger builds with parameters, inspect stage timing, and manage pending pipeline inputs.',
    },
    {
      icon: Zap,
      title: 'Fast & scriptable',
      body: 'Install a single prebuilt binary, stream logs in real time, pipe XML and scripts through stdin, and update from GitHub Releases.',
    },
    {
      icon: ListChecks,
      title: 'Administration included',
      body: 'Manage nodes, queues, views, plugins, credentials, users, quiet-down mode, restarts, and Groovy scripts from one CLI.',
    },
  ],
  exampleTitle: 'A seven-line tour',
  example: `# Authenticate with a username and API token
jenkins login
# Inspect the server and failed jobs as JSON
jenkins status -o json
jenkins job list --recursive --status FAILURE -o json
# Trigger a parameterized build and wait for its log
jenkins job build team/deploy --param ENV=staging --wait --follow`,
  compatible: [
    "REST API",
    "Pipelines",
    "Jobs",
    "Nodes",
    "Views",
    "CSRF crumbs",
  ],
};
