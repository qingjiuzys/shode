# 开发者工具集 (Developer Tools)

Shode 框架提供完整的开发者工具集，帮助开发者提高开发效率。

## 🛠️ 工具列表

### 1. 代码生成器 (Generator)

自动生成项目模板和代码骨架。

**命令**: `shode generate`

**功能**:
- 生成项目模板
- 生成代码骨架
- 生成配置文件
- 自定义模板

**示例**:
```bash
# 生成新项目
shode generate project my-app

# 生成控制器
shode generate controller User

# 生成模型
shode generate model Article
```

### 2. 调试器 (Debugger)

强大的断点调试工具。

**命令**: `shode debug`

**功能**:
- 断点设置
- 单步执行
- 变量查看
- 调用栈查看
- 表达式求值

**示例**:
```bash
# 启动调试模式
shode debug main.shode

# 设置断点
break main.shode:10

# 继续执行
continue

# 查看变量
print variable_name
```

### 3. 性能分析器 (Profiler)

CPU 和内存性能分析工具。

**命令**: `shode profile`

**功能**:
- CPU 性能分析
- 内存性能分析
- 火焰图生成
- 热点函数识别
- 性能报告

**示例**:
```bash
# CPU 性能分析
shode profile --cpu main.shode

# 内存性能分析
shode profile --memory main.shode

# 生成火焰图
shode profile --flamegraph main.shode
```

### 4. 代码格式化 (Formatter)

自动格式化代码。

**命令**: `shode fmt`

**功能**:
- 代码格式化
- 缩进规范化
- 空格处理
- 注释格式化

**示例**:
```bash
# 格式化当前目录
shode fmt .

# 检查格式但不修改
shode fmt --check .

# 显示差异
shode fmt --diff .
```

### 5. 代码检查 (Linter)

静态代码分析和质量检查。

**命令**: `shode lint`

**功能**:
- 静态分析
- 代码质量检查
- 最佳实践建议
- 潜在问题检测

**示例**:
```bash
# 检查所有文件
shode lint .

# 检查特定文件
shode lint main.shode

# 输出 JSON 格式
shode lint --json .
```

### 6. 包管理器 (Packager)

依赖和包管理工具。

**命令**: `shode package` 或 `shode pkg`

**功能**:
- 依赖安装
- 依赖更新
- 依赖树查看
- 包发布
- 版本管理

**示例**:
```bash
# 安装依赖
shode pkg install

# 添加依赖
shode pkg add github.com/user/repo

# 更新依赖
shode pkg update

# 查看依赖树
shode pkg tree
```

### 7. 测试工具 (Tester)

完整的测试框架和工具。

**命令**: `shode test`

**功能**:
- 单元测试
- 集成测试
- 覆盖率报告
- 基准测试
- 模拟和断言

**示例**:
```bash
# 运行所有测试
shode test

# 运行特定测试
shode test tests/main_test.shode

# 生成覆盖率报告
shode test --cover

# 详细输出
shode test --verbose
```

### 8. 文档生成器 (DocGen)

自动生成 API 和代码文档。

**命令**: `shode docs`

**功能**:
- API 文档生成
- 代码文档生成
- Markdown 输出
- HTML 输出
- 交互式文档

**示例**:
```bash
# 生成文档
shode docs

# 指定输出目录
shode docs --output docs/

# HTML 格式
shode docs --format html

# 启动文档服务器
shode docs --serve
```

### 9. REPL 交互环境

交互式执行环境。

**命令**: `shode repl`

**功能**:
- 交互式执行
- 代码补全
- 历史记录
- 多行输入

**示例**:
```bash
# 启动 REPL
shode repl

> print("Hello, World!")
Hello, World!
> 1 + 1
2
> .exit
```

### 10. 日志查看器 (Logger)

日志查询和分析工具。

**命令**: `shode logs`

**功能**:
- 日志查询
- 日志过滤
- 日志聚合
- 实时跟踪
- 日志统计

**示例**:
```bash
# 查看所有日志
shode logs

# 实时跟踪
shode logs --follow

# 过滤日志
shode logs --filter "level=error"

# 统计日志
shode logs --stats
```

## 📚 快速参考

### 常用命令

```bash
# 项目初始化
shode generate project my-app
cd my-app

# 开发模式运行
shode run main.shode

# 格式化代码
shode fmt .

# 运行测试
shode test

# 构建项目
shode build

# 调试
shode debug main.shode

# 性能分析
shode profile --cpu main.shode

# 生成文档
shode docs

# 检查代码质量
shode lint .
```

### 配置文件

开发者工具配置文件 `devtools.config`:

```javascript
{
  // 格式化配置
  formatter: {
    indent: 4,
    tab_width: 4,
    max_line_length: 100
  },

  // Linter 配置
  linter: {
    enable: ["all"],
    rules: {
      "no-unused-vars": "error",
      "no-console": "warn"
    }
  },

  // 测试配置
  test: {
    verbose: true,
    coverage: true,
    timeout: 30
  },

  // 文档配置
  docs: {
    format: "markdown",
    output: "docs/"
  }
}
```

## 🔧 集成开发环境 (IDE)

### VSCode 扩展

安装 VSCode 扩展获得更好的开发体验：

- Shode Language Support
- Shode Debugger
- Shode Formatter

### Vim/Neovim 插件

```vim
" 安装 vim-shode 插件
Plug 'shode/vim-shode'

" 启用语法高亮
syntax on

" 启用自动格式化
autocmd BufWritePre *.shode :ShodeFormat
```

## 📖 学习资源

- [快速入门](./getting-started.md)
- [工具详细文档](./tools/)
- [最佳实践](./best-practices.md)
- [故障排查](./troubleshooting.md)

## 🤝 贡献

欢迎贡献新的开发者工具！

## 📄 许可证

MIT License
