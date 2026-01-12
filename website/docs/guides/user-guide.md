# 用户指南

## 核心功能

### 1. 脚本执行

Shode 支持执行标准的 shell 脚本：

```shode
# script.sh
Println "Hello, World!"
ListFiles "."
ReadFile "file.txt"
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

```shode
# my-module/index.sh
export_hello() {
    Println "Hello from module"
}

export_world() {
    Println "World from module"
}
```

#### 使用模块

```shode
# 在脚本中导入
import my-module

# 使用导出的函数
hello
world
```

### 4. 控制流

#### If 语句

```shode
if FileExists "file.txt" {
    Println "File exists"
} else {
    Println "File not found"
}
```

#### For 循环

```shode
for item in 1 2 3 4 5 {
    Println "Item: " + item
}
```

#### While 循环

```shode
counter = 0
while counter < 10 {
    Println "Counter: " + counter
    counter = counter + 1
}
```

### 5. 变量系统

```shode
# 变量赋值
name = "Shode"
port = 9188

# 变量展开
Println "Hello, " + name
Println "Port: " + port

# 使用 ${VAR} 语法
Println "Value: ${name}"
```

### 6. HTTP 服务器

```shode
# 启动服务器
StartHTTPServer "9188"

# 定义处理函数
function handleRequest() {
    SetHTTPResponse 200 "Hello, World!"
}

# 注册路由
RegisterHTTPRoute "GET" "/api" "function" "handleRequest"
```

### 7. 缓存系统

```shode
# 设置缓存（TTL 60秒）
SetCache "key1" "value1" 60

# 获取缓存
value = GetCache "key1"

# 检查是否存在
exists = CacheExists "key1"

# 删除缓存
DeleteCache "key1"
```

### 8. 数据库操作

```shode
# 连接数据库
ConnectDB "sqlite" "app.db"

# 执行查询
QueryDB "SELECT * FROM users"
result = GetQueryResult

# 执行更新
ExecDB "INSERT INTO users (name) VALUES (?)" "Alice"

# 关闭连接
CloseDB
```

## 高级特性

### Spring 化功能

Shode 提供了类似 Spring 框架的功能：

- **IoC 容器**: Bean 管理和依赖注入
- **配置管理**: 多源配置、类型安全访问
- **Web 层**: 中间件、拦截器、控制器
- **数据访问**: 事务管理、Repository 模式
- **AOP**: 面向切面编程支持

详细内容请参考 [Spring 功能示例](../examples/advanced/spring-features.md)。

## 最佳实践

1. **使用标准库函数**: 优先使用内置函数而非外部命令
2. **错误处理**: 检查函数返回值，处理异常情况
3. **模块化**: 将复杂逻辑拆分为可复用的模块
4. **配置管理**: 使用配置文件而非硬编码
5. **缓存策略**: 合理使用缓存提升性能

## 相关文档

- [执行引擎指南](./execution-engine.md)
- [包管理指南](./package-registry.md)
- [API 参考](../api/stdlib.md)
- [示例集合](../examples/index.md)
