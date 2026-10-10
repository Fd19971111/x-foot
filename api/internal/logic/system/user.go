package system

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"x-foot/internal/model"
	"x-foot/internal/utility/passwd"
)

// 用户状态（与 site_user.status 枚举一致）
const (
	userStatusActive   = "active"
	userStatusDisabled = "disabled"
)

// validRoles 允许的角色
var validRoles = []string{"admin", "finance", "staff"}

// ListUsers 分页查询登录账号
func (s *sSystem) ListUsers(ctx context.Context, in *model.UserQuery) (*model.UserListResult, error) {
	page, pageSize := normalizePage(in.Page, in.PageSize)

	keyword := strings.TrimSpace(in.Keyword)

	// 统计与取数分别构造：
	// GoFrame 的 Count() 会沿用模型上已设置的 Fields，带别名的多字段会让 COUNT(...) 变成非法 SQL。
	newQuery := func() *gdb.Model {
		query := g.Model(userTable + " u").Ctx(ctx).
			LeftJoin(departmentTable+" d", "u.department_id = d.id").
			LeftJoin(employeeTable+" e", "u.employee_id = e.id")

		if keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where("u.username LIKE ? OR u.real_name LIKE ? OR u.phone LIKE ?", like, like, like)
		}

		return query
	}

	total, err := newQuery().Count()
	if err != nil {
		return nil, err
	}

	records, err := newQuery().
		Fields("u.id, u.username, u.real_name, u.phone, u.email, u.role, u.department_id, u.employee_id, u.status, u.last_login_at, u.created_at, d.dept_name department_name, e.emp_name employee_name").
		Page(page, pageSize).
		OrderAsc("u.id").
		All()
	if err != nil {
		return nil, err
	}

	list := make([]*model.UserItem, 0, records.Len())
	for _, record := range records {
		var lastLogin *gtime.Time
		if text := record["last_login_at"].String(); text != "" && !strings.HasPrefix(text, "0000-00-00") {
			lastLogin = gtime.New(text)
		}

		list = append(list, &model.UserItem{
			Id:             record["id"].Uint64(),
			Username:       record["username"].String(),
			RealName:       record["real_name"].String(),
			Phone:          record["phone"].String(),
			Email:          record["email"].String(),
			Role:           record["role"].String(),
			DepartmentId:   record["department_id"].Uint64(),
			DepartmentName: record["department_name"].String(),
			EmployeeId:     record["employee_id"].Uint64(),
			EmployeeName:   record["employee_name"].String(),
			Status:         record["status"].String(),
			LastLoginAt:    lastLogin,
			CreatedAt:      gtime.New(record["created_at"].String()),
		})
	}

	return &model.UserListResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// SaveUser 新增或编辑登录账号，返回用户ID
//
// Id 为 0 时新增（密码必填），否则按 Id 更新（密码留空表示不修改）。
func (s *sSystem) SaveUser(ctx context.Context, in *model.UserSaveInput) (uint64, error) {
	in.Username = strings.TrimSpace(in.Username)

	if in.Username == "" {
		return 0, gerror.New("登录用户名不能为空")
	}

	if in.Role == "" {
		in.Role = "staff"
	}

	var roleOk bool
	for _, role := range validRoles {
		if role == in.Role {
			roleOk = true
			break
		}
	}
	if !roleOk {
		return 0, gerror.Newf("角色【%s】不合法，可选值：%s", in.Role, strings.Join(validRoles, " / "))
	}

	if in.Status == "" {
		in.Status = userStatusActive
	}

	usernameQuery := g.Model(userTable).Ctx(ctx).Where("username", in.Username)
	if in.Id > 0 {
		usernameQuery = usernameQuery.WhereNot("id", in.Id)
	}

	exists, err := usernameQuery.Exist()
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, gerror.Newf("登录用户名【%s】已存在", in.Username)
	}

	data := g.Map{
		"username":      in.Username,
		"real_name":     in.RealName,
		"phone":         in.Phone,
		"email":         in.Email,
		"role":          in.Role,
		"department_id": nilIfZero(in.DepartmentId),
		"status":        in.Status,
	}

	if in.Id == 0 {
		if strings.TrimSpace(in.Password) == "" {
			return 0, gerror.New("新增用户时必须设置初始密码")
		}

		salt := passwd.NewSalt()
		data["salt"] = salt
		data["password"] = passwd.Hash(in.Password, salt)

		result, err := g.Model(userTable).Ctx(ctx).Data(data).InsertAndGetId()
		if err != nil {
			return 0, err
		}

		return uint64(result), nil
	}

	if strings.TrimSpace(in.Password) != "" {
		salt := passwd.NewSalt()
		data["salt"] = salt
		data["password"] = passwd.Hash(in.Password, salt)
	}

	if _, err := g.Model(userTable).Ctx(ctx).Where("id", in.Id).Data(data).Update(); err != nil {
		return 0, err
	}

	return in.Id, nil
}

