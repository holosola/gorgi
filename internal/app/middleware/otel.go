package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/holosola/gorgi/internal/pkg/config"
)

// OTel 在启用时返回 otelgin 中间件，否则返回 noop。
func OTel(cfg config.OTelConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }
	}
	name := cfg.ServiceName
	if name == "" {
		name = "gorgi"
	}
	return otelgin.Middleware(name)
}
