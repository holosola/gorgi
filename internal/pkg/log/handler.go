package log

import (
	"context"
	"log/slog"
)

// gorgiHandler 包装 slog.Handler，实现两件事：
//  1. 从 ctx 自动提取 request_id 写入日志；
//  2. 对命中脱敏列表的字段进行掩码处理。
type gorgiHandler struct {
	inner slog.Handler
	mask  *masker
}

// Enabled 透传到内部 handler。
func (h *gorgiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle 处理一条日志记录。
func (h *gorgiHandler) Handle(ctx context.Context, r slog.Record) error {
	if reqID := RequestIDFrom(ctx); reqID != "" {
		r.AddAttrs(slog.String(RequestIDLogKey, reqID))
	}
	if h.mask != nil {
		r = h.mask.maskRecord(r)
	}
	return h.inner.Handle(ctx, r)
}

// WithAttrs 透传时同时对 attrs 做脱敏。
func (h *gorgiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.mask != nil {
		out := make([]slog.Attr, len(attrs))
		for i, a := range attrs {
			out[i] = h.mask.maskAttr(a)
		}
		attrs = out
	}
	return &gorgiHandler{inner: h.inner.WithAttrs(attrs), mask: h.mask}
}

// WithGroup 透传 group。
func (h *gorgiHandler) WithGroup(name string) slog.Handler {
	return &gorgiHandler{inner: h.inner.WithGroup(name), mask: h.mask}
}
