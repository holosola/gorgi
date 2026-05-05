// Package account 是账户相关的 HTTP handler。
//
// 同一个业务方法可同时支持多个 API 版本：在 handler 上分别实现 GetUser、GetUserV1、GetUserV2，
// 路由通过 router.Dispatch 在请求到达时按 x-api-version 头分发。
package account

import (
	"github.com/gin-gonic/gin"

	"github.com/holosola/gorgi/internal/pkg/log"
	"github.com/holosola/gorgi/internal/pkg/response"
)

// UserAPI 是账户接口的 handler 集合。后续依赖通过构造函数注入。
type UserAPI struct{}

// NewUserAPI 构造 UserAPI。
func NewUserAPI() *UserAPI { return &UserAPI{} }

// GetUser 是 GET /api/user/:id 的基础版本。当请求未携带 x-api-version 头，
// 或携带的版本未在本 handler 上实现时，会回退到本方法。
func (h *UserAPI) GetUser(c *gin.Context) {
	id := c.Param("id")
	log.L(c.Request.Context()).Info("GetUser 基础版本", "id", id)
	response.OK(c, gin.H{
		"id":      id,
		"name":    "demo-base",
		"version": "base",
	})
}

// GetUserV1 是 v1 版本，请求头 x-api-version: v1 时命中。
func (h *UserAPI) GetUserV1(c *gin.Context) {
	id := c.Param("id")
	log.L(c.Request.Context()).Info("GetUser V1 版本", "id", id)
	response.OK(c, gin.H{
		"id":      id,
		"name":    "demo-v1",
		"version": "v1",
		// V1 版本相比基础版多返回的字段示例
		"phone": "13800000000",
	})
}
