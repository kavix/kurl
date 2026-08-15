import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';

import styles from './index.module.css';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)} style={{backgroundColor: '#0a0a0a', color: '#fff', borderBottom: '1px solid #333'}}>
      <div className="container">
        <Heading as="h1" className="hero__title" style={{fontSize: '4rem', fontWeight: 800}}>
          ⚡ {siteConfig.title}
        </Heading>
        <p className="hero__subtitle" style={{color: '#ccc', maxWidth: '800px', margin: '0 auto', fontSize: '1.5rem'}}>{siteConfig.tagline}</p>
        <div className={styles.buttons} style={{marginTop: '2rem'}}>
          <Link
            className="button button--secondary button--lg"
            to="/docs/"
            style={{backgroundColor: '#00ADD8', color: '#fff', border: 'none', borderRadius: '8px'}}>
            Get Started - 5min ⏱️
          </Link>
        </div>
      </div>
    </header>
  );
}

export default function Home(): JSX.Element {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`Hello from ${siteConfig.title}`}
      description="A modern, concurrent HTTP/GraphQL/SSE client for the terminal.">
      <HomepageHeader />
      <main style={{padding: '4rem 2rem', maxWidth: '1000px', margin: '0 auto'}}>
        <div className="row">
          <div className="col col--4" style={{marginBottom: '2rem'}}>
            <h3>⚡ Lightning Fast</h3>
            <p>Single static binary written in Go. Zero-config scheme probing and concurrent DNS racing.</p>
          </div>
          <div className="col col--4" style={{marginBottom: '2rem'}}>
            <h3>🔍 Modern Parsing</h3>
            <p>Built-in JSONPath filtering, smart HTML5 DOM pretty-printing, and a native GraphQL engine.</p>
          </div>
          <div className="col col--4" style={{marginBottom: '2rem'}}>
            <h3>📡 Real-time Streams</h3>
            <p>Native support for Server-Sent Events (SSE) and interactive duplex WebSockets.</p>
          </div>
        </div>
      </main>
    </Layout>
  );
}
