# 包管理指南

## 概述

Shode 包管理系统提供了完整的依赖管理和包分发功能，类似于 npm、PyPI 或 RubyGems，但专门为 shell 脚本设计。

## 核心功能

### 1. 包发布

发布包到注册表：

```bash
# 初始化包
shode pkg init my-package 1.0.0

# 发布包
shode pkg publish
```

### 2. 包发现

搜索和发现包：

```bash
# 搜索包
shode pkg search "http"

# 查看包信息
shode pkg info lodash
```

### 3. 包安装

安装包和依赖：

```bash
# 添加依赖
shode pkg add lodash 4.17.21

# 安装所有依赖
shode pkg install
```

### 4. 依赖管理

管理项目依赖：

```bash
# 添加开发依赖
shode pkg add --dev jest 29.7.0

# 列出依赖
shode pkg list

# 更新依赖
shode pkg update lodash
```

## 使用流程

### 初始化项目

```bash
# 创建新项目
mkdir my-project
cd my-project

# 初始化包配置
shode pkg init my-project 1.0.0
```

这会创建 `shode.json` 文件：

```json
{
  "name": "my-project",
  "version": "1.0.0",
  "dependencies": {},
  "devDependencies": {},
  "scripts": {}
}
```

### 添加依赖

```bash
# 添加生产依赖
shode pkg add lodash 4.17.21

# 添加开发依赖
shode pkg add --dev jest 29.7.0
```

### 安装依赖

```bash
# 安装所有依赖
shode pkg install
```

这会：
1. 读取 `shode.json`
2. 从注册表下载依赖
3. 安装到 `sh_modules/` 目录
4. 验证校验和

### 使用依赖

在脚本中导入模块：

```shode
# 导入已安装的包
import lodash

# 使用包中的函数
# (根据包的导出函数使用)
```

### 管理脚本

```bash
# 添加脚本
shode pkg script add test "echo 'Running tests'"
shode pkg script add build "echo 'Building...'"

# 运行脚本
shode pkg run test
shode pkg run build
```

## 包结构

### 标准包结构

```
my-package/
├── index.sh          # 主入口文件
├── package.json      # 包元数据（可选）
├── README.md         # 说明文档
└── ...               # 其他文件
```

### package.json 示例

```json
{
  "name": "my-package",
  "version": "1.0.0",
  "description": "My awesome package",
  "main": "index.sh",
  "exports": {
    "hello": "./functions/hello.sh",
    "world": "./functions/world.sh"
  }
}
```

## 本地缓存

包管理器使用本地缓存提升性能：

- **元数据缓存**: 24小时 TTL
- **Tarball 缓存**: 永久缓存，直到手动清理
- **自动清理**: 超出限制时自动清理旧缓存

## 最佳实践

1. **版本锁定**: 使用精确版本号而非范围
2. **依赖最小化**: 只添加必要的依赖
3. **定期更新**: 定期检查和更新依赖
4. **安全验证**: 验证包的校验和
5. **文档完善**: 为包提供清晰的文档

## 相关文档

- [用户指南](./user-guide.md)
- [API 参考](../api/cli.md)
