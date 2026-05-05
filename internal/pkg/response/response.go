// Package response 提供统一的 HTTP 响应封装，所有 handler 都应通过本包返回结果，
// 保证 {code, msg, data, request_id} 的一致结构。
package response

import (
	"github.com/gin-gonic/gin"

	"github.com/holosola/gorgi/internal/pkg/errs"
	"github.com/holosola/gorgi/internal/pkg/log"
)

// Body 是统一响应结构。
type Body struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// OK 写入成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(errs.OK.Status, Body{
		Code:      errs.OK.Code,
		Msg:       errs.OK.Message,
		Data:      data,
		RequestID: log.RequestIDFrom(c.Request.Context()),
	})
}

// Fail 写入失败响应。任意 error 都会被规整为 *errs.Err。
func Fail(c *gin.Context, err error) {
	e := errs.From(err)
	c.JSON(e.Status, Body{
		Code:      e.Code,
		Msg:       e.Message,
		RequestID: log.RequestIDFrom(c.Request.Context()),
	})
}

// AbortFail 与 Fail 相同，同时 Abort 后续中间件 / handler。
func AbortFail(c *gin.Context, err error) {
	Fail(c, err)
	c.Abort()
}
