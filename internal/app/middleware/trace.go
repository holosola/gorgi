package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/holosola/gorgi/internal/pkg/log"
)

// requestIDHeader 是 request_id 在 HTTP 头中的名字。
const requestIDHeader = "X-Request-ID"

// Trace 注入 / 透传 request_id：
//   - 优先沿用客户端 X-Request-ID；
//   - 缺失则生成新的 UUID；
//   - 写入 context（供日志中间件 / 业务取用）；
//   - 同时回写到响应头，便于客户端反馈问题。
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(requestIDHeader)
		if reqID == "" {
			reqID = uuid.NewString()
		}
		ctx := log.WithRequestID(c.Request.Context(), reqID)
		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set(requestIDHeader, reqID)
		c.Set("request_id", reqID)
		c.Next()
	}
}
