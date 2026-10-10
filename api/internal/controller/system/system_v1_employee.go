package system

import (
	"context"

	"x-foot/api/system/v1"
	"x-foot/internal/service"
)

// ListEmployees 员工分页列表
func (c *ControllerV1) ListEmployees(ctx context.Context, req *v1.ListEmployeesReq) (res *v1.ListEmployeesRes, err error) {
	data, err := service.System().ListEmployees(ctx, &req.EmployeeQuery)
	if err != nil {
		return nil, err
	}

	return &v1.ListEmployeesRes{EmployeeListResult: *data}, nil
}

// SaveEmployee 新增或编辑员工
func (c *ControllerV1) SaveEmployee(ctx context.Context, req *v1.SaveEmployeeReq) (res *v1.SaveEmployeeRes, err error) {
	id, err := service.System().SaveEmployee(ctx, &req.EmployeeSaveInput)
	if err != nil {
		return nil, err
	}

	return &v1.SaveEmployeeRes{Id: id}, nil
}

// DeleteEmployee 删除员工
func (c *ControllerV1) DeleteEmployee(ctx context.Context, req *v1.DeleteEmployeeReq) (res *v1.DeleteEmployeeRes, err error) {
	if err = service.System().DeleteEmployee(ctx, req.Id); err != nil {
		return nil, err
	}

	return &v1.DeleteEmployeeRes{}, nil
}

// GetEmployeeMenus 员工菜单权限树
func (c *ControllerV1) GetEmployeeMenus(ctx context.Context, req *v1.GetEmployeeMenusReq) (res *v1.GetEmployeeMenusRes, err error) {
	list, err := service.System().EmployeeMenuTree(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.GetEmployeeMenusRes{List: list}, nil
}

// SaveEmployeeMenus 保存员工菜单权限
func (c *ControllerV1) SaveEmployeeMenus(ctx context.Context, req *v1.SaveEmployeeMenusReq) (res *v1.SaveEmployeeMenusRes, err error) {
	if err = service.System().SaveEmployeeMenus(ctx, req.Id, &req.MenuIdsInput); err != nil {
		return nil, err
	}

	return &v1.SaveEmployeeMenusRes{}, nil
}

// ListEmployeeOptions 可绑定登录账号的员工列表
func (c *ControllerV1) ListEmployeeOptions(ctx context.Context, req *v1.ListEmployeeOptionsReq) (res *v1.ListEmployeeOptionsRes, err error) {
	list, err := service.System().EmployeeOptions(ctx, req.DeptId)
	if err != nil {
		return nil, err
	}

	return &v1.ListEmployeeOptionsRes{List: list}, nil
}
