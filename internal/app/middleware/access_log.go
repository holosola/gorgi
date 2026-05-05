package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/holosola/gorgi/internal/pkg/log"
)

// 默认的请求体记录上限，超出部分以 "..." 截断。
const defaultMaxBodyBytes = 4 * 1024

// 屏蔽以下请求 / 响应头，避免把敏感信息打到日志里。
var sensitiveHeaders = map[string]struct{}{
	"authorization": {},
	"cookie":        {},
	"set-cookie":    {},
	"x-app-key":     {},
	"x-sign":        {},
}

// bodyWriter 拦截响应体，便于在日志中输出。
type bodyWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

// Write 同时写入底层 ResponseWriter 和内存缓冲。
func (w *bodyWriter) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

// AccessLog 记录每个请求的 URL/Header/Body/状态码/耗时。
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}
		bw := &bodyWriter{ResponseWriter: c.Writer, buf: &bytes.Buffer{}}
		c.Writer = bw

		c.Next()

		log.L(c.Request.Context()).Info("access",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
			"req_headers", filterHeaders(c.Request.Header),
			"req_body", truncate(string(reqBody), defaultMaxBodyBytes),
			"resp_headers", filterHeaders(c.Writer.Header()),
			"resp_body", truncate(bw.buf.String(), defaultMaxBodyBytes),
		)
	}
}

// filterHeaders 把 http.Header 转成扁平 map 并屏蔽敏感字段。
func filterHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		if _, hit := sensitiveHeaders[strings.ToLower(k)]; hit {
			out[k] = "[REDACTED]"
			continue
		}
		out[k] = strings.Join(v, ",")
	}
	return out
}

// truncate 把字符串截断到最大长度。
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
