package system

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// menuSeed 默认菜单种子
//
// ParentCode 指向同一批种子里的上级菜单编码，插入时先父后子。
type menuSeed struct {
	MenuCode   string
	MenuName   string
	MenuType   string
	Path       string
	Icon       string
	SortOrder  int
	ParentCode string
}

// defaultMenuSeeds 系统内置的默认菜单
//
// MenuCode 必须与前端路由 name 完全一致（前端页面文件位于
// src/views/setting/<page>/index.vue，elegant-router 生成的 name 即 setting_<page>），
// 否则菜单权限无法翻译成前端路由。
var defaultMenuSeeds = []menuSeed{
	{
		MenuCode:  "setting",
		MenuName:  "系统设置",
		MenuType:  "catalog",
		Path:      "/setting",
		Icon:      "mdi:cog",
		SortOrder: 100,
	},
	{
		MenuCode:   "setting_department",
		MenuName:   "部门管理",
		MenuType:   "menu",
		Path:       "/setting/department",
		Icon:       "mdi:office-building",
		SortOrder:  1,
		ParentCode: "setting",
	},
	{
		MenuCode:   "setting_employee",
		MenuName:   "员工管理",
		MenuType:   "menu",
		Path:       "/setting/employee",
		Icon:       "mdi:account-group",
		SortOrder:  2,
		ParentCode: "setting",
	},
	{
		MenuCode:   "setting_user",
		MenuName:   "账号管理",
		MenuType:   "menu",
		Path:       "/setting/user",
		Icon:       "mdi:account-key",
		SortOrder:  3,
		ParentCode: "setting",
	},
}

// SeedDefaultMenus 补齐缺失的默认菜单
//
// 按 menu_code 幂等：已存在的菜单不覆盖，只补充缺失项，
// 因此在已有生产数据的环境里重复启动也是安全的。
func (s *sSystem) SeedDefaultMenus(ctx context.Context) error {
	// 编码 -> 菜单ID，用于建立父子关系
	ids := make(map[string]uint64, len(defaultMenuSeeds))

	for _, seed := range defaultMenuSeeds {
		existing, err := g.Model(menuTable).Ctx(ctx).
			Where("menu_code", seed.MenuCode).
			Fields("id").
			One()
		if err != nil {
			return err
		}

		if !existing.IsEmpty() {
			ids[seed.MenuCode] = existing["id"].Uint64()
			continue
		}

		var parentId uint64
		if seed.ParentCode != "" {
			parentId = ids[seed.ParentCode]
		}

		result, err := g.Model(menuTable).Ctx(ctx).Data(g.Map{
			"menu_code":  seed.MenuCode,
			"menu_name":  seed.MenuName,
			"parent_id":  nilIfZero(parentId),
			"menu_type":  seed.MenuType,
			"path":       seed.Path,
			"icon":       seed.Icon,
			"sort_order": seed.SortOrder,
			"is_active":  1,
			"created_at": gtime.Now(),
		}).InsertAndGetId()
		if err != nil {
			return err
		}

		ids[seed.MenuCode] = uint64(result)
		g.Log().Infof(ctx, "已补充默认菜单：%s（%s）", seed.MenuName, seed.MenuCode)
	}

	return nil
}
