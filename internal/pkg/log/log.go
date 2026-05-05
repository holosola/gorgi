// Package log 提供基于 slog 的日志封装，强约束 JSON 输出、敏感字段脱敏、
// 同一请求下日志带相同 request_id。业务统一通过 log.L(ctx) 获取 logger。
package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/holosola/gorgi/internal/pkg/config"
)

// requestIDKeyType 是放进 context 的 key，避免与其他包冲突。
type requestIDKeyType struct{}

// requestIDKey 是 context 中存储 request_id 的唯一 key。
var requestIDKey = requestIDKeyType{}

// RequestIDLogKey 是日志中 request_id 字段固定名字。
const RequestIDLogKey = "request_id"

// defaultLogger 持有当前默认 logger，初始化前默认输出到 stdout。
var defaultLogger atomic.Pointer[slog.Logger]

func init() {
	// 兜底 logger，避免业务在 Init 之前调用导致空指针。
	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	defaultLogger.Store(l)
}

// Init 按配置初始化全局 logger，应在 main 早期调用。
func Init(cfg config.LogConfig) error {
	w, err := buildWriter(cfg)
	if err != nil {
		return err
	}
	level := parseLevel(cfg.Level)

	jh := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     level,
	})

	h := &gorgiHandler{
		inner: jh,
		mask:  newMasker(cfg.MaskFields),
	}
	defaultLogger.Store(slog.New(h))
	slog.SetDefault(slog.New(h))
	return nil
}

// L 返回带 ctx 信息（含 request_id）的 logger，业务推荐用法：
//
//	log.L(ctx).Info("...", slog.String("k", "v"))
func L(ctx context.Context) *slog.Logger {
	return defaultLogger.Load()
}

// WithRequestID 把 request_id 写进 ctx，trace 中间件调用。
func WithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, requestIDKey, reqID)
}

// RequestIDFrom 从 ctx 取出 request_id，没有则返回空字符串。
func RequestIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// parseLevel 将字符串日志级别转成 slog.Level。
func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// buildWriter 根据 output 配置返回日志输出 writer。
func buildWriter(cfg config.LogConfig) (io.Writer, error) {
	if strings.ToLower(cfg.Output) != "file" {
		return os.Stdout, nil
	}
	path := cfg.FilePath
	if path == "" {
		path = "./logs/gorgi.log"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	return f, nil
}
