// Package middleware 汇总项目所有 gin 中间件实现。
//
// 注册顺序见 router 包：Recovery -> Trace -> AccessLog -> OTel -> CORS -> RateLimit -> Sign。
package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/holosola/gorgi/internal/pkg/errs"
	"github.com/holosola/gorgi/internal/pkg/log"
	"github.com/holosola/gorgi/internal/pkg/response"
)

// Recovery 兜底所有 panic，避免单次请求把整个进程带走。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.L(c.Request.Context()).Error("发生 panic",
					"recover", r,
					"stack", string(debug.Stack()),
				)
				response.AbortFail(c, errs.Internal)
			}
		}()
		c.Next()
	}
}
