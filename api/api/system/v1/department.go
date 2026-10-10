package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"x-foot/internal/model"
)

// ListDepartmentsReq 部门树
type ListDepartmentsReq struct {
	g.Meta `path:"/api/system/departments" method:"get" tags:"System" summary:"部门树"`
}

// ListDepartmentsRes 部门树
type ListDepartmentsRes struct {
	List []*model.DepartmentItem `json:"list"`
}

// SaveDepartmentReq 新增或编辑部门
type SaveDepartmentReq struct {
	g.Meta `path:"/api/system/departments" method:"post" tags:"System" summary:"新增或编辑部门"`
	model.DepartmentSaveInput
}

// SaveDepartmentRes 新增或编辑部门
type SaveDepartmentRes struct {
	Id uint64 `json:"id"`
}

// DeleteDepartmentReq 删除部门
type DeleteDepartmentReq struct {
	g.Meta `path:"/api/system/departments/:id" method:"delete" tags:"System" summary:"删除部门"`
	Id     uint64 `json:"id"`
}

// DeleteDepartmentRes 删除部门
type DeleteDepartmentRes struct {
	g.Meta `mime:"application/json"`
}

// GetDepartmentMenusReq 部门菜单权限树
type GetDepartmentMenusReq struct {
	g.Meta `path:"/api/system/departments/:id/menus" method:"get" tags:"System" summary:"部门菜单权限树"`
	Id     uint64 `json:"id"`
}

// GetDepartmentMenusRes 部门菜单权限树
type GetDepartmentMenusRes struct {
	List []*model.MenuItem `json:"list"`
}

// SaveDepartmentMenusReq 保存部门菜单权限
type SaveDepartmentMenusReq struct {
	g.Meta `path:"/api/system/departments/:id/menus" method:"put" tags:"System" summary:"保存部门菜单权限"`
	Id     uint64 `json:"id"`
	model.MenuIdsInput
}

// SaveDepartmentMenusRes 保存部门菜单权限
type SaveDepartmentMenusRes struct {
	g.Meta `mime:"application/json"`
}
