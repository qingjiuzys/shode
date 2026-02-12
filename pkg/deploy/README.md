# 部署与运维工具 (Deployment & Operations)

Shode 框架提供完整的部署和运维工具集。

## 🚀 部署工具

### 1. Docker 部署 (docker/)

容器化部署解决方案。

**功能**:
- ✅ Dockerfile 生成
- ✅ 镜像构建优化
- ✅ 多阶段构建
- ✅ Docker Compose 编排
- ✅ 容器健康检查
- ✅ 资源限制配置

**快速开始**:
```bash
# 生成 Dockerfile
shode deploy docker init

# 构建镜像
shode deploy docker build

# 运行容器
shode deploy docker up
```

---

### 2. Kubernetes 部署 (k8s/)

Kubernetes 部署解决方案。

**功能**:
- ✅ K8s 清单生成
- ✅ Helm Chart 管理
- ✅ 滚动更新
- ✅ 服务暴露 (Ingress/Service)
- ✅ 配置管理 (ConfigMap/Secret)
- ✅ 自动扩缩容 (HPA)

**快速开始**:
```bash
# 生成 K8s 清单
shode deploy k8s init

# 部署到 K8s
shode deploy k8s apply

# 查看状态
shode deploy k8s status
```

---

### 3. CI/CD 流水线 (cicd/)

持续集成和部署流水线。

**功能**:
- ✅ GitHub Actions 配置
- ✅ GitLab CI 配置
- ✅ Jenkins Pipeline 配置
- ✅ 自动化测试
- ✅ 自动化部署
- ✅ 灰度发布

**支持的 CI 平台**:
- GitHub Actions
- GitLab CI
- Jenkins
- CircleCI
- Travis CI

---

## 🔧 运维工具

### 4. 配置管理 (config/)

分布式配置管理中心。

**功能**:
- ✅ 配置文件管理
- ✅ 环境变量管理
- ✅ 敏感信息加密
- ✅ 配置版本控制
- ✅ 动态配置更新
- ✅ 配置共享

**特性**:
- 支持多种配置格式 (JSON, YAML, TOML, INI)
- 配置热更新
- 配置校验
- 配置差异对比

---

### 5. 服务发现 (discovery/)

服务注册与发现。

**功能**:
- ✅ 服务注册
- ✅ 健康检查
- ✅ 负载均衡
- ✅ 服务路由
- ✅ 故障转移
- ✅ 服务元数据

**集成支持**:
- Consul
- etcd
- Zookeeper
- Eureka

---

### 6. 日志收集 (logs/)

分布式日志收集和分析。

**功能**:
- ✅ 日志采集
- ✅ 日志解析
- ✅ 日志存储
- ✅ 日志查询
- ✅ 日志可视化
- ✅ 日志告警

**集成支持**:
- ELK Stack (Elasticsearch, Logstash, Kibana)
- Loki
- Fluentd
- Splunk

---

### 7. 监控告警 (monitor/)

系统监控和告警。

**功能**:
- ✅ 指标采集
- ✅ 性能监控
- ✅ 日志监控
- ✅ 链路追踪
- ✅ 告警规则
- ✅ 告警通知

**监控指标**:
- CPU 使用率
- 内存使用率
- 磁盘 I/O
- 网络流量
- 应用指标
- 业务指标

**告警通道**:
- Email
- Slack
- 钉钉
- 企业微信
- SMS
- Webhook

---

## 📖 快速参考

### Docker 部署流程

```bash
# 1. 初始化
shode deploy docker init

# 2. 构建镜像
shode deploy docker build -t myapp:v1.0

# 3. 运行容器
shode deploy docker run -p 8080:8080 myapp:v1.0

# 4. 查看日志
shode deploy docker logs myapp
```

### Kubernetes 部署流程

```bash
# 1. 初始化
shode deploy k8s init

# 2. 构建镜像
shode deploy docker build

# 3. 推送镜像
shode deploy docker push registry.example.com/myapp:v1.0

# 4. 部署
shode deploy k8s apply -f deployment.yaml

# 5. 查看状态
shode deploy k8s get pods
shode deploy k8s get services
```

### CI/CD 流程

```yaml
# .github/workflows/deploy.yml
name: Deploy
on: [push]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Build
        run: shode build
      - name: Test
        run: shode test
      - name: Deploy
        run: shode deploy k8s apply
```

## 🎯 典型场景

### 微服务部署

```bash
# 1. 为每个服务生成 Dockerfile
shode deploy docker init --service api
shode deploy docker init --service worker

# 2. 使用 Docker Compose 编排
shode deploy docker compose up

# 3. 部署到 K8s
shode deploy k8s apply -f k8s/
```

### 灰度发布

```bash
# 1. 部署新版本
shode deploy k8s apply -f deployment-v2.yaml

# 2. 逐步切换流量
shode deploy k8s rollout --service myapp --v2-weight 20

# 3. 完全切换
shode deploy k8s rollout --service myapp --v2-weight 100
```

### 监控和告警

```bash
# 1. 启动监控
shode monitor start

# 2. 配置告警规则
shode monitor alert add --name high_cpu --threshold 80

# 3. 查看监控大盘
shode monitor dashboard
```

## 📚 相关文档

- [Docker 部署指南](./docker/README.md)
- [Kubernetes 部署指南](./k8s/README.md)
- [CI/CD 配置指南](./cicd/README.md)
- [配置管理指南](./config/README.md)
- [监控告警指南](./monitor/README.md)

## 🤝 贡献

欢迎贡献新的部署和运维工具！

## 📄 许可证

MIT License
