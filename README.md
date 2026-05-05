# gorgi

`gorgi` 是一个 Go 后端 API 项目脚手架。开箱即用的能力包括：路由 / 版本分发 / 中间件全家桶 / 日志 + 链路 + 脱敏 / ent ORM / 出站熔断 / OpenTelemetry，目录按 [golang-standards/project-layout](https://github.com/golang-standards/project-layout) 组织。

---

## 1. 技术栈

| 类别 | 选型 | 版本 |
|---|---|---|
| 语言 | Go | 1.26.x |
| Web 框架 | gin | v1.12+ |
| ORM | [ent](https://entgo.io/) | latest |
| 配置 | viper | v1.21+ |
| 日志 | 标准库 `log/slog`（自封装） | — |
| Tracing | OpenTelemetry SDK | latest |
| 数据库 | MySQL | 8.0.18+ |
| 缓存 | Redis | 7.2+ |
| 限流 | `golang.org/x/time/rate` | — |
| 熔断 | `sony/gobreaker/v2` | — |

---

## 2. 目录结构

```
gorgi/
├── api/openapi/             # 对外 API 契约（OpenAPI），新增接口必须更新
├── build/                   # CI / 打包脚本（占位）
├── cmd/gorgi/               # 入口；只做参数解析
├── configs/                 # 配置文件模板
├── deployments/docker/      # Dockerfile
├── docs/                    # 架构 / 规范文档
├── internal/
│   ├── app/                 # 应用层
│   │   ├── api/             # HTTP handler（控制器）
│   │   ├── service/         # 业务编排层
│   │   ├── repository/      # 数据访问层
│   │   ├── middleware/      # gin 中间件
│   │   ├── router/          # 路由注册 + 版本分发器
│   │   ├── server/          # http.Server 装配
│   │   └── app.go           # 整个应用的装配入口
│   └── pkg/                 # 项目内通用基础设施
│       ├── breaker/         # 出站熔断
│       ├── config/          # 配置（强类型 + viper）
│       ├── ent/             # ent schema + 生成代码
│       ├── errs/            # 统一错误码
│       ├── log/             # slog 封装（json + 脱敏 + request_id）
│       ├── mysql/           # MySQL 连接
│       ├── redis/           # Redis 客户端
│       ├── response/        # 统一响应
│       └── tracing/         # OTel 初始化
├── pkg/                     # 对外公开库（暂无）
├── scripts/                 # build.sh / ent-gen.sh
├── test/e2e/                # 端到端测试（不依赖 MySQL/Redis）
├── Makefile
└── go.mod
```

---

## 3. 快速开始

### 3.1 环境要求

- Go 1.26+
- MySQL 8.0.18+（生产必需；本地开发可不装，DSN 留空即可跳过）
- Redis 7.2+（同上）
- 可选：Docker 24+

### 3.2 启动

```bash
cp configs/config.yaml.example configs/config.yaml
# 按需修改 mysql / redis 等连接串
make tidy           # 安装依赖
make ent            # 生成 ent 代码（首次）
make run            # 启动
```

启动成功后访问：

```bash
curl -i http://localhost:8080/api/user/42 \
  -H "X-App-Key: demo-app" \
  -H "X-Timestamp: $(date +%s)" \
  -H "X-Nonce: n1" \
  -H "X-Sign: <按算法生成>"
```

签名算法：`HMAC-SHA256(appSecret, appKey + timestamp + nonce + method + path + rawQuery + rawBody)`，hex 小写。`rawQuery` 即 URL `?` 之后的原始串（无 query 时传空），把它纳入签名是为了防止攻击者抓到合法签名后只改 query 参数就能重放。本地调试可在 config 里设 `middleware.sign.enabled: false` 关闭验签。

---

## 4. 开发流程（重要）

### 4.1 新增数据表

1. 在 `internal/pkg/ent/schema/` 新增 `xxx.go` 描述 schema；
2. `make ent` 触发代码生成；
3. 在 `internal/app/repository/<domain>/` 写 dao，封装对该表的访问；
4. service 层调用 dao，handler 调用 service。

### 4.2 新增接口

路由采用**模块化注册**：每个业务域自带一个 `routes.go`，router 包只在 `modules()` 里串联各模块，不需要也不应该感知具体业务。

**在已有 domain 下加接口**（不需要改动 `router` 包）：

1. 在 `internal/app/api/<domain>/` 新增 handler 方法；
   - 命名约定：基础方法叫 `Xxx`，版本方法叫 `XxxV1` / `XxxV2` …
   - 方法签名必须是 `func(c *gin.Context)`
2. 在该 domain 的 `routes.go` 里追加一行：

   ```go
   g.GET("/user/:id", router.Dispatch(api, "GetUser"))
   ```

   `Dispatch` 会自动支持 `GetUser` / `GetUserV1` / `GetUserV2` …
3. 业务实现写在 `internal/app/service/<domain>/`。
4. 在 `api/openapi/gorgi.yaml` 增加接口契约。

**新增一个 domain**（如 `order`）：

1. 在 `internal/app/api/order/` 下建 handler 与 `routes.go`，导出 `func Routes(deps router.Deps) router.Module`；
2. 在 `internal/app/modules.go::modules` 列表追加一行 `order.Routes(deps)`。

### 4.3 接口版本分发约定

请求头 `x-api-version` 决定调用哪个方法：

| 请求头 | 命中方法 | 说明 |
|---|---|---|
| 无 | `GetUser` | 基础方法 |
| `v1` | `GetUserV1` | 命中 v1 |
| `v2` | `GetUserV2` | 命中 v2 |
| `v99`（未实现） | `GetUser` | **回退到基础方法** |
| 非法值 | `GetUser` | 回退 |

启动期会扫描并缓存版本方法，运行时无反射开销；缺少基础方法会**启动时 panic**，把问题前置。

### 4.4 新增中间件

1. 在 `internal/app/middleware/xxx.go` 实现（`func(c *gin.Context)`）；
2. 在 `internal/app/router/router.go::New` 中按顺序 `engine.Use(...)`：
   `Recovery → Trace → AccessLog → OTel → CORS → RateLimit → Sign → 业务路由`。

### 4.5 新增配置项

1. 在 `internal/pkg/config/config.go` 的 struct 上加字段（带 `mapstructure` tag）；
2. 在 `configs/config.yaml.example` 同步示例值；
3. 业务通过依赖注入（`router.Deps` / 构造函数）读取，**不要**到处 `viper.GetString`。

---

## 5. 日志规范

- 业务统一通过 `log.L(c.Request.Context()).Info(...)` 获取 logger；不要直接用 `slog.Default`，更不要 `slog.SetDefault`。
- 日志固定 JSON 格式，所有同请求的日志自动带相同的 `request_id`（由 `Trace` 中间件注入）。
- 敏感字段在配置 `log.mask_fields` 中声明，自动按"首尾各保留一个字符，中间用 `*` 填充"的规则脱敏，递归处理 `slog.Group`。
- 默认脱敏字段：`password / token / id_card / phone / app_secret`，按需追加。
- 不要把 `Authorization` / `Cookie` 整段打到日志里（`AccessLog` 中间件已自动屏蔽常见敏感请求头）。

---

## 6. 出站熔断

业务在调用 MySQL / Redis / 外部 HTTP 时，统一通过 `breaker.Manager.Do`：

```go
err := mgr.Do(ctx, "mysql.user.get", func(ctx context.Context) error {
    return repo.Get(ctx, id)
})
```

熔断打开时返回 `errs.BreakerOpen`，框架会响应 HTTP 503。规则可在 `breaker.rules` 里按 name 覆盖默认值。

---

## 7. 测试规范

- **单元测试**：与源文件同目录、同包，文件名 `xxx_test.go`。
- **端到端测试**：放 `test/e2e/`，使用 `httptest` 驱动 `router.New(...)`，**不依赖外部组件**。
- 跑测试：`make test`（自带 `-race -cover`）。

已覆盖的关键路径：
- 版本分发：基础方法 / v1 命中 / v99 回退 / 非法版本 / 缺基础方法
- 验签：合法 / 缺头 / 错 appKey / 时间偏差 / 签名不匹配
- 日志脱敏：顶层字段 / 嵌套 group / request_id 注入

---

## 8. 编码规范

- 优先 [Effective Go](https://go.dev/doc/effective_go)，其次 [Uber Go Style Guide](https://github.com/uber-go/guide)。
- **所有注释使用中文**；导出标识符必须有以名字开头、句号结尾的中文注释。
- 错误统一通过 `*errs.Err`，HTTP 输出统一通过 `response.OK / response.Fail`。
- `panic` 仅用于不可恢复的初始化错误。

完整规范见 [`docs/conventions.md`](docs/conventions.md)。

---

## 9. 提交规范

[Conventional Commits](https://www.conventionalcommits.org/zh-hans/)：

```
feat:     新功能
fix:      bug 修复
refactor: 重构
docs:     文档
test:     测试
chore:    构建 / 依赖 / CI
perf:     性能优化
```

示例：`feat(account): 支持 GetUserV2 返回邮箱字段`

---

## 10. 部署

```bash
make docker
docker run -p 8080:8080 \
  -v $(pwd)/configs/config.yaml:/app/configs/config.yaml \
  gorgi:latest
```

`deployments/docker/Dockerfile` 是多阶段构建。

---

## 11. 常见问题

- **启动报"配置文件不存在"**：检查 `-c` 参数 / `GORGI_CONFIG` 环境变量 / `./configs/config.yaml` 是否存在。
- **`make ent` 失败**：先 `make tidy`，确保 `entgo.io/ent` 已下载到 module cache。
- **本地无 MySQL/Redis 想启动框架**：把 `mysql.dsn` 与 `redis.addr` 留空，应用会跳过这两个组件初始化（仅 dev 用）。
- **签名一直 401**：检查 timestamp 是不是秒级、签名串顺序、appKey/appSecret 是否在 `middleware.sign.apps` 中。
