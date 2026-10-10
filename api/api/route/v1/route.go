package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"x-foot/internal/model"
)

// GetUserRoutesReq 当前登录用户的动态路由
type GetUserRoutesReq struct {
	g.Meta `path:"/route/getUserRoutes" method:"get" tags:"Route" summary:"当前用户路由"`
}

// GetUserRoutesRes 当前登录用户的动态路由
type GetUserRoutesRes struct {
	model.UserRouteResult
}

// GetConstantRoutesReq 无需授权的内置路由
type GetConstantRoutesReq struct {
	g.Meta `path:"/route/getConstantRoutes" method:"get" tags:"Route" summary:"内置常量路由"`
}

// GetConstantRoutesRes 无需授权的内置路由（直接返回路由数组）
type GetConstantRoutesRes []*model.RouteNode

// IsRouteExistReq 路由是否属于当前用户授权范围
type IsRouteExistReq struct {
	g.Meta    `path:"/route/isRouteExist" method:"get" tags:"Route" summary:"路由是否存在"`
	RouteName string `json:"routeName"`
}

// IsRouteExistRes 路由是否属于当前用户授权范围（直接返回布尔值）
type IsRouteExistRes bool
