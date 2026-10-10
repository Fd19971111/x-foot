import { request } from '../request';

/* ------------------------------ 菜单 ------------------------------ */

/** 全部菜单树（用于权限分配） */
export function fetchMenuTree() {
  return request<Api.System.MenuTreeResult>({ url: '/api/system/menuTree' });
}

/** 当前登录用户可用的菜单树 */
export function fetchMyMenuTree() {
  return request<Api.System.MenuTreeResult>({ url: '/api/system/myMenuTree' });
}

/* ------------------------------ 部门 ------------------------------ */

/** 部门树 */
export function fetchDepartments() {
  return request<Api.System.DepartmentListResult>({ url: '/api/system/departments' });
}

/** 新增或编辑部门（id 为 0 表示新增） */
export function saveDepartment(data: Api.System.DepartmentSaveInput) {
  return request<Api.System.SaveResult>({
    url: '/api/system/departments',
    method: 'post',
    data
  });
}

/** 删除部门 */
export function fetchDeleteDepartment(id: number) {
  return request<null>({ url: `/api/system/departments/${id}`, method: 'delete' });
}

/** 部门的菜单权限树（含勾选状态） */
export function fetchDepartmentMenus(id: number) {
  return request<Api.System.MenuTreeResult>({ url: `/api/system/departments/${id}/menus` });
}

/** 保存部门菜单权限（全量覆盖） */
export function saveDepartmentMenus(id: number, data: Api.System.MenuIdsInput) {
  return request<null>({
    url: `/api/system/departments/${id}/menus`,
    method: 'put',
    data
  });
}

/* ------------------------------ 员工 ------------------------------ */

/** 员工分页列表 */
export function fetchEmployees(params: Api.System.EmployeeQuery) {
  return request<Api.System.EmployeeListResult>({ url: '/api/system/employees', params });
}

/** 新增或编辑员工（id 为 0 表示新增） */
export function saveEmployee(data: Api.System.EmployeeSaveInput) {
  return request<Api.System.SaveResult>({
    url: '/api/system/employees',
    method: 'post',
    data
  });
}

/** 删除员工 */
export function fetchDeleteEmployee(id: number) {
  return request<null>({ url: `/api/system/employees/${id}`, method: 'delete' });
}

/** 员工的菜单权限树（可选范围已限制在所属部门内） */
export function fetchEmployeeMenus(id: number) {
  return request<Api.System.MenuTreeResult>({ url: `/api/system/employees/${id}/menus` });
}

/** 保存员工菜单权限（全量覆盖，且不能超出部门权限） */
export function saveEmployeeMenus(id: number, data: Api.System.MenuIdsInput) {
  return request<null>({
    url: `/api/system/employees/${id}/menus`,
    method: 'put',
    data
  });
}

/** 可绑定登录账号的员工列表（已绑定的标记为不可选） */
export function fetchEmployeeOptions(deptId?: number) {
  return request<Api.System.EmployeeOptionResult>({
    url: '/api/system/employeeOptions',
    params: deptId ? { deptId } : {}
  });
}

/* ------------------------------ 登录账号 ------------------------------ */

/** 账号分页列表 */
export function fetchUsers(params: Api.System.UserQuery) {
  return request<Api.System.UserListResult>({ url: '/api/system/users', params });
}

/** 新增或编辑账号（id 为 0 表示新增） */
export function saveUser(data: Api.System.UserSaveInput) {
  return request<Api.System.SaveResult>({
    url: '/api/system/users',
    method: 'post',
    data
  });
}

/** 删除账号 */
export function fetchDeleteUser(id: number) {
  return request<null>({ url: `/api/system/users/${id}`, method: 'delete' });
}

/** 账号绑定员工（employeeId 为 0 表示解绑） */
export function bindUserEmployee(id: number, employeeId: number) {
  return request<null>({
    url: `/api/system/users/${id}/bindEmployee`,
    method: 'put',
    data: { employeeId }
  });
}

/** 重置登录密码 */
export function resetUserPassword(id: number, password: string) {
  return request<null>({
    url: `/api/system/users/${id}/password`,
    method: 'put',
    data: { password }
  });
}
