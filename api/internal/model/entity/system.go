package entity

import "github.com/gogf/gf/v2/os/gtime"

// SiteMenu 菜单/权限表 site_menu
//
// MenuCode 与前端路由的 route name 保持一致，例如 setting / setting_user，
// 这样菜单权限下发后可以直接翻译成前端路由，不需要额外的映射表。
type SiteMenu struct {
	Id        uint64      `json:"id"`
	MenuCode  string      `json:"menuCode"`
	MenuName  string      `json:"menuName"`
	ParentId  uint64      `json:"parentId"`
	MenuType  string      `json:"menuType"`
	Path      string      `json:"path"`
	Icon      string      `json:"icon"`
	SortOrder int         `json:"sortOrder"`
	IsActive  int         `json:"isActive"`
	CreatedAt *gtime.Time `json:"createdAt"`
}

// SiteDepartment 部门表 site_department
type SiteDepartment struct {
	Id        uint64      `json:"id"`
	DeptCode  string      `json:"deptCode"`
	DeptName  string      `json:"deptName"`
	ParentId  uint64      `json:"parentId"`
	Level     int         `json:"level"`
	CompanyId uint64      `json:"companyId"`
	LeaderId  uint64      `json:"leaderId"`
	SortOrder int         `json:"sortOrder"`
	IsActive  int         `json:"isActive"`
	CreatedAt *gtime.Time `json:"createdAt"`
}

// SiteEmployee 员工表 site_employee
//
// UserId 为该员工绑定的登录账号，为 0 表示还没有账号。
type SiteEmployee struct {
	Id        uint64      `json:"id"`
	EmpCode   string      `json:"empCode"`
	EmpName   string      `json:"empName"`
	DeptId    uint64      `json:"deptId"`
	UserId    uint64      `json:"userId"`
	Position  string      `json:"position"`
	Phone     string      `json:"phone"`
	Email     string      `json:"email"`
	HireDate  *gtime.Time `json:"hireDate"`
	IsActive  int         `json:"isActive"`
	CreatedAt *gtime.Time `json:"createdAt"`
}

// SiteUser 系统用户（登录账号）表 site_user
type SiteUser struct {
	Id             uint64      `json:"id"`
	Username       string      `json:"username"`
	Password       string      `json:"password"`
	Salt           string      `json:"salt"`
	RealName       string      `json:"realName"`
	Phone          string      `json:"phone"`
	Email          string      `json:"email"`
	Role           string      `json:"role"`
	DepartmentId   uint64      `json:"departmentId"`
	CompanyId      uint64      `json:"companyId"`
	EmployeeId     uint64      `json:"employeeId"`
	Status         string      `json:"status"`
	LastLoginAt    *gtime.Time `json:"lastLoginAt"`
	LastLoginIp    string      `json:"lastLoginIp"`
	LoginFailCount int         `json:"loginFailCount"`
	CreatedAt      *gtime.Time `json:"createdAt"`
	UpdatedAt      *gtime.Time `json:"updatedAt"`
}
