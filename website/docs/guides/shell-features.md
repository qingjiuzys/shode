# Shode Shell 特性清单

本文档详细列出了 Shode 已集成的所有 Shell 特性。

## 📋 目录

1. [控制流语句](#控制流语句)
2. [管道和重定向](#管道和重定向)
3. [变量系统](#变量系统)
4. [函数系统](#函数系统)
5. [模块系统](#模块系统)
6. [注解系统](#注解系统)
7. [注释支持](#注释支持)
8. [执行模式](#执行模式)
9. [安全特性](#安全特性)
10. [性能优化](#性能优化)

---

## 控制流语句

### ✅ If-Then-Else 语句

支持条件分支执行：

```bash
if test -f file.txt; then
    echo "File exists"
else
    echo "File not found"
fi
```

**实现位置**: `pkg/engine/engine.go:ExecuteIf()`

### ✅ For 循环

支持列表迭代：

```bash
for item in file1 file2 file3; do
    echo "Processing $item"
    cat "$item"
done
```

**实现位置**: `pkg/engine/engine.go:ExecuteFor()`

### ✅ While 循环

支持条件循环（带安全限制）：

```bash
count=0
while [ $count -lt 5 ]; do
    echo "Count: $count"
    count=$((count + 1))
done
```

**安全特性**:
- 最大迭代限制：10,000 次（防止无限循环）
- 支持上下文超时
- 正确的变量作用域

**实现位置**: `pkg/engine/engine.go:ExecuteWhile()`

### ✅ Break 语句

支持提前退出循环：

```bash
for item in a b c d e; do
    echo "Processing $item"
    if [ "$item" = "c" ]; then
        break
    fi
done
```

**实现位置**: `pkg/types/ast.go:BreakNode`

### ✅ Continue 语句

支持跳过当前迭代：

```bash
for item in a b c d e; do
    if [ "$item" = "c" ]; then
        continue
    fi
    echo "Processing $item"
done
```

**实现位置**: `pkg/types/ast.go:ContinueNode`

---

## 管道和重定向

### ✅ 管道 (|)

支持命令间的数据流传递：

```bash
# 简单管道
cat file.txt | grep "pattern" | wc -l

# 复杂管道
ls -la | awk '{print $9}' | sort | uniq
```

**工作原理**:
- 前一个命令的输出作为后一个命令的输入
- 如果任何命令失败，执行停止
- 返回最后一个命令的输出

**实现位置**: `pkg/engine/engine.go:ExecutePipeline()`

### ✅ 输出重定向 (>)

覆盖文件输出：

```bash
echo "Hello World" > output.txt
```

**实现位置**: `pkg/engine/engine.go:setupRedirect()`

### ✅ 追加重定向 (>>)

追加到文件：

```bash
echo "More text" >> output.txt
```

**实现位置**: `pkg/engine/engine.go:setupRedirect()`

### ✅ 输入重定向 (<)

从文件读取输入：

```bash
cat < input.txt
```

**实现位置**: `pkg/engine/engine.go:setupRedirect()`

### ✅ 错误重定向 (2>&1)

将标准错误重定向到标准输出：

```bash
command 2>&1
```

**实现位置**: `pkg/engine/engine.go:setupRedirect()`

### ✅ 合并重定向 (&>)

同时重定向标准输出和标准错误：

```bash
command &> all_output.txt
```

**实现位置**: `pkg/engine/engine.go:setupRedirect()`

### ✅ 文件描述符支持

支持指定文件描述符（0=stdin, 1=stdout, 2=stderr）：

```bash
command 1> output.txt 2> error.txt
```

**实现位置**: `pkg/types/ast.go:RedirectNode`

---

## 变量系统

### ✅ 变量赋值

支持简单的变量赋值：

```bash
NAME="John"
VERSION="0.2.0"
count=10
```

**实现位置**: 
- `pkg/types/ast.go:AssignmentNode`
- `pkg/parser/simple_parser.go:parseAssignment()`

### ✅ 变量展开 ($VAR)

支持标准变量展开语法：

```bash
echo "Hello, $NAME"
echo "Version: $VERSION"
```

**实现位置**: `pkg/engine/variable_expansion.go:expandVariables()`

### ✅ 变量展开 (${VAR})

支持花括号变量展开：

```bash
echo "Hello, ${NAME}"
echo "File: ${FILE}_backup.txt"
```

**实现位置**: `pkg/engine/variable_expansion.go:expandVariables()`

### ✅ 字符串拼接

支持字符串连接操作：

```bash
fullName = firstName + " " + lastName
message = "Hello, " + name
```

**实现位置**: `pkg/engine/variable_expansion.go:splitStringConcat()`

### ✅ 环境变量管理

支持环境变量的设置和获取：

```bash
export PATH="/usr/local/bin:$PATH"
export SHODE_ENV="production"
```

**实现位置**: `pkg/environment/manager.go`

---

## 函数系统

### ✅ 函数定义

支持用户自定义函数：

```bash
greet() {
    echo "Hello, $1"
    echo "Welcome to Shode!"
}
```

**实现位置**: 
- `pkg/types/ast.go:FunctionNode`
- `pkg/engine/engine.go:Execute()`

### ✅ 函数调用

支持函数调用：

```bash
greet "Alice"
```

**实现位置**: `pkg/engine/engine.go:executeUserFunction()`

### ✅ 函数参数

支持标准函数参数：

```bash
my_function() {
    echo "Function name: $0"
    echo "First argument: $1"
    echo "Second argument: $2"
    echo "All arguments: $@"
    echo "Argument count: $#"
}
```

**支持的参数变量**:
- `$0`: 函数名
- `$1, $2, ...`: 位置参数
- `$@`: 所有参数
- `$#`: 参数个数

**实现位置**: `pkg/engine/engine.go:executeUserFunction()`

### ✅ 函数作用域隔离

函数执行时具有独立的作用域，不会影响外部环境：

```bash
outer_var="outer"
my_function() {
    inner_var="inner"
    echo "$inner_var"
}
my_function
echo "$outer_var"  # 仍然可用
```

**实现位置**: `pkg/engine/engine.go:executeUserFunction()`

---

## 模块系统

### ✅ 模块导入/导出

支持模块的导入和导出：

```bash
# my-module/index.sh
export_hello() {
    echo "Hello from module!"
}

export_greet() {
    echo "Greetings, $1!"
}

# main.sh
import my-module
hello
greet "Alice"
```

**实现位置**: `pkg/module/manager.go`

### ✅ package.json 支持

支持 Node.js 风格的 package.json：

```json
{
  "name": "my-package",
  "version": "1.0.0",
  "main": "index.sh",
  "exports": {
    "hello": "./functions/hello.sh",
    "utils": "./utils.sh"
  }
}
```

**实现位置**: `pkg/module/manager.go`

### ✅ 路径解析

支持多种模块路径：
- 本地文件路径
- `node_modules` 包
- 相对路径和绝对路径

**实现位置**: `pkg/module/manager.go`

---

## 注解系统

### ✅ 简单注解

支持 `@AnnotationName` 语法：

```bash
@RestController
my_handler() {
    echo "Hello World"
}
```

**实现位置**: 
- `pkg/types/ast.go:AnnotationNode`
- `pkg/parser/simple_parser.go:parseAnnotation()`

### ✅ 带参数的注解

支持 `@AnnotationName(key=value, ...)` 语法：

```bash
@Route(path="/api/users", method="GET")
get_users() {
    echo "User list"
}
```

**实现位置**: `pkg/annotation/parser.go`

### ✅ 注解处理

支持注解的注册和处理：

```bash
@Transactional
transfer_money() {
    # 转账逻辑
}
```

**实现位置**: `pkg/annotation/processor.go`

---

## 注释支持

### ✅ 单行注释

支持 `#` 开头的单行注释：

```bash
# This is a comment
echo "Hello"  # Inline comment
```

**实现位置**: `pkg/parser/simple_parser.go`

---

## 执行模式

### ✅ 解释执行模式

标准库函数直接在内存中执行，无需创建进程：

```bash
Println "Hello World"      # 直接执行，快速
ReadFile "file.txt"         # 直接执行，快速
```

**优势**:
- 无进程创建开销
- 执行速度快
- 资源占用低

**实现位置**: `pkg/engine/engine.go:executeStdLibFunction()`

### ✅ 进程执行模式

外部命令通过进程执行：

```bash
ls -la                      # 创建进程执行
grep "pattern" file.txt     # 创建进程执行
```

**实现位置**: `pkg/engine/engine.go:executeProcess()`

### ✅ 混合模式

智能选择执行模式（未来增强）：

```bash
# 自动选择最优执行方式
command arg1 arg2
```

**实现位置**: `pkg/engine/engine.go:executeHybrid()`

---

## 安全特性

### ✅ 命令黑名单

自动拦截危险命令：

```bash
rm -rf /                    # 被阻止
dd if=/dev/zero            # 被阻止
shutdown -h now             # 被阻止
```

**被阻止的命令类型**:
- 破坏性操作：`rm`, `dd`, `mkfs`, `fdisk`
- 系统控制：`shutdown`, `reboot`, `halt`
- 权限修改：`chmod`, `chown`, `passwd`
- 网络操作：`iptables`, `ufw`, `route`

**实现位置**: `pkg/sandbox/security.go`

### ✅ 敏感文件保护

保护系统关键文件：

```bash
cat /etc/passwd             # 被阻止
rm /etc/shadow              # 被阻止
```

**受保护的文件/目录**:
- `/etc/passwd`, `/etc/shadow`, `/etc/sudoers`
- `/root/`, `/boot/`, `/dev/`, `/proc/`, `/sys/`

**实现位置**: `pkg/sandbox/security.go`

### ✅ 模式检测

检测 Shell 注入攻击：

```bash
command; rm -rf /           # 被检测
command $(rm -rf /)         # 被检测
```

**实现位置**: `pkg/sandbox/security.go`

---

## 性能优化

### ✅ 命令缓存

自动缓存命令执行结果：

```bash
# 第一次执行
cat large_file.txt | wc -l  # 执行并缓存

# 后续执行（相同命令）
cat large_file.txt | wc -l  # 从缓存读取
```

**特性**:
- TTL 过期机制
- 可配置缓存大小（默认 1000 条）
- 自动淘汰最旧条目

**实现位置**: `pkg/engine/command_cache.go`

### ✅ 进程池

重用进程以减少创建开销：

```bash
# 重复执行的命令会重用进程
for i in 1 2 3 4 5; do
    echo "Iteration $i"
done
```

**特性**:
- 可配置池大小（默认 10 个进程）
- 空闲超时清理
- 自动资源管理

**实现位置**: `pkg/engine/process_pool.go`

### ✅ 性能指标收集

收集执行性能数据：

```bash
# 自动收集：
# - 命令执行时间
# - 缓存命中率
# - 进程池使用率
# - 内存使用情况
# - 错误率
```

**实现位置**: `pkg/metrics/metrics.go`

---

## 总结

Shode 已集成以下 Shell 特性：

### ✅ 已实现
- ✅ 控制流：if/for/while/break/continue
- ✅ 管道和重定向：|, >, >>, <, 2>&1, &>
- ✅ 变量系统：赋值、展开、拼接
- ✅ 命令替换：`$(command)` 和 `` `command` ``
- ✅ 数组支持：`array=(value1 value2)`
- ✅ 后台任务：`command &`
- ✅ 函数系统：定义、调用、参数、作用域
- ✅ 模块系统：导入/导出、package.json
- ✅ 注解系统：简单注解、带参数注解
- ✅ 注释支持：单行注释
- ✅ 执行模式：解释执行、进程执行、混合模式
- ✅ 安全特性：命令黑名单、文件保护、模式检测
- ✅ 性能优化：命令缓存、进程池、性能指标

### ✅ 新增实现
- ✅ 后台任务支持 (`&`) - 命令后添加 `&` 在后台执行
- ✅ 命令替换 (`$(...)`) - 支持 `$(command)` 和 `` `command` `` 语法
- ✅ 数组支持 - 支持 `array=(value1 value2 value3)` 语法

### 🚧 计划中
- ⏳ 进程替换 (`<(...)`)
- ⏳ 关联数组支持
- ⏳ 信号处理
- ⏳ 调试器集成

---

## 相关文档

- [执行引擎指南](./execution-engine.md)
- [用户指南](../USER_GUIDE.md)
- [标准库文档](../stdlib/README.md)
