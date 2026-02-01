import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  description: ReactNode;
  icon: string;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Sub-Millisecond Redirects',
    icon: '\u26a1',
    description: (
      <>
        High-performance redirect service built in Go with multi-layer
        caching (L1/L2), achieving sub-millisecond response times at scale.
      </>
    ),
  },
  {
    title: 'Rich Analytics',
    icon: '\ud83d\udcca',
    description: (
      <>
        Real-time click analytics with geographic, device, and referrer
        tracking powered by ClickHouse. Customizable dashboards and data export.
      </>
    ),
  },
  {
    title: 'Enterprise Ready',
    icon: '\ud83c\udfe2',
    description: (
      <>
        SSO/SAML, audit logs, SCIM provisioning, white-label branding,
        and custom domains. Open core with AGPL-3.0 license.
      </>
    ),
  },
];

function Feature({title, icon, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center" style={{fontSize: '3rem'}}>
        {icon}
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
