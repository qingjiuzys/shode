# Shode 测试工具

Shode 提供了一套强大的测试工具，帮助您编写更简洁、更可靠的测试代码。

## 📦 工具列表

### 1. 断言库 (Assert)

提供丰富的断言方法，使测试代码更清晰易读。

**使用示例：**

```go
import "gitee.com/com_818cloud/shode/pkg/testing/assert"

func TestSomething(t *testing.T) {
    // 相等性断言
    assert.Equal(t, 1, 1)
    assert.NotEqual(t, 1, 2)

    // 布尔断言
    assert.True(t, true)
    assert.False(t, false)

    // 空值断言
    assert.Nil(t, nil)
    assert.NotNil(t, &value)

    // 字符串断言
    assert.Contains(t, "hello world", "hello")

    // 长度断言
    assert.Len(t, slice, 5)

    // 数值比较
    assert.Greater(t, 5, 3)
    assert.Less(t, 3, 5)

    // 错误断言
    assert.NoError(t, err)
    assert.Error(t, expectedErr)

    // Panic 断言
    assert.Panics(t, func() {
        panic("expected")
    })

    // JSON 断言
    assert.JSONEq(t, `{"name":"test"}`, `{"name":"test"}`)
}
```

**可用的断言方法：**
- `Equal/NotEqual` - 相等性断言
- `True/False` - 布尔断言
- `Nil/NotNil` - 空值断言
- `Contains` - 包含断言
- `Len` - 长度断言
- `Greater/Less` - 数值比较
- `Error/NoError` - 错误断言
- `Panics/NotPanics` - Panic 断言
- `JSONEq` - JSON 相等断言
- `Implements` - 接口实现断言

### 2. HTTP 测试工具 (HTTP Test)

简化 HTTP API 的测试。

**使用示例：**

```go
import (
    "net/http"
    httptest "gitee.com/com_818cloud/shode/pkg/testing/http"
)

func TestAPI(t *testing.T) {
    // 创建测试handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"message":"ok"}`))
    })

    h := httptest.NewHelper(t, handler)

    // GET 请求
    h.GET("/api/users").
        AssertOK().
        AssertJSON().
        AssertContains("ok")

    // POST 请求
    h.POST("/api/users", map[string]string{"name": "test"}).
        AssertCreated().
        AssertJSONEq(`{"id":1}`)

    // 测试不同状态码
    h.GET("/notfound").AssertNotFound()
    h.POST("/bad", nil).AssertBadRequest()
    h.GET("/unauthorized").AssertUnauthorized()

    // 断言 Headers
    h.GET("/api/data").
        AssertContentType("application/json").
        AssertHeader("X-Custom", "value")
}
```

**链式调用方法：**
- `AssertStatus(code)` - 断言状态码
- `AssertOK()` - 断言 200
- `AssertCreated()` - 断言 201
- `AssertNoContent()` - 断言 204
- `AssertBadRequest()` - 断言 400
- `AssertUnauthorized()` - 断言 401
- `AssertForbidden()` - 断言 403
- `AssertNotFound()` - 断言 404
- `AssertInternalServerError()` - 断言 500
- `AssertContentType(type)` - 断言 Content-Type
- `AssertJSON()` - 断言 JSON 响应
- `AssertBody(body)` - 断言响应体
- `AssertContains(substr)` - 断言包含
- `AssertJSONEq(json)` - 断言 JSON 相等
- `AssertHeader(key, val)` - 断言 Header
- `AssertCookie(name)` - 断言 Cookie

### 3. Mock 工具 (Mock)

创建 Mock 对象，模拟依赖项。

**使用示例：**

```go
import "gitee.com/com_818cloud/shode/pkg/testing/mock"

func TestService(t *testing.T) {
    m := mock.New()

    // 设置期望
    m.On("GetData", 1).Return("data1")
    m.On("GetData", 2).Return("data2")

    // 调用方法
    result := someService.GetData(1)

    // 验证结果
    assert.Equal(t, "data1", result)

    // 验证调用
    assert.True(t, m.Called("GetData"))
    assert.True(t, m.CalledWith("GetData", 1))
    assert.Equal(t, 2, m.CalledTimes("GetData"))

    // 断言所有期望
    err := m.AssertExpectations()
    assert.NoError(t, err)
}
```

**使用任意参数：**

```go
m.On("Process", mock.Any(), mock.Any()).Return("ok")
m.On("Handle", mock.AnyOfType("string"), mock.AnyOfType("int")).Return(true)

// 这些都会匹配
m.Recorded("Process", 1, 2)
m.Recorded("Process", "hello", "world")
m.Recorded("Handle", "test", 42)
```

**设置调用次数：**

```go
// 调用一次（默认）
m.On("Method").Once()

// 调用两次
m.On("Method").Twice()

// 调用 N 次
m.On("Method").Times(5)

// 至少调用 N 次
m.On("Method").AtLeast(2)

// 最多调用 N 次
m.On("Method").AtMost(5)

