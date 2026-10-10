package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"x-foot/internal/model"
)

// ListUsersReq 登录账号分页列表
type ListUsersReq struct {
	g.Meta `path:"/api/system/users" method:"get" tags:"System" summary:"登录账号分页列表"`
	model.UserQuery
}

// ListUsersRes 登录账号分页列表
type ListUsersRes struct {
	model.UserListResult
}

// SaveUserReq 新增或编辑登录账号
type SaveUserReq struct {
	g.Meta `path:"/api/system/users" method:"post" tags:"System" summary:"新增或编辑登录账号"`
	model.UserSaveInput
}

// SaveUserRes 新增或编辑登录账号
type SaveUserRes struct {
	Id uint64 `json:"id"`
}

// DeleteUserReq 删除登录账号
type DeleteUserReq struct {
	g.Meta `path:"/api/system/users/:id" method:"delete" tags:"System" summary:"删除登录账号"`
	Id     uint64 `json:"id"`
}

// DeleteUserRes 删除登录账号
type DeleteUserRes struct {
	g.Meta `mime:"application/json"`
}

// BindEmployeeReq 登录账号绑定员工（employeeId 为 0 表示解绑）
type BindEmployeeReq struct {
	g.Meta `path:"/api/system/users/:id/bindEmployee" method:"put" tags:"System" summary:"登录账号绑定员工"`
	Id     uint64 `json:"id"`
	model.BindEmployeeInput
}

// BindEmployeeRes 登录账号绑定员工
type BindEmployeeRes struct {
	g.Meta `mime:"application/json"`
}

// ResetUserPasswordReq 重置登录密码
type ResetUserPasswordReq struct {
	g.Meta   `path:"/api/system/users/:id/password" method:"put" tags:"System" summary:"重置登录密码"`
	Id       uint64 `json:"id"`
	Password string `json:"password"`
}

// ResetUserPasswordRes 重置登录密码
type ResetUserPasswordRes struct {
	g.Meta `mime:"application/json"`
}
