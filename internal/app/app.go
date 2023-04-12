package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holosola/gorgi/internal/app/router"
	"github.com/holosola/gorgi/internal/pkg/log"
)

func Start() error {
	engine := gin.New()
	log.InitLog()
	router.Init(engine)

	s := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           engine,
		ReadTimeout:       10,
		ReadHeaderTimeout: 10,
		WriteTimeout:      10,
		IdleTimeout:       10,
		MaxHeaderBytes:    1 << 20,
	}
	return s.ListenAndServe()
}
