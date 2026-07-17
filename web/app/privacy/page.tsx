import type { Metadata } from 'next';
import Link from 'next/link';
import { LegalPage } from '@/components/legal-page';
import { createPageMetadata } from '@/lib/metadata';

export const metadata: Metadata = createPageMetadata({
  title: 'Privacy Policy',
  description:
    'Privacy practices for Jenkins CLI, an independent, unofficial open-source command-line client for Jenkins.',
  path: '/privacy',
});

export default function PrivacyPage() {
  return (
    <LegalPage
      title="Privacy Policy"
      intro={
        <>
          Jenkins CLI (the <code>jenkins</code> command-line tool) is an open-source
          program that runs entirely on your own computer. It does <strong>not</strong>{' '}
          collect, transmit, or store your personal data on any server operated by the
          maintainer.
        </>
      }
    >
      <h2>1. No data collection</h2>
      <p>
        The CLI runs locally on your machine. The maintainer operates <strong>no backend
        servers</strong> and receives <strong>no data</strong> from your use of the tool.
        There is no analytics, no telemetry, no tracking, and no advertising.
      </p>

      <h2>2. Credentials &amp; local storage</h2>
      <p>
        Any credentials you provide, API tokens, personal access tokens, OAuth
        access/refresh tokens, or passwords, are stored <strong>only on your device</strong>,
        in a local configuration file in your home directory (typically{' '}
        <code>~/.config/jenkins-cli/config.yaml</code>, created with owner-only{' '}
        <code>0600</code> permissions). They are never sent to the maintainer or any third
        party.
      </p>

      <h2>3. Network connections</h2>
      <p>The CLI makes outbound network requests only to:</p>
      <ul>
        <li>
          <strong>Your Jenkins instance</strong>, the server you configure, to perform the
          actions you explicitly request.
        </li>
        <li>
          <strong>GitHub&apos;s public API</strong>, to check whether a newer release of the CLI
          is available. This request contains no personal data.
        </li>
      </ul>
      <p>The maintainer is not a party to, and cannot observe, these connections.</p>

      <h2>4. Data you access through the tool</h2>
      <p>
        When authenticated, the CLI reads and writes data in your Jenkins instance using
        the permissions you grant, solely to execute the commands you run. That data is
        shown in your terminal (or written to files you specify) and is <strong>not</strong>{' '}
        retained, copied, or transmitted anywhere by the maintainer.
      </p>

      <h2>5. Third parties</h2>
      <p>
        The maintainer does not sell, rent, or share any data. The tool integrates no
        third-party analytics or tracking SDKs. Your use of Jenkins is governed by that
        service provider&apos;s own privacy policy.
      </p>

      <h2>6. Children</h2>
      <p>This tool is a developer utility and is not directed at children under 13.</p>

      <h2>7. Changes to this policy</h2>
      <p>Any changes will be posted on this page with an updated effective date.</p>

      <h2>8. Contact</h2>
      <p>
        Questions about this policy? Contact <strong>Piyush Gambhir</strong> at{' '}
        <a href="mailto:developer.piyushgambhir@gmail.com">
          developer.piyushgambhir@gmail.com
        </a>
        , or see the <Link href="/contact">contact page</Link>.
      </p>
    </LegalPage>
  );
}
