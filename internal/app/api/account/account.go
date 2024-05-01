package account

import (
	"net/http"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/holosola/gorgi/internal/app/config"
)

type Resp struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data map[string]any `json:"data"`
}

func Hello(c *gin.Context) {
	conf := config.GetConfig()
	slog.Info("account Hello Function", slog.String("date", "2023-04-30"), slog.Int("conf int", conf.GetInt("mysql.port")))
	rs := Resp{
		Code: conf.GetInt("mysql.port"),
		Msg:  conf.GetString("redis.password"),
		Data: map[string]any{"ping": "pong", "query": c.Query("aaa")},
	}
	c.JSON(http.StatusOK, rs)
}
