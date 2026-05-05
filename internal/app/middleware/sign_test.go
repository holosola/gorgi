package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/holosola/gorgi/internal/pkg/config"
)

// newSignedRequest 构造一个带合法签名的请求。path 可以包含 query string。
func newSignedRequest(t *testing.T, method, path, secret, appKey string, body []byte, ts time.Time) *http.Request {
	t.Helper()
	r := httptest.NewRequest(method, path, bytes.NewReader(body))
	tsStr := strconv.FormatInt(ts.Unix(), 10)
	nonce := "n-1"
	sig := buildSignature(secret, appKey, tsStr, nonce, method, r.URL.Path, r.URL.RawQuery, body)
	r.Header.Set(headerAppKey, appKey)
	r.Header.Set(headerTimestamp, tsStr)
	r.Header.Set(headerNonce, nonce)
	r.Header.Set(headerSign, sig)
	return r
}

// runWithSign 用 Sign 中间件包一个最简 handler 并跑一次。
func runWithSign(cfg config.SignConfig, r *http.Request) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(Sign(cfg))
	engine.Any("/*any", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, r)
	return w
}

func TestSign_Disabled(t *testing.T) {
	w := runWithSign(config.SignConfig{Enabled: false},
		httptest.NewRequest(http.MethodGet, "/x", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("禁用签名时应直接放行, got %d", w.Code)
	}
}

func TestSign_Valid(t *testing.T) {
	cfg := config.SignConfig{
		Enabled: true,
		Apps:    []config.SignApp{{AppKey: "k", AppSecret: "s"}},
	}
	r := newSignedRequest(t, http.MethodPost, "/api/x", "s", "k", []byte(`{"a":1}`), time.Now())
	w := runWithSign(cfg, r)
	if w.Code != http.StatusOK {
		t.Fatalf("合法签名应放行, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestSign_BadAppKey(t *testing.T) {
	cfg := config.SignConfig{Enabled: true, Apps: []config.SignApp{{AppKey: "k", AppSecret: "s"}}}
	r := newSignedRequest(t, http.MethodGet, "/api/x", "s", "unknown", nil, time.Now())
	w := runWithSign(cfg, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("未知 appKey 应被拒, got %d", w.Code)
	}
}

func TestSign_TimestampOutOfSkew(t *testing.T) {
	cfg := config.SignConfig{
		Enabled:       true,
		TimestampSkew: 1 * time.Minute,
		Apps:          []config.SignApp{{AppKey: "k", AppSecret: "s"}},
	}
	r := newSignedRequest(t, http.MethodGet, "/api/x", "s", "k", nil, time.Now().Add(-10*time.Minute))
	w := runWithSign(cfg, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("超出时间偏差应被拒, got %d", w.Code)
	}
}

func TestSign_SignatureMismatch(t *testing.T) {
	cfg := config.SignConfig{Enabled: true, Apps: []config.SignApp{{AppKey: "k", AppSecret: "s"}}}
	r := newSignedRequest(t, http.MethodGet, "/api/x", "s", "k", nil, time.Now())
	r.Header.Set(headerSign, "deadbeef")
	w := runWithSign(cfg, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("签名错误应被拒, got %d", w.Code)
	}
}

func TestSign_MissingHeaders(t *testing.T) {
	cfg := config.SignConfig{Enabled: true, Apps: []config.SignApp{{AppKey: "k", AppSecret: "s"}}}
	r := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	w := runWithSign(cfg, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("缺头应被拒, got %d", w.Code)
	}
}

// TestSign_QueryTamperingRejected 验证：捕获了一个合法签名后，仅修改 query string
// 不重算签名，必须被拒（防止 query 参数重放篡改攻击）。
func TestSign_QueryTamperingRejected(t *testing.T) {
	cfg := config.SignConfig{Enabled: true, Apps: []config.SignApp{{AppKey: "k", AppSecret: "s"}}}
	r := newSignedRequest(t, http.MethodGet, "/api/users?role=user", "s", "k", nil, time.Now())
	// 模拟攻击者把 query 改成 role=admin，但保留原签名头。
	r.URL.RawQuery = "role=admin"
	w := runWithSign(cfg, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("篡改 query 后必须被拒, got %d", w.Code)
	}
}

// TestSign_QueryIncludedInSignature 反向验证：带 query 的请求按新算法签名后能通过。
func TestSign_QueryIncludedInSignature(t *testing.T) {
	cfg := config.SignConfig{Enabled: true, Apps: []config.SignApp{{AppKey: "k", AppSecret: "s"}}}
	r := newSignedRequest(t, http.MethodGet, "/api/users?role=admin&page=2", "s", "k", nil, time.Now())
	w := runWithSign(cfg, r)
	if w.Code != http.StatusOK {
		t.Fatalf("带 query 的合法签名应放行, got %d body=%s", w.Code, w.Body.String())
	}
}
