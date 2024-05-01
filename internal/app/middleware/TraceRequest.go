package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/holosola/gorgi/internal/pkg/log"
)

func TraceRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqId := getReqId(ctx)
		ctx.Set("reqId", reqId)
		slog.SetDefault(log.GetLogger().With("reqId", reqId))
	}
}

func getReqId(ctx *gin.Context) string {
	reqId := ctx.GetString("reqId")
	if reqId != "" {
		return reqId
	}
	reqId = ctx.GetHeader("X-Request-ID")
	if reqId == "" {
		reqId = uuid.New().String()
	}
	return reqId
}
