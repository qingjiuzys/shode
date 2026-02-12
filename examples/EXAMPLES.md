# Shode 框架 - 现代化示例项目集合

欢迎来到 Shode 框架的现代化示例项目集合！这里展示了如何使用 Shode v0.6.0 构建各种类型的应用。

## 📚 示例列表

### 1. Web 应用示例 (web-app/)

**现代化的 RESTful API 服务**

- ✅ RESTful API 设计
- ✅ 用户认证和授权 (JWT)
- ✅ 数据库集成 (PostgreSQL)
- ✅ 缓存优化 (Redis)
- ✅ 请求限流
- ✅ 日志记录
- ✅ 性能监控 (OpenTelemetry)
- ✅ Docker 部署
- ✅ 完整测试用例

**快速开始:**
```bash
cd examples/web-app
docker-compose up
# 访问 http://localhost:8080
```

[查看完整文档 →](./web-app/README.md)

---

### 2. CLI 工具示例 (cli-tool/)

**功能完整的命令行应用**

- ✅ 命令行参数解析
- ✅ 交互式命令
- ✅ 配置文件管理
- ✅ 进度条显示
- ✅ 表格输出
- ✅ 颜色输出
- ✅ 自动补全 (Bash/Zsh)
- ✅ 插件系统

**快速开始:**
```bash
cd examples/cli-tool
shode run main.shode --help
shode run main.shode init my-project
```

[查看完整文档 →](./cli-tool/README.md)

---

### 3. 数据处理示例 (data-processing/)

**ETL 数据处理流程**

- ✅ 数据提取 (Extract)
- ✅ 数据转换 (Transform)
- ✅ 数据加载 (Load)
- ✅ 流式处理
- ✅ 批处理
- ✅ 错误处理
- ✅ 进度监控

**使用场景:** 数据同步、报表生成、数据清洗

```bash
cd examples/data-processing
shode run etl.shode --config config.json
```

---

### 4. 微服务示例 (microservice/)

**分布式微服务架构**

- ✅ 服务注册与发现
- ✅ API 网关
- ✅ 服务间通信 (gRPC)
- ✅ 负载均衡
- ✅ 熔断降级
- ✅ 分布式追踪
- ✅ 配置中心

**快速开始:**
```bash
cd examples/microservice
docker-compose up
```

---

### 5. 实时聊天示例 (realtime-chat/)

**WebSocket 实时通信应用**

- ✅ WebSocket 服务器
- ✅ 房间管理
- ✅ 在线用户
- ✅ 消息广播
- ✅ 私聊功能
- ✅ 消息历史
- ✅ 多媒体支持

**快速开始:**
```bash
cd examples/realtime-chat
shode run server.shode
# 访问 http://localhost:3000
```

---

### 6. DevOps 工具示例 (devops-tool/)

**自动化运维工具**

- ✅ CI/CD 流水线
- ✅ 容器管理
- ✅ 监控告警
- ✅ 日志收集
- ✅ 自动化部署
- ✅ 健康检查

**使用场景:** 自动化部署、系统监控、日志分析

```bash
cd examples/devops-tool
shode run deploy.shode
```

---

### 7. AI 应用示例 (ai-app/)

**人工智能应用**

- ✅ 模型训练
- ✅ 模型推理
- ✅ 数据预处理
- ✅ 特征工程
- ✅ 模型评估
- ✅ 模型部署
- ✅ A/B 测试

**技术栈:** TensorFlow, PyTorch, ONNX

```bash
cd examples/ai-app
shode run inference.shode --model model.onnx
```

---

### 8. IoT 网关示例 (iot-gateway/)

**物联网网关服务**

- ✅ 设备连接
- ✅ 数据采集
- ✅ 协议转换 (MQTT, CoAP)
- ✅ 边缘计算
- ✅ 数据上报
- ✅ 远程控制
- ✅ OTA 升级

```bash
cd examples/iot-gateway
shode run gateway.shode
```

---

## 🆚 新旧示例对比

### 旧示例（基础示例）

位置: `examples/*.sh`

这些是 Shode 框架的早期基础示例，展示了核心功能：

- `http_server.sh` - HTTP 服务器
- `cache_example.sh` - 缓存使用
- `database_example.sh` - 数据库操作
- `ecommerce_api.sh` - 电商 API
- `blog_api.sh` - 博客 API
- `user_management.sh` - 用户管理
- `session_management.sh` - 会话管理
- `rate_limiting.sh` - API 限流
- `data_aggregation.sh` - 数据聚合
- `account_transfer.sh` - 账户转账

**特点:**
- ✅ 简单易懂
- ✅ 单文件示例
- ✅ 快速上手
- ✅ 核心功能演示

**适合:** 初学者学习 Shode 基础语法和功能

### 新示例（现代化示例）

位置: `examples/<category>/`

这些是基于 Shode v0.6.0 的完整项目示例：

- `web-app/` - Web 应用
- `cli-tool/` - CLI 工具
- `data-processing/` - 数据处理
- `microservice/` - 微服务
- `realtime-chat/` - 实时聊天
- `devops-tool/` - DevOps 工具
- `ai-app/` - AI 应用
- `iot-gateway/` - IoT 网关

