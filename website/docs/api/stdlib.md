# 标准库 API 参考

Shode 标准库提供了丰富的内置函数，涵盖字符串处理、文件操作、系统信息、网络请求等多个领域。所有函数都经过优化，提供比传统 shell 命令更好的性能和安全性。

## 字符串处理函数

### `upper(string) → string`
将字符串转换为大写。

**参数**:
- `string`: 要转换的字符串

**返回值**: 大写字符串

**示例**:
```bash
upper("hello")        # "HELLO"
upper("Hello World")  # "HELLO WORLD"
```

### `lower(string) → string`
将字符串转换为小写。

**参数**:
- `string`: 要转换的字符串

**返回值**: 小写字符串

**示例**:
```bash
lower("HELLO")        # "hello"
lower("Hello World")  # "hello world"
```

### `trim(string) → string`
去除字符串两端的空白字符。

**参数**:
- `string`: 要处理的字符串

**返回值**: 修剪后的字符串

**示例**:
```bash
trim("   hello   ")   # "hello"
trim("  test  \n")    # "test"
```

### `contains(haystack, needle) → bool`
检查字符串是否包含子串。

**参数**:
- `haystack`: 要搜索的字符串
- `needle`: 要查找的子串

**返回值**: 布尔值，true 表示包含

**示例**:
```bash
contains("hello", "ell")   # true
contains("hello", "world") # false
```

### `replace(str, old, new) → string`
替换字符串中的子串。

**参数**:
- `str`: 原始字符串
- `old`: 要替换的子串
- `new`: 替换后的子串

**返回值**: 替换后的字符串

**示例**:
```bash
replace("hello", "l", "x")   # "hexxo"
replace("aabbcc", "bb", "dd") # "aaddcc"
```

### `split(str, delimiter) → []string`
将字符串按分隔符分割为数组。

**参数**:
- `str`: 要分割的字符串
- `delimiter`: 分隔符

**返回值**: 字符串数组

**示例**:
```bash
split("a,b,c", ",")    # ["a", "b", "c"]
split("one two three", " ") # ["one", "two", "three"]
```

### `join(array, separator) → string`
将字符串数组连接为一个字符串。

**参数**:
- `array`: 字符串数组
- `separator`: 连接分隔符

**返回值**: 连接后的字符串

**示例**:
```bash
join(["a", "b", "c"], ",")  # "a,b,c"
join(["2024", "01", "15"], "-") # "2024-01-15"
```

## 文件系统函数

### `readfile(filename) → string`
读取文件内容。

**参数**:
- `filename`: 文件名

**返回值**: 文件内容字符串

**异常**: 文件不存在时返回空字符串

**示例**:
```bash
content = readfile("config.txt")
println("文件内容:", content)
```

### `write(filename, content) → bool`
写入内容到文件。

**参数**:
- `filename`: 文件名
- `content`: 要写入的内容

**返回值**: 布尔值，true 表示成功

**示例**:
```bash
success = write("output.txt", "Hello, World!")
if success {
    println("写入成功")
}
```

### `exists(filename) → bool`
检查文件或目录是否存在。

**参数**:
- `filename`: 文件或目录路径

**返回值**: 布尔值，true 表示存在

**示例**:
```bash
if exists("config.json") {
    println("配置文件存在")
}
```

### `delete(filename) → bool`
删除文件。

**参数**:
- `filename`: 文件名

**返回值**: 布尔值，true 表示成功

**安全限制**: 沙箱模式下有权限限制

**示例**:
```bash
if delete("temp.txt") {
    println("文件已删除")
}
```

### `list(directory) → []string`
列出目录中的文件和子目录。

**参数**:
- `directory`: 目录路径

**返回值**: 文件名数组

**示例**:
```bash
files = list(".")
for file in files {
    println("文件:", file)
}
```

### `size(filename) → int`
获取文件大小（字节）。

**参数**:
- `filename`: 文件名

**返回值**: 文件大小（字节）

**示例**:
```bash
file_size = size("data.bin")
println("文件大小:", file_size, "字节")
```

## 系统信息函数

### `whoami() → string`
获取当前用户名。

**返回值**: 用户名字符串

**示例**:
```bash
user = whoami()
println("当前用户:", user)
```

### `hostname() → string`
获取主机名。

**返回值**: 主机名字符串

**示例**:
```bash
name = hostname()
println("主机名:", name)
```

### `pid() → int`
获取当前进程ID。

