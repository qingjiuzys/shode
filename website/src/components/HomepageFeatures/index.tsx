import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  Svg: React.ComponentType<React.ComponentProps<'svg'>>;
  description: ReactNode;
};

const FeatureList: FeatureItem[] = [
  {
    title: '极致安全',
    Svg: require('@site/static/img/undraw_docusaurus_mountain.svg').default,
    description: (
      <>
        内置安全沙箱，从根源上杜绝注入攻击和危险操作，
        为脚本执行提供安全护栏。执行前安全检查，自动拦截危险命令。
      </>
    ),
  },
  {
    title: '卓越性能',
    Svg: require('@site/static/img/undraw_docusaurus_tree.svg').default,
    description: (
      <>
        标准库函数直接内存访问，无进程生成开销，启动时间毫秒级。
        比传统shell命令快数倍，比Spring Boot应用轻量10-100倍。
      </>
    ),
  },
  {
    title: '现代生态',
    Svg: require('@site/static/img/undraw_docusaurus_react.svg').default,
    description: (
      <>
        完整的包管理和模块系统，支持依赖管理和代码复用。
        内置HTTP服务器、数据库、缓存，开箱即用，无需复杂配置。
      </>
    ),
  },
];

function Feature({title, Svg, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
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
