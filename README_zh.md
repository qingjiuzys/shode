# Shode - 安全的 Shell 脚本运行时平台

Shode 是一个现代化的 Shell 脚本运行时平台，旨在解决传统 Shell 脚本固有的混乱、不可维护和安全问题。它提供了一个统一、安全、高性能的环境，用于编写和管理自动化脚本。

## 🎯 愿景

将 Shell 脚本从手工作坊模式提升到现代工程学科，创建一个统一、安全、高性能的平台，为 AI 时代的运维提供基础。

## 🌐 [官方网址](http://shode.818cloud.com/)

## ✨ 核心特性

- **完整的 Shell 语法支持**: 控制流、管道、重定向、变量、函数
- **执行引擎**: 支持管道、重定向、控制流、变量赋值
- **包管理**: 基于 `shode.json` 的依赖管理和包注册表
- **模块系统**: 模块导入/导出，支持本地和远程包
- **安全沙箱**: 危险命令黑名单、敏感文件保护、模式检测
- **标准库**: 文件系统、网络、字符串、环境管理等内置函数
- **交互式 REPL**: 带命令历史的交互式环境

## 🚀 快速开始

### 安装

```bash
git clone https://gitee.com/com_818cloud/shode.git
cd shode
go build -o shode ./cmd/shode
```

### 基本用法

```bash
# 运行脚本文件
./shode run examples/test.sh

# 执行命令
./shode exec "echo hello world"

# 交互式 REPL
./shode repl

# 包管理
./shode pkg init my-project 1.0.0
./shode pkg add lodash 4.17.21
./shode pkg install
```

## 📁 项目结构

```
shode/
├── cmd/shode/          # 主 CLI 应用
├── pkg/                # 核心包（parser, engine, stdlib, sandbox 等）
├── examples/           # 示例脚本
└── docs/               # 文档
```

## 🛠️ 技术栈

- **语言**: Go 1.21+
- **CLI 框架**: Cobra
- **平台**: 跨平台（macOS, Linux, Windows）

## 📝 许可证

MIT 许可证 - 详见 LICENSE 文件

## 🤝 贡献

欢迎贡献和反馈！项目已可用于生产环境。

## 🌟 为什么选择 Shode？

1. **安全性**: 防止危险操作，保护敏感系统
2. **可维护性**: 现代化的代码组织和依赖管理
3. **可移植性**: 跨平台兼容，行为一致
4. **生产力**: 丰富的标准库和开发工具
5. **现代化**: 将 Shell 脚本带入现代开发时代
