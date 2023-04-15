package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holosola/gorgi/internal/app/middleware"
)

func Init(r *gin.Engine) {
	r.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, "{}")
	})

	r.NoMethod(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, "{}")
	})
	r.Use(middleware.TraceRequest())
	r.Use(middleware.LogRequest())
	RegisterInternal(r)
	RegisterAPIRouters(r)
}
