package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/holosola/gorgi/internal/pkg/config"
)

// CORS 按配置返回 gin-contrib/cors 中间件。
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     defaultIfEmpty(cfg.AllowOrigins, []string{"*"}),
		AllowMethods:     defaultIfEmpty(cfg.AllowMethods, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		AllowHeaders:     defaultIfEmpty(cfg.AllowHeaders, []string{"Origin", "Content-Type", "Authorization"}),
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
}

// defaultIfEmpty 在配置缺失时给一个默认值。
func defaultIfEmpty(v, def []string) []string {
	if len(v) == 0 {
		return def
	}
	return v
}
