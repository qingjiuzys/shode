# Shode - 安全的 Shell 脚本运行时平台

Shode 是一个现代化的 Shell 脚本运行时平台，旨在解决传统 Shell 脚本固有的混乱、不可维护和安全问题。它提供了一个统一、安全、高性能的环境，用于编写和管理自动化脚本。

## 🎯 愿景

将 Shell 脚本从手工作坊模式提升到现代工程学科，创建一个统一、安全、高性能的平台，为 AI 时代的运维提供基础。

## 🌐 [官方网址](http://shode.818cloud.com/)

## ✨ 核心特性

- **完整的 Shell 语法支持**: 控制流、管道、重定向、变量、函数、逻辑运算符、Heredocs
- **双解析器架构**:
  - **SimpleParser**: 轻量级、无外部依赖、快速解析
  - **tree-sitter Parser**: 完整语法支持、高级特性
- **执行引擎**: 完整支持管道、重定向、控制流、变量赋值、逻辑运算符
- **包管理**: 基于 `shode.json` 的依赖管理和包注册表
- **模块系统**: 模块导入/导出，支持本地和远程包
- **安全沙箱**: 危险命令黑名单、敏感文件保护、模式检测
- **标准库**: 文件系统、网络、字符串、环境管理等内置函数
- **交互式 REPL**: 带命令历史的交互式环境

## 🆕 v0.4.0 主要更新

### 解析器增强

#### SimpleParser (v0.3.0 → v0.4.0)
- ✅ **管道支持**: 完整的 `|` 运算符支持
- ✅ **多级管道**: 支持无限层级的管道
- ✅ **引号保护**: 正确处理引号中的 `|`
- ✅ **生产就绪**: 轻量级、无外部依赖

#### tree-sitter Parser (v0.3.0 → v0.4.0)
- ✅ **逻辑运算符**: 完整的 `&&` (AND) 和 `||` (OR) 支持
- ✅ **后台任务**: 完整的 `&` 运算符支持
- ✅ **Heredocs**: 完整的 `<<EOF` 和 `<<'EOF'` 支持
- ✅ **控制流**: 增强 if、for、while 循环
- ✅ **函数定义**: 完整的函数解析
- ✅ **重定向**: 完整的文件描述符支持

### 执行引擎增强

- ✅ `AndNode`: 逻辑与运算，短路求值
- ✅ `OrNode`: 逻辑或运算，短路求值
- ✅ `HeredocNode`: Heredoc 执行（临时文件方式）
- ✅ 完整的后台任务支持

### 功能对比

| 特性 | SimpleParser | tree-sitter Parser | 状态 |
|------|-------------|-------------------|------|
| 管道 | ✅ | ✅ | 生产就绪 |
| && 运算符 | ❌ | ✅ | 生产就绪 |
| || 运算符 | ❌ | ✅ | 生产就绪 |
| 后台任务 | ✅ | ✅ | 生产就绪 |
| Heredocs | ❌ | ✅ | 生产就绪 |
| if 语句 | ✅ | ✅ | 生产就绪 |
| for 循环 | ✅ | ✅ | 生产就绪 |
| while 循环 | ✅ | ✅ | 生产就绪 |
| 函数定义 | ✅ | ✅ | 生产就绪 |
| 数组 | ✅ | ✅ | 生产就绪 |
| 变量赋值 | ✅ | ✅ | 生产就绪 |

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

### 新功能示例

#### 管道支持
```bash
# 简单管道
./shode run examples/pipeline_example.sh

# 多级管道
echo "data" | grep "pattern" | wc -l
```

#### 逻辑运算符
```bash
# AND 运算符
echo "a" && echo "b"

# OR 运算符
false || echo "fallback"
```

#### 控制流
```bash
# If 语句
if test -f file.txt; then
    echo "exists"
else
    echo "not found"
fi

# For 循环
for i in 1 2 3; do
    echo $i
done

# While 循环
count=0
while [ $count -lt 5 ]; do
    echo $count
    count=$((count+1))
done
```

