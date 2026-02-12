# 可观测性系统 (Observability System)

Shode 框架提供完整的可观测性解决方案。

## 📊 功能特性

### 1. 指标收集 (Metrics/)
- ✅ Prometheus 指标导出
- ✅ 自定义指标注册
- ✅ HTTP 指标端点
- ✅ 性能指标收集
- ✅ 业务指标统计

### 2. 健康检查 (Health/)
- ✅ 服务健康检查
- ✅ 依赖健康检查
- ✅ 就绪探针
- ✅ 存活探针
- ✅ 健康状态报告

### 3. 分布式追踪 (Tracing/)
- ✅ OpenTelemetry 集成
- ✅ 请求链路追踪
- ✅ 性能分析
- ✅ 依赖关系图
- ✅ 延迟分析

### 4. 结构化日志 (Logging/)
- ✅ 结构化日志输出
- ✅ 日志级别管理
- ✅ 上下文日志
- ✅ 日志聚合
- ✅ 日志查询

## 🚀 快速开始

### 启用指标收集

```go
import (
    "gitee.com/com_818cloud/shode/pkg/observability/metrics"
    "github.com/prometheus/client_golang/prometheus"
)

func main() {
    // 创建指标注册器
    registry := metrics.NewRegistry()

    // 注册 HTTP 指标
    httpMetrics := metrics.NewHTTPMetrics("api")
    registry.MustRegister(httpMetrics)

    // 启动指标端点
    metrics.ServeMetrics(":9090", registry)
}
```

### 配置健康检查

```go
import "gitee.com/com_818cloud/shode/pkg/observability/health"

func main() {
    checker := health.NewChecker()

    // 添加健康检查
    checker.AddCheck("database", health.CheckFunc(func() error {
        return db.Ping()
    }))

    checker.AddCheck("redis", health.CheckFunc(func() error {
        return redis.Ping().Err()
    }))

    // 启动健康检查端点
    http.Handle("/health", checker.Handler())
}
```

### 启用分布式追踪

```go
import "gitee.com/com_818cloud/shode/pkg/observability/tracing"

func main() {
    // 初始化追踪器
    tracer, err := tracing.InitTracer(tracing.Config{
        ServiceName: "my-service",
        Endpoint:    "http://jaeger:14268/api/traces",
        Sampler:     1.0,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer tracer.Close()

    // 创建带追踪的 HTTP 处理器
    http.Handle("/", tracing.WrapHandler(tracer, myHandler))
}
```

### 结构化日志

```go
import "gitee.com/com_818cloud/shode/pkg/observability/logging"

func main() {
    logger := logging.NewLogger(logging.Config{
        Level:      "info",
        Format:     "json",
        Output:     []string{"stdout"},
    })

    // 使用日志
    logger.Info("Starting service",
        "port", 8080,
        "env", "production",
    )
}
```

## 📈 Prometheus 指标

### 内置指标

#### HTTP 指标
- `http_requests_total` - HTTP 请求总数
- `http_request_duration_seconds` - 请求处理时间
- `http_requests_in_flight` - 当前处理中的请求数
- `http_response_size_bytes` - 响应大小

#### 系统指标
- `process_cpu_seconds_total` - CPU 使用时间
- `process_resident_memory_bytes` - 内存使用量
- `process_open_fds` - 打开的文件描述符数量
- `go_goroutines` - Goroutine 数量

### 自定义指标

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    requestCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_requests_total",
            Help: "Total number of API requests",
        },
        []string{"method", "endpoint", "status"},
    )

    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "api_request_duration_seconds",
            Help:    "API request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(requestCounter)
    prometheus.MustRegister(requestDuration)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    // 处理请求
    // ...

    duration := time.Since(start).Seconds()
    requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    requestCounter.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
}
```

## 🔍 健康检查

### 检查类型

#### Liveness Probe (存活探针)
检查服务是否正在运行。

```go
http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "OK")
})
```

#### Readiness Probe (就绪探针)
检查服务是否准备好接收流量。

```go
http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
    if isReady() {
        fmt.Fprintf(w, "OK")
    } else {
        http.Error(w, "Service not ready", http.StatusServiceUnavailable)
    }
})
```

#### Startup Probe (启动探针)
检查服务是否已启动。

```go
http.HandleFunc("/started", func(w http.ResponseWriter, r *http.Request) {
    if isStarted() {
        fmt.Fprintf(w, "OK")
    } else {
        http.Error(w, "Service not started", http.StatusServiceUnavailable)
    }
})
```

### 健康检查示例

```go
type healthChecker struct {
    db    *sql.DB
    redis *redis.Client
}

