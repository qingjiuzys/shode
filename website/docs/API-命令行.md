# 命令行 API 参考

Shode 提供了完整的命令行工具集，支持脚本执行、命令执行、包管理等功能。所有命令都支持丰富的参数和选项。

## 命令概览

### 主要命令
```bash
shode run      [script-file]    # 运行脚本文件
shode exec     [command]        # 执行单条命令
shode repl                      # 启动交互式 REPL
shode pkg      [subcommand]     # 包管理功能
shode version                   # 显示版本信息
shode completion                # 生成自动补全脚本
```

## shode run - 运行脚本

运行 Shode 脚本文件。

### 语法
```bash
shode run [options] <script-file>
```

### 参数
- `<script-file>`: 要执行的脚本文件路径

### 选项
- `-v, --verbose`: 详细模式，显示解析和执行详情
- `-s, --safe`: 安全模式，启用严格沙箱限制
- `-e, --env`: 设置环境变量（格式：KEY=VALUE）
- `-h, --help`: 显示帮助信息

### 示例
```bash
# 运行简单脚本
shode run hello.sh

# 详细模式运行
shode run -v script.sh

# 安全模式运行
shode run -s untrusted_script.sh

# 设置环境变量
shode run -e DEBUG=true -e TIMEOUT=1000 script.sh
```

### 退出代码
- `0`: 执行成功
- `1`: 脚本语法错误
- `2`: 运行时错误
- `3`: 文件不存在
- `4`: 权限不足

## shode exec - 执行命令

执行单条 Shode 命令或外部命令。

### 语法
```bash
shode exec [options] <command>
```

### 参数
- `<command>`: 要执行的命令字符串

### 选项
- `-v, --verbose`: 详细模式，显示执行详情
- `-s, --safe`: 安全模式，启用严格沙箱
- `-h, --help`: 显示帮助信息

### 示例
```bash
# 执行标准库函数
shode exec "upper hello world"
shode exec "trim '   test   '"

# 执行外部命令
shode exec "echo Hello from Shode"
shode exec "pwd && ls -la"

# 组合命令
shode exec "upper hello && lower WORLD"

# 使用管道
echo "input text" | shode exec "upper"
```

### 特殊语法支持
- `&&`: 命令串联（前一个成功才执行下一个）
- `||`: 命令或（前一个失败才执行下一个）
- `;`: 命令顺序执行
- `|`: 管道传递

## shode repl - 交互式环境

启动交互式 Read-Eval-Print Loop 环境。

### 语法
```bash
shode repl [options]
```

### 选项
- `-h, --help`: 显示帮助信息
- `--no-color`: 禁用颜色输出
- `--history-size`: 设置历史记录大小（默认1000）

### REPL 特性
- **代码补全**: Tab 键补全函数和变量名
- **历史记录**: 上下箭头浏览历史命令
- **多行输入**: 自动检测多行语句
- **语法高亮**: 彩色语法高亮显示
- **内联帮助**: 输入 `help()` 查看帮助

### REPL 命令
- `exit` 或 `quit`: 退出 REPL
- `clear`: 清空屏幕
- `history`: 显示命令历史
- `help([function])`: 显示帮助信息

### 示例
```bash
# 启动 REPL
shode repl

# REPL 会话示例
> name = "Shode"
> println("Hello, " + upper(name) + "!")
> result = 10 + 5 * 2
> println("结果:", result)
> exit
```

## shode pkg - 包管理

管理 Shode 包依赖和模块。

### 子命令
```bash
shode pkg install    [package]    # 安装包
shode pkg uninstall  [package]    # 卸载包  
shode pkg list                    # 列出已安装包
shode pkg search     [query]      # 搜索包
shode pkg update     [package]    # 更新包
shode pkg init                    # 初始化包配置
```

### shode pkg install
安装包或依赖。

**语法**:
```bash
shode pkg install [options] <package[@version]>
```

**选项**:
- `-g, --global`: 全局安装
- `-S, --save`: 保存到 dependencies
- `-D, --save-dev`: 保存到 devDependencies
- `-h, --help`: 显示帮助

**示例**:
```bash
# 安装最新版本
shode pkg install lodash

# 安装指定版本
shode pkg install lodash@1.2.0

# 全局安装
shode pkg install -g jest

# 保存到依赖
shode pkg install -S axios
```

### shode pkg uninstall
卸载包。

**语法**:
```bash
shode pkg uninstall [options] <package>
```

**选项**:
- `-g, --global`: 卸载全局包
- `-S, --save`: 从 dependencies 移除
- `-D, --save-dev`: 从 devDependencies 移除
- `-h, --help`: 显示帮助

**示例**:
```bash
# 卸载本地包
shode pkg uninstall lodash

# 卸载全局包
shode pkg uninstall -g jest

# 移除依赖记录
shode pkg uninstall -S axios
```

### shode pkg list
列出已安装的包。

**语法**:
```bash
shode pkg list [options]
```

**选项**:
- `-g, --global`: 列出全局包
- `--depth`: 显示依赖深度
- `-h, --help`: 显示帮助

**示例**:
```bash
# 列出本地包
shode pkg list

# 列出全局包
shode pkg list -g

# 显示依赖树
shode pkg list --depth=2
```

