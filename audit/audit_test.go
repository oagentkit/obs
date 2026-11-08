package audit

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/oekit/logx"
)

// TestFallbackLoggerImmuneToLogLevel 审计兜底 logger 不随 LOG_LEVEL 丢留痕：
// logx v0.3.0 起 NewSlogLogger 尊重 LOG_LEVEL 环境变量，生产常见 LOG_LEVEL=error
// 会吞掉 Info 级事件——New(nil, ...) 回退 logx.NewAuditLogger（级别锁 Info，v0.5.1
// 起门面承诺）。对照组证明裸 NewSlogLogger 确实被吞、守卫是必要的。
func TestFallbackLoggerImmuneToLogLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "error")
	out := captureStdout(t, func() {
		New(nil, "test-audit").Log("task.create", "task_id", "t1")
	})
	if !strings.Contains(out, "audit.event") {
		t.Fatalf("LOG_LEVEL=error 下审计事件应仍落盘，实际输出: %q", out)
	}
	ctl := captureStdout(t, func() {
		logx.NewSlogLogger("ctl").Info("plain.info")
	})
	if strings.Contains(ctl, "plain.info") {
		t.Fatalf("对照组：LOG_LEVEL=error 下未钉级别的裸 logger 应吞 Info，实际输出: %q", ctl)
	}
}

// captureStdout 截获 fn 执行期间写入 os.Stdout 的内容（logx slog 后端写时取 os.Stdout）。
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	fn()
	os.Stdout = orig
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}

func TestLoggerLog(t *testing.T) {
	log := logx.NewSlogLogger("test")
	logger := New(log, "test-audit")

	// 不应 panic
	logger.Log(Action("task.create"), "task_id", "t1", "platform", "gitea")
	logger.Log(Action("finding.suppress"), "fp", "abc123")
}

func TestLoggerNilSafe(t *testing.T) {
	// nil receiver 与 nil logger 都不应 panic（审计留痕缺失好过进程崩溃）
	var l *Logger
	l.Log("test.action", "k", "v")

	logger := New(nil, "test-audit") // nil logger → 回退缺省 logger
	if logger == nil {
		t.Fatal("New 应返回非 nil")
	}
	logger.Log("test.action", "k", "v")
}

func TestActionConstants(t *testing.T) {
	// Action 类型是 string 的别名
	var a Action = "test.action"
	if string(a) != "test.action" {
		t.Fatal("Action 应可转为 string")
	}
}

// 凭据字段值必须打码，普通业务字段/非字符串保留。
func TestRedactFields(t *testing.T) {
	in := []any{
		"url", "postgres://user:pw@db/x",
		"auth", "Bearer sk-abc123secret",
		"n", 42,
		"ok", "plain",
		"pem", "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADAN\n-----END PRIVATE KEY-----",
	}
	out := redactFields(in)
	if out[1] != "postgres://user:****@db/x" {
		t.Fatalf("URL 凭据未打码: %v", out[1])
	}
	if s, _ := out[3].(string); s == "" || s == "Bearer sk-abc123secret" {
		t.Fatalf("Bearer 未打码: %v", out[3])
	}
	if out[5] != 42 {
		t.Fatalf("非字符串字段被改动: %v", out[5])
	}
	if out[7] != "plain" {
		t.Fatalf("普通串被误伤: %v", out[7])
	}
	if s, _ := out[9].(string); len(s) == 0 || s == in[9] {
		t.Fatalf("PEM 私钥未打码: %v", out[9])
	}
}

