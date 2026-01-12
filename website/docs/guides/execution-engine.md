# 执行引擎指南

## 概述

Shode 的执行引擎提供了完整的运行时环境，支持管道、重定向、控制流和安全沙箱等高级特性。

## 核心特性

### 1. 管道支持

执行命令时支持数据流通过管道传递：

```shode
# 简单管道
cat file.txt | grep "pattern" | wc -l

# 复杂管道
ListFiles "." | grep "test" | upper
```

**工作原理：**
- 每个命令的输出作为下一个命令的输入
- 如果任何命令失败，执行停止
- 返回最后一个命令的输出

### 2. 输入/输出重定向

支持所有标准重定向操作符：

```shode
# 输出重定向（覆盖）
Println "Hello World" > output.txt

# 输出重定向（追加）
Println "More text" >> output.txt

# 输入重定向
ReadFile < input.txt
```

### 3. 控制流

#### If-Then-Else 语句

```shode
if FileExists "file.txt" {
    Println "File exists"
} else {
    Println "File not found"
}
```

#### For 循环

```shode
for item in file1 file2 file3 {
    Println "Processing " + item
    ReadFile item
}
```

#### While 循环

```shode
count = 0
while count < 5 {
    Println "Count: " + count
    count = count + 1
}
```

**安全特性：**
- 最大迭代限制（10,000）防止无限循环
- 支持上下文超时
- 正确的变量作用域

### 4. 变量赋值

```shode
# 简单赋值
name = "John"

# 在命令中使用
Println "Hello, " + name

# 环境变量
SetEnv "PATH" "/usr/local/bin:" + GetEnv "PATH"
```

### 5. 安全沙箱

所有命令都会经过安全检查：

**被阻止的危险命令：**
- `rm -rf /`
- `format`
- `dd if=/dev/zero`
- 其他危险操作

**保护的文件：**
- `/etc/passwd`
- `/etc/shadow`
- 系统关键文件

## 执行模式

### 解释执行模式

标准库函数直接在内存中执行，无需创建进程：

```shode
upper("hello")      # 直接执行，快速
ReadFile "file.txt" # 直接执行，快速
```

### 进程执行模式

外部命令通过进程执行：

```shode
echo "hello"        # 创建进程执行
ls -la              # 创建进程执行
```

### 智能选择

执行引擎会根据命令类型自动选择最优执行方式。

## 性能优化

### 命令缓存

频繁执行的命令会被缓存，提升性能：

```shode
# 第一次执行：创建进程
ls -la

# 第二次执行：使用缓存
ls -la
```

### 进程池

复用进程资源，减少创建开销。

## 错误处理

### 错误类型

- **SecurityViolation**: 安全违规
- **CommandNotFound**: 命令未找到
- **ExecutionFailed**: 执行失败
- **Timeout**: 超时

### 错误恢复

- 超时处理和资源清理
- 优雅降级（缓存失败时直接执行）
- 部分失败处理

## 相关文档

- [用户指南](./user-guide.md)
- [API 参考](../api/stdlib.md)
