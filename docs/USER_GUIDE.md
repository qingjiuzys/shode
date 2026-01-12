# Shode 用户指南

## 快速入门

### 安装

```bash
# 从源码构建
git clone https://gitee.com/com_818cloud/shode.git
cd shode
go build -o shode ./cmd/shode
```

### 基本使用

```bash
# 执行脚本文件
shode run script.sh

# 交互式 REPL
shode repl

# 初始化包
shode pkg init my-package 1.0.0

# 添加依赖
shode pkg add lodash 4.17.21

# 运行脚本
shode pkg run test
```

## 核心功能

### 1. 脚本执行

Shode 支持执行标准的 shell 脚本：

```bash
# script.sh
echo "Hello, World!"
ls -la
cat file.txt
```

```bash
shode run script.sh
```

### 2. 包管理

#### 初始化包

```bash
shode pkg init my-package 1.0.0
```

这会创建 `shode.json` 配置文件。

#### 添加依赖

```bash
shode pkg add lodash 4.17.21
shode pkg add jest 29.0.0 --dev
```

#### 安装依赖

```bash
shode pkg install
```

#### 运行脚本

```bash
# 添加脚本到 shode.json
shode pkg script add test "echo 'Running tests'"

# 运行脚本
shode pkg run test
```

### 3. 模块系统

#### 创建模块

```bash
# 创建模块目录
mkdir my-module
cd my-module

# 创建 index.sh
cat > index.sh << 'EOF'
export_hello() {
    echo "Hello from module"
}

export_world() {
    echo "World from module"
}
EOF
```

#### 使用模块

```bash
# 在脚本中导入
import my-module

# 使用导出的函数
hello
world
```

### 4. 控制流

#### If 语句

```bash
if test -f file.txt; then
    echo "File exists"
else
    echo "File not found"
fi
```

#### For 循环

```bash
for item in 1 2 3 4 5; do
    echo "Item: $item"
done
```

#### While 循环

```bash
counter=0
while [ $counter -lt 10 ]; do
    echo "Counter: $counter"
    counter=$((counter + 1))
done
```

### 5. 管道和重定向

```bash
# 管道
echo "hello" | cat | grep "h"

# 输出重定向
echo "test" > output.txt

# 追加重定向
echo "more" >> output.txt

# 输入重定向
cat < input.txt
```

## 最佳实践

### 1. 错误处理

```bash
# 检查命令执行结果
if ! command; then
    echo "Command failed"
    exit 1
fi
```

### 2. 函数定义

```bash
# 定义函数
my_function() {
    local arg1=$1
    local arg2=$2
    echo "Args: $arg1 $arg2"
}

# 调用函数
my_function "hello" "world"
```

### 3. 环境变量

```bash
# 设置环境变量
export MY_VAR="value"

# 使用环境变量
echo $MY_VAR
```

### 4. 模块化

将功能拆分为可重用的模块：

```bash
# utils.sh
export_format_date() {
    date +"%Y-%m-%d"
}

# main.sh
import utils
format_date
```

## 故障排除

### 常见问题

#### 1. 命令未找到

**问题**: `command not found`

**解决方案**:
- 检查命令是否在 PATH 中
- 使用完整路径执行命令
- 检查命令权限

#### 2. 权限 denied

**问题**: `Permission denied`

**解决方案**:
- 检查文件权限: `chmod +x script.sh`
- 检查目录权限
- 使用 `sudo`（谨慎使用）

#### 3. 模块加载失败

**问题**: `Module not found`

**解决方案**:
- 检查模块路径是否正确
- 确认模块目录存在 `index.sh` 或 `package.json`
- 检查模块导出语法

#### 4. 超时错误

**问题**: `Operation timed out`

**解决方案**:
- 检查命令执行时间
- 增加超时时间（如果支持）
- 优化长时间运行的命令

### 调试技巧

#### 1. 启用详细输出

```bash
# 使用 -v 标志
shode run -v script.sh
```

#### 2. 检查执行结果

```bash
# 查看命令退出码
echo $?

# 查看输出
command > output.txt 2>&1
```

#### 3. 使用调试工具

```bash
# 使用 shode-debug 工具
shode-debug script.sh
```

## 性能优化

### 1. 使用缓存

Shode 自动缓存命令执行结果。确保：
- 命令参数稳定
- 避免使用随机参数
- 合理设置缓存过期时间

### 2. 减少进程创建

- 使用内置函数替代外部命令
- 批量处理数据
- 使用管道而非临时文件

### 3. 优化循环

```bash
# 避免在循环中执行外部命令
# 不好
for i in 1 2 3; do
    echo $i > file.txt
done

# 更好
echo -e "1\n2\n3" > file.txt
```

## 安全建议

### 1. 输入验证

```bash
# 验证用户输入
if [ -z "$input" ]; then
    echo "Input required"
    exit 1
fi
```

### 2. 避免命令注入

```bash
# 危险
eval "$user_input"

# 安全
echo "$user_input"
```

### 3. 文件操作

```bash
# 检查文件存在
if [ -f "$file" ]; then
    cat "$file"
fi
```

## 示例脚本

### 示例 1: 文件处理

```bash
#!/usr/bin/env shode

# 处理文件列表
for file in *.txt; do
    if [ -f "$file" ]; then
        echo "Processing: $file"
        cat "$file" | grep "pattern"
    fi
done
```

### 示例 2: 数据处理

```bash
#!/usr/bin/env shode

# 读取数据并处理
cat data.txt | \
    grep "filter" | \
    sort | \
    uniq > output.txt
```

### 示例 3: 模块使用

```bash
#!/usr/bin/env shode

import utils
import logger

logger.info "Starting process"
utils.process_data
logger.info "Process complete"
```

## 更多资源

- [API 文档](API.md)
- [开发指南](DEVELOPMENT.md)
- [贡献指南](CONTRIBUTING.md)
