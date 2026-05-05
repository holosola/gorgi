package log

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

// newTestHandler 构造可写到 buffer 的 gorgi handler，便于断言。
func newTestHandler(buf *bytes.Buffer, fields []string) *gorgiHandler {
	return &gorgiHandler{
		inner: slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
		mask:  newMasker(fields),
	}
}

// decode 把单行 JSON 解析为 map，方便断言。
func decode(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(b), &m); err != nil {
		t.Fatalf("解析日志 JSON 失败: %v, 原始: %s", err, b)
	}
	return m
}

func TestMaskString(t *testing.T) {
	cases := map[string]string{
		"":         "",
		"a":        "*",
		"ab":       "**",
		"abc":      "a*c",
		"password": "p******d",
	}
	for in, want := range cases {
		if got := maskString(in); got != want {
			t.Errorf("maskString(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestHandlerMasksTopLevelField(t *testing.T) {
	buf := &bytes.Buffer{}
	h := newTestHandler(buf, []string{"password"})
	logger := slog.New(h)

	logger.Info("login", slog.String("user", "alice"), slog.String("password", "secret123"))

	got := decode(t, buf.Bytes())
	if got["user"] != "alice" {
		t.Errorf("非脱敏字段被改动: %v", got["user"])
	}
	if got["password"] == "secret123" {
		t.Errorf("password 未被脱敏: %v", got["password"])
	}
	if !strings.Contains(got["password"].(string), "*") {
		t.Errorf("脱敏结果应包含 *, got %v", got["password"])
	}
}

func TestHandlerMasksNestedGroup(t *testing.T) {
	buf := &bytes.Buffer{}
	h := newTestHandler(buf, []string{"token"})
	logger := slog.New(h)

	logger.Info("req",
		slog.Group("auth",
			slog.String("token", "abcdefg"),
			slog.String("user", "bob"),
		),
	)

	got := decode(t, buf.Bytes())
	auth, ok := got["auth"].(map[string]any)
	if !ok {
		t.Fatalf("auth 不是 group: %v", got["auth"])
	}
	if auth["token"] == "abcdefg" {
		t.Errorf("group 内 token 未被脱敏: %v", auth["token"])
	}
	if auth["user"] != "bob" {
		t.Errorf("group 内非敏感字段被改动: %v", auth["user"])
	}
}

func TestHandlerInjectsRequestID(t *testing.T) {
	buf := &bytes.Buffer{}
	h := newTestHandler(buf, nil)
	logger := slog.New(h)

	ctx := WithRequestID(context.Background(), "req-xyz")
	logger.InfoContext(ctx, "hello")

	got := decode(t, buf.Bytes())
	if got[RequestIDLogKey] != "req-xyz" {
		t.Errorf("request_id 未注入: %v", got[RequestIDLogKey])
	}
}

func TestHandlerNoRequestIDWhenAbsent(t *testing.T) {
	buf := &bytes.Buffer{}
	h := newTestHandler(buf, nil)
	logger := slog.New(h)

	logger.InfoContext(context.Background(), "hello")
	got := decode(t, buf.Bytes())
	if _, exists := got[RequestIDLogKey]; exists {
		t.Errorf("不应当注入 request_id, got %v", got[RequestIDLogKey])
	}
}
