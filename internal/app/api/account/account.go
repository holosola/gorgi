package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/holosola/gorgi/internal/app/config"
	"github.com/holosola/gorgi/internal/pkg/log"
	"golang.org/x/exp/slog"
)

type Resp struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data map[string]any `json:"data"`
}

func Hello(c *gin.Context) {
	conf := config.GetConfig()
	log.Info("account Hello Function", slog.String("date", "2023-04-30"), slog.Int("conf int", conf.GetInt("mysql.port")))
	rs := Resp{
		Code: conf.GetInt("mysql.port"),
		Msg:  conf.GetString("redis.password"),
		Data: map[string]any{"ping": "pong"},
	}
	c.JSON(http.StatusOK, rs)
}
