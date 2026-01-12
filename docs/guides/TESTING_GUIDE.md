# 测试指南

## 📋 当前测试状态

### ✅ 已完成的测试

1. **执行引擎测试** (`pkg/engine/engine_test.go`)
   - ✅ 基本命令执行测试
   - ✅ 标准库函数测试
   - ✅ 函数定义和执行测试
   - ✅ For 循环和 break 测试
   - ✅ While 循环和 continue 测试
   - ✅ Pipeline 执行测试
   - ✅ If 语句测试
   - ✅ 变量赋值测试
   - ✅ 超时处理测试
   - ✅ 用户定义函数检测测试

2. **包管理器测试** (`pkg/pkgmgr/manager_test.go`)
   - ✅ 包配置初始化测试
   - ✅ 依赖添加测试
   - ✅ Tarball 创建测试
   - ✅ 脚本管理测试

3. **模块系统测试** (`pkg/module/manager_test.go`)
   - ✅ package.json 加载测试
   - ✅ 默认 main 入口点测试
   - ✅ 模块加载测试

### ⏳ 待完成的测试

1. **Tarball 解压测试** (`pkg/registry/client_test.go`)
   - [ ] 解压功能测试
   - [ ] 路径安全测试
   - [ ] 文件权限测试
   - [ ] 符号链接测试

2. **安全检查器测试** (`pkg/sandbox/security_test.go`)
   - [ ] 危险命令黑名单测试
   - [ ] 敏感文件保护测试
   - [ ] 模式匹配测试
   - [ ] 边界情况测试

3. **环境管理器测试** (`pkg/environment/manager_test.go`)
   - [ ] 环境变量管理测试
   - [ ] 工作目录管理测试
   - [ ] 会话隔离测试

4. **解析器测试** (`pkg/parser/parser_test.go`)
   - [ ] 字符串解析测试
   - [ ] 文件解析测试
   - [ ] 复杂脚本解析测试

5. **集成测试**
   - [ ] 端到端脚本执行测试
   - [ ] 包管理完整流程测试
   - [ ] 模块系统完整流程测试

---

## 🚀 运行测试

### 运行所有测试
```bash
go test ./...
```

### 运行特定包的测试
```bash
# 执行引擎测试
go test ./pkg/engine -v

# 包管理器测试
go test ./pkg/pkgmgr -v

# 模块系统测试
go test ./pkg/module -v
```

### 运行特定测试
```bash
go test ./pkg/engine -v -run TestExecuteStdLibFunction
```

### 查看测试覆盖率
```bash
go test ./pkg/engine -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## 📝 编写新测试的最佳实践

### 1. 测试文件命名
- 测试文件必须以 `_test.go` 结尾
- 测试文件应该与被测试文件在同一包中

### 2. 测试函数命名
- 测试函数必须以 `Test` 开头
- 测试函数名应该描述测试的内容
- 示例: `TestExecuteCommand`, `TestFunctionDefinition`

### 3. 测试结构
```go
func TestFeatureName(t *testing.T) {
    // 1. 设置测试环境
    setup := setupTestEnvironment()
    defer cleanup(setup)
    
    // 2. 执行测试
    result, err := functionUnderTest()
    
    // 3. 验证结果
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
    
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### 4. 使用表驱动测试
```go
func TestMultipleCases(t *testing.T) {
    testCases := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "input1", "output1"},
        {"case2", "input2", "output2"},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := functionUnderTest(tc.input)
            if result != tc.expected {
                t.Errorf("Expected %s, got %s", tc.expected, result)
            }
        })
    }
}
```

### 5. 测试辅助函数
```go
// setupTestEngine 创建测试用的执行引擎
func setupTestEngine() *ExecutionEngine {
    // ... 设置代码
}

// 在多个测试中复用
func TestFeature1(t *testing.T) {
    ee := setupTestEngine()
    // ... 测试代码
}
```

---

## 🎯 测试覆盖率目标

### 当前目标
- **单元测试覆盖率**: > 80%
- **关键功能覆盖率**: > 90%

### 检查覆盖率
```bash
# 生成覆盖率报告
go test ./... -coverprofile=coverage.out

# 查看详细覆盖率
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html
```

---

## 🔍 测试类型说明

### 单元测试
- 测试单个函数或方法
- 使用 mock 对象隔离依赖
- 快速执行

### 集成测试
- 测试多个组件协作
- 使用真实依赖
- 验证完整流程

### 端到端测试
- 测试完整用户场景
- 从输入到输出的完整流程
- 验证系统行为

---

## 📚 测试示例

### 示例 1: 基本功能测试
```go
func TestBasicFunction(t *testing.T) {
    result := functionUnderTest("input")
    if result != "expected" {
        t.Errorf("Expected 'expected', got '%s'", result)
    }
}
```

### 示例 2: 错误处理测试
```go
func TestErrorHandling(t *testing.T) {
    _, err := functionThatCanFail("invalid")
    if err == nil {
        t.Error("Expected error, got nil")
    }
}
```

### 示例 3: 并发测试
```go
func TestConcurrency(t *testing.T) {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // 测试代码
        }()
    }
    wg.Wait()
}
```

---

## 🐛 调试测试

### 使用 t.Log 输出调试信息
```go
func TestDebug(t *testing.T) {
    t.Log("Debug information")
    // 测试代码
}
```

### 使用 -v 标志查看详细输出
```bash
go test -v ./pkg/engine
```

### 使用调试器
```bash
# 使用 delve 调试器
dlv test ./pkg/engine
```

---

## ✅ 测试检查清单

在提交代码前，确保：
- [ ] 所有新功能都有测试
- [ ] 测试通过 (`go test ./...`)
- [ ] 测试覆盖率满足要求
- [ ] 没有跳过测试 (`t.Skip()`)
- [ ] 测试名称清晰描述测试内容
- [ ] 测试代码简洁易读
- [ ] 测试独立，不依赖执行顺序

---

## 📖 相关资源

- [Go Testing 官方文档](https://golang.org/pkg/testing/)
- [Go Test 命令文档](https://golang.org/cmd/go/#hdr-Test_packages)
- [测试最佳实践](https://github.com/golang/go/wiki/TestComments)

---

**最后更新**: 2025-01-XX  
**维护者**: Shode Team
