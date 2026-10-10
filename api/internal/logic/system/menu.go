package system

import (
	"context"
	"sort"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"x-foot/internal/model"
)

// roleAdmin 超级管理员角色，拥有全部菜单权限
const roleAdmin = "admin"

// ---------------------------------------------------------------------------------
// 内置菜单
// ---------------------------------------------------------------------------------

// builtinMenus 前端内置、不参与授权分配的页面
//
//   - home：登录后默认首页，任何人都要能进
//   - init：公司未完成初始化时的强制入口，菜单里隐藏
//
// 它们没有写进 site_menu，因此不可能被授权回收，不会出现「没有菜单可点」的死角。
func builtinMenus() []*model.MenuItem {
	return []*model.MenuItem{
		{
			MenuCode:  "home",
			MenuName:  "首页",
			MenuType:  model.MenuTypeMenu,
			Path:      "/home",
			Icon:      "mdi:monitor-dashboard",
			SortOrder: 1,
			IsActive:  1,
		},
		{
			MenuCode:  "init",
			MenuName:  "初始化配置",
			MenuType:  model.MenuTypeMenu,
			Path:      "/init",
			SortOrder: 9999,
			IsActive:  1,
		},
	}
}

// ---------------------------------------------------------------------------------
// 菜单树读取
// ---------------------------------------------------------------------------------

// MenuTree 查询全部启用状态的菜单树（不回填勾选状态）
func (s *sSystem) MenuTree(ctx context.Context) ([]*model.MenuItem, error) {
	menus, err := activeMenus(ctx)
	if err != nil {
		return nil, err
	}

	return buildMenuTree(menus), nil
}

// GetUserMenus 计算登录用户最终可用的菜单树
//
// 取值优先级：
//  1. 超级管理员：全部启用菜单
//  2. 已绑定员工的账号：员工权限 ∩ 部门权限（部门收权后员工自动失效）
//  3. 未绑定员工但有部门的账号：部门权限
//  4. 都没有：只有内置页面
//
// 最后统一合并内置页面，保证首页与初始化页永远可达。
func (s *sSystem) GetUserMenus(ctx context.Context, userId uint64) ([]*model.MenuItem, error) {
	record, err := g.Model(userTable).Ctx(ctx).Where("id", userId).One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return buildMenuTree(builtinMenus()), nil
	}

	var menus []*model.MenuItem

	switch role := record["role"].String(); {
	case role == roleAdmin:
		menus, err = activeMenus(ctx)
	case record["employee_id"].Uint64() > 0:
		menus, err = employeeEffectiveMenus(ctx, record["employee_id"].Uint64())
	case record["department_id"].Uint64() > 0:
		menus, err = departmentMenus(ctx, record["department_id"].Uint64())
	default:
		menus = []*model.MenuItem{}
	}
	if err != nil {
		return nil, err
	}

	menus = append(menus, builtinMenus()...)

	return buildMenuTree(menus), nil
}

// DepartmentMenuTree 查询菜单树，并回填部门当前已授权的菜单
func (s *sSystem) DepartmentMenuTree(ctx context.Context, deptId uint64) ([]*model.MenuItem, error) {
	menus, err := activeMenus(ctx)
	if err != nil {
		return nil, err
	}

	ids, err := departmentMenuIds(ctx, deptId)
	if err != nil {
		return nil, err
	}

	checked := make(map[uint64]bool, len(ids))
	for _, id := range ids {
		checked[id] = true
	}

	tree := buildMenuTree(menus)
	fillChecked(tree, checked)

	return tree, nil
}

// EmployeeMenuTree 查询菜单树，并回填员工当前已授权的菜单
//
// 可选范围限制在所属部门被授权的菜单之内，员工不可能拥有部门之外的权限。
func (s *sSystem) EmployeeMenuTree(ctx context.Context, employeeId uint64) ([]*model.MenuItem, error) {
	employee, err := g.Model(employeeTable).Ctx(ctx).Where("id", employeeId).One()
	if err != nil {
		return nil, err
	}
	if employee.IsEmpty() {
		return nil, gerror.Newf("员工不存在：%d", employeeId)
	}

	scoped, err := departmentMenus(ctx, employee["dept_id"].Uint64())
	if err != nil {
		return nil, err
	}

	records, err := g.Model(employeeMenuTable).Ctx(ctx).Where("employee_id", employeeId).Fields("menu_id").All()
	if err != nil {
		return nil, err
	}

	checked := make(map[uint64]bool, records.Len())
	for _, record := range records {
		checked[record["menu_id"].Uint64()] = true
	}

	tree := buildMenuTree(scoped)
	fillChecked(tree, checked)

	return tree, nil
}

