package system

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"x-foot/internal/model"
)

// ListEmployees 分页查询员工
//
// DeptId 传入时会包含该部门的所有子部门；Keyword 匹配员工编号、姓名、手机号。
func (s *sSystem) ListEmployees(ctx context.Context, in *model.EmployeeQuery) (*model.EmployeeListResult, error) {
	page, pageSize := normalizePage(in.Page, in.PageSize)

	var deptIds []uint64
	if in.DeptId > 0 {
		ids, err := childDeptIds(ctx, in.DeptId)
		if err != nil {
			return nil, err
		}

		deptIds = ids
	}

	keyword := strings.TrimSpace(in.Keyword)

	// 统计与取数分别构造：
	// GoFrame 的 Count() 会沿用模型上已设置的 Fields，带别名的多字段（e.*, d.dept_name ...）
	// 会生成 COUNT(e.*, d.dept_name ...) 这样的非法 SQL。
	newQuery := func() *gdb.Model {
		query := g.Model(employeeTable+" e").Ctx(ctx).
			LeftJoin(departmentTable+" d", "e.dept_id = d.id").
			LeftJoin(userTable+" u", "e.user_id = u.id")

		if len(deptIds) > 0 {
			query = query.Where("e.dept_id", deptIds)
		}

		if keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where("e.emp_code LIKE ? OR e.emp_name LIKE ? OR e.phone LIKE ?", like, like, like)
		}

		return query
	}

	total, err := newQuery().Count()
	if err != nil {
		return nil, err
	}

	records, err := newQuery().
		Fields("e.*, d.dept_name dept_name, u.username username").
		Page(page, pageSize).
		OrderDesc("e.id").
		All()
	if err != nil {
		return nil, err
	}

	menuCount, err := countByField(ctx, employeeMenuTable, "employee_id")
	if err != nil {
		return nil, err
	}

	list := make([]*model.EmployeeItem, 0, records.Len())
	for _, record := range records {
		id := record["id"].Uint64()

		var hireDate *gtime.Time
		if text := record["hire_date"].String(); text != "" && !strings.HasPrefix(text, "0000-00-00") {
			hireDate = gtime.New(text)
		}

		list = append(list, &model.EmployeeItem{
			Id:        id,
			EmpCode:   record["emp_code"].String(),
			EmpName:   record["emp_name"].String(),
			DeptId:    record["dept_id"].Uint64(),
			DeptName:  record["dept_name"].String(),
			UserId:    record["user_id"].Uint64(),
			Username:  record["username"].String(),
			Position:  record["position"].String(),
			Phone:     record["phone"].String(),
			Email:     record["email"].String(),
			HireDate:  hireDate,
			MenuCount: menuCount[id],
			IsActive:  record["is_active"].Int(),
			CreatedAt: gtime.New(record["created_at"].String()),
		})
	}

	return &model.EmployeeListResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// SaveEmployee 新增或编辑员工，返回员工ID
//
// Id 为 0 时新增，否则按 Id 更新；员工必须归属到一个已存在的部门。
func (s *sSystem) SaveEmployee(ctx context.Context, in *model.EmployeeSaveInput) (uint64, error) {
	in.EmpCode = strings.TrimSpace(in.EmpCode)
	in.EmpName = strings.TrimSpace(in.EmpName)

	if in.EmpCode == "" {
		return 0, gerror.New("员工编号不能为空")
	}
	if in.EmpName == "" {
		return 0, gerror.New("员工姓名不能为空")
	}
	if in.DeptId == 0 {
		return 0, gerror.New("请选择员工所属部门")
	}

	deptExists, err := g.Model(departmentTable).Ctx(ctx).Where("id", in.DeptId).Exist()
	if err != nil {
		return 0, err
	}
	if !deptExists {
		return 0, gerror.Newf("部门不存在：%d", in.DeptId)
	}

	codeQuery := g.Model(employeeTable).Ctx(ctx).Where("emp_code", in.EmpCode)
	if in.Id > 0 {
		codeQuery = codeQuery.WhereNot("id", in.Id)
	}

	codeExists, err := codeQuery.Exist()
	if err != nil {
		return 0, err
	}
	if codeExists {
		return 0, gerror.Newf("员工编号【%s】已存在", in.EmpCode)
	}

	if in.IsActive == 0 {
		in.IsActive = 1
	}

	data := g.Map{
		"emp_code":  in.EmpCode,
		"emp_name":  in.EmpName,
		"dept_id":   in.DeptId,
		"position":  in.Position,
		"phone":     in.Phone,
		"email":     in.Email,
		"hire_date": nilIfEmptyDate(in.HireDate),
		"is_active": in.IsActive,
	}

	if in.Id == 0 {
		result, err := g.Model(employeeTable).Ctx(ctx).Data(data).InsertAndGetId()
		if err != nil {
			return 0, err
		}

		return uint64(result), nil
	}

	if _, err := g.Model(employeeTable).Ctx(ctx).Where("id", in.Id).Data(data).Update(); err != nil {
		return 0, err
	}

	return in.Id, nil
}

// DeleteEmployee 删除员工
//
// 同时清理该员工的菜单授权，并解除其登录账号的绑定关系。
func (s *sSystem) DeleteEmployee(ctx context.Context, employeeId uint64) error {
	if employeeId == 0 {
		return gerror.New("员工ID不能为空")
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(employeeMenuTable).Ctx(ctx).Where("employee_id", employeeId).Delete(); err != nil {
			return err
		}

		if _, err := tx.Model(userTable).Ctx(ctx).
			Where("employee_id", employeeId).
			Data(g.Map{"employee_id": nil}).
			Update(); err != nil {
			return err
		}

		_, err := tx.Model(employeeTable).Ctx(ctx).Where("id", employeeId).Delete()

		return err
	})
}

// EmployeeOptions 查询可用于绑定登录账号的员工
//
// 已经绑定过登录账号的员工标记为不可选，避免一对一关系被破坏。
func (s *sSystem) EmployeeOptions(ctx context.Context, deptId uint64) ([]*model.EmployeeOption, error) {
	query := g.Model(employeeTable+" e").Ctx(ctx).
		LeftJoin(departmentTable+" d", "e.dept_id = d.id").
		Fields("e.id, e.emp_code, e.emp_name, e.user_id, d.dept_name").
		Where("e.is_active", 1)

	if deptId > 0 {
		deptIds, err := childDeptIds(ctx, deptId)
		if err != nil {
			return nil, err
		}

		query = query.Where("e.dept_id", deptIds)
	}

	records, err := query.OrderAsc("e.id").All()
	if err != nil {
		return nil, err
	}

	options := make([]*model.EmployeeOption, 0, records.Len())
	for _, record := range records {
		label := record["emp_name"].String() + "（" + record["emp_code"].String() + "）"
		if deptName := record["dept_name"].String(); deptName != "" {
			label += " · " + deptName
		}

		options = append(options, &model.EmployeeOption{
			Id:       record["id"].Uint64(),
			Label:    label,
			Disabled: record["user_id"].Uint64() > 0,
		})
	}

	return options, nil
}

// nilIfEmptyDate 入职日期为空时存 NULL
func nilIfEmptyDate(date string) any {
	date = strings.TrimSpace(date)
	if date == "" {
		return nil
	}

	return date
}
