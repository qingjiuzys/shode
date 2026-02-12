# CLI 工具示例 - 命令行应用

一个功能完整的命令行工具示例，展示 Shode 框架在 CLI 应用中的能力。

## 功能特性

- ✅ 命令行参数解析
- ✅ 交互式命令
- ✅ 配置文件管理
- ✅ 日志输出
- ✅ 进度条显示
- ✅ 表格输出
- ✅ 颜色输出
- ✅ 自动补全
- ✅ 插件系统
- ✅ 文件操作

## 技术栈

- **框架**: Shode v0.6.0
- **终端**: pkg/terminal
- **文件系统**: pkg/storage
- **配置**: pkg/config
- **日志**: pkg/logger

## 项目结构

```
cli-tool/
├── main.shode          # 主程序
├── commands/           # 命令定义
│   ├── init.shode
│   ├── build.shode
│   ├── deploy.shode
│   └── test.shode
├── config/             # 配置文件
│   └── default.conf
├── plugins/            # 插件
│   └── hello.shode
├── completions/        # 自动补全脚本
│   ├── bash
│   └── zsh
└── README.md           # 说明文档
```

## 快速开始

### 安装

```bash
# 编译 CLI 工具
cd examples/cli-tool
shode build main.shode -o shode-cli

# 添加到 PATH
export PATH=$PATH:$(pwd)
```

### 基本使用

```bash
# 查看帮助
shode-cli --help
shode-cli --help

# 初始化项目
shode-cli init my-project

# 构建项目
shode-cli build

# 部署项目
shode-cli deploy --env production

# 运行测试
shode-cli test --verbose

# 查看版本
shode-cli --version
```

## 命令列表

### init - 初始化项目

```bash
shode-cli init <project-name> [options]

选项:
  --template, -t    模板名称 (default, api, microservice)
  --description     项目描述
  --author          作者名称
  --license         许可证 (MIT, Apache-2.0)
  --git             初始化 Git 仓库

示例:
  shode-cli init my-app
  shode-cli init my-api --template api --git
```

### build - 构建项目

```bash
shode-cli build [options]

选项:
  --output, -o      输出目录
  --minify          压缩输出
  --watch           监听文件变化
  --verbose, -v     详细输出

示例:
  shode-cli build
  shode-cli build --output dist --minify
```

### deploy - 部署项目

```bash
shode-cli deploy [options]

选项:
  --env, -e         环境 (development, staging, production)
  --config, -c      配置文件
  --dry-run         模拟运行
  --skip-tests      跳过测试

示例:
  shode-cli deploy --env production
  shode-cli deploy --env staging --dry-run
```

### test - 运行测试

```bash
shode-cli test [options]

选项:
  --verbose, -v     详细输出
  --cover           显示覆盖率
  --watch           监听模式
  --filter          过滤测试

示例:
  shode-cli test
  shode-cli test --cover --verbose
```

## 配置文件

`config/default.conf`:

```javascript
{
  // 项目配置
  project: {
    name: "my-project",
    version: "1.0.0",
    description: "My Shode Project"
  },

  // 构建配置
  build: {
    output: "dist",
    minify: false,
    source_map: true
  },

  // 部署配置
  deploy: {
    provider: "docker",
    registry: "localhost:5000",
    namespace: "my-org"
  },

  // 测试配置
  test: {
    verbose: false,
    coverage: true,
    timeout: 30
  }
}
```

## 插件开发

创建自定义插件：

```javascript
// plugins/hello.shode
plugin {
    name: "hello"
    version: "1.0.0"
    description: "Say hello"

    command "hello" {
        description = "Say hello to someone"
        arguments = ["name"]

        execute = func(args) {
            name = args[0] || "World"
            print("Hello, ${name}!")
        }
    }
}
```

使用插件：

```bash
shode-cli plugin install ./plugins/hello.shode
shode-cli hello Shode
```

## 自动补全

### Bash

```bash
# 启用补全
source <(shode-cli completion bash)

# 永久启用
shode-cli completion bash > /etc/bash_completion.d/shode-cli
```

### Zsh

```bash
# 启用补全
source <(shode-cli completion zsh)

# 永久启用
shode-cli completion zsh > /usr/local/share/zsh/site-functions/_shode-cli
```

## 高级特性

### 交互式模式

```bash
shode-cli interactive

> init my-project
> build
> deploy --env production
> exit
```

### 管道操作

```bash
# 查找文件并处理
find . -name "*.shode" | shode-cli parse | shode-cli validate

# 批量处理
cat files.txt | shode-cli process --batch
```

### 远程执行

```bash
# 在远程服务器执行命令
shode-cli deploy --ssh user@server --remote-path /app
```

## 输出格式

### 表格输出

```bash
$ shode-cli list

NAME          STATUS    VERSION    CREATED
my-project    running   1.0.0     2026-01-15
api-service   stopped   2.1.0     2026-01-14
web-app       running   1.5.2     2026-01-13
```

### JSON 输出

```bash
$ shode-cli list --format json

[
  {
    "name": "my-project",
    "status": "running",
    "version": "1.0.0",
    "created": "2026-01-15T10:30:00Z"
  }
]
```

## 故障排查

### 常见问题

1. **命令未找到**
   - 检查 PATH 环境变量
   - 确认工具已正确安装

2. **权限错误**
   - 检查文件权限
   - 使用 sudo（如需要）

3. **配置文件错误**
   - 验证配置文件语法
   - 检查配置文件路径

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
