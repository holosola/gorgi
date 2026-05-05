// Package ent 通过 go:generate 触发 ent 代码生成。
//
// 在添加 schema 后，请执行：
//
//	make ent
//
// 等价于：
//
//	go generate ./internal/pkg/ent/...
package ent

//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./schema
