package router

import "github.com/gin-gonic/gin"

// Module 表示一个业务模块的路由注册函数。
//
// 每个业务域（account / order / billing ...）通过实现一个 Module，把自己的接口
// 挂到给定的 *gin.RouterGroup 上。router 包只负责按顺序调用各 Module，
// 不感知具体业务。这样新增业务域时无需改动 router 包内部，只需要在
// modules() 列表中追加一行。
type Module func(g *gin.RouterGroup)
