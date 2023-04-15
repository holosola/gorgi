package middleware

import (
	"bytes"
	"time"

	"github.com/bytedance/sonic"
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

func LogRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		reqData, _ := ctx.GetRawData()
		cw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = cw
		ctx.Next()

		requestHeaders, _ := sonic.MarshalString(ctx.Request.Header)
		responseHeadrs, _ := sonic.MarshalString(cw.Header())

		slog.Info("log request and response",
			slog.Group("Request",
				slog.String("URL", ctx.Request.URL.String()),
				slog.String("Headers", requestHeaders),
				slog.String("Body", string(reqData))),
			slog.Group("Response",
				slog.Int("Status", ctx.Writer.Status()),
				slog.String("Headers", responseHeadrs),
				slog.String("Data", cw.body.String())),
			slog.Duration("Duration", time.Since(start)),
		)
	}
}
