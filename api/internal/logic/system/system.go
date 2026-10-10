// Package system 系统设置业务逻辑：部门、员工、登录账号与菜单权限
//
// 文件划分：
//   system.go      —— 注册、表名常量、内部通用辅助
//   menu.go        —— 菜单树读取、内置菜单、用户最终权限计算
//   department.go  —— 部门树维护与部门授权
//   employee.go    —— 员工维护与员工授权
//   user.go        —— 登录账号维护与员工绑定
package system

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"x-foot/internal/service"
)

// 表名常量
const (
	companyTable        = "finance_company"
	menuTable           = "site_menu"
	departmentTable     = "site_department"
	departmentMenuTable = "site_department_menu"
	employeeTable       = "site_employee"
	employeeMenuTable   = "site_employee_menu"
	userTable           = "site_user"
)

// 分页默认值
const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 200
)

type sSystem struct{}

func init() {
	service.RegisterSystem(New())
}

// New 创建系统设置业务实例
func New() service.ISystem {
	return &sSystem{}
}

// normalizePage 规范分页参数
func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	return page, pageSize
}

// currentCompanyId 取当前账套ID，未初始化时为 0
//
// 部门、员工都归属到当前账套；单账套场景下取第一条公司记录即可。
func currentCompanyId(ctx context.Context) uint64 {
	value, err := g.Model(companyTable).Ctx(ctx).Fields("id").OrderAsc("id").Value()
	if err != nil {
		return 0
	}

	return value.Uint64()
}

// childDeptIds 递归收集某部门及其所有子部门的ID
func childDeptIds(ctx context.Context, deptId uint64) ([]uint64, error) {
	records, err := g.Model(departmentTable).Ctx(ctx).Fields("id,parent_id").All()
	if err != nil {
		return nil, err
	}

	childrenOf := make(map[uint64][]uint64)
	for _, record := range records {
		parentId := record["parent_id"].Uint64()
		childrenOf[parentId] = append(childrenOf[parentId], record["id"].Uint64())
	}

	result := []uint64{deptId}
	queue := []uint64{deptId}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, child := range childrenOf[current] {
			result = append(result, child)
			queue = append(queue, child)
		}
	}

	return result, nil
}
