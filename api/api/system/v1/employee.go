package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"x-foot/internal/model"
)

// ListEmployeesReq 员工分页列表
type ListEmployeesReq struct {
	g.Meta `path:"/api/system/employees" method:"get" tags:"System" summary:"员工分页列表"`
	model.EmployeeQuery
}

// ListEmployeesRes 员工分页列表
type ListEmployeesRes struct {
	model.EmployeeListResult
}

// SaveEmployeeReq 新增或编辑员工
type SaveEmployeeReq struct {
	g.Meta `path:"/api/system/employees" method:"post" tags:"System" summary:"新增或编辑员工"`
	model.EmployeeSaveInput
}

// SaveEmployeeRes 新增或编辑员工
type SaveEmployeeRes struct {
	Id uint64 `json:"id"`
}

// DeleteEmployeeReq 删除员工
type DeleteEmployeeReq struct {
	g.Meta `path:"/api/system/employees/:id" method:"delete" tags:"System" summary:"删除员工"`
	Id     uint64 `json:"id"`
}

// DeleteEmployeeRes 删除员工
type DeleteEmployeeRes struct {
	g.Meta `mime:"application/json"`
}

// GetEmployeeMenusReq 员工菜单权限树（可选范围受限于所属部门）
type GetEmployeeMenusReq struct {
	g.Meta `path:"/api/system/employees/:id/menus" method:"get" tags:"System" summary:"员工菜单权限树"`
	Id     uint64 `json:"id"`
}

// GetEmployeeMenusRes 员工菜单权限树
type GetEmployeeMenusRes struct {
	List []*model.MenuItem `json:"list"`
}

// SaveEmployeeMenusReq 保存员工菜单权限
type SaveEmployeeMenusReq struct {
	g.Meta `path:"/api/system/employees/:id/menus" method:"put" tags:"System" summary:"保存员工菜单权限"`
	Id     uint64 `json:"id"`
	model.MenuIdsInput
}

// SaveEmployeeMenusRes 保存员工菜单权限
type SaveEmployeeMenusRes struct {
	g.Meta `mime:"application/json"`
}

// ListEmployeeOptionsReq 可绑定登录账号的员工列表
type ListEmployeeOptionsReq struct {
	g.Meta `path:"/api/system/employeeOptions" method:"get" tags:"System" summary:"可绑定的员工列表"`
	DeptId uint64 `json:"deptId"`
}

// ListEmployeeOptionsRes 可绑定登录账号的员工列表
type ListEmployeeOptionsRes struct {
	List []*model.EmployeeOption `json:"list"`
}