// ---------------------------------------------------------------------------------
// 授权写入
// ---------------------------------------------------------------------------------

// SaveDepartmentMenus 保存部门菜单权限（全量覆盖）
func (s *sSystem) SaveDepartmentMenus(ctx context.Context, deptId uint64, in *model.MenuIdsInput) error {
	if deptId == 0 {
		return gerror.New("部门ID不能为空")
	}

	exists, err := g.Model(departmentTable).Ctx(ctx).Where("id", deptId).Exist()
	if err != nil {
		return err
	}
	if !exists {
		return gerror.Newf("部门不存在：%d", deptId)
	}

	menuIds, err := withAncestors(ctx, in.MenuIds)
	if err != nil {
		return err
	}

	return replaceMenus(ctx, departmentMenuTable, "dept_id", deptId, menuIds)
}

// SaveEmployeeMenus 保存员工菜单权限（全量覆盖，且必须落在部门权限内）
func (s *sSystem) SaveEmployeeMenus(ctx context.Context, employeeId uint64, in *model.MenuIdsInput) error {
	employee, err := g.Model(employeeTable).Ctx(ctx).Where("id", employeeId).One()
	if err != nil {
		return err
	}
	if employee.IsEmpty() {
		return gerror.Newf("员工不存在：%d", employeeId)
	}

	deptIds, err := departmentMenuIds(ctx, employee["dept_id"].Uint64())
	if err != nil {
		return err
	}

	allowed := make(map[uint64]bool, len(deptIds))
	for _, id := range deptIds {
		allowed[id] = true
	}

	// 员工的每一个授权都必须能在部门权限里找到出处
	for _, id := range in.MenuIds {
		if !allowed[id] {
			return gerror.New("员工的菜单权限必须在其所属部门的权限范围内，请先在部门管理中授权")
		}
	}

	menuIds, err := withAncestors(ctx, in.MenuIds)
	if err != nil {
		return err
	}

	return replaceMenus(ctx, employeeMenuTable, "employee_id", employeeId, menuIds)
}

// ---------------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------------

// activeMenus 读取全部启用状态的菜单，按排序升序
func activeMenus(ctx context.Context) ([]*model.MenuItem, error) {
	records, err := g.Model(menuTable).
		Ctx(ctx).
		Where("is_active", 1).
		OrderAsc("sort_order").
		OrderAsc("id").
		All()
	if err != nil {
		return nil, err
	}

	menus := make([]*model.MenuItem, 0, records.Len())
	for _, record := range records {
		menus = append(menus, &model.MenuItem{
			Id:        record["id"].Uint64(),
			MenuCode:  record["menu_code"].String(),
			MenuName:  record["menu_name"].String(),
			ParentId:  record["parent_id"].Uint64(),
			MenuType:  record["menu_type"].String(),
			Path:      record["path"].String(),
			Icon:      record["icon"].String(),
			SortOrder: record["sort_order"].Int(),
			IsActive:  record["is_active"].Int(),
		})
	}

	return menus, nil
}

// departmentMenus 部门被授权的菜单
func departmentMenus(ctx context.Context, deptId uint64) ([]*model.MenuItem, error) {
	ids, err := departmentMenuIds(ctx, deptId)
	if err != nil {
		return nil, err
	}

	return menusByIds(ctx, ids)
}

// employeeEffectiveMenus 员工最终生效的菜单：员工授权 ∩ 部门授权
//
// 取交集是为了兜底：即便员工被赋予了部门之外的菜单，也不会真正生效。
func employeeEffectiveMenus(ctx context.Context, employeeId uint64) ([]*model.MenuItem, error) {
	employee, err := g.Model(employeeTable).Ctx(ctx).Where("id", employeeId).One()
	if err != nil {
		return nil, err
	}
	if employee.IsEmpty() {
		return []*model.MenuItem{}, nil
	}

	deptIds, err := departmentMenuIds(ctx, employee["dept_id"].Uint64())
	if err != nil {
		return nil, err
	}

	records, err := g.Model(employeeMenuTable).Ctx(ctx).Where("employee_id", employeeId).Fields("menu_id").All()
	if err != nil {
		return nil, err
	}

	visible := make(map[uint64]bool, len(records))
	for _, record := range records {
		visible[record["menu_id"].Uint64()] = true
	}

	effective := make([]uint64, 0, len(deptIds))
	for _, id := range deptIds {
		if visible[id] {
			effective = append(effective, id)
		}
	}

	return menusByIds(ctx, effective)
}

