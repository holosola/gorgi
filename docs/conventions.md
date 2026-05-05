# 编码与协作规范

## 代码风格

- 优先遵循 [Effective Go](https://go.dev/doc/effective_go)，其次 [Uber Go Style Guide](https://github.com/uber-go/guide)。
- 所有注释使用中文。
- 导出标识符（包 / 类型 / 函数 / 变量）必须有句号结尾的中文注释，且以名字开头（Effective Go 强约束）。
- 错误信息使用小写中文，不带句号；包装错误用 `fmt.Errorf("xxx: %w", err)`。
- `panic` 仅用于"无法恢复的初始化错误"或"明显的编程错误"。业务错误一律返回 `error`。

## 命名

- 包名：小写、单词、不带下划线 / 复数。
- 文件名：小写下划线 `snake_case.go`。
- 接口：以 `er` 结尾或语义化命名，避免 `IXxx` 这类前缀。
- 缩写词整体大小写一致：`URL`、`HTTP`、`ID`，例如 `userID` 而不是 `userId`。

## 错误处理

- 所有边界错误必须用 `*errs.Err` 包装，便于统一响应。
- 不允许在业务里直接 `c.JSON(...)`，必须用 `response.OK / response.Fail`。

## 日志

- 业务获取 logger：`log.L(ctx)`。**不要**调用 `slog.Default` 或 `slog.SetDefault`。
- 敏感字段（密码 / token / 手机号）必须命中 `mask_fields` 配置；新增敏感字段时同步更新 `configs/config.yaml.example`。
- 不要把整段 `Authorization` / `Cookie` 打到日志里。

## 提交规范（Conventional Commits）

| 类型 | 用途 |
|---|---|
| `feat:` | 新功能 |
| `fix:` | bug 修复 |
| `refactor:` | 重构（不影响行为） |
| `docs:` | 文档 |
| `test:` | 测试 |
| `chore:` | 构建 / 依赖 / CI 等杂项 |
| `perf:` | 性能优化 |

示例：

```
feat(account): 支持 GetUserV2 返回邮箱字段
```

## PR 流程

1. 从 `master` 切出 `feat/xxx` 或 `fix/xxx` 分支
2. 提交前必须通过 `make vet && make test`
3. 接口变更必须更新 `api/openapi/gorgi.yaml`
4. 影响数据表的变更必须附带 ent schema 修改 + `make ent` 后的生成代码
