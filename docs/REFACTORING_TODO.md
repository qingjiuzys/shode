# 重构任务清单

## 优先级 P0 - 立即重构（圈复杂度 > 15）

### pkg/stdlib/stdlib.go (2027 行)

#### 高优先级函数（需要拆分）

1. **serveStaticFile** (~250 行)
   - 问题：过长，包含太多逻辑
   - 建议：拆分为多个小函数
   - 子函数：
     - `serveStaticFileWithCompression()`
     - `serveStaticFileRange()`
     - `serveStaticFileETag()`

2. **RegisterHTTPRoute** (~150 行)
   - 问题：逻辑复杂，不易测试
   - 建议：提取路由注册逻辑到单独函数

3. **createRequestContext** (~80 行)
   - 问题：查询参数解析逻辑复杂
   - 建议：拆分为 `parseQueryParams()` 和 `validateParams()`

### pkg/engine/engine.go (2303 行)

#### 高优先级函数

1. **Execute** (~400 行)
   - 问题：核心执行函数过长
   - 建议：按节点类型拆分处理函数
     - `executeCommandNode()`
     - `executePipelineNode()`
     - `executeControlFlowNode()`

2. **executeUserFunction** (~150 行)
   - 问题：函数执行逻辑复杂
   - 建议：提取环境设置和清理逻辑

## 优先级 P1 - 计划重构

### pkg/parser/parser.go (972 行)

1. **walkTree** (~300 行)
   - 问题：递归深度过大
   - 建议：使用迭代替代部分递归

2. **parseList** (~100 行)
   - 问题：逻辑分支过多
   - 建议：使用策略模式处理不同节点类型

### pkg/pkgmgr/manager.go (501 行)

1. **CreatePackage** (~200 行)
   - 问题：打包逻辑耦合
   - 建议：分离文件收集和打包逻辑

## 重构原则

### 1. 单一职责原则
每个函数只做一件事：
```go
// ❌ 不好 - 做多件事
func ProcessFile(path string) error {
    // 读取文件
    // 解析内容
    // 验证数据
    // 保存到数据库
    // 发送通知
}

// ✅ 好 - 单一职责
func ReadFile(path string) ([]byte, error) { ... }
func ParseContent(data []byte) (Content, error) { ... }
func ValidateData(content Content) error { ... }
func SaveToDB(content Content) error { ... }
func SendNotification(content Content) error { ... }
```

### 2. 函数长度限制
- 目标：每个函数 < 50 行
- 最大：< 100 行（必须有充分理由）

### 3. 参数数量限制
- 目标：< 5 个参数
- 超过 5 个：使用结构体
```go
// ❌ 不好 - 参数太多
func Process(name, type, status, priority, owner, created string) error { ... }

// ✅ 好 - 使用结构体
type Config struct {
    Name     string
    Type     string
    Status   string
    Priority string
    Owner    string
    Created  string
}
func Process(cfg Config) error { ... }
```

### 4. 嵌套层级限制
- 最大嵌套深度：3 层
- 超过 3 层：提前返回或提取函数
```go
// ❌ 不好 - 嵌套过深
func process(data []string) {
    for _, d := range data {
        if d != "" {
            if len(d) > 10 {
                for i, c := range d {
                    if c == 'x' {
                        // 4 层嵌套
                    }
                }
            }
        }
    }
}

// ✅ 好 - 提前返回
func process(data []string) {
    for _, d := range data {
        if d == "" {
            continue
        }
        if len(d) <= 10 {
            continue
        }
        processX(d)
    }
}
```

## 重构步骤

### 准备阶段
1. ✅ 编写测试覆盖现有功能
2. ✅ 确认所有测试通过
3. ✅ 创建功能分支

### 执行阶段
1. ✅ 识别重构目标（函数/类）
2. ✅ 提取小函数，保持功能不变
3. ✅ 运行测试确保无破坏
4. ✅ 提交代码

### 验证阶段
1. ✅ 运行完整测试套件
2. ✅ 性能基准测试
3. ✅ 代码审查

## 当前进度

- [ ] pkg/stdlib/stdlib.go - serveStaticFile 重构
- [ ] pkg/stdlib/stdlib.go - RegisterHTTPRoute 重构
- [ ] pkg/engine/engine.go - Execute 重构
- [ ] pkg/parser/parser.go - walkTree 重构

## 注意事项

1. **小步快跑**：每次只重构一个小函数
2. **保持测试**：确保测试始终通过
3. **备份代码**：重构前创建 git 分支
4. **文档同步**：更新相关文档和注释

## 工具推荐

```bash
# 安装代码复杂度分析工具
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

# 分析复杂度
gocyclo -over 15 pkg/

# 安装重构工具
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
```

## 参考资源

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://go.dev/doc/effective_go)
- [Refactoring Guru](https://refactoring.guru/)