// menusByIds 按ID批量读取菜单，保留菜单自身的排序
func menusByIds(ctx context.Context, ids []uint64) ([]*model.MenuItem, error) {
	if len(ids) == 0 {
		return []*model.MenuItem{}, nil
	}

	all, err := activeMenus(ctx)
	if err != nil {
		return nil, err
	}

	wanted := make(map[uint64]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}

	menus := make([]*model.MenuItem, 0, len(ids))
	for _, menu := range all {
		if wanted[menu.Id] {
			menus = append(menus, menu)
		}
	}

	return menus, nil
}

// departmentMenuIds 部门已授权的菜单ID
func departmentMenuIds(ctx context.Context, deptId uint64) ([]uint64, error) {
	records, err := g.Model(departmentMenuTable).
		Ctx(ctx).
		Where("dept_id", deptId).
		Fields("menu_id").
		All()
	if err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, records.Len())
	for _, record := range records {
		ids = append(ids, record["menu_id"].Uint64())
	}

	return ids, nil
}

// buildMenuTree 将扁平菜单列表组装成树
//
// 父节点缺失（例如上级菜单未被授权）时节点自动提升为顶级，避免菜单丢失。
func buildMenuTree(items []*model.MenuItem) []*model.MenuItem {
	index := make(map[uint64]*model.MenuItem, len(items))
	for _, item := range items {
		index[item.Id] = item
	}

	roots := make([]*model.MenuItem, 0)
	for _, item := range items {
		parent, ok := index[item.ParentId]
		if item.ParentId == 0 || !ok {
			roots = append(roots, item)
			continue
		}

		parent.Children = append(parent.Children, item)
	}

	sortTree(roots)

	return roots
}

// sortTree 按 SortOrder 递归排序
func sortTree(items []*model.MenuItem) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].SortOrder < items[j].SortOrder
	})

	for _, item := range items {
		if len(item.Children) > 0 {
			sortTree(item.Children)
		}
	}
}

// fillChecked 按给定集合回填树的勾选状态
func fillChecked(items []*model.MenuItem, checked map[uint64]bool) {
	for _, item := range items {
		item.Checked = checked[item.Id]

		if len(item.Children) > 0 {
			fillChecked(item.Children, checked)
		}
	}
}

// withAncestors 补齐所选节点的全部祖先节点
//
// 勾选子菜单却漏掉上级目录时菜单树会丢失层级，这里自动补上，免去手工勾选目录。
func withAncestors(ctx context.Context, ids []uint64) ([]uint64, error) {
	if len(ids) == 0 {
		return ids, nil
	}

	all, err := activeMenus(ctx)
	if err != nil {
		return nil, err
	}

	parentOf := make(map[uint64]uint64, len(all))
	for _, menu := range all {
		parentOf[menu.Id] = menu.ParentId
	}

	result := make([]uint64, 0, len(ids))
	seen := make(map[uint64]bool, len(ids))

	var walk func(id uint64)
	walk = func(id uint64) {
		if id == 0 || seen[id] {
			return
		}

		seen[id] = true

		if parentId, ok := parentOf[id]; ok {
			walk(parentId)
		}

		result = append(result, id)
	}

	for _, id := range ids {
		walk(id)
	}

	return result, nil
}

// replaceMenus 全量覆盖某个对象的菜单授权（先清后写，同一事务内完成）
func replaceMenus(ctx context.Context, table, keyField string, ownerId uint64, menuIds []uint64) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(table).Ctx(ctx).Where(keyField, ownerId).Delete(); err != nil {
			return err
		}

		for _, menuId := range menuIds {
			if menuId == 0 {
				continue
			}

			if _, err := tx.Model(table).Ctx(ctx).Data(g.Map{
				keyField:  ownerId,
				"menu_id": menuId,
			}).Insert(); err != nil {
				return err
			}
		}

		return nil
	})
}
