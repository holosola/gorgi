// Package app 是应用装配层。modules.go 列出所有挂到 /api 下的业务模块，
// 新增业务域只需在 modules() 列表里追加一行。
package app

import (
	"github.com/holosola/gorgi/internal/app/api/account"
	"github.com/holosola/gorgi/internal/app/router"
)

// modules 返回挂到 /api 下的业务模块。
// router 包不感知具体业务，模块列表由本层组装并注入。
func modules(deps router.Deps) []router.Module {
	return []router.Module{
		account.Routes(deps),
		// order.Routes(deps),
		// billing.Routes(deps),
	}
}
