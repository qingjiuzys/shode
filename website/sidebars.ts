import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  tutorialSidebar: [
    {
      type: 'category',
      label: '开始使用',
      items: [
        'getting-started/introduction',
        'getting-started/installation',
        'getting-started/quick-start',
      ],
    },
    {
      type: 'category',
      label: '使用指南',
      items: [
        'guides/user-guide',
        'guides/execution-engine',
        'guides/package-registry',
      ],
    },
    {
      type: 'category',
      label: 'API参考',
      items: [
        'api/stdlib',
        'api/cli',
      ],
    },
    {
      type: 'category',
      label: '示例',
      items: [
        'examples/index',
        {
          type: 'category',
          label: '基础示例',
          items: [
            'examples/basic/http-server',
            'examples/basic/cache',
            'examples/basic/database',
          ],
        },
        {
          type: 'category',
          label: '高级示例',
          items: [
            'examples/advanced/ecommerce-api',
            'examples/advanced/blog-api',
            'examples/advanced/spring-features',
          ],
        },
      ],
    },
  ],
};

export default sidebars;