**返回值**: 进程ID整数

**示例**:
```bash
process_id = pid()
println("进程ID:", process_id)
```

### `now() → string`
获取当前时间戳。

**返回值**: ISO格式时间字符串

**示例**:
```bash
current_time = now()
println("当前时间:", current_time)
```

### `sleep(milliseconds) → void`
暂停执行指定毫秒数。

**参数**:
- `milliseconds`: 毫秒数

**示例**:
```bash
println("开始等待...")
sleep(1000)  # 等待1秒
println("等待结束")
```

## 加密哈希函数

### `md5(data) → string`
计算字符串的MD5哈希值。

**参数**:
- `data`: 输入数据

**返回值**: MD5哈希字符串

**示例**:
```bash
hash = md5("password123")
println("MD5哈希:", hash)
```

### `sha1(data) → string`
计算字符串的SHA1哈希值。

**参数**:
- `data`: 输入数据

**返回值**: SHA1哈希字符串

**示例**:
```bash
hash = sha1("sensitive data")
println("SHA1哈希:", hash)
```

### `sha256(data) → string`
计算字符串的SHA256哈希值。

**参数**:
- `data`: 输入数据

**返回值**: SHA256哈希字符串

**示例**:
```bash
hash = sha256("important file")
println("SHA256哈希:", hash)
```

## 数据处理函数

### `json(data) → string`
将数据转换为JSON字符串。

**参数**:
- `data`: 要转换的数据（对象或数组）

**返回值**: JSON字符串

**示例**:
```bash
user = {
    "name": "张三",
    "age": 30,
    "city": "北京"
}
json_str = json(user)
println("JSON:", json_str)
```

### `jsonparse(json_str) → any`
解析JSON字符串为数据对象。

**参数**:
- `json_str`: JSON字符串

**返回值**: 解析后的数据对象

**示例**:
```bash
data = jsonparse('{"name":"李四","age":25}')
println("姓名:", data.name)
println("年龄:", data.age)
```

## 网络函数

### `httpget(url) → string`
发送HTTP GET请求。

**参数**:
- `url`: 请求URL

**返回值**: 响应内容字符串

**安全限制**: 沙箱模式下需要网络权限

**示例**:
```bash
response = httpget("https://httpbin.org/json")
println("响应:", substr(response, 0, 100))
```

### `httppost(url, data) → string`
发送HTTP POST请求。

**参数**:
- `url`: 请求URL
- `data`: 提交的数据

**返回值**: 响应内容字符串

**示例**:
```bash
data = json({"name": "test", "value": 123})
response = httppost("https://httpbin.org/post", data)
println("POST响应:", response)
```

## 工具函数

### `println(...args) → void`
打印输出到控制台。

**参数**:
- `...args`: 要打印的参数

**示例**:
```bash
println("Hello", "World", 123)
```

### `type(value) → string`
获取值的类型。

**参数**:
- `value`: 任意值

**返回值**: 类型字符串（"string", "number", "bool", "array", "object"）

**示例**:
```bash
println(type("hello"))   # "string"
println(type(123))       # "number"
println(type([1,2,3]))   # "array"
```

### `len(value) → int`
获取字符串或数组的长度。

**参数**:
- `value`: 字符串或数组

**返回值**: 长度整数

**示例**:
```bash
println(len("hello"))     # 5
println(len([1,2,3,4]))   # 4
```

## 函数列表获取

### `list functions() → []string`
获取所有可用标准库函数列表。

**返回值**: 函数名数组

**示例**:
```bash
functions = list functions()
for func in functions {
    println("函数:", func)
}
```

### `function help(func_name) → string`
获取函数的帮助信息。

**参数**:
- `func_name`: 函数名

**返回值**: 函数帮助信息

**示例**:
```bash
help_text = function help("upper")
println(help_text)
```

## 使用建议

1. **性能优先**: 尽量使用标准库函数而不是外部命令
2. **错误处理**: 使用 try-catch 处理可能出错的函数
3. **安全性**: 注意沙箱限制，特别是文件操作和网络请求
4. **类型检查**: 使用 `type()` 函数检查参数类型

## 完整函数列表

要查看完整的函数列表和详细帮助，可以在 REPL 中运行：
```bash
shode repl
> list functions()
> function help("函数名")
```

或者使用命令行：
```bash
shode exec "list functions"
```

标准库持续更新中，欢迎贡献新的函数功能！
