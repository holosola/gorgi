package log

import (
	"os"
	"time"

	"github.com/holosola/gorgi/internal/app/config"
	"github.com/holosola/gorgi/internal/pkg/ex"
	"golang.org/x/exp/slog"
)

func InitLog() {
	conf := config.GetConfig()
	path := conf.GetString("app.log.path")
	prefix := conf.GetString("app.log.prefix")
	logFile := path + "/" + prefix + time.Now().Format("20060102") + ".log"
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0640))
	if err != nil {
		ex.Handle(err)
	}

	myLog := slog.New(
		slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05.000"))
				}
				return a
			},
		}.NewJSONHandler(f))
	slog.SetDefault(myLog)
}