// 结构化字段（map/slice/struct）必须同样下探打码：审计载荷常整传 header
// map / 请求体，只打顶层字符串会让嵌套凭据原文落盘（obs#57）。路径键形态
// 与顶层一致——字符串叶子过文本规则、凭证词键整值打码、普通值保留。
func TestRedactFieldsStructured(t *testing.T) {
	type req struct {
		Password string
		Note     string
	}
	in := []any{
		"payload", map[string]any{
			"authorization": "Bearer sk-secret-token-9999",
			"user":          "alice",
		},
		"headers", map[string]string{
			"x-api-key": "sk-nested-api-key-123456",
			"x-trace":   "trace-1",
		},
		"items", []any{"postgres://user:pw@db/x", "ok"},
		"typed", []string{"webhook_token=sk-slice-token-987654", "plain"},
		"req", req{Password: "super-secret-pass-123", Note: "hello"},
	}
	out := redactFields(in)

	// map[string]any：凭证键整值打码，普通键保留
	m, ok := out[1].(map[string]any)
	if !ok {
		t.Fatalf("map 字段类型应保持, 实际 %T", out[1])
	}
	if m["authorization"] != "****" {
		t.Fatalf("嵌套 authorization 未整值打码: %v", m["authorization"])
	}
	if m["user"] != "alice" {
		t.Fatalf("嵌套普通值被误伤: %v", m["user"])
	}

	// map[string]string（类型化容器，反射下探）
	h, ok := out[3].(map[string]string)
	if !ok {
		t.Fatalf("类型化 map 字段类型应保持, 实际 %T", out[3])
	}
	if h["x-api-key"] != "****" {
		t.Fatalf("类型化 map 的 api-key 未打码: %v", h["x-api-key"])
	}
	if h["x-trace"] != "trace-1" {
		t.Fatalf("类型化 map 普通值被误伤: %v", h["x-trace"])
	}

	// []any：URL 内嵌凭据按文本规则打码
	items, ok := out[5].([]any)
	if !ok {
		t.Fatalf("slice 字段类型应保持, 实际 %T", out[5])
	}
	if items[0] != "postgres://user:****@db/x" {
		t.Fatalf("嵌套 URL 凭据未打码: %v", items[0])
	}
	if items[1] != "ok" {
		t.Fatalf("嵌套普通串被误伤: %v", items[1])
	}

	// []string（类型化 slice）：kv 文本规则下探，密文不得原文出现
	typed, ok := out[7].([]string)
	if !ok {
		t.Fatalf("类型化 slice 字段类型应保持, 实际 %T", out[7])
	}
	if strings.Contains(typed[0], "sk-slice-token-987654") {
		t.Fatalf("类型化 slice 内凭据未打码: %v", typed[0])
	}
	if typed[1] != "plain" {
		t.Fatalf("类型化 slice 普通串被误伤: %v", typed[1])
	}

	// struct 导出字段：字段名过键名判定整值打码，普通字段保留
	r, ok := out[9].(req)
	if !ok {
		t.Fatalf("struct 字段类型应保持, 实际 %T", out[9])
	}
	if r.Password != "****" {
		t.Fatalf("struct Password 字段未打码: %v", r.Password)
	}
	if r.Note != "hello" {
		t.Fatalf("struct 普通字段被误伤: %v", r.Note)
	}

	// 入参不得被原地污染（输出为副本）
	src := in[1].(map[string]any)
	if src["authorization"] != "Bearer sk-secret-token-9999" {
		t.Fatalf("入参 map 被原地修改: %v", src["authorization"])
	}
}

// 端到端（issue#57 复现形态）：Log 的结构化字段打码后落盘，嵌套密文不得
// 原文进入审计日志；顶层字符串路径行为不变。
func TestLoggerLogStructuredSecret(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := New(logx.NewSlogLogger("test", logx.WithHandler(h)), "test-audit")
	logger.Log("api.call",
		"bearer", "Bearer sk-top-level-token-0000",
		"payload", map[string]any{"authorization": "Bearer sk-secret-token-9999", "user": "alice"},
	)
	out := buf.String()
	if out == "" {
		t.Fatal("审计事件未落盘")
	}
	if strings.Contains(out, "sk-secret-token-9999") {
		t.Fatalf("嵌套密文原文落盘（obs#57 缺陷）: %s", out)
	}
	if strings.Contains(out, "sk-top-level-token-0000") {
		t.Fatalf("顶层密文原文落盘: %s", out)
	}
	if !strings.Contains(out, `"user":"alice"`) && !strings.Contains(out, `"user": "alice"`) {
		t.Fatalf("嵌套普通值应保留以便检索: %s", out)
	}
}

// Log 路径不 panic 且会走 redactFields（字符串含 token= 时被替换）。
func TestLoggerLogWithSecret(t *testing.T) {
	logger := New(logx.NewSlogLogger("test"), "test-audit")
	logger.Log(Action("qa.reply"), "question", "见 token=sk-should-not-appear", "task_id", "t-1")
}
