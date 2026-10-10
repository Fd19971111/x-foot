package model

// RouteNode 前端路由节点
//
// 结构对齐前端 elegant-router 的 ElegantConstRoute：
//   - 一级且有子节点：Component = "layout.base"，Children 的 Component = "view.xxx"
//   - 一级且无子节点：Component = "layout.base$view.xxx"
type RouteNode struct {
	Id        string       `json:"id"`
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	Component string       `json:"component,omitempty"`
	Redirect  string       `json:"redirect,omitempty"`
	Props     bool         `json:"props,omitempty"`
	Meta      *RouteMeta   `json:"meta"`
	Children  []*RouteNode `json:"children,omitempty"`
}

// RouteMeta 前端路由元信息
type RouteMeta struct {
	Title      string `json:"title"`
	I18nKey    string `json:"i18nKey,omitempty"`
	Icon       string `json:"icon,omitempty"`
	Order      int    `json:"order,omitempty"`
	HideInMenu bool   `json:"hideInMenu,omitempty"`
	KeepAlive  bool   `json:"keepAlive,omitempty"`
	Constant   bool   `json:"constant,omitempty"`
}

// UserRouteResult 当前用户的动态路由
//
// Routes 为已授权的路由树，Home 为登录后的默认落地路由 key。
type UserRouteResult struct {
	Routes []*RouteNode `json:"routes"`
	Home   string       `json:"home"`
}
