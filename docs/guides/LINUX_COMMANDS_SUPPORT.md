# Shode Linux 命令支持情况

本文档详细说明 Shode 对 Linux 命令的支持情况，包括标准库函数、被阻止的命令和可执行的外部命令。

## 📋 目录

1. [标准库函数（推荐使用）](#标准库函数推荐使用)
2. [被安全沙箱阻止的命令](#被安全沙箱阻止的命令)
3. [可执行的外部命令](#可执行的外部命令)
4. [命令执行模式](#命令执行模式)
5. [最佳实践](#最佳实践)

---

## 标准库函数（推荐使用）

Shode 提供了丰富的标准库函数来替代常见的 Linux 命令，这些函数直接在内存中执行，性能更好、更安全。

### 文件系统操作

| Linux 命令 | Shode 标准库函数 | 说明 |
|-----------|-----------------|------|
| `cat file` | `ReadFile "file"` | 读取文件内容 |
| `echo "text" > file` | `WriteFile "file" "text"` | 写入文件 |
| `ls -la` | `ListFiles "."` | 列出目录文件 |
| `test -f file` | `FileExists "file"` | 检查文件是否存在 |
| `pwd` | `WorkingDir` | 获取当前工作目录 |
| `cd /path` | `ChangeDir "/path"` | 切换目录 |

**示例**:
```bash
# 读取文件
content = ReadFile "config.json"
Println content

# 写入文件
WriteFile "output.txt" "Hello World"

# 列出文件
files = ListFiles "."
Println files

# 检查文件存在
if FileExists "config.json" {
    Println "Config file exists"
}
```

### 字符串操作

| Linux 命令 | Shode 标准库函数 | 说明 |
|-----------|-----------------|------|
| `grep "pattern"` | `Contains "text" "pattern"` | 检查字符串包含 |
| `sed 's/old/new/g'` | `Replace "text" "old" "new"` | 字符串替换 |
| `tr '[:lower:]' '[:upper:]'` | `ToUpper "text"` | 转大写 |
| `tr '[:upper:]' '[:lower:]'` | `ToLower "text"` | 转小写 |
| `xargs` | `Trim "text"` | 去除首尾空格 |

**示例**:
```bash
text = "Hello World"
Println ToUpper text        # HELLO WORLD
Println ToLower text        # hello world
Println Replace text "World" "Shode"  # Hello Shode
Println Contains text "Hello"  # true
```

### 环境变量操作

| Linux 命令 | Shode 标准库函数 | 说明 |
|-----------|-----------------|------|
| `echo $VAR` | `GetEnv "VAR"` | 获取环境变量 |
| `export VAR=value` | `SetEnv "VAR" "value"` | 设置环境变量 |
| `unset VAR` | `UnsetEnv "VAR"` | 删除环境变量 |

**示例**:
```bash
SetEnv "APP_ENV" "production"
env = GetEnv "APP_ENV"
Println "Environment: " + env
```

### 输出操作

| Linux 命令 | Shode 标准库函数 | 说明 |
|-----------|-----------------|------|
| `echo -n "text"` | `Print "text"` | 打印（不换行） |
| `echo "text"` | `Println "text"` | 打印（换行） |
| `echo "error" >&2` | `Error "error"` | 打印到stderr |

**示例**:
```bash
Print "Loading..."
Println "Complete"
Error "Error occurred"
```

### HTTP 服务器

| 功能 | Shode 标准库函数 | 说明 |
|-----|-----------------|------|
| 启动服务器 | `StartHTTPServer "9188"` | 启动HTTP服务器 |
| 注册路由 | `RegisterHTTPRoute "GET" "/api" "function" "handler"` | 注册HTTP路由 |
| 停止服务器 | `StopHTTPServer` | 停止HTTP服务器 |
| 获取请求信息 | `GetHTTPMethod`, `GetHTTPPath`, `GetHTTPQuery`, `GetHTTPHeader`, `GetHTTPBody` | 获取HTTP请求信息 |
| 设置响应 | `SetHTTPResponse 200 "OK"` | 设置HTTP响应 |

**示例**:
```bash
StartHTTPServer "9188"
RegisterHTTPRoute "GET" "/api/users" "function" "handleGetUsers"

function handleGetUsers() {
    method = GetHTTPMethod
    path = GetHTTPPath
    SetHTTPResponse 200 "Users list"
}
```

### 缓存操作

| 功能 | Shode 标准库函数 | 说明 |
|-----|-----------------|------|
| 设置缓存 | `SetCache "key" "value" 3600` | 设置缓存（带TTL） |
| 获取缓存 | `GetCache "key"` | 获取缓存 |
| 删除缓存 | `DeleteCache "key"` | 删除缓存 |
| 清空缓存 | `ClearCache` | 清空所有缓存 |
| 检查存在 | `CacheExists "key"` | 检查缓存是否存在 |

**示例**:
```bash
SetCache "user:1" "John Doe" 3600
user = GetCache "user:1"
if CacheExists "user:1" {
    Println "User found in cache"
}
```

### 数据库操作

| 功能 | Shode 标准库函数 | 说明 |
|-----|-----------------|------|
| 连接数据库 | `ConnectDB "mysql" "user:pass@tcp(host:3306)/db"` | 连接数据库 |
| 查询 | `QueryDB "SELECT * FROM users"` | 执行查询 |
| 单行查询 | `QueryRowDB "SELECT * FROM users WHERE id = ?" 1` | 查询单行 |
| 执行 | `ExecDB "INSERT INTO users (name) VALUES (?)" "John"` | 执行SQL |
| 获取结果 | `GetQueryResult` | 获取查询结果（JSON） |

**示例**:
```bash
ConnectDB "sqlite" "app.db"
ExecDB "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)"
ExecDB "INSERT INTO users (name) VALUES (?)" "John"
QueryDB "SELECT * FROM users"
result = GetQueryResult
Println result
```

### IoC 容器

| 功能 | Shode 标准库函数 | 说明 |
|-----|-----------------|------|
| 注册Bean | `RegisterBean "userService" "UserService"` | 注册Bean |
| 获取Bean | `GetBean "userService"` | 获取Bean |
| 检查存在 | `ContainsBean "userService"` | 检查Bean是否存在 |

### 配置管理

| 功能 | Shode 标准库函数 | 说明 |
|-----|-----------------|------|
| 加载配置 | `LoadConfig "application.json"` | 加载配置文件 |
| 获取配置 | `GetConfigString "server.port" "8080"` | 获取字符串配置 |
| 获取整数 | `GetConfigInt "server.port" 8080` | 获取整数配置 |
| 获取布尔 | `GetConfigBool "cache.enabled" false` | 获取布尔配置 |
| 设置配置 | `SetConfig "server.port" "9188"` | 设置配置值 |

**示例**:
```bash
LoadConfig "application.json"
port = GetConfigString "server.port" "8080"
enabled = GetConfigBool "cache.enabled" false
```

---

## 被安全沙箱阻止的命令

Shode 的安全沙箱会自动阻止以下危险命令，以保护系统安全。

### 破坏性操作命令

| 命令 | 说明 | 替代方案 |
|-----|------|---------|
| `rm` | 文件删除 | 使用标准库函数或谨慎使用外部命令 |
| `dd` | 磁盘操作 | 不允许 |
| `mkfs` | 文件系统操作 | 不允许 |
| `fdisk` | 分区操作 | 不允许 |

### 系统控制命令

| 命令 | 说明 | 替代方案 |
|-----|------|---------|
| `shutdown` | 系统关机 | 不允许 |
| `reboot` | 系统重启 | 不允许 |
| `halt` | 系统停机 | 不允许 |
| `poweroff` | 系统断电 | 不允许 |

### 权限管理命令

| 命令 | 说明 | 替代方案 |
|-----|------|---------|
| `chmod` | 修改文件权限 | 不允许 |
| `chown` | 修改文件所有者 | 不允许 |
| `useradd` | 添加用户 | 不允许 |
| `userdel` | 删除用户 | 不允许 |
| `groupadd` | 添加组 | 不允许 |
| `groupdel` | 删除组 | 不允许 |
| `passwd` | 修改密码 | 不允许 |

### 网络操作命令

| 命令 | 说明 | 替代方案 |
|-----|------|---------|
| `iptables` | 防火墙规则 | 不允许 |
| `ufw` | 防火墙管理 | 不允许 |
| `route` | 路由管理 | 不允许 |
| `ifconfig` | 网络接口配置 | 不允许 |
| `ip` | 网络配置 | 不允许 |
| `nc` | Netcat | 不允许 |
| `nmap` | 网络扫描 | 不允许 |
| `tcpdump` | 网络抓包 | 不允许 |

### 受保护的文件和目录

以下文件和目录的访问会被阻止：

- `/etc/passwd` - 用户账户信息
- `/etc/shadow` - 密码哈希
- `/etc/sudoers` - sudo配置
- `/root/` - root用户目录
- `/boot/` - 启动文件
- `/dev/` - 设备文件
- `/proc/` - 进程信息
- `/sys/` - 系统信息
- `/var/log/` - 日志文件

### 模式检测

安全沙箱还会检测以下危险模式：

- **递归删除根目录**: `rm -rf /`
- **命令行密码**: `-p password` 或 `--password secret`
- **Shell注入**: `;`, `&`, `|`, `` ` ``, `$()`

---

## 可执行的外部命令

除了标准库函数和被阻止的命令外，Shode 可以执行系统中可用的其他 Linux 命令。

### 常用可执行命令

| 命令类别 | 示例命令 | 说明 |
|---------|---------|------|
| 文本处理 | `grep`, `awk`, `sed`, `sort`, `uniq`, `wc` | 文本处理工具 |
| 文件操作 | `cp`, `mv`, `mkdir`, `rmdir`, `touch` | 文件操作（注意：`rm`被阻止） |
| 系统信息 | `uname`, `hostname`, `whoami`, `date`, `uptime` | 系统信息查询 |
| 进程管理 | `ps`, `top`, `kill`（有限制） | 进程查看和管理 |
| 网络工具 | `curl`, `wget`, `ping` | 网络工具（注意：某些网络配置命令被阻止） |
| 压缩工具 | `tar`, `gzip`, `zip`, `unzip` | 压缩和解压 |
| 查找工具 | `find`, `locate`, `which` | 文件查找 |
| 其他工具 | `echo`, `printf`, `cat`（可用但推荐用标准库） | 基础工具 |

### 执行示例

```bash
# 文本处理
grep "pattern" file.txt
awk '{print $1}' file.txt
sort file.txt | uniq

# 系统信息
date
hostname
whoami

# 网络工具
curl https://api.example.com/data
ping -c 3 8.8.8.8

# 文件操作（安全）
cp source.txt dest.txt
mkdir -p /tmp/test
touch /tmp/test/file.txt
```

### 注意事项

1. **安全检查**: 所有外部命令都会经过安全检查
2. **敏感文件**: 访问敏感文件会被阻止
3. **性能**: 外部命令需要创建进程，性能不如标准库函数
4. **跨平台**: 某些Linux特定命令在macOS/Windows上可能不可用

---

## 命令执行模式

Shode 支持三种命令执行模式：

### 1. 解释执行模式（标准库函数）

标准库函数直接在内存中执行，无需创建进程：

```bash
ReadFile "file.txt"    # 直接内存执行，快速
Println "Hello"         # 直接内存执行，快速
SetCache "key" "value"  # 直接内存执行，快速
```

**优势**:
- ✅ 无进程创建开销
- ✅ 执行速度快（毫秒级）
- ✅ 资源占用低
- ✅ 跨平台一致

### 2. 进程执行模式（外部命令）

外部命令通过创建进程执行：

```bash
ls -la                  # 创建进程执行
grep "pattern" file.txt # 创建进程执行
curl https://api.com    # 创建进程执行
```

**特点**:
- ⚠️ 需要创建进程（较慢）
- ⚠️ 资源占用较高
- ⚠️ 依赖系统环境

### 3. 混合模式（智能选择）

Shode 会自动选择最优执行方式（未来增强）：

```bash
# 自动选择标准库函数或外部命令
command arg1 arg2
```

---

## 最佳实践

### 1. 优先使用标准库函数

```bash
# ❌ 不推荐：使用外部命令
cat file.txt
echo "text" > file.txt
ls -la

# ✅ 推荐：使用标准库函数
ReadFile "file.txt"
WriteFile "file.txt" "text"
ListFiles "."
```

**原因**:
- 性能更好（无进程开销）
- 更安全（经过安全检查）
- 跨平台一致

### 2. 避免使用被阻止的命令

```bash
# ❌ 会被阻止
rm -rf /
chmod 777 file.txt
iptables -A INPUT -j DROP

# ✅ 使用替代方案
# 文件删除：谨慎使用，或使用标准库函数
# 权限修改：不允许，使用系统管理工具
# 防火墙：不允许，使用系统管理工具
```

### 3. 处理敏感文件

```bash
# ❌ 会被阻止
cat /etc/passwd
rm /etc/shadow

# ✅ 使用安全的文件路径
ReadFile "/tmp/config.json"
WriteFile "/tmp/output.txt" "data"
```

### 4. 使用管道和重定向

```bash
# ✅ 支持管道
cat file.txt | grep "pattern" | wc -l

# ✅ 支持重定向
echo "Hello" > output.txt
cat < input.txt > output.txt
```

### 5. 错误处理

```bash
# ✅ 检查命令执行结果
if FileExists "config.json" {
    content = ReadFile "config.json"
    Println content
} else {
    Error "Config file not found"
}
```

---

## 总结

### ✅ 推荐使用
- **标准库函数**: 性能好、安全、跨平台
- **常用外部命令**: `grep`, `awk`, `sed`, `curl`, `date` 等
- **Shell特性**: 管道、重定向、变量展开、命令替换

### ❌ 被阻止
- **危险命令**: `rm`, `dd`, `shutdown`, `chmod` 等
- **网络配置**: `iptables`, `ifconfig`, `route` 等
- **敏感文件**: `/etc/passwd`, `/root/` 等

### ⚠️ 注意事项
- 外部命令需要创建进程，性能较低
- 某些命令可能在不同系统上不可用
- 所有命令都会经过安全检查

---

## 相关文档

- [Shell特性清单](./SHELL_FEATURES.md)
- [执行引擎指南](./execution-engine.md)
- [安全沙箱文档](../sandbox/security.go)
