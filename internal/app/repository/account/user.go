// Package account 是账户数据访问层（repository / dao）。
//
// 负责直接操作 ent / Redis；service 层调用本层接口。
// 当前仅留占位，作为分层模板。
package account

// UserRepository 封装对 user 表的访问。
type UserRepository struct{}

// NewUserRepository 构造 UserRepository。
func NewUserRepository() *UserRepository { return &UserRepository{} }
