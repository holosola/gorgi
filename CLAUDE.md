# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 常用命令

所有命令在仓库根目录执行：

- `make run` — 本地启动（读取 `./configs/config.yaml`，需先 `cp configs/config.yaml.example configs/config.yaml`）。
- `make test` — 全量单元测试（`go test ./... -race -cover`）。
- `make vet` — 静态检查；提交前必须通过。
- `make ent` — 新增 / 修改 `internal/pkg/ent/schema/` 后执行，重新生成 ent 代码。
- `make build` / `make build-linux` — 编译。
- `make tidy` — 整理 `go.mod`。
- `make docker` — 多阶段构建镜像（`deployments/docker/Dockerfile`）。

跑单个测试：`go test ./internal/app/middleware -run TestSign_Valid -race -count=1`。

**本地无 MySQL/Redis 也能起服务**：把 `mysql.dsn` 与 `redis.addr` 留空，`internal/app/app.go::initStores` 会跳过对应初始化并打 WARN 日志。仅 dev 用，生产必须填完整。

## 架构要点

### 分层

```
cmd/gorgi/main.go          仅参数解析
  └─ internal/app/app.go   装配入口（config/log/tracing/db/router/server）
      └─ internal/app/router/   路由 + 中间件链装配
          ├─ internal/app/middleware/   gin 中间件
          └─ internal/app/api/<domain>/ HTTP handler
              └─ internal/app/service/<domain>/   业务编排
                  └─ internal/app/repository/<domain>/   ent 数据访问
internal/pkg/   项目内通用基础设施（config/log/errs/response/breaker/tracing/mysql/redis/ent）
```

### 中间件注册顺序（强约束）

`internal/app/router/router.go::middlewares` 按此顺序注册，**改动需慎重**：

```
Recovery → Trace → AccessLog → OTel → CORS → RateLimit → Sign
```

理由：Recovery 必须最外层；Trace 必须在 AccessLog 之前注入 `request_id`；Sign 在 RateLimit 之后，避免无效签名消耗令牌。

### 模块化路由注册（关键）

`router` 包不感知具体业务，避免循环依赖。流程：

1. 每个业务域在 `internal/app/api/<domain>/routes.go` 导出 `Routes(deps router.Deps) router.Module`。
2. `internal/app/modules.go::modules` 列表把所有域 `Routes(deps)` 拼起来。
3. `app.Run` 调用 `router.New(deps, modules(deps)...)`。

**新增 domain 只需要：建 `routes.go` + 在 `modules.go` 加一行**；新增同 domain 下接口完全不需要碰 router 包。

### 接口版本分发（核心创新点）

`internal/app/router/version_dispatcher.go::Dispatch(handler, "GetUser")`：

- 启动期反射扫描 handler 上 `GetUser`、`GetUserV1`、`GetUserV2` 等方法并缓存为 `func(*gin.Context)`，**运行期零反射**。
- 请求 header `x-api-version: v1` → 命中 `GetUserV1`；无 header / 未实现的版本 → 回退到基础方法 `GetUser`；签名必须是 `func(c *gin.Context)`，否则启动期 panic。
- 命名约定：基础方法 `Xxx`，版本方法 `XxxV1` / `XxxV2` …

### 日志（slog 自封装）

- 业务统一通过 `log.L(c.Request.Context()).Info(...)` 取 logger，**不要** `slog.Default` / `slog.SetDefault`。
- `internal/pkg/log/handler.go` 自动从 ctx 注入 `request_id`（由 Trace 中间件埋入）。
- `internal/pkg/log/mask.go` 按 `log.mask_fields` 配置脱敏 slog 属性 key（递归处理 `slog.Group`）。**注意**：脱敏作用于 attr key，不会扫描 body 字符串内部。

### 签名算法

`HMAC-SHA256(appSecret, appKey + ts + nonce + method + path + rawQuery + rawBody)`，hex 小写，由 `X-App-Key / X-Timestamp / X-Nonce / X-Sign` 四头携带。`rawQuery` 必须纳入签名（防 query 参数重放篡改），客户端 SDK 须保持一致。

### 出站熔断

业务调 MySQL/Redis/外部 HTTP 一律走 `breaker.Manager.Do(ctx, name, fn)`（基于 `sony/gobreaker/v2`）。熔断打开返回 `errs.BreakerOpen` → HTTP 503。规则在 `breaker.rules.<name>` 配置覆盖默认值。

### 错误与响应

- 业务错误一律 `*errs.Err`（`internal/pkg/errs/errs.go` 预定义码）。
- handler **禁止**直接 `c.JSON`，必须 `response.OK / response.Fail / response.AbortFail`，保证 `{code, msg, data, request_id}` 一致结构。

### 配置

`internal/pkg/config/config.go` 是强类型 struct（`mapstructure` tag）。路径优先级：`-c` 命令行 → `GORGI_CONFIG` 环境变量 → `./configs/config.yaml`。新增字段需同步 `configs/config.yaml.example`，并通过 `router.Deps` 注入业务，**禁止**到处 `viper.GetString`。

## 编码约定

- **所有注释中文**；导出标识符必须有以名字开头、句号结尾的中文注释（Effective Go 强约束）。
- 缩写词整体大小写一致：`URL` / `HTTP` / `ID`，例如 `userID` 而非 `userId`。
- 错误信息使用小写中文不带句号；包装用 `fmt.Errorf("xxx: %w", err)`。
- `panic` 仅用于不可恢复的初始化错误。
- 提交规范 Conventional Commits：`feat / fix / refactor / docs / test / chore / perf`。
- 完整规范见 `docs/conventions.md`。

## 新增数据表流程

1. 在 `internal/pkg/ent/schema/` 加 schema 文件；
2. `make ent` 生成代码；
3. 在 `internal/app/repository/<domain>/` 写 dao；
4. service 调 dao，handler 调 service。
