package system

import (
	"context"

	"x-foot/api/system/v1"
	"x-foot/internal/service"
)

// ListDepartments 部门树
func (c *ControllerV1) ListDepartments(ctx context.Context, req *v1.ListDepartmentsReq) (res *v1.ListDepartmentsRes, err error) {
	list, err := service.System().ListDepartments(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.ListDepartmentsRes{List: list}, nil
}

// SaveDepartment 新增或编辑部门
func (c *ControllerV1) SaveDepartment(ctx context.Context, req *v1.SaveDepartmentReq) (res *v1.SaveDepartmentRes, err error) {
	id, err := service.System().SaveDepartment(ctx, &req.DepartmentSaveInput)
	if err != nil {
		return nil, err
	}

	return &v1.SaveDepartmentRes{Id: id}, nil
}

// DeleteDepartment 删除部门
func (c *ControllerV1) DeleteDepartment(ctx context.Context, req *v1.DeleteDepartmentReq) (res *v1.DeleteDepartmentRes, err error) {
	if err = service.System().DeleteDepartment(ctx, req.Id); err != nil {
		return nil, err
	}

	return &v1.DeleteDepartmentRes{}, nil
}

// GetDepartmentMenus 部门菜单权限树
func (c *ControllerV1) GetDepartmentMenus(ctx context.Context, req *v1.GetDepartmentMenusReq) (res *v1.GetDepartmentMenusRes, err error) {
	list, err := service.System().DepartmentMenuTree(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.GetDepartmentMenusRes{List: list}, nil
}

// SaveDepartmentMenus 保存部门菜单权限
func (c *ControllerV1) SaveDepartmentMenus(ctx context.Context, req *v1.SaveDepartmentMenusReq) (res *v1.SaveDepartmentMenusRes, err error) {
	if err = service.System().SaveDepartmentMenus(ctx, req.Id, &req.MenuIdsInput); err != nil {
		return nil, err
	}

	return &v1.SaveDepartmentMenusRes{}, nil
}
