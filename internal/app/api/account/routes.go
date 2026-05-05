package account

import (
	"github.com/gin-gonic/gin"

	"github.com/holosola/gorgi/internal/app/router"
)

// Routes 返回 account 模块的路由注册函数。
//
// 新增 account 域下的接口时，只需要在本文件追加路由，不需要改动 router 包。
// deps 用于注入未来可能依赖的 dao / service / 配置等，当前示例未使用。
func Routes(_ router.Deps) router.Module {
	api := NewUserAPI()
	return func(g *gin.RouterGroup) {
		// 同一路由可同时支持多个版本：根据 x-api-version 自动分发到 GetUser / GetUserV1 / GetUserV2…
		g.GET("/user/:id", router.Dispatch(api, "GetUser"))
	}
}
