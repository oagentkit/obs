package audit

// 2026-11-11 第二轮清零回归（#87 #89）。

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestNilRawMessageSurvivesRedaction_Issue87 nil RawMessage 打码后仍是 nil
// ——MarshalJSON 恒 "null"，slog JSON 渲染不炸。
func TestNilRawMessageSurvivesRedaction_Issue87(t *testing.T) {
	out := redactValueField(json.RawMessage(nil))
	rm, ok := out.(json.RawMessage)
	if !ok {
		t.Fatalf("nil RawMessage 应保型返回，实得 %T", out)
	}
	if rm != nil {
		t.Fatalf("nil 应保持 nil（空非 nil 会让 MarshalJSON 产出空字节）: %#v", rm)
	}
	b, err := rm.MarshalJSON()
	if err != nil || string(b) != "null" {
		t.Fatalf("MarshalJSON 应为 null: %q %v", b, err)
	}
}

// TestNonJSONBytesTextStillRedacted_Issue89 UTF-8 非 JSON 文本载荷（表单体/
// PEM/token 串）必须过文本规则——不再借 []byte 承载绕开打码；非 UTF-8 二进制
// 仍原样。
func TestNonJSONBytesTextStillRedacted_Issue89(t *testing.T) {
	out := redactValueField([]byte("token=ghp_formsecret&user=1"))
	if bs, ok := out.([]byte); !ok || strings.Contains(string(bs), "ghp_formsecret") {
		t.Fatalf("UTF-8 文本 []byte 应过文本规则: %#v", out)
	}
	if out := redactValueField([]byte{0x00, 0xff, 0xfe, 0xfd}); len(out.([]byte)) != 4 {
		t.Fatalf("非 UTF-8 二进制应原样: %#v", out)
	}
	if out := redactValueField([]byte(`{"password":"x"}`)); strings.Contains(string(out.([]byte)), `"x"`) {
		t.Fatalf("JSON []byte 键感知打码漂移: %#v", out)
	}
}
