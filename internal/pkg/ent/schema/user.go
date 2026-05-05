// Package schema 存放 ent 实体定义。新增表时在本目录新增一个文件，跑 `make ent` 生成代码。
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User 是用户表的 schema，仅作为示例。
type User struct {
	ent.Schema
}

// Fields 定义 User 的字段。
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(64),
		field.String("email").Unique().MaxLen(128),
		field.String("phone").Optional().MaxLen(32),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges 定义 User 与其他实体的关系。当前没有。
func (User) Edges() []ent.Edge {
	return nil
}
