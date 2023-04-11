package middleware

import (
	"bytes"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w CustomResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w CustomResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func RequestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now().Nanosecond()
		reqData, _ := ctx.GetRawData()
		cus := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = cus
		ctx.Next()
		ctx.Writer.Status()
		rep := cus.body.String()
		slog.Info("RequestLogger", slog.String("req:", string(reqData)), slog.Int("cost", time.Now().Nanosecond()-start), slog.String("resp", rep), slog.String("requestId", ctx.GetString("requestId")))
	}
}
