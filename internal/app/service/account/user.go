// Package account 是账户业务编排层（service）。
//
// 在 handler 与 repository 之间充当编排者：组合多个 repository 调用 / 处理事务 / 业务规则。
// 当前仅留占位，作为分层模板。
package account

// UserService 编排账户相关的业务逻辑。
type UserService struct{}

// NewUserService 构造 UserService。
func NewUserService() *UserService { return &UserService{} }
