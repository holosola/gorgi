package router

import (
	"github.com/gin-gonic/gin"

	"github.com/holosola/gorgi/internal/app/middleware"
	"github.com/holosola/gorgi/internal/pkg/breaker"
	"github.com/holosola/gorgi/internal/pkg/config"
)

// Deps 是装配路由需要的全部外部依赖。
//
// 这里只暴露真正会被路由 / handler 用到的部分，而不是把整个 Config 透传，
// 便于在测试中构造最小依赖。
type Deps struct {
	Cfg     *config.Config
	Breaker *breaker.Manager
	// 后续可扩展：EntClient、RedisClient 等。
}

// New 装配整个 gin.Engine：注册中间件、挂载业务模块、设置 NoRoute / NoMethod。
//
// modules 是要挂到 /api 下的业务模块列表，由上层（app 包）按需组装传入；
// router 包不感知具体业务，避免循环依赖。
func New(deps Deps, modules ...Module) *gin.Engine {
	engine := gin.New()
	engine.NoRoute(MethodNotFound)
	engine.NoMethod(MethodNotFound)
	engine.Use(middlewares(deps)...)

	api := engine.Group("/api")
	for _, register := range modules {
		register(api)
	}
	return engine
}

// middlewares 返回挂在 engine 根上的中间件链。
//
// 注册顺序至关重要：Recovery → Trace → AccessLog → OTel → CORS → RateLimit → Sign。
// Recovery 必须最外层；Trace 必须在 AccessLog 之前注入 request_id；
// Sign 放在限流之后，避免无效签名请求消耗令牌。
func middlewares(deps Deps) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.Recovery(),
		middleware.Trace(),
		middleware.AccessLog(),
		middleware.OTel(deps.Cfg.OTel),
		middleware.CORS(deps.Cfg.Middleware.CORS),
		middleware.RateLimit(deps.Cfg.Middleware.RateLimit),
		middleware.Sign(deps.Cfg.Middleware.Sign),
	}
}
