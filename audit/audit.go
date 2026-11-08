// Package audit 审计日志：对敏感操作留结构化审计痕迹，与业务日志分离。
// 字段值统一经 logredact 打码——审计载荷常含平台 token、连接串、Bearer 头。
package audit

import (
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
// 此前只打顶层字符串，header map / 请求体里的嵌套凭据原文落盘）。
func (l *Logger) Log(action Action, fields ...any) {
	if l == nil || l.log == nil {
		return
	}
	args := append([]any{"action", string(action)}, redactFields(fields)...)
	l.log.Info("audit.event", args...)
}

// redactFields 对字段列表打码（键不处理，保持可检索）：字符串值过 Redact
// 文本规则；非字符串值交 logredact.RedactValue 递归下探——map/slice/struct
// 容器逐层拷贝，字符串叶子过同一套文本规则，凭证词键（authorization/
// password/token…）整值打码，普通值保留。输出为副本，入参不动。
func redactFields(fields []any) []any {
	if len(fields) == 0 {
		return fields
	}
	out := make([]any, len(fields))
	for i, f := range fields {
		if s, ok := f.(string); ok {
			out[i] = logredact.Redact(s)
			continue
		}
		out[i] = logredact.RedactValue(f)
	}
	return out
}