// DeleteUser 删除登录账号
//
// 删除前解除与该账号绑定的员工，保留员工档案本身。
func (s *sSystem) DeleteUser(ctx context.Context, userId uint64) error {
	if userId == 0 {
		return gerror.New("用户ID不能为空")
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(employeeTable).Ctx(ctx).
			Where("user_id", userId).
			Data(g.Map{"user_id": nil}).
			Update(); err != nil {
			return err
		}

		_, err := tx.Model(userTable).Ctx(ctx).Where("id", userId).Delete()

		return err
	})
}

// BindEmployee 登录账号绑定 / 解绑员工
//
// EmployeeId 为 0 表示解绑。账号与员工是一对一关系：
// 绑定时会先清掉该账号原来的绑定关系，避免一名员工被多个账号占用。
func (s *sSystem) BindEmployee(ctx context.Context, in *model.BindEmployeeInput) error {
	if in.UserId == 0 {
		return gerror.New("用户ID不能为空")
	}

	userExists, err := g.Model(userTable).Ctx(ctx).Where("id", in.UserId).Exist()
	if err != nil {
		return err
	}
	if !userExists {
		return gerror.Newf("登录账号不存在：%d", in.UserId)
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 先解绑该账号原来关联的员工
		if _, err := tx.Model(employeeTable).Ctx(ctx).
			Where("user_id", in.UserId).
			Data(g.Map{"user_id": nil}).
			Update(); err != nil {
			return err
		}

		if in.EmployeeId == 0 {
			_, err := tx.Model(userTable).Ctx(ctx).
				Where("id", in.UserId).
				Data(g.Map{"employee_id": nil}).
				Update()

			return err
		}

		employee, err := tx.Model(employeeTable).Ctx(ctx).Where("id", in.EmployeeId).One()
		if err != nil {
			return err
		}
		if employee.IsEmpty() {
			return gerror.Newf("员工不存在：%d", in.EmployeeId)
		}

		// 员工已被别的账号占用时拒绝，保证一对一
		boundUserId := employee["user_id"].Uint64()
		if boundUserId != 0 && boundUserId != in.UserId {
			return gerror.New("该员工已经绑定了其它登录账号，请先解绑")
		}

		if _, err := tx.Model(employeeTable).Ctx(ctx).
			Where("id", in.EmployeeId).
			Data(g.Map{"user_id": in.UserId}).
			Update(); err != nil {
			return err
		}

		_, err = tx.Model(userTable).Ctx(ctx).Where("id", in.UserId).Data(g.Map{
			"employee_id":   in.EmployeeId,
			"department_id": nilIfZero(employee["dept_id"].Uint64()),
		}).Update()

		return err
	})
}

// ResetUserPassword 重置登录密码
func (s *sSystem) ResetUserPassword(ctx context.Context, userId uint64, password string) error {
	password = strings.TrimSpace(password)
	if password == "" {
		return gerror.New("新密码不能为空")
	}
	if len(password) < 6 {
		return gerror.New("密码至少 6 位")
	}

	exists, err := g.Model(userTable).Ctx(ctx).Where("id", userId).Exist()
	if err != nil {
		return err
	}
	if !exists {
		return gerror.Newf("登录账号不存在：%d", userId)
	}

	salt := passwd.NewSalt()
	if _, err := g.Model(userTable).Ctx(ctx).Where("id", userId).Data(g.Map{
		"salt":     salt,
		"password": passwd.Hash(password, salt),
	}).Update(); err != nil {
		return err
	}

	return nil
}
