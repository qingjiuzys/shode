import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type AdvantageItem = {
  title: string;
  description: ReactNode;
  icon: string;
};

const AdvantageList: AdvantageItem[] = [
  {
    title: '轻量级部署',
    icon: '⚡',
    description: (
      <>
        单二进制文件，无需JVM或运行时环境。内存占用仅数MB，
        启动时间毫秒级，比Spring Boot应用快10-100倍。
      </>
    ),
  },
  {
    title: '零编译开发',
    icon: '🚀',
    description: (
      <>
        脚本即代码，无需编译打包。修改即运行，开发效率提升数倍。
        特别适合快速迭代和原型验证场景。
      </>
    ),
  },
  {
    title: '原生Shell能力',
    icon: '🔧',
    description: (
      <>
        直接执行Shell命令，无需额外包装。完美兼容现有Shell脚本生态，
        学习成本低，运维人员零门槛上手。
      </>
    ),
  },
  {
    title: '专为运维设计',
    icon: '🎯',
    description: (
      <>
        内置HTTP服务器、数据库连接、缓存系统，开箱即用。
        无需复杂的框架配置，专注业务逻辑实现。
      </>
    ),
  },
  {
    title: '跨平台一致',
    icon: '🌍',
    description: (
      <>
        一套代码，Linux/macOS/Windows(WSL)统一运行。
        标准库函数保证跨平台行为一致性，告别环境差异问题。
      </>
    ),
  },
  {
    title: 'Spring化能力',
    icon: '🌸',
    description: (
      <>
        提供IoC容器、配置管理、Web层、事务管理等企业级特性。
        既有Spring的便利，又保持Shell脚本的简洁。
      </>
    ),
  },
];

function Advantage({title, icon, description}: AdvantageItem) {
  return (
    <div className={clsx('col col--4', styles.advantage)}>
      <div className={styles.advantageIcon}>{icon}</div>
      <Heading as="h3" className={styles.advantageTitle}>{title}</Heading>
      <p className={styles.advantageDescription}>{description}</p>
    </div>
  );
}

export default function HomepageComparison(): ReactNode {
  return (
    <section className={styles.comparison}>
      <div className="container">
        <div className="row">
          <div className="col col--12">
            <div className={styles.comparisonHeader}>
              <Heading as="h2" className={styles.comparisonTitle}>
                Shode vs Spring Boot
              </Heading>
              <p className={styles.comparisonSubtitle}>
                Shode 专为脚本和运维场景设计，在轻量级、快速开发、运维友好等方面具有显著优势
              </p>
            </div>
          </div>
        </div>
        <div className="row">
          {AdvantageList.map((props, idx) => (
            <Advantage key={idx} {...props} />
          ))}
        </div>
        <div className="row">
          <div className="col col--12">
            <div className={styles.comparisonTable}>
              <table>
                <thead>
                  <tr>
                    <th>特性</th>
                    <th>Shode</th>
                    <th>Spring Boot</th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td><strong>部署方式</strong></td>
                    <td>✅ 单二进制文件，零依赖</td>
                    <td>❌ 需要JVM + JAR包</td>
                  </tr>
                  <tr>
                    <td><strong>启动时间</strong></td>
                    <td>✅ 毫秒级（&lt;100ms）</td>
                    <td>❌ 秒级（1-10秒）</td>
                  </tr>
                  <tr>
                    <td><strong>内存占用</strong></td>
                    <td>✅ 数MB（5-50MB）</td>
                    <td>❌ 数百MB（200-500MB+）</td>
                  </tr>
                  <tr>
                    <td><strong>开发流程</strong></td>
                    <td>✅ 修改即运行，无需编译</td>
                    <td>❌ 需要编译打包</td>
                  </tr>
                  <tr>
                    <td><strong>学习成本</strong></td>
                    <td>✅ Shell语法，运维友好</td>
                    <td>❌ Java + Spring框架</td>
                  </tr>
                  <tr>
                    <td><strong>适用场景</strong></td>
                    <td>✅ 脚本、运维、API服务</td>
                    <td>✅ 大型企业应用</td>
                  </tr>
                  <tr>
                    <td><strong>Shell兼容</strong></td>
                    <td>✅ 原生支持，完美兼容</td>
                    <td>❌ 需要额外集成</td>
                  </tr>
                  <tr>
                    <td><strong>跨平台</strong></td>
                    <td>✅ 统一行为，标准库保证</td>
                    <td>✅ JVM跨平台</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
        <div className="row">
          <div className="col col--12">
            <div className={styles.comparisonFooter}>
              <p>
                <strong>选择 Shode 如果：</strong>你需要快速开发脚本、构建轻量级API服务、
                进行系统运维自动化，或者希望零编译、低资源占用的解决方案。
              </p>
              <p>
                <strong>选择 Spring Boot 如果：</strong>你需要构建大型企业级应用、
                复杂的业务系统，或者团队已有深厚的Java技术栈。
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
