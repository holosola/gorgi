package router

import (
	"github.com/gin-gonic/gin"
	"github.com/holosola/gorgi/internal/app/api/account"
)

func RegisterAPIRouters(r *gin.Engine) {
	api := r.Group("/api/")
	{
		api.GET("/hello", account.Hello)
	}
}
