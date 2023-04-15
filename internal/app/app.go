package app

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/holosola/gorgi/internal/app/router"
)

func Start() error {
	engine := gin.New()
	router.Init(engine)

	s := &http.Server{
		Addr:              ":8080",
		Handler:           engine,
		ReadTimeout:       time.Second * 60,
		ReadHeaderTimeout: time.Second * 10,
		WriteTimeout:      time.Second * 60,
		IdleTimeout:       time.Second * 120,
		MaxHeaderBytes:    40960,
	}
	return s.ListenAndServe()
}