**特点:**
- ✅ 完整项目结构
- ✅ 现代化技术栈
- ✅ 生产就绪
- ✅ 最佳实践
- ✅ 详细文档
- ✅ 测试用例
- ✅ Docker 支持

**适合:** 构建实际生产应用

## 🚀 快速开始

### 前置要求

- Shode 框架 v0.6.0+
- Go 1.21+ (如需编译)
- Docker (可选，用于容器化部署)

### 安装 Shode

```bash
# 克隆仓库
git clone https://github.com/shode/shode.git
cd shode

# 编译安装
go build -o shode ./cmd/shode
sudo mv shode /usr/local/bin/

# 验证安装
shode --version
```

### 运行示例

#### 方式 1: 使用 shode run

```bash
# 进入示例目录
cd examples/web-app

# 运行示例
shode run main.shode
```

#### 方式 2: 使用 Docker

```bash
# 进入示例目录
cd examples/web-app

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

## 📖 学习路径

### 🌱 初级（入门）

**目标**: 学习 Shode 基础语法和核心概念

1. **基础示例** - 运行 `examples/*.sh` 中的单文件示例
2. **Web 应用** - 学习 HTTP 服务器和路由
3. **CLI 工具** - 学习命令行应用开发

**预计时间**: 1-2 周

### 🌿 中级（进阶）

**目标**: 学习构建实际应用

4. **数据处理** - 学习数据流和批处理
5. **实时聊天** - 学习 WebSocket 和实时通信
6. **DevOps 工具** - 学习自动化运维

**预计时间**: 2-4 周

### 🌳 高级（专家）

**目标**: 学习构建复杂系统

7. **微服务** - 学习分布式系统设计
8. **AI 应用** - 学习机器学习集成
9. **IoT 网关** - 学习物联网协议和边缘计算

**预计时间**: 4-8 周

## 💡 使用建议

### 学习建议

1. **从简单开始** - 建议从基础示例开始
2. **阅读代码** - 仔细阅读源码和注释
3. **修改实验** - 在示例基础上进行修改和实验
4. **查看文档** - 遇到问题查看框架文档
5. **社区交流** - 加入社区讨论和分享

### 项目开发建议

1. **选择合适的模板** - 根据项目类型选择示例
2. **复制示例代码** - 复制示例作为项目起点
3. **修改配置** - 根据需要修改配置
4. **添加业务逻辑** - 实现具体业务需求
5. **编写测试** - 为新功能编写测试
6. **部署上线** - 使用示例的部署脚本

## 🎯 示例特点

所有现代化示例项目都遵循：

- ✅ **完整代码** - 可直接运行的完整示例
- ✅ **详细文档** - 包含 README 和代码注释
- ✅ **测试用例** - 包含完整的测试代码
- ✅ **配置文件** - 提供配置示例和说明
- ✅ **部署脚本** - 支持 Docker 和 K8s 部署
- ✅ **最佳实践** - 遵循框架和行业最佳实践

## 📊 示例统计

| 类别 | 示例数量 | 代码行数 | 文档数量 |
|------|---------|---------|---------|
| 基础示例 | 10+ | ~2000 | 内嵌注释 |
| 现代化示例 | 8 | ~8000 | 独立文档 |
| **总计** | **18+** | **~10000** | **20+** |

## 🤝 贡献指南

欢迎贡献新的示例！

### 提交示例

1. Fork 本仓库
2. 在 `examples/` 下创建新的目录
3. 编写代码和文档
4. 添加测试和配置
5. 提交 Pull Request

### 示例要求

- ✅ 代码清晰易懂
- ✅ 包含完整文档 (README.md)
- ✅ 提供测试用例
- ✅ 遵循框架规范
- ✅ 添加必要注释
- ✅ 支持容器化部署

### 示例模板

创建新示例时，请包含以下文件：

```
examples/my-example/
├── README.md           # 详细说明文档
├── main.shode          # 主程序
├── config.shode        # 配置文件
├── Dockerfile          # Docker 镜像
├── docker-compose.yml  # Docker Compose
└── tests/              # 测试用例
    └── test.shode
```

## 📞 获取帮助

- 📖 [框架文档](https://shode.dev/docs)
- 💬 [社区论坛](https://forum.shode.dev)
- 🐛 [问题追踪](https://github.com/shode/shode/issues)
- 💬 [Discord 社区](https://discord.gg/shode)
- 📧 [邮件支持](support@shode.dev)

## 🔗 相关资源

- [Shode 主页](https://shode.dev)
- [GitHub 仓库](https://github.com/shode/shode)
- [API 文档](https://shode.dev/api)
- [插件市场](https://shode.dev/plugins)
- [视频教程](https://shode.dev/videos)

## 📄 许可证

所有示例代码采用 MIT 许可证，可自由使用和修改。

---

**开始探索 Shode 框架的强大功能吧！** 🚀

*最后更新: 2026-02-01*
