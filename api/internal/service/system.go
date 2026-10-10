package service

import (
	"context"

	"x-foot/internal/model"
)

// ISystem 系统设置业务接口（部门 / 员工 / 用户 / 菜单权限）
type ISystem interface {
	// SeedDefaultMenus 补齐缺失的默认菜单（按 menu_code 幂等）
	SeedDefaultMenus(ctx context.Context) error

	// MenuTree 查询可授权的菜单树（不回填勾选状态）
	MenuTree(ctx context.Context) ([]*model.MenuItem, error)
	// DepartmentMenuTree 查询菜单树，并回填部门已授权的菜单
	DepartmentMenuTree(ctx context.Context, deptId uint64) ([]*model.MenuItem, error)
	// EmployeeMenuTree 查询菜单树，并回填员工已授权的菜单
	//
	// 可选范围限制在所属部门的菜单之内，员工的权限不可能超出部门。
	EmployeeMenuTree(ctx context.Context, employeeId uint64) ([]*model.MenuItem, error)
	// SaveDepartmentMenus 保存部门菜单权限（全量覆盖）
	SaveDepartmentMenus(ctx context.Context, deptId uint64, in *model.MenuIdsInput) error
	// SaveEmployeeMenus 保存员工菜单权限（全量覆盖，且必须落在部门权限内）
	SaveEmployeeMenus(ctx context.Context, employeeId uint64, in *model.MenuIdsInput) error

	// ListDepartments 查询部门树
	ListDepartments(ctx context.Context) ([]*model.DepartmentItem, error)
	// SaveDepartment 新增或编辑部门，返回部门ID
	SaveDepartment(ctx context.Context, in *model.DepartmentSaveInput) (uint64, error)
	// DeleteDepartment 删除部门（存在子部门或员工时拒绝）
	DeleteDepartment(ctx context.Context, deptId uint64) error

	// ListEmployees 分页查询员工
	ListEmployees(ctx context.Context, in *model.EmployeeQuery) (*model.EmployeeListResult, error)
	// SaveEmployee 新增或编辑员工，返回员工ID
	SaveEmployee(ctx context.Context, in *model.EmployeeSaveInput) (uint64, error)
	// DeleteEmployee 删除员工
	DeleteEmployee(ctx context.Context, employeeId uint64) error
	// EmployeeOptions 查询可用于绑定登录账号的员工（未绑定的优先）
	EmployeeOptions(ctx context.Context, deptId uint64) ([]*model.EmployeeOption, error)

	// ListUsers 分页查询登录账号
	ListUsers(ctx context.Context, in *model.UserQuery) (*model.UserListResult, error)
	// SaveUser 新增或编辑登录账号，返回用户ID
	SaveUser(ctx context.Context, in *model.UserSaveInput) (uint64, error)
	// DeleteUser 删除登录账号
	DeleteUser(ctx context.Context, userId uint64) error
	// BindEmployee 登录账号绑定 / 解绑员工（employeeId 为 0 表示解绑）
	BindEmployee(ctx context.Context, in *model.BindEmployeeInput) error
	// ResetUserPassword 重置登录密码
	ResetUserPassword(ctx context.Context, userId uint64, password string) error

	// GetUserMenus 计算某个登录用户最终可用的菜单（已合并内置菜单）
	GetUserMenus(ctx context.Context, userId uint64) ([]*model.MenuItem, error)
}

var localSystem ISystem

// System 获取系统设置业务实例
func System() ISystem {
	if localSystem == nil {
		panic("implement not found for interface ISystem, forgot register?")
	}

	return localSystem
}

// RegisterSystem 注册系统设置业务实现
func RegisterSystem(i ISystem) {
	localSystem = i
}
