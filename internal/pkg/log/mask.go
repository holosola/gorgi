package log

import (
	"log/slog"
	"strings"
)

// masker 负责把命中字段名的值替换为脱敏后的形式。
// 字段名的匹配是大小写不敏感的，匹配集合在初始化时确定。
type masker struct {
	fields map[string]struct{}
}

// newMasker 构造 masker；列表为空返回 nil 表示不做脱敏。
func newMasker(fields []string) *masker {
	if len(fields) == 0 {
		return nil
	}
	m := &masker{fields: make(map[string]struct{}, len(fields))}
	for _, f := range fields {
		f = strings.ToLower(strings.TrimSpace(f))
		if f != "" {
			m.fields[f] = struct{}{}
		}
	}
	if len(m.fields) == 0 {
		return nil
	}
	return m
}

// maskRecord 返回一条所有字段都已检查并按需脱敏的新 Record。
func (m *masker) maskRecord(r slog.Record) slog.Record {
	out := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		out.AddAttrs(m.maskAttr(a))
		return true
	})
	return out
}

// maskAttr 检查并按需脱敏单个属性，自动递归处理 group。
func (m *masker) maskAttr(a slog.Attr) slog.Attr {
	if a.Value.Kind() == slog.KindGroup {
		gs := a.Value.Group()
		out := make([]slog.Attr, 0, len(gs))
		for _, g := range gs {
			out = append(out, m.maskAttr(g))
		}
		return slog.Attr{Key: a.Key, Value: slog.GroupValue(out...)}
	}
	if _, hit := m.fields[strings.ToLower(a.Key)]; hit {
		a.Value = slog.StringValue(maskString(a.Value.String()))
	}
	return a
}

// maskString 按统一规则脱敏字符串：
//   - 空串保持空串；
//   - 长度 <=2：全部用 * 代替；
//   - 长度  >2：保留首尾各一个字符，中间用 * 填充。
func maskString(s string) string {
	r := []rune(s)
	switch n := len(r); {
	case n == 0:
		return ""
	case n <= 2:
		return strings.Repeat("*", n)
	default:
		return string(r[0]) + strings.Repeat("*", n-2) + string(r[n-1])
	}
}