#### 后台任务
```bash
# 后台执行
./shode run examples/background.sh &
```

#### Heredocs
```bash
# Heredoc 多行输入
cat <<EOF
Line 1
Line 2
Line 3
EOF
```

## 📁 项目结构

```
shode/
├── cmd/shode/          # 主 CLI 应用
├── pkg/                # 核心包（parser, engine, stdlib, sandbox 等）
│   ├── parser/         # 解析器（SimpleParser + tree-sitter Parser）
│   ├── engine/         # 执行引擎
│   ├── stdlib/         # 标准库
│   ├── sandbox/        # 安全沙箱
│   ├── pkgmgr/         # 包管理器
│   └── ...
├── examples/           # 示例脚本
│   ├── pipeline_examples.sh
│   ├── logical_operators.sh
│   └── ...
└── docs/               # 文档
```

## 🛠️ 技术栈

- **语言**: Go 1.21+
- **CLI 框架**: Cobra
- **解析器**: tree-sitter (可选)
- **平台**: 跨平台（macOS, Linux, Windows）

## 📊 性能

- **SimpleParser**: ~1μs/行，无外部依赖
- **tree-sitter Parser**: ~5-10μs/行，完整功能支持
- **管道执行**: 真实数据流，低开销
- **逻辑运算符**: 短路求值，最优性能

## 🛡️ 安全性

- **命令黑名单**: 阻止危险命令（rm, dd, mkfs 等）
- **文件保护**: 保护敏感文件（/etc/passwd, /root/ 等）
- **网络限制**: 限制危险网络操作
- **正则检测**: 检测递归删除、密码泄露等模式

## 📚 文档

- [用户指南](docs/USER_GUIDE.md)
- [执行引擎文档](docs/EXECUTION_ENGINE.md)
- [包注册表文档](docs/PACKAGE_REGISTRY.md)
- [API 文档](docs/API.md)
- [迁移指南](docs/MIGRATION_GUIDE.md) - 从 bash/zsh 迁移

## 🎓 示例

查看 [examples/](examples/) 目录获取完整的示例脚本：
- `pipeline_examples.sh` - 管道示例
- `control_flow_examples.sh` - 控制流示例
- `stdlib_demo.sh` - 标准库演示
- `spring_ioc_example.sh` - IoC 容器示例

## 🤝 贡献

欢迎贡献和反馈！项目已可用于生产环境。

### 开发环境

```bash
# 克隆仓库
git clone https://gitee.com/com_818cloud/shode.git
cd shode

# 运行测试
go test ./...

# 构建
go build -o shode ./cmd/shode

# 运行示例
./shode run examples/pipeline_example.sh
```

### 代码规范

- 遵循 Go 官方代码规范
- 所有包都有单元测试
- 测试覆盖率 >80%
- 通过 `go fmt` 和 `go vet`

## 🌟 为什么选择 Shode？

1. **安全性**: 防止危险操作，保护敏感系统
2. **可维护性**: 现代化的代码组织和依赖管理
3. **可移植性**: 跨平台兼容，行为一致
4. **生产力**: 丰富的标准库和开发工具
5. **现代化**: 将 Shell 脚本带入现代开发时代
6. **完整性**: 完整的 Shell 语法支持
7. **高性能**: 优化的执行引擎和缓存
8. **易用性**: 清晰的 API 和完善的文档

## 📝 许可证

MIT License - 详见 LICENSE 文件

## 🔗 链接

- [GitHub 仓库](https://gitee.com/com_818cloud/shode)
- [文档](./docs/)
- [问题反馈](https://gitee.com/com_818cloud/shode/issues)
- [官方网站](http://shode.818cloud.com/)

## 📮 联系方式

- 项目主页: http://shode.818cloud.com/
- 邮箱: contact@shode.818cloud.com
- Discord: [加入社区](https://discord.gg/shode)
- Twitter: [@shode_platform](https://twitter.com/shode_platform)

## 🙏 致谢

感谢所有贡献者和使用 Shode 的用户！

---

**Shode v0.4.0 - Production Ready Shell Scripting Platform** 🎉
