// Package audit 审计日志：对敏感操作留结构化审计痕迹，与业务日志分离。
// 字段值统一经 logredact 打码——审计载荷常含平台 token、连接串、Bearer 头。
package audit

import (
	"encoding/json"

	"github.com/oagentkit/guardx/logredact"
	"github.com/oekit/logx"
)

// Action 审计操作类型（调用方按需扩展）。
type Action string

// Logger 审计日志器。
type Logger struct {
	log logx.Logger
}

// New 创建审计日志器。service 用于日志前缀（如 "argus-audit"）。
// log 为 nil 时回退到 logx.NewAuditLogger（v0.5.1 起门面级承诺：级别锁 Info、
// 不受 LOG_LEVEL 降级、SetLevel 调到 WARN/ERROR 被拒、With* 派生保持守卫）——
// 审计留痕不随业务日志级别配置丢失。
func New(log logx.Logger, service string) *Logger {
	if log == nil {
		log = logx.NewAuditLogger(service)
	}
	return &Logger{log: log.WithService(service)}
}

// Log 记录一条审计事件（nil receiver 安全）。字段值统一经 logredact 打码：
// 字符串过 Redact 文本规则，结构化值（map/slice/struct）递归下探（obs#57——
// 此前只打顶层字符串，header map / 请求体里的嵌套凭据原文落盘）；JSON 文本
// 的字节载荷（[]byte / json.RawMessage）过 RedactJSON（obs#76）；error 值先
// 取文本再走掩码（obs#77）。
func (l *Logger) Log(action Action, fields ...any) {
	if l == nil || l.log == nil {
		return
	}
	args := append([]any{"action", string(action)}, redactFields(fields)...)
	l.log.Info("audit.event", args...)
}

// redactFields 对字段列表打码（键不处理，保持可检索）。输出为副本，入参不动。
// 各形态的打码路径：
//   - string：过 Redact 文本规则。
//   - json.RawMessage / []byte：JSON 文本的字节载荷交 RedactJSON（键感知
//     打码，obs#76）——反射下探豁免字节切片（guardx：二进制非文本），而
//     json.RawMessage 正是请求体在 Go 侧的规范承载类型，[]byte 则以可逆
//     base64 落盘，同一份凭证体换承载类型即漏掩；json.Valid 的 []byte 才
//     走 JSON 路径，非 JSON 二进制原样（base64 落盘，不是检测面）。
//   - error：先 Error() 取文本再过 Redact（obs#77）——标准 error 的文本在
//     unexported 字段里，反射下探够不到，而 slog 渲染时才 err.Error()，
//     克隆 URL 内嵌 token / 连接串原文落盘。
//   - 其余值交 logredact.RedactValue 递归下探：map/slice/struct 容器逐层
//     拷贝，字符串叶子过同一套文本规则，凭证词键（authorization/password/
//     token…）整值打码，普通值保留。
func redactFields(fields []any) []any {
	if len(fields) == 0 {
		return fields
	}
	out := make([]any, len(fields))
	for i, f := range fields {
		switch v := f.(type) {
		case string:
			out[i] = logredact.Redact(v)
		case json.RawMessage:
			out[i] = json.RawMessage(logredact.RedactJSON(v))
		case []byte:
			if json.Valid(v) {
				out[i] = logredact.RedactJSON(v)
				continue
			}
			out[i] = v
		case error:
			out[i] = logredact.Redact(v.Error())
		default:
			out[i] = logredact.RedactValue(f)
		}
	}
	return out
}
