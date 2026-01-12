# 安装指南

## 系统要求

- **操作系统**: Linux, macOS, Windows (WSL2)
- **架构**: x86_64, ARM64
- **内存**: 至少 512MB RAM
- **磁盘空间**: 至少 100MB 可用空间

## 安装方式

### 方式一：从源码编译（推荐）

1. **安装 Go 语言环境**（需要 Go 1.18+）：
   ```bash
   # Ubuntu/Debian
   sudo apt install golang-go
   
   # macOS
   brew install go
   ```

2. **克隆源码**：
   ```bash
   git clone https://gitee.com/com_818cloud/shode.git
   cd shode
   ```

3. **编译项目**：
   ```bash
   go build -o shode ./cmd/shode/
   ```

4. **安装到系统**：
   ```bash
   sudo mv shode /usr/local/bin/
   ```

### 方式二：二进制文件安装

1. **下载最新版本**：
   ```bash
   # 从 Gitee Releases 下载对应平台的二进制文件
   wget https://gitee.com/com_818cloud/shode/releases/latest/download/shode-linux-amd64
   ```

2. **设置执行权限并安装**：
   ```bash
   chmod +x shode-linux-amd64
   sudo mv shode-linux-amd64 /usr/local/bin/shode
   ```

## 验证安装

运行以下命令验证安装是否成功：

```bash
# 检查版本
shode version

# 测试基本功能
shode exec "echo Hello World"
```

## 配置环境变量（可选）

```bash
# 设置缓存目录
export SHODE_CACHE_DIR="$HOME/.shode/cache"

# 配置日志级别
export SHODE_LOG_LEVEL="info"  # debug, info, warn, error
```

## 常见问题

### Q: 安装后命令找不到？
A: 确保 `/usr/local/bin` 在您的 `PATH` 环境变量中：
```bash
echo $PATH
```

### Q: 权限被拒绝？
A: 使用 `sudo` 或调整文件权限：
```bash
sudo chmod 755 /usr/local/bin/shode
```

## 下一步

- [快速开始](./quick-start.md)
