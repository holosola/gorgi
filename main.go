package main

import (
	"github.com/gin-gonic/gin"
	"github.com/holosola/gorgi/internal/app"
	"github.com/holosola/gorgi/internal/pkg/ex"
)

func main() {
	gin.SetMode(gin.DebugMode)
	//系统配置读取
	//日志
	e := app.Start()
	if e != nil {
		ex.Handle(e)
	}
}
