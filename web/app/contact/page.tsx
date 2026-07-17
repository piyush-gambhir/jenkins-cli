import type { Metadata } from 'next';
import Link from 'next/link';
import { LegalPage } from '@/components/legal-page';
import { createPageMetadata } from '@/lib/metadata';

export const metadata: Metadata = createPageMetadata({
  title: 'Contact',
  description:
    'Contact the maintainer of Jenkins CLI, an independent, unofficial open-source command-line client for Jenkins.',
  path: '/contact',
});

export default function ContactPage() {
  return (
    <LegalPage
      title="Contact"
      intro={
        <>
          Jenkins CLI is a free, open-source project maintained by{' '}
          <strong>Piyush Gambhir</strong>. Support is best-effort, here are the best ways
          to get in touch.
        </>
      }
    >
      <div className="contact-grid">
        <div className="contact-card">
          <p className="contact-card__label">Email</p>
          <p className="contact-card__value">
            <a href="mailto:developer.piyushgambhir@gmail.com">
              developer.piyushgambhir@gmail.com
            </a>
          </p>
          <p className="contact-card__note">General questions, privacy, and security reports.</p>
        </div>
        <div className="contact-card">
          <p className="contact-card__label">Bugs &amp; features</p>
          <p className="contact-card__value">
            <a
              href="https://github.com/piyush-gambhir/jenkins-cli/issues"
              target="_blank"
              rel="noreferrer"
            >
              GitHub Issues ↗
            </a>
          </p>
          <p className="contact-card__note">
            The fastest way to report a bug or request a feature.
          </p>
        </div>
        <div className="contact-card">
          <p className="contact-card__label">Source</p>
          <p className="contact-card__value">
            <a
              href="https://github.com/piyush-gambhir/jenkins-cli"
              target="_blank"
              rel="noreferrer"
            >
              piyush-gambhir/jenkins-cli ↗
            </a>
          </p>
          <p className="contact-card__note">Read the code, open a pull request, or fork it.</p>
        </div>
      </div>

      <h2>Security issues</h2>
      <p>
        If you believe you&apos;ve found a security vulnerability, please email{' '}
        <a href="mailto:developer.piyushgambhir@gmail.com">
          developer.piyushgambhir@gmail.com
        </a>{' '}
        with the details rather than opening a public issue. Jenkins CLI stores
        credentials only on your own device and operates no servers, but responsible
        disclosure is always appreciated.
      </p>

      <h2>Response time</h2>
      <p>
        This is an independent side project, not a commercial product. The maintainer
        aims to respond when possible, but no response time or level of support is
        guaranteed. See the <Link href="/terms">Terms of Service</Link> for the full
        no-warranty terms.
      </p>

      <h2>Not affiliated with Jenkins</h2>
      <p>
        Jenkins CLI is an independent, unofficial tool and is not affiliated with,
        endorsed by, or sponsored by Jenkins or its vendor. For issues with Jenkins
        itself, contact that vendor&apos;s own support channels.
      </p>
    </LegalPage>
  );
}
