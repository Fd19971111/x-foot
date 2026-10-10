// Package route 前端动态路由
//
// 路由不再写死在前端，而是由这里根据当前登录用户的菜单权限下发：
//   - GetUserRoutes     —— 用户可见的菜单路由（决定了侧边栏与可访问页面）
//   - GetConstantRoutes —— 无需授权的内置路由（登录页、错误页等）
//   - IsRouteExist      —— 判断某个路由是否属于当前用户的授权范围
//
// 路由结构对齐前端 elegant-router：
//   - 一级且有子节点：Component = "layout.base"，子级 Component = "view.xxx"
//   - 一级且无子节点：Component = "layout.base$view.xxx"
// 其中 layout/view 的名称来自前端自动生成的路由表，两者必须保持一致。
package route

import (
	"context"

	"github.com/gogf/gf/v2/util/gconv"

	"x-foot/internal/model"
	"x-foot/internal/service"
)

const (
	// homeRouteKey 登录后默认落地路由
	homeRouteKey = "home"
	// initRouteKey 公司初始化配置页，使用空白布局且不出现在菜单里
	initRouteKey = "init"

	layoutBase  = "layout.base"
	layoutBlank = "layout.blank"
)

type sRoute struct{}

func init() {
	service.RegisterRoute(New())
}

// New 创建动态路由业务实例
func New() service.IRoute {
	return &sRoute{}
}

// GetUserRoutes 取当前登录用户可用的路由树与默认落地路由
func (s *sRoute) GetUserRoutes(ctx context.Context, userId uint64) (*model.UserRouteResult, error) {
	menus, err := service.System().GetUserMenus(ctx, userId)
	if err != nil {
		return nil, err
	}

	menus = routableMenus(menus)

	routes := make([]*model.RouteNode, 0, len(menus))
	for _, menu := range menus {
		routes = append(routes, buildNode(menu))
	}

	return &model.UserRouteResult{
		Routes: routes,
		Home:   homeRouteKey,
	}, nil
}

// GetConstantRoutes 取无需授权的内置路由
//
// 与前端 src/router/elegant/routes.ts 中 meta.constant 为 true 的路由保持一致，
// 前端新增/修改这类路由时，这里需要同步。
func (s *sRoute) GetConstantRoutes(ctx context.Context) ([]*model.RouteNode, error) {
	return []*model.RouteNode{
		{
			Id:        "403",
			Name:      "403",
			Path:      "/403",
			Component: layoutBlank + "$view.403",
			Meta:      constantMeta("403"),
		},
		{
			Id:        "404",
			Name:      "404",
			Path:      "/404",
			Component: layoutBlank + "$view.404",
			Meta:      constantMeta("404"),
		},
		{
			Id:        "500",
			Name:      "500",
			Path:      "/500",
			Component: layoutBlank + "$view.500",
			Meta:      constantMeta("500"),
		},
		{
			Id:        "iframe-page",
			Name:      "iframe-page",
			Path:      "/iframe-page/:url",
			Component: layoutBase + "$view.iframe-page",
			Meta: &model.RouteMeta{
				Title:      "iframe-page",
				I18nKey:    "route.iframe-page",
				Constant:   true,
				HideInMenu: true,
				KeepAlive:  true,
			},
		},
		{
			Id:        "login",
			Name:      "login",
			Path:      "/login/:module(pwd-login|code-login|register|reset-pwd|bind-wechat)?",
			Component: layoutBlank + "$view.login",
			Meta:      constantMeta("login"),
		},
	}, nil
}

// IsRouteExist 判断某个路由是否属于当前用户的授权范围
func (s *sRoute) IsRouteExist(ctx context.Context, userId uint64, routeName string) (bool, error) {
	menus, err := service.System().GetUserMenus(ctx, userId)
	if err != nil {
		return false, err
	}

	return containsMenu(routableMenus(menus), routeName), nil
}

// ---------------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------------

// routableMenus 剔除不能生成前端路由的节点
//
//   - button：按钮级权限点只用于鉴权，前端没有对应页面
//   - 空目录：目录下没有任何可访问页面时不会生成路由，下发出去只会变成死链
//
// 返回的是副本，不会改动调用方拿到的菜单树。
func routableMenus(items []*model.MenuItem) []*model.MenuItem {
	result := make([]*model.MenuItem, 0, len(items))

	for _, item := range items {
		if item.MenuType == model.MenuTypeButton {
			continue
		}

		node := *item
		node.Children = routableMenus(item.Children)

		if item.MenuType == model.MenuTypeCatalog && len(node.Children) == 0 {
			continue
		}

		result = append(result, &node)
	}

	return result
}

// buildNode 把一个菜单节点翻译成前端路由节点
func buildNode(menu *model.MenuItem) *model.RouteNode {
	node := &model.RouteNode{
		Id:   gconv.String(menu.Id),
		Name: menu.MenuCode,
		Path: menu.Path,
		Meta: &model.RouteMeta{
			Title:   menu.MenuName,
			I18nKey: "route." + menu.MenuCode,
			Icon:    menu.Icon,
			Order:   menu.SortOrder,
		},
	}

	// 初始化配置页使用空白布局，且不在菜单里出现
	if menu.MenuCode == initRouteKey {
		node.Component = layoutBlank + "$view." + menu.MenuCode
		node.Meta.HideInMenu = true

		return node
	}

	// 有子节点的一级菜单：自身只作为目录容器，路由节点由子级承载
	if len(menu.Children) > 0 {
		node.Component = layoutBase
		node.Redirect = menu.Children[0].Path

		node.Children = make([]*model.RouteNode, 0, len(menu.Children))
		for _, child := range menu.Children {
			childNode := buildNode(child)
			childNode.Component = "view." + child.MenuCode
			node.Children = append(node.Children, childNode)
		}

		return node
	}

	node.Component = layoutBase + "$view." + menu.MenuCode

	return node
}

// constantMeta 内置路由的通用元信息
func constantMeta(key string) *model.RouteMeta {
	return &model.RouteMeta{
		Title:      key,
		I18nKey:    "route." + key,
		Constant:   true,
		HideInMenu: true,
	}
}

// containsMenu 递归查找菜单编码是否存在
func containsMenu(menus []*model.MenuItem, menuCode string) bool {
	for _, menu := range menus {
		if menu.MenuCode == menuCode {
			return true
		}

		if len(menu.Children) > 0 && containsMenu(menu.Children, menuCode) {
			return true
		}
	}

	return false
}
