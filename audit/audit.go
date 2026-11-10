// Package audit 审计日志：对敏感操作留结构化审计痕迹，与业务日志分离。
// 字段值统一经 logredact 打码——审计载荷常含平台 token、连接串、Bearer 头。
package audit

import (
	"encoding/json"
	"unicode/utf8"

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
//     token…）整值打码，普通值保留。嵌套层里的 []byte 与 error 叶子同语义
//     （obs#88：guardx redactReflect 单点收口，需 guardx v0.24.3+ 地板）。
//
// 键感知（obs#83）：fields 是 logx 风格的交替 kv 序列，值位打码必须参考
// **同对的键名**——此前逐元素处理只看值类型：顶层平铺的 `"password",
// "P@ssw0rd-2026"` 原文落盘，同一凭证嵌一层 map（"payload", map{password…}）
// 反而打码——守卫依赖载荷形态而非字段语义，包头承诺「字段值统一经 logredact
// 打码」在最自然的调用形态上落空。kv 对的值经单键 map 走 RedactValue 的键名
// 判定单源（凭证词表/复合键/复数形态全在 logredact 内部，本包不复制）。
// JSON 文本/字节/error 三类载体不能整体走 walk（[]byte 被当二进制原样、
// RawMessage 丢 JSON 路径），凭证键命中时整值打码、未命中走各自类型化路径。
func redactFields(fields []any) []any {
	if len(fields) == 0 {
		return fields
	}
	out := make([]any, len(fields))
	for i := 0; i < len(fields); i += 2 {
		// 尾部悬挂元素（奇数长度）：没有配对键，按值位兜底路径处理。
		if i+1 >= len(fields) {
			out[i] = redactValueField(fields[i])
			break
		}
		out[i] = fields[i] // 键不处理，保持可检索
		key, isKey := fields[i].(string)
		if !isKey {
			out[i+1] = redactValueField(fields[i+1])
			continue
		}
		switch fields[i+1].(type) {
		case json.RawMessage, []byte, error:
			// 特殊载体：凭证键命中即整值打码（掩码标记与嵌套形态同一份），
			// 未命中走各自的类型化路径（RedactJSON / 原样 / err 文本）。
			if masked, ok := credentialValueMask(key); ok {
				out[i+1] = masked
			} else {
				out[i+1] = redactValueField(fields[i+1])
			}
		default:
			// 常规值：单键 map 走 RedactValue 键名判定——凭证词键整值打码；
			// 非凭证键的输出与旧路径等价（string→RedactWith、容器→递归下探）。
			if m, ok := logredact.RedactValue(map[string]any{key: fields[i+1]}).(map[string]any); ok {
				if mv, present := m[key]; present {
					out[i+1] = mv
					continue
				}
			}
			out[i+1] = redactValueField(fields[i+1])
		}
	}
	return out
}

// credentialValueMask 键名是否命中 logredact 的凭证词表（哨兵探测）：词表与
// 复合键/复数形态判定都在 logredact 内部（未导出），经公开 API RedactValue
// 用哨兵值探测——凭证词键整值打码（哨兵被替换）、普通键原样透传（文本规则
// 不命中哨兵串）。审计不是热路径，每次探测一次单键 map 分配可接受。
func credentialValueMask(key string) (any, bool) {
	const sentinel = "audit-sentinel-value"
	m, ok := logredact.RedactValue(map[string]any{key: sentinel}).(map[string]any)
	if !ok {
		return nil, false
	}
	if v, present := m[key]; present && v != sentinel {
		return v, true
	}
	return nil, false
}

// redactValueField 单值位的形态开关（无键可依的兜底路径——奇数长度尾元素、
// 非字符串键位的值）：string 过 Redact 文本规则；JSON 文本的字节载荷交
// RedactJSON（obs#76）；error 先取文本再过 Redact（obs#77）；其余交
// RedactValue 递归下探。
func redactValueField(f any) any {
	switch v := f.(type) {
	case string:
		return logredact.Redact(v)
	case json.RawMessage:
		// nil 短路（obs#87）：RedactJSON(nil) 产出空非 nil 切片，重铸回
		// RawMessage 后 MarshalJSON 从 "null" 变空字节——slog JSON 渲染炸
		// unexpected end of JSON input。空 body 请求以 RawMessage(nil) 入
		// 审计是常态。
		if v == nil {
			return v
		}
		return json.RawMessage(logredact.RedactJSON(v))
	case []byte:
		// 非 JSON 分支收窄（obs#89）：#76 的豁免条款是「非 JSON 二进制」，
		// json.Valid 预检把非 JSON 一切（UTF-8 文本：PEM/表单体/token 串）
		// 都豁免了——base64 可逆落盘，同一载荷换承载类型即绕开打码。收口：
		// UTF-8（JSON→键感知 / 非 JSON→RedactJSON 自带文本回落）；
		// 非 UTF-8 才是真二进制、检测面为空、原样。
		if !utf8.Valid(v) {
			return v
		}
		return logredact.RedactJSON(v)
	case error:
		return logredact.Redact(v.Error())
	default:
		return logredact.RedactValue(f)
	}
}
