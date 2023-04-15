package log

import (
	"context"
	"os"
	"time"

	"github.com/holosola/gorgi/internal/app/config"
	"github.com/holosola/gorgi/internal/pkg/ex"
	"golang.org/x/exp/slog"
)

var logger *slog.Logger

func init() {
	conf := config.GetConfig()
	path := conf.GetString("app.log.path")
	prefix := conf.GetString("app.log.prefix")
	logFile := path + "/" + prefix + time.Now().Format("20060102") + ".log"
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0640))
	if err != nil {
		ex.Handle(err)
	}

	logger = slog.New(
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
	SetDefault(logger)
}

func Clone() *slog.Logger {
	cp := *logger
	return &cp
}

func SetDefault(l *slog.Logger) {
	slog.SetDefault(l)
}

func With(args ...any) *slog.Logger {
	cp := Clone()
	return cp.With(args...)
}

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func DebugCtx(ctx context.Context, msg string, args ...any) {
	slog.DebugCtx(ctx, msg, args...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func InfoCtx(ctx context.Context, msg string, args ...any) {
	slog.InfoCtx(ctx, msg, args...)
}

func Warn(msg string, args ...any) {
	//do sth
	slog.Warn(msg, args...)
}

func WarnCtx(ctx context.Context, msg string, args ...any) {
	//do sth
	slog.WarnCtx(ctx, msg, args...)
}

func Error(msg string, args ...any) {
	//do sth
	slog.Error(msg, args...)
}
func ErrorCtx(ctx context.Context, msg string, args ...any) {
	//do sth
	slog.ErrorCtx(ctx, msg, args...)
}

func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	slog.Log(ctx, level, msg, args...)
}

func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, level, msg, attrs...)
}
