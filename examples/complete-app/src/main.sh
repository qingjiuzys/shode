#!/bin/sh
# Shode 完整应用示例
# 展示所有官方包的使用

# =====================
# 加载依赖
# =====================

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 加载官方包
. "$PROJECT_ROOT/sh_modules/@shode/logger/index.sh"
. "$PROJECT_ROOT/sh_modules/@shode/config/index.sh"
. "$PROJECT_ROOT/sh_modules/@shode/http/index.sh"

# =====================
# 应用程序
# =====================

main() {
    # 初始化日志
    SetLogLevel "info"
    LogInfo "=========================================="
    LogInfo "Shode 完整应用示例启动"
    LogInfo "=========================================="

    # 加载配置
    LogInfo "加载应用配置..."
    ConfigLoad "$PROJECT_ROOT/config/app.json"

    # 获取配置
    app_name=$(ConfigGet "app.name" "Shode App")
    server_port=$(ConfigGet "server.port" "8080")
    log_level=$(ConfigGet "logging.level" "info")

    LogInfo "应用名称: $app_name"
    LogInfo "服务器端口: $server_port"
    LogInfo "日志级别: $log_level"

    # 演示日志功能
    LogDebug "这是调试信息（默认不显示）"
    LogInfo "这是普通信息"
    LogWarn "这是警告信息"

    # 演示 HTTP 客户端
    LogInfo "测试 HTTP 功能..."
    test_http_client

    # 演示配置管理
    LogInfo "测试配置管理功能..."
    test_config_management

    LogInfo "=========================================="
    LogInfo "应用运行完成！"
    LogInfo "=========================================="
}

# 测试 HTTP 客户端
test_http_client() {
    LogInfo "→ 测试 HTTP GET 请求"

    # 注意：这只是一个演示，实际需要有 HTTP 服务器运行
    # response=$(HttpGet "http://httpbin.org/get")
    # LogInfo "HTTP 响应: $response"

    LogInfo "  HTTP 客户端功能已就绪"
}

# 测试配置管理
test_config_management() {
    LogInfo "→ 测试配置读取"

    # 读取嵌套配置
    app_env=$(ConfigGet "app.env" "production")
    server_host=$(ConfigGet "server.host" "localhost")

    LogInfo "  应用环境: $app_env"
    LogInfo "  服务器主机: $server_host"

    # 检查配置键是否存在
    if ConfigHas "server.port"; then
        LogInfo "  配置键 server.port 存在"
    fi

    # 设置新的配置值
    ConfigSet "test.key" "test.value"
    test_value=$(ConfigGet "test.key" "")
    LogInfo "  设置的测试值: $test_value"
}

# =====================
# 错误处理
# =====================

error_handler() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        LogError "应用异常退出，退出码: $exit_code"
    fi
}

trap error_handler EXIT

# =====================
# 主入口
# =====================

main "$@"
