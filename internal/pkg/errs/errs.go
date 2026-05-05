// Package errs 定义业务错误码与错误类型。
//
// 错误的传递链路：业务返回 *Err -> response.Fail 写出 HTTP 响应。
// 一个错误码同时承担两件事：HTTP 状态由 Status 决定，业务码由 Code 决定。
package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// Err 是项目内统一的错误类型。
type Err struct {
	Status  int    // HTTP 状态码
	Code    int    // 业务错误码
	Message string // 错误描述（面向客户端）
	Cause   error  // 原始错误（不会序列化给客户端）
}

// Error 实现 error 接口。
func (e *Err) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 暴露底层错误，配合 errors.Is/As 使用。
func (e *Err) Unwrap() error { return e.Cause }

// Wrap 返回一个新的 *Err，附带额外的原始错误信息。
func (e *Err) Wrap(cause error) *Err {
	cp := *e
	cp.Cause = cause
	return &cp
}

// New 构造一个 *Err。
func New(status, code int, message string) *Err {
	return &Err{Status: status, Code: code, Message: message}
}

// From 从 error 中提取 *Err，找不到则用 Internal 作为兜底。
func From(err error) *Err {
	if err == nil {
		return nil
	}
	var e *Err
	if errors.As(err, &e) {
		return e
	}
	return Internal.Wrap(err)
}

// 预定义错误。业务可继续在自己的包里 New 新错误码。
var (
	OK             = New(http.StatusOK, 0, "ok")
	Internal       = New(http.StatusInternalServerError, 10000, "服务器内部错误")
	InvalidParam   = New(http.StatusBadRequest, 10001, "请求参数不合法")
	Unauthorized   = New(http.StatusUnauthorized, 10002, "未授权")
	Forbidden      = New(http.StatusForbidden, 10003, "禁止访问")
	NotFound       = New(http.StatusNotFound, 10004, "资源不存在")
	MethodNotAllow = New(http.StatusMethodNotAllowed, 10005, "方法不允许")
	RateLimited    = New(http.StatusTooManyRequests, 10006, "请求过于频繁")
	SignInvalid    = New(http.StatusUnauthorized, 10007, "签名校验失败")
	BreakerOpen    = New(http.StatusServiceUnavailable, 10008, "下游服务不可用")
)