func (h *healthChecker) Check() error {
    // 检查数据库
    if err := h.db.Ping(); err != nil {
        return fmt.Errorf("database unhealthy: %w", err)
    }

    // 检查 Redis
    if err := h.redis.Ping().Err(); err != nil {
        return fmt.Errorf("redis unhealthy: %w", err)
    }

    return nil
}
```

## 🎯 分布式追踪

### OpenTelemetry 集成

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 创建 span
    ctx, span := otel.Tracer("api").Start(ctx, "handleRequest")
    defer span.End()

    // 添加属性
    span.SetAttributes(
        attribute.String("http.method", r.Method),
        attribute.String("http.path", r.URL.Path),
    )

    // 处理请求
    processRequest(ctx)

    // 记录事件
    span.AddEvent("request_processed")
}
```

### 传播上下文

```go
import (
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/baggage"
)

func makeRequest(ctx context.Context, url string) error {
    // 创建请求
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return err
    }

    // 注入追踪上下文
    propagator := propagation.TraceContext{}
    propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

    // 发送请求
    return http.DefaultClient.Do(req)
}
```

## 📝 结构化日志

### 日志级别

- **Debug**: 详细的调试信息
- **Info**: 一般信息
- **Warn**: 警告信息
- **Error**: 错误信息
- **Fatal**: 致命错误

### 日志格式

```json
{
  "level": "info",
  "time": "2024-01-01T12:00:00Z",
  "message": "Request received",
  "context": {
    "request_id": "abc123",
    "user_id": "user1",
    "path": "/api/users"
  },
  "fields": {
    "method": "GET",
    "status": 200,
    "duration": "10ms"
  }
}
```

### 上下文日志

```go
import (
    "context"
    "go.opentelemetry.io/otel/trace"
)

func handleRequest(ctx context.Context) {
    span := trace.SpanFromContext(ctx)
    traceID := span.SpanContext().TraceID()

    logger.Info("Processing request",
        "trace_id", traceID.String(),
        "span_id", span.SpanContext().SpanID().String(),
    )
}
```

## 🔧 配置选项

### Metrics 配置

```go
type MetricsConfig struct {
    Enabled    bool
    Endpoint   string
    Namespace  string
    Subsystem  string
    Buckets    []float64
    Labels     []string
}
```

### Health 配置

```go
type HealthConfig struct {
    Enabled       bool
    LivenessPath  string
    ReadinessPath string
    Interval      time.Duration
    Timeout       time.Duration
}
```

### Tracing 配置

```go
type TracingConfig struct {
    Enabled     bool
    ServiceName string
    Endpoint    string
    Sampler     float64
    Batcher     string
}
```

### Logging 配置

```go
type LoggingConfig struct {
    Level      string
    Format     string
    Output     []string
    TimeFormat string
    Color      bool
}
```

## 📊 监控集成

### Prometheus

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'shode'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: /metrics
    scrape_interval: 15s
```

### Grafana

导入预配置的仪表板：
- Go 应用监控
- HTTP 服务监控
- 数据库连接池监控
- 缓存性能监控

### Jaeger

```yaml
# jaeger.yml
collector:
  zipkin:
    host-port: :9411
```

## 🎯 最佳实践

1. **尽早添加可观测性**: 在开发早期就集成监控和日志
2. **使用结构化日志**: 便于查询和分析
3. **添加上下文**: 在日志和追踪中包含请求 ID、用户 ID 等
4. **设置合理的采样率**: 避免过多的追踪数据
5. **监控关键指标**: 关注延迟、错误率、吞吐量
6. **使用 SLI/SLO**: 定义服务水平指标和目标
7. **告警要精确**: 避免告警疲劳

## 📚 相关文档

- [Prometheus 文档](https://prometheus.io/docs/)
- [OpenTelemetry 文档](https://opentelemetry.io/docs/)
- [Grafana 文档](https://grafana.com/docs/)
- [Jaeger 文档](https://www.jaegertracing.io/docs/)

## 🤝 贡献

欢迎贡献新的可观测性功能！

## 📄 许可证

MIT License
