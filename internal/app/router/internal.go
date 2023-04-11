package router

import (
	"github.com/gin-gonic/gin"
	"github.com/holosola/gorgi/internal/app/config"
)

func RegisterInternal(r *gin.Engine) {
	r.POST("/internal.action", func(ctx *gin.Context) {
		conf := config.GetConfig()

		ctx.JSON(200, gin.H{
			"internal": "asdasd",
			"username": conf.GetString("mysql.username"),
		})
	})
}
