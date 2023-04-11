package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestId() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("requestId", uuid.New().String())
	}
}
