import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    {
      type: 'doc',
      id: 'index',
      label: 'Introduction',
    },
    {
      type: 'doc',
      id: 'USAGE',
      label: 'Usage Guide',
    },
    {
      type: 'doc',
      id: 'ARCHITECTURE',
      label: 'Architecture',
    },
    {
      type: 'doc',
      id: 'CONTRIBUTING',
      label: 'Contributing',
    },
  ],
};

export default sidebars;
