package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holosola/gorgi/internal/app/config"
	"golang.org/x/exp/slog"
)

type Resp struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data map[string]any `json:"data"`
}

func Hello(c *gin.Context) {
	conf := config.GetConfig()
	slog.Info("i will got at ", slog.String("date", "2023-04-30"), slog.String("8160000", "rmb"))

	rs := Resp{
		Code: conf.GetInt("mysql.port"),
		Msg:  conf.GetString("redis.password"),
		Data: map[string]any{"ping": "pong"},
	}
	c.JSON(http.StatusOK, rs)
}
