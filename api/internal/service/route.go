package service

import (
	"context"

	"x-foot/internal/model"
)

// IRoute 前端动态路由业务接口
type IRoute interface {
	// GetUserRoutes 取当前登录用户可用的路由树与默认落地路由
	GetUserRoutes(ctx context.Context, userId uint64) (*model.UserRouteResult, error)
	// GetConstantRoutes 取无需授权的内置路由（登录页、错误页等）
	GetConstantRoutes(ctx context.Context) ([]*model.RouteNode, error)
	// IsRouteExist 判断某个路由是否属于当前用户的授权范围
	IsRouteExist(ctx context.Context, userId uint64, routeName string) (bool, error)
}

var localRoute IRoute

// Route 获取动态路由业务实例
func Route() IRoute {
	if localRoute == nil {
		panic("implement not found for interface IRoute, forgot register?")
	}

	return localRoute
}

// RegisterRoute 注册动态路由业务实现
func RegisterRoute(i IRoute) {
	localRoute = i
}
