package system

import (
	"context"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"x-foot/internal/model"
)

// maxDeptLevel 部门最大层级
const maxDeptLevel = 5

// ListDepartments 查询部门树
//
// 每一行都会带上：部门负责人的姓名、在册员工数、已授权的菜单数。
func (s *sSystem) ListDepartments(ctx context.Context) ([]*model.DepartmentItem, error) {
	records, err := g.Model(departmentTable).
		Ctx(ctx).
		OrderAsc("sort_order").
		OrderAsc("id").
		All()
	if err != nil {
		return nil, err
	}

	employeeCount, err := countByField(ctx, employeeTable, "dept_id")
	if err != nil {
		return nil, err
	}

	menuCount, err := countByField(ctx, departmentMenuTable, "dept_id")
	if err != nil {
		return nil, err
	}

	leaderName, err := employeeNameMap(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*model.DepartmentItem, 0, records.Len())
	for _, record := range records {
		id := record["id"].Uint64()

		items = append(items, &model.DepartmentItem{
			Id:            id,
			DeptCode:      record["dept_code"].String(),
			DeptName:      record["dept_name"].String(),
			ParentId:      record["parent_id"].Uint64(),
			Level:         record["level"].Int(),
			LeaderId:      record["leader_id"].Uint64(),
			LeaderName:    leaderName[record["leader_id"].Uint64()],
			EmployeeCount: employeeCount[id],
			MenuCount:     menuCount[id],
			SortOrder:     record["sort_order"].Int(),
			IsActive:      record["is_active"].Int(),
		})
	}

	return buildDeptTree(items), nil
}

// SaveDepartment 新增或编辑部门，返回部门ID
//
// Id 为 0 时新增，否则按 Id 更新；上级变化时自动重算层级。
func (s *sSystem) SaveDepartment(ctx context.Context, in *model.DepartmentSaveInput) (uint64, error) {
	in.DeptCode = strings.TrimSpace(in.DeptCode)
	in.DeptName = strings.TrimSpace(in.DeptName)

	if in.DeptCode == "" {
		return 0, gerror.New("部门编码不能为空")
	}
	if in.DeptName == "" {
		return 0, gerror.New("部门名称不能为空")
	}

	// 上级部门必须存在，且层级不能过深
	level := 1
	if in.ParentId > 0 {
		parent, err := g.Model(departmentTable).Ctx(ctx).Where("id", in.ParentId).One()
		if err != nil {
			return 0, err
		}
		if parent.IsEmpty() {
			return 0, gerror.Newf("上级部门不存在：%d", in.ParentId)
		}

		level = parent["level"].Int() + 1
		if level > maxDeptLevel {
			return 0, gerror.Newf("部门层级最多 %d 级", maxDeptLevel)
		}
	}

	// 不能把部门挂到自己或自己的子孙下面
	if in.Id > 0 && in.ParentId > 0 {
		descendants, err := childDeptIds(ctx, in.Id)
		if err != nil {
			return 0, err
		}
		for _, id := range descendants[1:] {
			if id == in.ParentId {
				return 0, gerror.New("不能把部门移动到自己的子部门之下")
			}
		}
	}

	// 编码唯一
	existQuery := g.Model(departmentTable).Ctx(ctx).Where("dept_code", in.DeptCode)
	if in.Id > 0 {
		existQuery = existQuery.WhereNot("id", in.Id)
	}
	exists, err := existQuery.Exist()
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, gerror.Newf("部门编码【%s】已存在", in.DeptCode)
	}

	if in.IsActive == 0 {
		in.IsActive = 1
	}

	data := g.Map{
		"dept_code":  in.DeptCode,
		"dept_name":  in.DeptName,
		"parent_id":  nilIfZero(in.ParentId),
		"level":      level,
		"leader_id":  nilIfZero(in.LeaderId),
		"sort_order": in.SortOrder,
		"is_active":  in.IsActive,
	}

	if in.Id == 0 {
		data["company_id"] = currentCompanyId(ctx)

		result, err := g.Model(departmentTable).Ctx(ctx).Data(data).InsertAndGetId()
		if err != nil {
			return 0, err
		}

		return uint64(result), nil
	}

	if _, err := g.Model(departmentTable).Ctx(ctx).Where("id", in.Id).Data(data).Update(); err != nil {
		return 0, err
	}

	return in.Id, nil
}

// DeleteDepartment 删除部门
//
// 存在子部门或在册员工时拒绝删除，避免出现无主的部门与员工。
func (s *sSystem) DeleteDepartment(ctx context.Context, deptId uint64) error {
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

	childrenCount, err := g.Model(departmentTable).Ctx(ctx).Where("parent_id", deptId).Count()
	if err != nil {
		return err
	}
	if childrenCount > 0 {
		return gerror.New("该部门下还有子部门，请先删除子部门")
	}

	employeeIds := make([]uint64, 0)
	if err := g.Model(employeeTable).Ctx(ctx).Where("dept_id", deptId).Fields("id").Scan(&employeeIds); err != nil {
		return err
	}
	if len(employeeIds) > 0 {
		return gerror.Newf("该部门下还有 %d 名员工，请先调整员工归属", len(employeeIds))
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(departmentMenuTable).Ctx(ctx).Where("dept_id", deptId).Delete(); err != nil {
			return err
		}

		_, err := tx.Model(departmentTable).Ctx(ctx).Where("id", deptId).Delete()

		return err
	})
}

// ---------------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------------

// countByField 按指定字段分组计数，返回 值 -> 数量
func countByField(ctx context.Context, table, field string) (map[uint64]int, error) {
	records, err := g.Model(table).Ctx(ctx).Fields(field, "count(*) total").Group(field).All()
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]int, records.Len())
	for _, record := range records {
		result[record[field].Uint64()] = record["total"].Int()
	}

	return result, nil
}

// employeeNameMap 员工ID -> 姓名
func employeeNameMap(ctx context.Context) (map[uint64]string, error) {
	records, err := g.Model(employeeTable).Ctx(ctx).Fields("id,emp_name").All()
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]string, records.Len())
	for _, record := range records {
		result[record["id"].Uint64()] = record["emp_name"].String()
	}

	return result, nil
}

// buildDeptTree 组装部门树
func buildDeptTree(items []*model.DepartmentItem) []*model.DepartmentItem {
	index := make(map[uint64]*model.DepartmentItem, len(items))
	for _, item := range items {
		index[item.Id] = item
	}

	roots := make([]*model.DepartmentItem, 0)
	for _, item := range items {
		parent, ok := index[item.ParentId]
		if item.ParentId == 0 || !ok {
			roots = append(roots, item)
			continue
		}

		parent.Children = append(parent.Children, item)
	}

	sortDeptTree(roots)

	return roots
}

// sortDeptTree 按 SortOrder 递归排序
func sortDeptTree(items []*model.DepartmentItem) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].SortOrder < items[j].SortOrder
	})

	for _, item := range items {
		if len(item.Children) > 0 {
			sortDeptTree(item.Children)
		}
	}
}

// nilIfZero 顶级部门的 parent_id 存 NULL，保持与建表语义一致
func nilIfZero(value uint64) any {
	if value == 0 {
		return nil
	}

	return value
}
