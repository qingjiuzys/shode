import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureCategory = {
  title: string;
  icon: string;
  features: string[];
  link?: string;
};

const FeatureCategories: FeatureCategory[] = [
  {
    title: '控制流语句',
    icon: '🔄',
    features: ['If-Then-Else', 'For 循环', 'While 循环', 'Break/Continue'],
    link: '/docs/guides/shell-features#控制流语句',
  },
  {
    title: '管道和重定向',
    icon: '🔀',
    features: ['管道 (|)', '输出重定向 (> >>)', '输入重定向 (<)', '错误重定向 (2>&1)'],
    link: '/docs/guides/shell-features#管道和重定向',
  },
  {
    title: '变量系统',
    icon: '📝',
    features: ['变量赋值', '变量展开 ($VAR, ${VAR})', '字符串拼接', '环境变量管理'],
    link: '/docs/guides/shell-features#变量系统',
  },
  {
    title: '函数系统',
    icon: '⚙️',
    features: ['函数定义', '函数调用', '参数传递 ($1, $2, $@, $#)', '作用域隔离'],
    link: '/docs/guides/shell-features#函数系统',
  },
  {
    title: '模块系统',
    icon: '📦',
    features: ['模块导入/导出', 'package.json 支持', '路径解析'],
    link: '/docs/guides/shell-features#模块系统',
  },
  {
    title: '注解系统',
    icon: '🏷️',
    features: ['简单注解 (@Annotation)', '带参数注解', '注解处理'],
    link: '/docs/guides/shell-features#注解系统',
  },
  {
    title: '执行模式',
    icon: '⚡',
    features: ['解释执行（标准库）', '进程执行（外部命令）', '混合模式'],
    link: '/docs/guides/shell-features#执行模式',
  },
  {
    title: '安全特性',
    icon: '🔒',
    features: ['命令黑名单', '敏感文件保护', '模式检测'],
    link: '/docs/guides/shell-features#安全特性',
  },
  {
    title: '性能优化',
    icon: '🚀',
    features: ['命令缓存', '进程池', '性能指标收集'],
    link: '/docs/guides/shell-features#性能优化',
  },
];

function FeatureCategory({category}: {category: FeatureCategory}) {
  const content = (
    <div className={clsx('card', styles.featureCard)}>
      <div className={styles.featureHeader}>
        <span className={styles.featureIcon}>{category.icon}</span>
        <Heading as="h3" className={styles.featureTitle}>
          {category.title}
        </Heading>
      </div>
      <ul className={styles.featureList}>
        {category.features.map((feature, idx) => (
          <li key={idx}>{feature}</li>
        ))}
      </ul>
    </div>
  );

  if (category.link) {
    return (
      <Link to={category.link} className={styles.featureLink}>
        {content}
      </Link>
    );
  }

  return content;
}

export default function HomepageShellFeatures(): ReactNode {
  return (
    <section className={styles.shellFeatures}>
      <div className="container">
        <div className="row">
          <div className="col col--12">
            <div className={styles.header}>
              <Heading as="h2" className={styles.title}>
                完整的 Shell 特性支持
              </Heading>
              <p className={styles.subtitle}>
                Shode 支持完整的 Shell 语法和特性，兼容传统 Shell 脚本，同时提供现代化的增强功能
              </p>
              <Link
                className="button button--primary button--lg"
                to="/docs/guides/shell-features">
                查看完整特性清单 →
              </Link>
            </div>
          </div>
        </div>
        <div className="row">
          {FeatureCategories.map((category, idx) => (
            <div key={idx} className={clsx('col col--4', styles.featureCol)}>
              <FeatureCategory category={category} />
            </div>
          ))}
        </div>
        <div className="row">
          <div className="col col--12">
            <div className={styles.footer}>
              <p>
                <strong>💡 提示：</strong>
                所有特性均已实现并通过测试，可直接使用。查看{' '}
                <Link to="/docs/guides/shell-features">完整特性文档</Link>{' '}
                了解更多详情和代码示例。
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
