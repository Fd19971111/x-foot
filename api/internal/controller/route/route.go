package route

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"x-foot/api/route/v1"
	"x-foot/internal/service"
)

// ControllerV1 前端动态路由控制器
type ControllerV1 struct{}

// NewV1 创建动态路由控制器
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

// GetUserRoutes 当前登录用户可用的路由
func (c *ControllerV1) GetUserRoutes(ctx context.Context, req *v1.GetUserRoutesReq) (res *v1.GetUserRoutesRes, err error) {
	userId := g.RequestFromCtx(ctx).GetCtxVar("userId").Uint64()

	data, err := service.Route().GetUserRoutes(ctx, userId)
	if err != nil {
		return nil, err
	}

	return &v1.GetUserRoutesRes{UserRouteResult: *data}, nil
}

// GetConstantRoutes 无需授权的内置路由
func (c *ControllerV1) GetConstantRoutes(ctx context.Context, req *v1.GetConstantRoutesReq) (res *v1.GetConstantRoutesRes, err error) {
	routes, err := service.Route().GetConstantRoutes(ctx)
	if err != nil {
		return nil, err
	}

	res = (*v1.GetConstantRoutesRes)(&routes)

	return res, nil
}

// IsRouteExist 路由是否属于当前用户授权范围
func (c *ControllerV1) IsRouteExist(ctx context.Context, req *v1.IsRouteExistReq) (res *v1.IsRouteExistRes, err error) {
	userId := g.RequestFromCtx(ctx).GetCtxVar("userId").Uint64()

	exist, err := service.Route().IsRouteExist(ctx, userId, req.RouteName)
	if err != nil {
		return nil, err
	}

	res = (*v1.IsRouteExistRes)(&exist)

	return res, nil
}
