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
		start := time.Now()
		reqData, _ := ctx.GetRawData()
		cw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = cw
		ctx.Next()
		slog.Info("RequestLogger",
			slog.Group("request",
				slog.String("URL", ctx.Request.URL.String()),
				slog.String("rawData", string(reqData))),
			slog.Group("response",
				slog.Int("status", ctx.Writer.Status()),
				slog.String("data", cw.body.String())),
			slog.Duration("duration", time.Since(start)),
		)
	}
}
