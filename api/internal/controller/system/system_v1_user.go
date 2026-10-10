package system

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"x-foot/api/system/v1"
	"x-foot/internal/model"
	"x-foot/internal/service"
)

// ListUsers 登录账号分页列表
func (c *ControllerV1) ListUsers(ctx context.Context, req *v1.ListUsersReq) (res *v1.ListUsersRes, err error) {
	data, err := service.System().ListUsers(ctx, &req.UserQuery)
	if err != nil {
		return nil, err
	}

	return &v1.ListUsersRes{UserListResult: *data}, nil
}

// SaveUser 新增或编辑登录账号
func (c *ControllerV1) SaveUser(ctx context.Context, req *v1.SaveUserReq) (res *v1.SaveUserRes, err error) {
	id, err := service.System().SaveUser(ctx, &req.UserSaveInput)
	if err != nil {
		return nil, err
	}

	return &v1.SaveUserRes{Id: id}, nil
}

// DeleteUser 删除登录账号
func (c *ControllerV1) DeleteUser(ctx context.Context, req *v1.DeleteUserReq) (res *v1.DeleteUserRes, err error) {
	if err = service.System().DeleteUser(ctx, req.Id); err != nil {
		return nil, err
	}

	return &v1.DeleteUserRes{}, nil
}

// BindEmployee 登录账号绑定 / 解绑员工
func (c *ControllerV1) BindEmployee(ctx context.Context, req *v1.BindEmployeeReq) (res *v1.BindEmployeeRes, err error) {
	input := &model.BindEmployeeInput{
		UserId:     req.Id,
		EmployeeId: req.EmployeeId,
	}

	if err = service.System().BindEmployee(ctx, input); err != nil {
		return nil, err
	}

	return &v1.BindEmployeeRes{}, nil
}

// ResetUserPassword 重置登录密码
func (c *ControllerV1) ResetUserPassword(ctx context.Context, req *v1.ResetUserPasswordReq) (res *v1.ResetUserPasswordRes, err error) {
	if err = service.System().ResetUserPassword(ctx, req.Id, req.Password); err != nil {
		return nil, err
	}

	return &v1.ResetUserPasswordRes{}, nil
}

// GetMenuTree 全部菜单树
func (c *ControllerV1) GetMenuTree(ctx context.Context, req *v1.GetMenuTreeReq) (res *v1.GetMenuTreeRes, err error) {
	list, err := service.System().MenuTree(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.GetMenuTreeRes{List: list}, nil
}

// GetMyMenuTree 当前登录用户可用的菜单树
func (c *ControllerV1) GetMyMenuTree(ctx context.Context, req *v1.GetMyMenuTreeReq) (res *v1.GetMyMenuTreeRes, err error) {
	userId := g.RequestFromCtx(ctx).GetCtxVar("userId").Uint64()

	list, err := service.System().GetUserMenus(ctx, userId)
	if err != nil {
		return nil, err
	}

	return &v1.GetMyMenuTreeRes{List: list}, nil
}
