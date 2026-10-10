package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"x-foot/internal/model"
)

// GetMenuTreeReq 全部菜单树
type GetMenuTreeReq struct {
	g.Meta `path:"/api/system/menuTree" method:"get" tags:"System" summary:"全部菜单树"`
}

// GetMenuTreeRes 全部菜单树
type GetMenuTreeRes struct {
	List []*model.MenuItem `json:"list"`
}

// GetMyMenuTreeReq 当前登录用户可用的菜单树（已合并内置页面）
type GetMyMenuTreeReq struct {
	g.Meta `path:"/api/system/myMenuTree" method:"get" tags:"System" summary:"当前用户菜单树"`
}

// GetMyMenuTreeRes 当前登录用户可用的菜单树
type GetMyMenuTreeRes struct {
	List []*model.MenuItem `json:"list"`
}
