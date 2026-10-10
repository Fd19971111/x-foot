package system

// ControllerV1 系统设置控制器：部门 / 员工 / 登录账号 / 菜单权限
type ControllerV1 struct{}

// NewV1 创建系统设置控制器
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}
