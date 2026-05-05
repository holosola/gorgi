package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/holosola/gorgi/internal/pkg/config"
	"github.com/holosola/gorgi/internal/pkg/errs"
	"github.com/holosola/gorgi/internal/pkg/response"
)

// Sign 相关的请求头名字。
const (
	headerAppKey    = "X-App-Key"
	headerTimestamp = "X-Timestamp"
	headerNonce     = "X-Nonce"
	headerSign      = "X-Sign"
)

// Sign 校验请求签名。
//
// 签名算法：HMAC-SHA256(appSecret, appKey + timestamp + nonce + method + path + rawQuery + rawBody)
// 客户端把上述字符串拼接后 HMAC-SHA256，hex(lowercase) 编码后通过 X-Sign 提交。
// rawQuery 必须把 query 串原样（不解码、不排序）参与签名，避免攻击者
// 抓到合法签名后只改 query 参数就能重放。
func Sign(cfg config.SignConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }
	}
	secrets := make(map[string]string, len(cfg.Apps))
	for _, app := range cfg.Apps {
		if app.AppKey != "" && app.AppSecret != "" {
			secrets[app.AppKey] = app.AppSecret
		}
	}
	skew := cfg.TimestampSkew
	if skew <= 0 {
		skew = 5 * time.Minute
	}

	return func(c *gin.Context) {
		appKey := c.GetHeader(headerAppKey)
		ts := c.GetHeader(headerTimestamp)
		nonce := c.GetHeader(headerNonce)
		sign := c.GetHeader(headerSign)
		if appKey == "" || ts == "" || nonce == "" || sign == "" {
			response.AbortFail(c, errs.SignInvalid)
			return
		}
		secret, ok := secrets[appKey]
		if !ok {
			response.AbortFail(c, errs.SignInvalid)
			return
		}
		if !checkTimestamp(ts, skew) {
			response.AbortFail(c, errs.SignInvalid)
			return
		}
		body, err := readAndRestoreBody(c)
		if err != nil {
			response.AbortFail(c, errs.InvalidParam.Wrap(err))
			return
		}
		expect := buildSignature(secret, appKey, ts, nonce, c.Request.Method, c.Request.URL.Path, c.Request.URL.RawQuery, body)
		if !hmac.Equal([]byte(expect), []byte(sign)) {
			response.AbortFail(c, errs.SignInvalid)
			return
		}
		c.Next()
	}
}

// readAndRestoreBody 读出 body 并放回去，使后续 handler 能再次读取。
func readAndRestoreBody(c *gin.Context) ([]byte, error) {
	if c.Request.Body == nil {
		return nil, nil
	}
	b, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(b))
	return b, nil
}

// checkTimestamp 校验时间戳秒级数值与当前时间差不超过 skew。
func checkTimestamp(ts string, skew time.Duration) bool {
	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return false
	}
	delta := time.Since(time.Unix(sec, 0))
	if delta < 0 {
		delta = -delta
	}
	return delta <= skew
}

// buildSignature 构造签名值，业务代码与本函数共享算法。
// rawQuery 即 URL 的 "?" 之后部分（不含 "?"），无 query 时传空串。
func buildSignature(secret, appKey, ts, nonce, method, path, rawQuery string, body []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(appKey))
	h.Write([]byte(ts))
	h.Write([]byte(nonce))
	h.Write([]byte(method))
	h.Write([]byte(path))
	h.Write([]byte(rawQuery))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
