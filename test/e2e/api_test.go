// Package e2e 是黑盒端到端测试，使用 httptest 直接驱动 gin.Engine，
// 不依赖 MySQL / Redis 等外部组件。
package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/holosola/gorgi/internal/app/api/account"
	"github.com/holosola/gorgi/internal/app/router"
	"github.com/holosola/gorgi/internal/pkg/breaker"
	"github.com/holosola/gorgi/internal/pkg/config"
)

// minimalDeps 构造一个不依赖外部组件、跳过签名 / 限流 / OTel 的最小依赖集，
// 仅用于 e2e 验证路由 / 中间件 / 版本分发的端到端正确性。
func minimalDeps() router.Deps {
	cfg := &config.Config{
		Middleware: config.MiddlewareConfig{
			Sign:      config.SignConfig{Enabled: false},
			RateLimit: config.RateLimitConfig{Enabled: false},
		},
	}
	return router.Deps{Cfg: cfg, Breaker: breaker.NewManager(cfg.Breaker)}
}

// newEngine 构造一个带 account 模块的最小 engine。
func newEngine() http.Handler {
	deps := minimalDeps()
	return router.New(deps, account.Routes(deps))
}

// doGet 发起一个 GET 请求，返回响应记录。
func doGet(engine http.Handler, path, version string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if version != "" {
		req.Header.Set(router.VersionHeader, version)
	}
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

// decodeData 把 response.Body 中的 data 字段拿出来。
func decodeData(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]any         `json:"data"`
		Raw  map[string]any         `json:"-"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("解析响应失败: %v, body=%s", err, body)
	}
	return resp.Data
}

func TestE2E_GetUser_NoVersion_HitsBase(t *testing.T) {
	engine := newEngine()
	w := doGet(engine, "/api/user/42", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	data := decodeData(t, w.Body.Bytes())
	if data["version"] != "base" {
		t.Errorf("无版本头应命中 base, got %v", data["version"])
	}
	if data["id"] != "42" {
		t.Errorf("path 参数解析错误, got %v", data["id"])
	}
	// 同时应当在响应头里能看到 request_id
	if w.Header().Get("X-Request-ID") == "" {
		t.Errorf("响应缺少 X-Request-ID")
	}
}

func TestE2E_GetUser_V1(t *testing.T) {
	engine := newEngine()
	w := doGet(engine, "/api/user/42", "v1")
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	data := decodeData(t, w.Body.Bytes())
	if data["version"] != "v1" {
		t.Errorf("v1 应命中 GetUserV1, got %v", data["version"])
	}
}

func TestE2E_GetUser_V99_FallsBackToBase(t *testing.T) {
	engine := newEngine()
	w := doGet(engine, "/api/user/42", "v99")
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	data := decodeData(t, w.Body.Bytes())
	if data["version"] != "base" {
		t.Errorf("无对应版本应回退到 base, got %v", data["version"])
	}
}

func TestE2E_NotFound(t *testing.T) {
	engine := newEngine()
	w := doGet(engine, "/no/such/route", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在的路由应 404, got %d", w.Code)
	}
}
