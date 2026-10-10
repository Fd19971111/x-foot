// Package model 系统设置模块（部门 / 员工 / 用户 / 菜单权限）的交互结构体
package model

import "github.com/gogf/gf/v2/os/gtime"

// ---------------------------------------------------------------------------------
// 菜单
// ---------------------------------------------------------------------------------

// MenuType 菜单类型：与 site_menu.menu_type 枚举一致
const (
	MenuTypeCatalog = "catalog" // 目录
	MenuTypeMenu    = "menu"    // 菜单（页面）
	MenuTypeButton  = "button"  // 按钮级权限点
)

// MenuItem 菜单树节点
//
// 用于「授权树」展示：Checked 表示该角色是否已勾选，接口返回时按当前授权对象回填。
type MenuItem struct {
	Id        uint64      `json:"id"`
	MenuCode  string      `json:"menuCode"`
	MenuName  string      `json:"menuName"`
	ParentId  uint64      `json:"parentId"`
	MenuType  string      `json:"menuType"`
	Path      string      `json:"path"`
	Icon      string      `json:"icon"`
	SortOrder int         `json:"sortOrder"`
	IsActive  int         `json:"isActive"`
	Checked   bool        `json:"checked"`
	Children  []*MenuItem `json:"children,omitempty"`
}

// MenuIdsInput 保存授权的入参
type MenuIdsInput struct {
	// MenuIds 选中的菜单ID集合，全量覆盖式保存（不在集合内即取消授权）
	MenuIds []uint64 `json:"menuIds"`
}

// ---------------------------------------------------------------------------------
// 部门
// ---------------------------------------------------------------------------------

// DepartmentItem 部门树节点
type DepartmentItem struct {
	Id            uint64            `json:"id"`
	DeptCode      string            `json:"deptCode"`
	DeptName      string            `json:"deptName"`
	ParentId      uint64            `json:"parentId"`
	Level         int               `json:"level"`
	LeaderId      uint64            `json:"leaderId"`
	LeaderName    string            `json:"leaderName"`
	EmployeeCount int               `json:"employeeCount"`
	MenuCount     int               `json:"menuCount"`
	SortOrder     int               `json:"sortOrder"`
	IsActive      int               `json:"isActive"`
	Children      []*DepartmentItem `json:"children,omitempty"`
}

// DepartmentSaveInput 新增/编辑部门
//
// Id 为 0 时新增，否则按 Id 更新。
type DepartmentSaveInput struct {
	Id        uint64 `json:"id"`
	DeptCode  string `json:"deptCode"`
	DeptName  string `json:"deptName"`
	ParentId  uint64 `json:"parentId"`
	LeaderId  uint64 `json:"leaderId"`
	SortOrder int    `json:"sortOrder"`
	// IsActive 是否启用：1=启用，0=停用（界面传布尔时由逻辑层转换）
	IsActive int `json:"isActive"`
}

// DepartmentOption 部门下拉选项
type DepartmentOption struct {
	Id       uint64              `json:"id"`
	Label    string              `json:"label"`
	Children []*DepartmentOption `json:"children,omitempty"`
}

// ---------------------------------------------------------------------------------
// 员工
// ---------------------------------------------------------------------------------

// EmployeeQuery 员工列表查询条件
type EmployeeQuery struct {
	// DeptId 部门ID，传入时同时包含所有子部门；为 0 表示不限制
	DeptId uint64 `json:"deptId"`
	// Keyword 编号 / 姓名 / 手机号 模糊匹配
	Keyword string `json:"keyword"`
	// Page 页码，默认 1
	Page int `json:"page"`
	// PageSize 每页条数，默认 20
	PageSize int `json:"pageSize"`
}

// EmployeeItem 员工列表行
type EmployeeItem struct {
	Id        uint64      `json:"id"`
	EmpCode   string      `json:"empCode"`
	EmpName   string      `json:"empName"`
	DeptId    uint64      `json:"deptId"`
	DeptName  string      `json:"deptName"`
	UserId    uint64      `json:"userId"`
	Username  string      `json:"username"`
	Position  string      `json:"position"`
	Phone     string      `json:"phone"`
	Email     string      `json:"email"`
	HireDate  *gtime.Time `json:"hireDate"`
	MenuCount int         `json:"menuCount"`
	IsActive  int         `json:"isActive"`
	CreatedAt *gtime.Time `json:"createdAt"`
}

// EmployeeListResult 员工分页列表
type EmployeeListResult struct {
	List     []*EmployeeItem `json:"list"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// EmployeeSaveInput 新增/编辑员工
type EmployeeSaveInput struct {
	Id       uint64 `json:"id"`
	EmpCode  string `json:"empCode"`
	EmpName  string `json:"empName"`
	DeptId   uint64 `json:"deptId"`
	Position string `json:"position"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	HireDate string `json:"hireDate"`
	IsActive int    `json:"isActive"`
}

// EmployeeOption 员工下拉选项（用于用户管理绑定员工）
type EmployeeOption struct {
	Id       uint64 `json:"id"`
	Label    string `json:"label"`
	Disabled bool   `json:"disabled"`
}

// ---------------------------------------------------------------------------------
// 用户（登录账号）
// ---------------------------------------------------------------------------------

// UserQuery 用户列表查询条件
type UserQuery struct {
	Keyword  string `json:"keyword"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

// UserItem 用户列表行
type UserItem struct {
	Id             uint64      `json:"id"`
	Username       string      `json:"username"`
	RealName       string      `json:"realName"`
	Phone          string      `json:"phone"`
	Email          string      `json:"email"`
	Role           string      `json:"role"`
	DepartmentId   uint64      `json:"departmentId"`
	DepartmentName string      `json:"departmentName"`
	EmployeeId     uint64      `json:"employeeId"`
	EmployeeName   string      `json:"employeeName"`
	Status         string      `json:"status"`
	LastLoginAt    *gtime.Time `json:"lastLoginAt"`
	CreatedAt      *gtime.Time `json:"createdAt"`
}

// UserListResult 用户分页列表
type UserListResult struct {
	List     []*UserItem `json:"list"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// UserSaveInput 新增/编辑用户
//
// Password 仅在新增时必填；编辑时留空表示不修改密码。
type UserSaveInput struct {
	Id           uint64 `json:"id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	RealName     string `json:"realName"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	DepartmentId uint64 `json:"departmentId"`
	Status       string `json:"status"`
}

// BindEmployeeInput 用户绑定 / 解绑员工
//
// EmployeeId 为 0 表示解绑。
type BindEmployeeInput struct {
	UserId     uint64 `json:"userId"`
	EmployeeId uint64 `json:"employeeId"`
}
