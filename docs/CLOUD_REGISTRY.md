# Shode Cloud Registry

Shode Cloud Registry 提供一个可部署在任意云环境的包管理后端，采用 **Go + PostgreSQL + S3** 架构，实现水平扩展与对象存储解耦。

## 架构

```
┌─────────────────────────────┐
│         HTTP API            │
│  cmd/registry-cloud         │
└──────────────┬──────────────┘
               │
     ┌─────────┴─────────┐
     │                   │
┌────▼─────┐       ┌─────▼────┐
│PostgreSQL│       │  S3/OSS  │
│Metadata  │       │Tarballs  │
└──────────┘       └──────────┘
```

- **PostgreSQL** 存储包、版本、依赖、签名等结构化信息  
- **S3 兼容存储** 保存 tarball，返回预签名下载 URL  
- **REST API** 兼容原有 Registry 接口（search、get package、publish）

## 配置

`cmd/registry-cloud` 使用环境变量配置：

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| `DATABASE_URL` | PostgreSQL 连接串 (pgx) | 必填 |
| `S3_ENDPOINT` | S3/MinIO 访问域名 | `s3.amazonaws.com` |
| `S3_BUCKET` | 存储包的 bucket 名 | `shode-packages` |
| `S3_ACCESS_KEY` / `S3_SECRET_KEY` | S3 凭证 | 必填 |
| `S3_USE_SSL` | 是否使用 HTTPS | `true` |
| `S3_REGION` | bucket 区域 | `us-east-1` |
| `REGISTRY_TOKEN` | 发布鉴权 Token（可选） | 空=关闭鉴权 |
| `LISTEN_ADDR` | HTTP 监听地址 | `:8080` |

示例（Docker/Kubernetes）：

```bash
export DATABASE_URL=postgres://user:pass@postgres:5432/shode?sslmode=disable
export S3_ENDPOINT=minio:9000
export S3_BUCKET=shode
export S3_ACCESS_KEY=minio
export S3_SECRET_KEY=minio123
export S3_USE_SSL=false
export REGISTRY_TOKEN=super-secret

./shode-registry-cloud
```

## API

- `POST /api/search`：请求体为 `registry.SearchQuery`，返回 `[]SearchResult`
- `GET /api/packages/{name}`：返回 `registry.PackageMetadata`
- `POST /api/packages`：请求体为 `registry.PublishRequest`，需要 `Authorization: Bearer <token>`
- `GET /api/packages/{name}/versions/download?version=x.y.z`：返回预签名下载 URL
- `GET /health`：健康检查

## 部署建议

- 使用 Kubernetes/Helm 将服务与 PostgreSQL、MinIO 一起部署
- 反向代理（Ingress/Nginx）层终止 TLS
- 借助对象存储版本化能力实现包回滚
- 可通过 PostgreSQL 读副本和 S3 多 region 提升可用性
