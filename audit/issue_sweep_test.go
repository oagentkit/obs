package audit

// 2026-11-11 issue 清零战回归测试（#83）：钉住顶层 kv 对的键感知打码。

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// TestFlatKVPairCredentialMasked_Issue83 顶层平铺 kv 对的凭证词键值必须整值
// 打码——此前只看值类型不看键名，同一凭证平铺形态原文落盘、嵌套形态反而
// 打码（守卫依赖形态而非语义）。
func TestFlatKVPairCredentialMasked_Issue83(t *testing.T) {
	fields := redactFields([]any{
		"user", "alice",
		"password", "P@ssw0rd-2026",
		"api_key", "sk-1234567890abcdef",
		"db_password", "hunter2secret",
		"payload", map[string]any{"password": "P@ssw0rd-2026"},
		"note", "普通说明不带凭证",
		"body", []byte(`{"password":"in-json"}`),
		"err", errors.New("connect failed: password=x"),
	})
	out := jsonFieldsToString(t, fields)
	for _, leak := range []string{"P@ssw0rd-2026", "sk-1234567890abcdef", "hunter2secret", "in-json"} {
		if strings.Contains(out, leak) {
			t.Errorf("凭证值 %q 不得原文出现在打码结果里: %s", leak, out)
		}
	}
	for _, keep := range []string{"alice", "普通说明不带凭证"} {
		if !strings.Contains(out, keep) {
			t.Errorf("非凭证值 %q 应保留: %s", keep, out)
		}
	}
	// 键本身保持可检索。
	for _, key := range []string{"password", "api_key", "db_password"} {
		if !strings.Contains(out, `"`+key+`"`) {
			t.Errorf("键 %q 应保持可检索", key)
		}
	}
	// JSON 载体在凭证键下也死：整值打码成掩码标记（与嵌套形态同一份），
	// 原文不落盘。
	rawOut := redactFields([]any{"token", json.RawMessage(`{"secret":"raw-credential"}`)})
	switch v := rawOut[1].(type) {
	case json.RawMessage:
		if strings.Contains(string(v), "raw-credential") {
			t.Errorf("RawMessage 载体的凭证键值应整值打码: %s", v)
		}
	case string:
		if strings.Contains(v, "raw-credential") {
			t.Errorf("凭证键下的 RawMessage 不得原文落盘: %q", v)
		}
	default:
		t.Fatalf("凭证键下 RawMessage 应被掩码或保型，实得 %T", rawOut[1])
	}
}

// TestOddLengthFieldsStillRedacted_Issue83 奇数长度字段列表的尾元素走值位
// 兜底，不 panic、凭证文本仍被文本规则打码。
func TestOddLengthFieldsStillRedacted_Issue83(t *testing.T) {
	out := jsonFieldsToString(t, redactFields([]any{"key1", "v1", "password=https://u:p@h/x"}))
	if strings.Contains(out, "p@h") {
		t.Errorf("尾元素 URL 凭证应被文本规则打码: %s", out)
	}
}

// jsonFieldsToString 渲染字段列表为 JSON 文本（与 slog JSON handler 同形态，
// 断言子串）。
func jsonFieldsToString(t *testing.T, fields []any) string {
	t.Helper()
	b, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
