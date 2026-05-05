package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// fakeHandler 提供基础方法和两个版本方法。
type fakeHandler struct{}

func (h *fakeHandler) Get(c *gin.Context)   { c.String(http.StatusOK, "base") }
func (h *fakeHandler) GetV1(c *gin.Context) { c.String(http.StatusOK, "v1") }
func (h *fakeHandler) GetV2(c *gin.Context) { c.String(http.StatusOK, "v2") }

// 名字看起来像版本但其实不是版本（GetVer），不应被当作版本方法。
func (h *fakeHandler) GetVer(c *gin.Context) { c.String(http.StatusOK, "ver") }

// fakeBaseOnly 只有基础方法。
type fakeBaseOnly struct{}

func (h *fakeBaseOnly) Get(c *gin.Context) { c.String(http.StatusOK, "base") }

func newEngineWithDispatch(handler any, base string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", Dispatch(handler, base))
	return r
}

func doGet(r *gin.Engine, version string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	if version != "" {
		req.Header.Set(VersionHeader, version)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestDispatch_NoVersionHeader_HitsBase(t *testing.T) {
	r := newEngineWithDispatch(&fakeHandler{}, "Get")
	w := doGet(r, "")
	if w.Body.String() != "base" {
		t.Fatalf("无版本头应命中 base, got %q", w.Body.String())
	}
}

func TestDispatch_V1_HitsV1(t *testing.T) {
	r := newEngineWithDispatch(&fakeHandler{}, "Get")
	w := doGet(r, "v1")
	if w.Body.String() != "v1" {
		t.Fatalf("v1 应命中 GetV1, got %q", w.Body.String())
	}
}

func TestDispatch_V99_FallsBackToBase(t *testing.T) {
	r := newEngineWithDispatch(&fakeHandler{}, "Get")
	w := doGet(r, "v99")
	if w.Body.String() != "base" {
		t.Fatalf("无对应版本应回退到 base, got %q", w.Body.String())
	}
}

func TestDispatch_GarbageVersion_FallsBackToBase(t *testing.T) {
	r := newEngineWithDispatch(&fakeHandler{}, "Get")
	w := doGet(r, "abc")
	if w.Body.String() != "base" {
		t.Fatalf("非法版本头应回退到 base, got %q", w.Body.String())
	}
}

func TestDispatch_PanicsWhenBaseMissing(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("缺基础方法应启动期 panic")
		}
	}()
	Dispatch(&fakeBaseOnly{}, "NoSuchBase")
}

func TestDispatch_BaseOnly_Works(t *testing.T) {
	r := newEngineWithDispatch(&fakeBaseOnly{}, "Get")
	w := doGet(r, "v1")
	if w.Body.String() != "base" {
		t.Fatalf("只有 base 时所有版本都应回退, got %q", w.Body.String())
	}
}

func TestNormalizeVersion(t *testing.T) {
	cases := map[string]string{
		"":     "",
		"v1":   "v1",
		"V1":   "v1",
		" v2 ": "v2",
		"v":    "",
		"vx":   "",
		"abc":  "",
		"v01":  "v01",
		"v100": "v100",
	}
	for in, want := range cases {
		if got := normalizeVersion(in); got != want {
			t.Errorf("normalizeVersion(%q)=%q, want %q", in, got, want)
		}
	}
}
