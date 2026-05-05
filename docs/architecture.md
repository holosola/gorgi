# 架构说明

## 分层

```
HTTP -> Middleware -> Router -> Handler(api) -> Service -> Repository -> DB / Cache / 外部 HTTP
                                                            └── 出站调用统一经过 breaker 包保护
```

| 层 | 目录 | 职责 |
|---|---|---|
| Handler | `internal/app/api/<domain>/` | 解析参数 / 校验 / 统一响应；不写业务规则 |
| Service | `internal/app/service/<domain>/` | 业务规则、事务编排、跨 repository 协作 |
| Repository | `internal/app/repository/<domain>/` | 直接操作 ent / Redis；不依赖 gin |
| 基础设施 | `internal/pkg/*` | 日志 / 配置 / DB 连接 / 熔断 / Tracing 等通用能力 |

## 请求生命周期

1. `Recovery`：兜底所有 panic
2. `Trace`：注入 / 透传 `X-Request-ID`，写入 ctx
3. `AccessLog`：记录请求 / 响应（敏感头脱敏）
4. `OTel`：可选，启用后给当前请求开 span
5. `CORS`：跨域
6. `RateLimit`：按 IP 入站限流
7. `Sign`：HMAC-SHA256 验签
8. `Router.Dispatch`：根据 `x-api-version` 头分发到 `Method` / `MethodVx`

## 接口版本分发

- 在 handler 上同时实现 `User`、`UserV1`、`UserV2`…
- 路由用 `router.Dispatch(handler, "User")` 一次性注册
- 启动时缓存所有版本方法，运行时零反射开销
- 找不到对应版本时**回退到基础方法**

## 出站熔断

业务在调用 MySQL / Redis / 外部 HTTP 时，统一通过：

```go
err := mgr.Do(ctx, "mysql.user.query", func(ctx context.Context) error {
    // 真正的调用
    return nil
})
```

熔断打开时返回 `errs.BreakerOpen`，由 response.Fail 统一转 503。
