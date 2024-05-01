package log

import (
	"io"
	"os"
	"time"

	"log/slog"

	"github.com/holosola/gorgi/internal/app/config"
	"github.com/holosola/gorgi/internal/pkg/ex"
	"github.com/spf13/viper"
)

var appConf *viper.Viper
var logeer *slog.Logger

func setOpts() *slog.HandlerOptions {
	return &slog.HandlerOptions{
		AddSource: appConf.GetBool("app.log.addSource"),
		Level:     slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05.000"))
			}
			return a
		},
	}
}

func setOutput() io.Writer {
	outMode := appConf.GetString("app.log.outMode")
	if outMode != "file" {
		return os.Stdout
	}
	path := appConf.GetString("app.log.path")
	if path == "" {
		path, err := os.Getwd()
		if err != nil {
			ex.Handle(err)
		}
		path += "/logs"
	}
	prefix := appConf.GetString("app.log.prefix")
	if prefix == "" {
		prefix = "gorgi_"
	}
	logFile := path + "/" + prefix + time.Now().Format("20060102") + ".log"
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0640))
	if err != nil {
		ex.Handle(err)
	}
	return f
}

func init() {
	appConf = config.GetConfig()
	setOpts()
	setOutput()
	logeer = slog.New(slog.NewJSONHandler(setOutput(), setOpts()))
	slog.SetDefault(logeer)
}

func GetLogger() *slog.Logger {
	cp := *logeer
	return &cp
}