### shode pkg search
搜索包。

**语法**:
```bash
shode pkg search [options] <query>
```

**选项**:
- `--limit`: 结果数量限制
- `--sort`: 排序方式（name, version, date）
- `-h, --help`: 显示帮助

**示例**:
```bash
# 搜索包
shode pkg search "http"

# 限制结果数量
shode pkg search --limit=10 "test"

# 按日期排序
shode pkg search --sort=date "utility"
```

### shode pkg update
更新包。

**语法**:
```bash
shode pkg update [options] [package]
```

**选项**:
- `-g, --global`: 更新全局包
- `-h, --help`: 显示帮助

**示例**:
```bash
# 更新所有包
shode pkg update

# 更新指定包
shode pkg update lodash

# 更新全局包
shode pkg update -g jest
```

### shode pkg init
初始化包配置。

**语法**:
```bash
shode pkg init [options]
```

**选项**:
- `-y, --yes`: 使用默认配置
- `-h, --help`: 显示帮助

**示例**:
```bash
# 交互式初始化
shode pkg init

# 使用默认配置
shode pkg init -y
```

## shode version - 版本信息

显示版本信息。

### 语法
```bash
shode version [options]
```

### 选项
- `-h, --help`: 显示帮助信息

### 输出信息
- 版本号
- Go 版本
- 构建时间
- Git 提交哈希
- 平台信息

### 示例
```bash
shode version
# 输出示例:
# Shode v0.1.0
# Go: go1.21.0
# Build: 2024-01-15T10:30:00Z
# Commit: a1b2c3d4
# Platform: linux/amd64
```

## shode completion - 自动补全

生成 shell 自动补全脚本。

### 语法
```bash
shode completion [shell]
```

### 支持的 shell
- `bash`: Bash 补全脚本
- `zsh`: Zsh 补全脚本
- `fish`: Fish 补全脚本
- `powershell`: PowerShell 补全脚本

### 示例
```bash
# 生成 Bash 补全
shode completion bash > /etc/bash_completion.d/shode

# 生成 Zsh 补全
shode completion zsh > "${fpath[1]}/_shode"

# 当前 shell 生效
source <(shode completion bash)
```

## 全局选项

所有命令都支持的全局选项：

### `-h, --help`
显示命令帮助信息。

### `-v, --version`
显示版本信息（等同于 `shode version`）。

### `--config`
指定配置文件路径。

### `--log-level`
设置日志级别：debug, info, warn, error。

### `--color`
控制颜色输出：auto, always, never。

## 配置系统

### 配置文件位置
- 全局配置: `~/.shode/config.json`
- 项目配置: `./shode.json`
- 环境变量: `SHODE_*`

### 配置示例
```json
{
  "logLevel": "info",
  "safeMode": false,
  "modulePath": "./sh_modules",
  "cacheDir": "./.shode/cache",
  "environment": {
    "NODE_ENV": "production"
  }
}
```

## 环境变量

### 核心配置
- `SHODE_LOG_LEVEL`: 日志级别
- `SHODE_SAFE_MODE`: 安全模式（true/false）
- `SHODE_MODULE_PATH`: 模块搜索路径
- `SHODE_CACHE_DIR`: 缓存目录

### 网络配置
- `HTTP_PROXY`: HTTP 代理
- `HTTPS_PROXY`: HTTPS 代理
- `NO_PROXY`: 不代理的主机

### 开发配置
- `SHODE_DEV`: 开发模式（true/false）
- `SHODE_DEBUG`: 调试模式（true/false）

## 使用技巧

### 命令组合
```bash
# 执行脚本并处理结果
result=$(shode run process.sh)
echo "处理结果: $result"

# 批量处理文件
for file in *.sh; do
    shode run "$file"
done

# 条件执行
shode run setup.sh && shode run main.sh
```

### 性能监控
```bash
# 计时执行
time shode run large_script.sh

# 内存监控
/usr/bin/time -v shode run memory_intensive.sh
```

### 调试技巧
```bash
# 详细模式跟踪
shode run -v script.sh

# 只解析不执行
shode run --dry-run script.sh

# 输出 AST 结构
shode run --ast script.sh
```

## 错误处理

### 常见错误代码
- `127`: 命令未找到
- `126`: 权限被拒绝
- `125`: 无效退出状态
- `124`: 超时

### 错误处理示例
```bash
# 检查命令是否存在
if command -v shode >/dev/null 2>&1; then
    echo "Shode 已安装"
else
    echo "Shode 未安装"
fi

# 错误处理
if ! shode run script.sh; then
    echo "脚本执行失败"
    exit 1
fi
```

## 最佳实践

1. **使用脚本文件**: 复杂逻辑使用 `.sh` 脚本文件
2. **参数验证**: 在脚本中验证输入参数
3. **错误处理**: 使用 try-catch 处理预期错误
4. **资源清理**: 确保文件描述符和网络连接正确关闭
5. **性能考虑**: 避免在循环中频繁创建进程

通过熟练掌握这些命令行工具，您可以高效地使用 Shode 进行脚本开发和自动化任务。
