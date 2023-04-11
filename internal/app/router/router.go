package router

import (
	"github.com/gin-gonic/gin"
	"github.com/holosola/gorgi/internal/app/middleware"
)

func Init(r *gin.Engine) {
	r.NoRoute(func(ctx *gin.Context) {
		ctx.String(404, "Not router!!")
	})

	r.NoMethod(func(ctx *gin.Context) {
		ctx.String(404, "Not method!!")
	})
	r.Use(middleware.RequestId())
	r.Use(middleware.RequestLogger())
	RegisterInternal(r)
	RegisterAPIRouters(r)
}
