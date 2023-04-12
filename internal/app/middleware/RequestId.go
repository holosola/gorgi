package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestId() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqId := ctx.GetHeader("X-Request-ID")
		if reqId == "" {
			reqId = uuid.New().String()
		}
		ctx.Set("requestId", reqId)
	}
}
