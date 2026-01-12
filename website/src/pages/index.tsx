import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import HomepageComparison from '@site/src/components/HomepageComparison';
import Heading from '@theme/Heading';

import styles from './index.module.css';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <Heading as="h1" className="hero__title">
          {siteConfig.title}
        </Heading>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link
            className="button button--primary button--lg"
            to="/docs/getting-started/quick-start">
            快速开始 🚀
          </Link>
          <Link
            className="button button--secondary button--lg margin-left--md"
            href="https://gitee.com/com_818cloud/shode">
            Gitee ⭐
          </Link>
        </div>
      </div>
    </header>
  );
}

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={siteConfig.title}
      description="Shode - 下一代Shell脚本运行时平台，提供安全、高效、现代化的脚本执行环境，内置标准库、包管理和安全沙箱">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
        <HomepageComparison />
      </main>
    </Layout>
  );
}
