// Package router 负责装配 gin 路由树，包含中间件注册顺序与版本分发器。
package router

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

// VersionHeader 是接口版本约定的请求头名字。
const VersionHeader = "x-api-version"

// handlerFuncType 是合法 handler 方法的反射类型 func(*gin.Context)。
var handlerFuncType = reflect.TypeOf((func(*gin.Context))(nil))

// Dispatch 根据 handler 实例 + 基础方法名生成 gin.HandlerFunc。
//
// 行为：
//   - 启动期一次性扫描 handler 上所有 "<base>" 与 "<base>V<x>" 命名的方法，
//     校验签名并缓存，避免请求时反射查找；
//   - 请求期读取 x-api-version 头，命中则调用对应版本方法；
//   - 找不到时回退到基础方法（按用户确认的回退策略）；
//   - 基础方法不存在直接 panic，问题在启动期暴露，不留到运行时。
//
// handler 必须以指针接收者实现方法，且方法签名为 func(*gin.Context)。
func Dispatch(handler any, base string) gin.HandlerFunc {
	if handler == nil {
		panic("router.Dispatch: handler 为空")
	}
	if base == "" {
		panic("router.Dispatch: base 为空")
	}
	versions := scanVersions(handler, base)
	baseFn, ok := versions[""]
	if !ok {
		panic(fmt.Sprintf("router.Dispatch: handler %T 上找不到基础方法 %s", handler, base))
	}
	return func(c *gin.Context) {
		v := normalizeVersion(c.GetHeader(VersionHeader))
		if v != "" {
			if fn, ok := versions[v]; ok {
				fn(c)
				return
			}
		}
		baseFn(c)
	}
}

// scanVersions 扫描 handler 类型上的方法，返回 map：版本号（小写，不含 v）-> 方法。
// 版本号 "" 表示基础方法。
func scanVersions(handler any, base string) map[string]func(*gin.Context) {
	t := reflect.TypeOf(handler)
	out := make(map[string]func(*gin.Context))
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if !strings.HasPrefix(m.Name, base) {
			continue
		}
		suffix := m.Name[len(base):]
		// 基础方法：suffix 为空。
		// 版本方法：suffix 形如 V1 / V2 / V12，必须以大写 V 开头，且后面是非空数字。
		var version string
		switch {
		case suffix == "":
			version = ""
		case len(suffix) >= 2 && suffix[0] == 'V' && allDigits(suffix[1:]):
			version = strings.ToLower(suffix)
		default:
			continue
		}
		fn, ok := toGinHandler(handler, m)
		if !ok {
			panic(fmt.Sprintf("router.Dispatch: 方法 %s.%s 签名不是 func(*gin.Context)",
				t.String(), m.Name))
		}
		out[version] = fn
	}
	return out
}

// toGinHandler 把反射得到的 method 转换成 gin.HandlerFunc。
// 第二个返回值表示签名是否合法。
func toGinHandler(handler any, m reflect.Method) (func(*gin.Context), bool) {
	mv := reflect.ValueOf(handler).MethodByName(m.Name)
	if !mv.IsValid() {
		return nil, false
	}
	fn, ok := mv.Interface().(func(*gin.Context))
	if !ok {
		// 兜底：用反射类型也校验一次，便于错误信息更明确。
		if mv.Type() != handlerFuncType {
			return nil, false
		}
		return nil, false
	}
	return fn, true
}

// normalizeVersion 把 "V1" / "v1" / " v1 " 都规整成 "v1"。
// 不合法（如 "abc"）返回空串，由 Dispatch 兜底回退。
func normalizeVersion(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return ""
	}
	if !strings.HasPrefix(v, "v") || len(v) < 2 || !allDigits(v[1:]) {
		return ""
	}
	return v
}

// allDigits 判断字符串是否全部为数字（且非空）。
func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// MethodNotFound 是给 NoRoute / NoMethod 用的统一处理函数。
func MethodNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"code": 10004, "msg": "资源不存在"})
}
