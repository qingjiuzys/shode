# Shode 最佳实践

本文档提供了使用 Shode 进行开发的最佳实践和建议。

## 📁 项目结构

### 推荐的项目结构

```
my-shode-project/
├── shode.json              # 项目配置
├── shode-lock.json         # 依赖锁定（自动生成）
├── main.sh                 # 入口文件
├── src/                    # 源代码
├── tests/                  # 测试
└── docs/                   # 文档
```

## 📦 包管理最佳实践

### 依赖版本规范

```bash
# 使用语义版本范围
shode pkg add @shode/logger ^1.0.0
shode pkg add lodash ^4.17.0
```

## 🔒 安全最佳实践

### 敏感信息管理

```bash
# 使用环境变量
API_KEY=$(ConfigGet "API_KEY" "")
```

### 命令注入防护

```bash
# 使用数组
command_array=($user_input)
"${command_array[@]}"
```

## 🧪 测试最佳实践

```bash
test_example_success_case() {
    local input="test"
    local expected="output"
    local result=$(MyFunction "$input")
    
    if [ "$result" != "$expected" ]; then
        echo "FAIL: Expected $expected, got $result"
        return 1
    fi
    
    echo "PASS: test_example_success_case"
}
```

---

遵循这些最佳实践，可以让你的 Shode 项目更加健壮、安全、易于维护！