// 可选调用（0次或1次）
m.On("Method").Maybe()
```

### 4. 测试夹具 (Fixtures)

管理测试数据，支持数据库夹具。

**使用示例：**

```go
import "gitee.com/com_818cloud/shode/pkg/testing/fixtures"

func TestWithData(t *testing.T) {
    f := fixtures.New(t)

    // 加载 JSON 夹具
    f.MustLoad("users")

    // 获取夹具数据
    var users []User
    f.MustGetAs("users", &users)

    // 集合操作
    usersCollection := f.Collection("users")
    usersCollection.Add(user1)
    usersCollection.Add(user2)
    assert.Equal(t, 2, usersCollection.Count())

    // 清理
    f.Reset()
}
```

**数据库夹具：**

```go
func TestWithDatabase(t *testing.T) {
    db := setupTestDB()
    defer db.Close()

    f := fixtures.New(t)
    f.SetDB(db)

    // 创建表夹具
    usersTable := fixtures.NewTable(t, db, "users")
    usersTable.Create(`id INT PRIMARY KEY, name VARCHAR(100)`)

    // 插入测试数据
    usersTable.Insert(map[string]interface{}{
        "id":   1,
        "name": "Alice",
    })

    // 验证数据
    assert.Equal(t, 1, usersTable.Count())
    assert.True(t, usersTable.Exists("id = ?", 1))

    // 清理
    usersTable.Drop()
}
```

## 🚀 快速开始

### 安装

```bash
go get gitee.com/com_818cloud/shode/pkg/testing/...
```

### 基础测试示例

```go
package myapp

import (
    "testing"
    "gitee.com/com_818cloud/shode/pkg/testing/assert"
)

func TestAdd(t *testing.T) {
    result := Add(1, 2)
    assert.Equal(t, 3, result)
}

func TestDivision(t *testing.T) {
    result, err := Divide(10, 2)
    assert.NoError(t, err)
    assert.Equal(t, 5.0, result)

    // 测试错误情况
    _, err = Divide(10, 0)
    assert.Error(t, err)
}
```

### HTTP API 测试示例

```go
package api

import (
    "net/http"
    "testing"
    httptest "gitee.com/com_818cloud/shode/pkg/testing/http"
    "gitee.com/com_818cloud/shode/pkg/testing/assert"
)

func TestUserAPI(t *testing.T) {
    router := setupRouter()
    h := httptest.NewHelper(t, router)

    // 测试创建用户
    h.POST("/api/users", map[string]string{
        "name":  "Alice",
        "email": "alice@example.com",
    }).AssertCreated().
      AssertJSON().
      AssertContains("\"id\":1")

    // 测试获取用户列表
    h.GET("/api/users").
        AssertOK().
        AssertJSON()

    // 测试获取单个用户
    h.GET("/api/users/1").
        AssertOK().
        AssertJSONEq(`{"id":1,"name":"Alice","email":"alice@example.com"}`)

    // 测试404
    h.GET("/api/users/999").AssertNotFound()
}
```

### Mock 测试示例

```go
package service

import (
    "testing"
    "gitee.com/com_818cloud/shode/pkg/testing/mock"
    "gitee.com/com_818cloud/shode/pkg/testing/assert"
)

func TestUserService(t *testing.T) {
    m := mock.New()
    repo := &MockRepository{mock: m}
    service := NewUserService(repo)

    // 设置 Mock 期望
    m.On("GetByID", 1).Return(&User{ID: 1, Name: "Alice"}, nil)
    m.On("GetByID", 2).Return(nil, errors.New("not found"))

    // 测试成功情况
    user, err := service.GetUser(1)
    assert.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)

    // 测试错误情况
    user, err = service.GetUser(2)
    assert.Error(t, err)
    assert.Nil(t, user)

    // 验证所有期望都被满足
    err = m.AssertExpectations()
    assert.NoError(t, err)
}
```

## 📚 最佳实践

### 1. 使用表驱动测试

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 2. 使用测试辅助函数

```go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    assert.NoError(t, err)

    // 运行迁移
    _, err = db.Exec(migrationSQL)
    assert.NoError(t, err)

    return db
}

func TestWithDB(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    // 测试代码...
}
```

### 3. 使用子测试

```go
func TestAPI(t *testing.T) {
    router := setupRouter()
    h := httptest.NewHelper(t, router)

    t.Run("Create", func(t *testing.T) {
        h.POST("/api/users", userData).AssertCreated()
    })

    t.Run("Get", func(t *testing.T) {
        h.GET("/api/users/1").AssertOK()
    })

    t.Run("Update", func(t *testing.T) {
        h.PUT("/api/users/1", updateData).AssertOK()
    })

    t.Run("Delete", func(t *testing.T) {
        h.DELETE("/api/users/1").AssertNoContent()
    })
}
```

## 🎯 特性总结

- ✅ 丰富的断言方法
- ✅ 链式调用 API
- ✅ HTTP 测试简化
- ✅ Mock 对象支持
- ✅ 测试数据管理
- ✅ 数据库夹具
- ✅ 表驱动测试支持
- ✅ 清晰的错误信息

## 📖 更多示例

查看 [examples/testing](../examples/testing) 获取更多测试示例。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
