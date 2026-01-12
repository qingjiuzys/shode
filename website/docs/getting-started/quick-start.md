# 快速开始

## 第一个 Shode 脚本

### 1. 简单的问候脚本

创建一个名为 `hello.sh` 的文件：

```shode
# 使用 Shode 标准库函数
Println "=== 欢迎使用 Shode ==="

# 字符串操作示例
name = "世界"
Println "你好, " + name + "!"
```

运行脚本：
```bash
shode run hello.sh
```

### 2. 文件操作示例

```shode
# 创建测试文件
content = "这是测试文件内容"
WriteFile "test.txt" content

# 读取文件
fileContent = ReadFile "test.txt"
Println "文件内容: " + fileContent

# 检查文件是否存在
if FileExists "test.txt" {
    Println "文件存在"
}
```

### 3. HTTP 服务器示例

```shode
# 启动 HTTP 服务器
StartHTTPServer "9188"

# 注册路由
RegisterRouteWithResponse "/" "Hello, Shode!"

Println "服务器运行在 http://localhost:9188"
```

运行：
```bash
shode run http_server.sh
```

## 命令行使用

### 执行单条命令

```bash
# 执行标准库函数
shode exec "upper hello world"
shode exec "trim '   测试   '"

# 执行外部命令
shode exec "echo Hello from Shode"
```

### 交互式 REPL 环境

启动交互式 Shell：
```bash
shode repl
```

在 REPL 中尝试：
```bash
> Println "Hello, Shode"
> name = "Shode"
> Println "Hello, " + name
> exit
```

## 常用标准库函数

### 字符串处理
```shode
upper("hello")        # HELLO
lower("WORLD")        # world  
trim("   text   ")    # text
contains("hello", "ell")  # true
```

### 文件操作
```shode
WriteFile "file.txt" "内容"    # 写文件
ReadFile "file.txt"            # 读文件
FileExists "file.txt"          # 检查存在
ListFiles "."                  # 列出文件
```

### HTTP 服务器
```shode
StartHTTPServer "9188"                    # 启动服务器
RegisterHTTPRoute "GET" "/api" "function" "handler"  # 注册路由
```

## 实用技巧

### 1. 调试脚本

使用 `-v` 参数查看详细执行信息：
```bash
shode run -v script.sh
```

### 2. 错误处理

```shode
# 安全执行
content = ReadFile "nonexistent.txt"
if content == "" {
    Println "文件不存在或读取失败"
}
```

## 下一步学习

1. **掌握基础**: 熟练使用常用标准库函数
2. **脚本编写**: 编写复杂的自动化脚本
3. **查看示例**: 浏览 [示例文档](../examples/index.md)
4. **API参考**: 查阅 [完整API文档](../api/stdlib.md)

## 获取帮助

- 查看所有命令：`shode --help`
- 查看具体命令帮助：`shode run --help`
- 查阅完整文档：[API参考](../api/stdlib.md)

现在您已经掌握了 Shode 的基本用法，开始编写您的第一个脚本吧！
