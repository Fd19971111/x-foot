import type { TreeOption } from 'naive-ui';

/**
 * 设置模块公用工具
 *
 * 说明：本文件是 .ts 而非 .vue，elegant-router 只扫描页面组件，不会为它生成路由。
 */

/* ------------------------------ 部门 ------------------------------ */

/** 部门树 -> NTreeSelect 选项；excludeIds 用于编辑时排除自身及其子孙 */
export function toDeptTreeOptions(items: Api.System.DepartmentItem[], excludeIds: number[] = []): TreeOption[] {
  return items
    .filter(item => !excludeIds.includes(item.id))
    .map(item => {
      const children = item.children?.length ? toDeptTreeOptions(item.children, excludeIds) : undefined;

      return {
        key: item.id,
        label: `${item.deptName}（${item.deptCode}）`,
        children
      } as TreeOption;
    });
}

/** 收集部门树里的全部ID */
export function collectDeptIds(items: Api.System.DepartmentItem[]): number[] {
  const ids: number[] = [];

  const walk = (nodes: Api.System.DepartmentItem[]) => {
    nodes.forEach(node => {
      ids.push(node.id);
      if (node.children?.length) walk(node.children);
    });
  };

  walk(items);

  return ids;
}

/** 在部门树里按ID查找节点 */
export function findDepartment(items: Api.System.DepartmentItem[], id: number): Api.System.DepartmentItem | null {
  for (const item of items) {
    if (item.id === id) return item;

    if (item.children?.length) {
      const found = findDepartment(item.children, id);
      if (found) return found;
    }
  }

  return null;
}

/** 部门下拉选项（用于员工查询、账号所属部门） */
export function toDeptSelectOptions(items: Api.System.DepartmentItem[]): TreeOption[] {
  return toDeptTreeOptions(items);
}

/* ------------------------------ 菜单 ------------------------------ */

/** 菜单树 -> NTree 选项 */
export function toMenuTreeOptions(items: Api.System.MenuItem[]): TreeOption[] {
  return items.map(item => {
    const children = item.children?.length ? toMenuTreeOptions(item.children) : undefined;

    return {
      key: item.id,
      label: item.menuName,
      children
    } as TreeOption;
  });
}

/** 收集菜单树里已被勾选的节点ID */
export function collectCheckedMenuIds(items: Api.System.MenuItem[]): number[] {
  const keys: number[] = [];

  const walk = (nodes: Api.System.MenuItem[]) => {
    nodes.forEach(node => {
      if (node.checked) keys.push(node.id);
      if (node.children?.length) walk(node.children);
    });
  };

  walk(items);

  return keys;
}

/** 统计菜单树里的节点总数 */
export function countMenus(items: Api.System.MenuItem[]): number {
  return items.reduce((total, item) => total + 1 + (item.children?.length ? countMenus(item.children) : 0), 0);
}

/* ------------------------------ 角色与状态 ------------------------------ */

/** 账号角色选项（与后端 validRoles 保持一致） */
export const roleOptions = [
  { label: '超级管理员', value: 'admin' },
  { label: '财务主管', value: 'finance' },
  { label: '普通员工', value: 'staff' }
];

/** 角色显示名 */
export function roleLabel(role: string) {
  return roleOptions.find(item => item.value === role)?.label || role;
}

/** 账号状态选项（与 site_user.status 枚举一致） */
export const userStatusOptions = [
  { label: '正常', value: 'active' },
  { label: '停用', value: 'disabled' }
];

/** 账号状态显示名 */
export function userStatusLabel(status: string) {
  return userStatusOptions.find(item => item.value === status)?.label || status;
}

/** 账号状态标签颜色 */
export function userStatusTagType(status: string): 'success' | 'error' | 'default' {
  if (status === 'active') return 'success';
  if (status === 'disabled') return 'error';

  return 'default';
}

/** 启用状态标签颜色 */
export function activeTagType(isActive: number): 'success' | 'default' {
  return isActive === 1 ? 'success' : 'default';
}
