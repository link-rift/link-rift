import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'Linkrift Docs',
  tagline: 'Enterprise-grade URL shortener with analytics, custom domains, and team collaboration',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://docs.linkrift.io',
  baseUrl: '/',

  organizationName: 'linkrift',
  projectName: 'linkrift',

  onBrokenLinks: 'warn',

  markdown: {
    format: 'md',
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          routeBasePath: 'docs',
          editUrl: 'https://github.com/linkrift/linkrift/tree/main/docs-site/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/linkrift-social-card.png',
    colorMode: {
      defaultMode: 'dark',
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'Linkrift',
      logo: {
        alt: 'Linkrift Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Documentation',
        },
        {
          to: '/docs/api/API_DOCUMENTATION',
          label: 'API',
          position: 'left',
        },
        {
          href: 'https://github.com/linkrift/linkrift',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Quick Start',
              to: '/docs/getting-started/QUICK_START',
            },
            {
              label: 'Architecture',
              to: '/docs/architecture/',
            },
            {
              label: 'API Reference',
              to: '/docs/api/API_DOCUMENTATION',
            },
          ],
        },
        {
          title: 'Features',
          items: [
            {
              label: 'Analytics',
              to: '/docs/features/ANALYTICS_PIPELINE',
            },
            {
              label: 'Custom Domains',
              to: '/docs/features/CUSTOM_DOMAINS',
            },
            {
              label: 'Authentication',
              to: '/docs/features/AUTHENTICATION',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/linkrift/linkrift',
            },
            {
              label: 'Contributing',
              to: '/docs/contributing/',
            },
          ],
        },
      ],
      copyright: `Copyright \u00a9 ${new Date().getFullYear()} Linkrift. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'go', 'json', 'yaml', 'sql', 'typescript'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
