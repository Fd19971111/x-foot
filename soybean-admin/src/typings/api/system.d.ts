declare namespace Api {
  /**
   * namespace System
   *
   * backend api module: "system"（部门 / 员工 / 账号 / 菜单权限）
   */
  namespace System {
    /** 菜单类型：catalog=目录，menu=页面，button=按钮级权限点 */
    type MenuType = 'catalog' | 'menu' | 'button';

    /** 菜单树节点（授权树） */
    interface MenuItem {
      id: number;
      menuCode: string;
      menuName: string;
      parentId: number;
      menuType: MenuType;
      path: string;
      icon: string;
      sortOrder: number;
      isActive: number;
      /** 当前授权对象是否已勾选 */
      checked: boolean;
      children?: MenuItem[];
    }

    /** 菜单树返回 */
    interface MenuTreeResult {
      list: MenuItem[];
    }

    /** 保存菜单授权入参（全量覆盖） */
    interface MenuIdsInput {
      menuIds: number[];
    }

    /** 部门树节点 */
    interface DepartmentItem {
      id: number;
      deptCode: string;
      deptName: string;
      parentId: number;
      level: number;
      leaderId: number;
      leaderName: string;
      employeeCount: number;
      menuCount: number;
      sortOrder: number;
      isActive: number;
      children?: DepartmentItem[];
    }

    /** 部门树返回 */
    interface DepartmentListResult {
      list: DepartmentItem[];
    }

    /** 新增/编辑部门入参 */
    interface DepartmentSaveInput {
      id: number;
      deptCode: string;
      deptName: string;
      parentId: number;
      leaderId: number;
      sortOrder: number;
      /** 1=启用，0=停用 */
      isActive: number;
    }

    /** 员工列表行 */
    interface EmployeeItem {
      id: number;
      empCode: string;
      empName: string;
      deptId: number;
      deptName: string;
      /** 绑定的登录账号ID，0 表示未绑定 */
      userId: number;
      username: string;
      position: string;
      phone: string;
      email: string;
      hireDate: string | null;
      menuCount: number;
      isActive: number;
      createdAt: string | null;
    }

    /** 员工分页查询条件 */
    interface EmployeeQuery {
      /** 部门ID，传入时包含其所有子部门 */
      deptId?: number;
      /** 编号 / 姓名 / 手机号 模糊匹配 */
      keyword?: string;
      page?: number;
      pageSize?: number;
    }

    /** 员工分页列表 */
    interface EmployeeListResult {
      list: EmployeeItem[];
      total: number;
      page: number;
      pageSize: number;
    }

    /** 新增/编辑员工入参 */
    interface EmployeeSaveInput {
      id: number;
      empCode: string;
      empName: string;
      deptId: number;
      position: string;
      phone: string;
      email: string;
      hireDate: string;
      isActive: number;
    }

    /** 员工下拉选项（绑定登录账号用） */
    interface EmployeeOption {
      id: number;
      label: string;
      /** 已绑定其它账号的员工不可选 */
      disabled: boolean;
    }

    /** 员工下拉选项返回 */
    interface EmployeeOptionResult {
      list: EmployeeOption[];
    }

    /** 登录账号列表行 */
    interface UserItem {
      id: number;
      username: string;
      realName: string;
      phone: string;
      email: string;
      role: string;
      departmentId: number;
      departmentName: string;
      employeeId: number;
      employeeName: string;
      status: string;
      lastLoginAt: string | null;
      createdAt: string | null;
    }

    /** 账号分页查询条件 */
    interface UserQuery {
      keyword?: string;
      page?: number;
      pageSize?: number;
    }

    /** 账号分页列表 */
    interface UserListResult {
      list: UserItem[];
      total: number;
      page: number;
      pageSize: number;
    }

    /** 新增/编辑账号入参（编辑时 password 留空表示不修改） */
    interface UserSaveInput {
      id: number;
      username: string;
      password: string;
      realName: string;
      phone: string;
      email: string;
      role: string;
      departmentId: number;
      status: string;
    }

    /** 保存返回的自增ID */
    interface SaveResult {
      id: number;
    }
  }
}
