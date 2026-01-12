import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  tutorialSidebar: [
    {
      type: 'category',
      label: '开始使用',
      items: [
        '介绍',
        '安装',
        '快速开始',
      ],
    },
    {
      type: 'category',
      label: 'API参考',
      items: [
        'API-标准库',
        'API-命令行',
      ],
    },
  ],
};

export default sidebars;
