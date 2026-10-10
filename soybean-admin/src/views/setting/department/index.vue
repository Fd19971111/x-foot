<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, FormInst, FormRules, TreeOption } from 'naive-ui';
import { NButton, NSpace, NTag } from 'naive-ui';
import {
  fetchDeleteDepartment,
  fetchDepartmentMenus,
  fetchDepartments,
  fetchEmployeeOptions,
  saveDepartment,
  saveDepartmentMenus
} from '@/service/api';
import {
  activeTagType,
  collectCheckedMenuIds,
  collectDeptIds,
  findDepartment,
  toDeptTreeOptions,
  toMenuTreeOptions
} from '../shared';

defineOptions({
  name: 'SettingDepartment'
});

const loading = ref(false);
const submitting = ref(false);

/** 原始部门树 */
const deptTree = ref<Api.System.DepartmentItem[]>([]);
/** 员工下拉（用于选择部门负责人） */
const employeeOptions = ref<{ label: string; value: number }[]>([]);

/* ------------------------------ 新增 / 编辑 ------------------------------ */

const modalVisible = ref(false);
const formRef = ref<FormInst | null>(null);
const editingId = ref(0);

const form = reactive({
  deptCode: '',
  deptName: '',
  parentId: null as number | null,
  leaderId: null as number | null,
  sortOrder: 0,
  isActive: true
});

const rules: FormRules = {
  deptCode: { required: true, message: '请输入部门编码', trigger: ['input', 'blur'] },
  deptName: { required: true, message: '请输入部门名称', trigger: ['input', 'blur'] }
};

const modalTitle = computed(() => (editingId.value ? '编辑部门' : '新增部门'));

/**
 * 可选上级部门
 *
 * 编辑时排除自身及其所有子孙，避免把部门挂到自己的下级形成环（后端也会校验）。
 */
const parentOptions = computed(() => {
  let excludeIds: number[] = [];

  if (editingId.value) {
    const self = findDepartment(deptTree.value, editingId.value);
    if (self) excludeIds = collectDeptIds([self]);
  }

  return toDeptTreeOptions(deptTree.value, excludeIds);
});

/* ------------------------------ 菜单权限 ------------------------------ */

const menuVisible = ref(false);
const menuLoading = ref(false);
const menuSaving = ref(false);
const menuDept = ref<Api.System.DepartmentItem | null>(null);
const menuOptions = ref<TreeOption[]>([]);
const checkedMenuKeys = ref<number[]>([]);
const allMenuKeys = ref<number[]>([]);

async function loadDepartments() {
  loading.value = true;
  const { data, error } = await fetchDepartments();
  loading.value = false;

  if (error || !data) return;

  deptTree.value = data.list || [];
}

async function loadEmployeeOptions() {
  const { data, error } = await fetchEmployeeOptions();

  if (error || !data) return;

  // 该接口的 disabled 只表示「已绑定登录账号」，与能否担任部门负责人无关，这里统一放开
  employeeOptions.value = (data.list || []).map(item => ({ label: item.label, value: item.id }));
}

function resetForm() {
  form.deptCode = '';
  form.deptName = '';
  form.parentId = null;
  form.leaderId = null;
  form.sortOrder = 0;
  form.isActive = true;
}

function openCreate(parent?: Api.System.DepartmentItem) {
  editingId.value = 0;
  resetForm();
  form.parentId = parent ? parent.id : null;
  modalVisible.value = true;
}

function openEdit(row: Api.System.DepartmentItem) {
  editingId.value = row.id;
  form.deptCode = row.deptCode;
  form.deptName = row.deptName;
  form.parentId = row.parentId || null;
  form.leaderId = row.leaderId || null;
  form.sortOrder = row.sortOrder;
  form.isActive = row.isActive === 1;
  modalVisible.value = true;
}

async function submitForm() {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  const payload: Api.System.DepartmentSaveInput = {
    id: editingId.value,
    deptCode: form.deptCode.trim(),
    deptName: form.deptName.trim(),
    parentId: form.parentId || 0,
    leaderId: form.leaderId || 0,
    sortOrder: Number(form.sortOrder) || 0,
    isActive: form.isActive ? 1 : 0
  };

  submitting.value = true;
  const { error } = await saveDepartment(payload);
  submitting.value = false;

  if (error) return;

  window.$message?.success(editingId.value ? '修改成功' : '新增成功');
  modalVisible.value = false;
  await loadDepartments();
}

function handleDelete(row: Api.System.DepartmentItem) {
  window.$dialog?.warning({
    title: '删除部门',
    content: `确认删除部门「${row.deptName}」吗？该部门的菜单授权会一并清除。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { error } = await fetchDeleteDepartment(row.id);
      if (error) return false;

      window.$message?.success('删除成功');
      await loadDepartments();

      return true;
    }
  });
}

/* ------------------------------ 菜单权限操作 ------------------------------ */

/** 递归收集树里的全部 key */
function flattenKeys(options: TreeOption[]): number[] {
  return options.reduce<number[]>((keys, option) => {
    keys.push(option.key as number);
    if (option.children?.length) keys.push(...flattenKeys(option.children));

    return keys;
  }, []);
}

async function openMenuModal(row: Api.System.DepartmentItem) {
  menuDept.value = row;
  menuVisible.value = true;
  menuLoading.value = true;
  menuOptions.value = [];
  checkedMenuKeys.value = [];
  allMenuKeys.value = [];

  const { data, error } = await fetchDepartmentMenus(row.id);
  menuLoading.value = false;

  if (error || !data) return;

  menuOptions.value = toMenuTreeOptions(data.list || []);
  checkedMenuKeys.value = collectCheckedMenuIds(data.list || []);
  allMenuKeys.value = flattenKeys(menuOptions.value);
}

async function submitMenus() {
  if (!menuDept.value) return;

  menuSaving.value = true;
  const { error } = await saveDepartmentMenus(menuDept.value.id, { menuIds: checkedMenuKeys.value });
  menuSaving.value = false;

  if (error) return;

  window.$message?.success('菜单权限已保存');
  menuVisible.value = false;
  await loadDepartments();
}

/* ------------------------------ 表格列 ------------------------------ */

const columns: DataTableColumns<Api.System.DepartmentItem> = [
  { key: 'deptName', title: '部门名称', minWidth: 200 },
  { key: 'deptCode', title: '部门编码', width: 130 },
  {
    key: 'leaderName',
    title: '负责人',
    width: 120,
    render: row => row.leaderName || '—'
  },
  {
    key: 'employeeCount',
    title: '员工数',
    width: 90,
    render: row => h(NTag, { size: 'small', bordered: false }, { default: () => row.employeeCount })
  },
  {
    key: 'menuCount',
    title: '已授权菜单',
    width: 110,
    render: row =>
      h(
        NTag,
        { size: 'small', bordered: false, type: row.menuCount ? 'info' : 'default' },
        { default: () => row.menuCount }
      )
  },
  { key: 'level', title: '层级', width: 80 },
  { key: 'sortOrder', title: '排序', width: 80 },
  {
    key: 'isActive',
    title: '状态',
    width: 90,
    render: row =>
      h(
        NTag,
        { size: 'small', bordered: false, type: activeTagType(row.isActive) },
        { default: () => (row.isActive === 1 ? '启用' : '停用') }
      )
  },
  {
    key: 'action',
    title: '操作',
    width: 220,
    fixed: 'right',
    render: row =>
      h(
        NSpace,
        { size: 4, wrap: false },
        {
          default: () => [
            h(
              NButton,
              { size: 'tiny', quaternary: true, type: 'primary', onClick: () => openEdit(row) },
              { default: () => '编辑' }
            ),
            h(
              NButton,
              { size: 'tiny', quaternary: true, type: 'info', onClick: () => openCreate(row) },
              { default: () => '添加下级' }
            ),
            h(
              NButton,
              { size: 'tiny', quaternary: true, type: 'primary', onClick: () => openMenuModal(row) },
              { default: () => '菜单权限' }
            ),
            h(
              NButton,
              { size: 'tiny', quaternary: true, type: 'error', onClick: () => handleDelete(row) },
              { default: () => '删除' }
            )
          ]
        }
      )
  }
];

onMounted(async () => {
  await Promise.all([loadDepartments(), loadEmployeeOptions()]);
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard title="部门管理" :bordered="false" size="small" class="sm:flex-1-hidden card-wrapper">
      <template #header-extra>
        <NSpace :size="12">
          <NButton size="small" @click="loadDepartments">
            <template #icon>
              <icon-ic-round-refresh class="text-icon" />
            </template>
            刷新
          </NButton>
          <NButton size="small" type="primary" @click="openCreate()">
            <template #icon>
              <icon-ic-round-plus class="text-icon" />
            </template>
            新增部门
          </NButton>
        </NSpace>
      </template>

      <NAlert type="info" :bordered="false" class="mb-12px">
        先维护部门层级，再对部门分配菜单权限；员工的可用菜单不能超出所属部门的授权范围。
      </NAlert>

      <NDataTable
        :columns="columns"
        :data="deptTree"
        :loading="loading"
        :row-key="row => row.id"
        :bordered="false"
        :single-line="false"
        size="small"
        flex-height
        class="h-full"
      />
    </NCard>

    <!-- 新增 / 编辑部门 -->
    <NModal v-model:show="modalVisible" preset="card" :title="modalTitle" class="w-560px">
      <NForm ref="formRef" :model="form" :rules="rules" label-placement="top">
        <NGrid :x-gap="16" :y-gap="0" responsive="screen" item-responsive>
          <NGridItem span="24 s:12">
            <NFormItem label="部门编码" path="deptCode">
              <NInput v-model:value="form.deptCode" placeholder="如：D001，需唯一" clearable />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="部门名称" path="deptName">
              <NInput v-model:value="form.deptName" placeholder="如：财务部" clearable />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="上级部门">
              <NTreeSelect
                v-model:value="form.parentId"
                :options="parentOptions"
                placeholder="留空表示顶级部门"
                clearable
                filterable
                default-expand-all
              />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="部门负责人">
              <NSelect
                v-model:value="form.leaderId"
                :options="employeeOptions"
                placeholder="可留空，需先维护员工"
                clearable
                filterable
              />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="排序">
              <NInputNumber v-model:value="form.sortOrder" class="w-full" :min="0" :show-button="false" />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="启用状态">
              <NSwitch v-model:value="form.isActive">
                <template #checked>启用</template>
                <template #unchecked>停用</template>
              </NSwitch>
            </NFormItem>
          </NGridItem>
        </NGrid>
      </NForm>

      <template #footer>
        <NSpace justify="end" :size="12">
          <NButton @click="modalVisible = false">取消</NButton>
          <NButton type="primary" :loading="submitting" @click="submitForm">保存</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 菜单权限分配 -->
    <NModal v-model:show="menuVisible" preset="card" class="w-560px" :title="`菜单权限 · ${menuDept?.deptName || ''}`">
      <NSpin :show="menuLoading">
        <NAlert type="info" :bordered="false" class="mb-12px">
          勾选该部门可用的页面。勾选子菜单时，服务端会自动补齐上级目录。
        </NAlert>

        <NSpace :size="8" class="mb-8px">
          <NButton size="tiny" quaternary @click="checkedMenuKeys = [...allMenuKeys]">全选</NButton>
          <NButton size="tiny" quaternary @click="checkedMenuKeys = []">全不选</NButton>
          <NTag size="small" :bordered="false" type="info">已选 {{ checkedMenuKeys.length }}</NTag>
        </NSpace>

        <div class="max-h-420px overflow-auto rounded-6px bg-layout p-12px">
          <NTree
            v-model:checked-keys="checkedMenuKeys"
            :data="menuOptions"
            checkable
            :cascade="false"
            default-expand-all
            block-line
          />
          <NEmpty v-if="!menuLoading && !menuOptions.length" size="small" description="暂无可授权的菜单" />
        </div>
      </NSpin>

      <template #footer>
        <NSpace justify="end" :size="12">
          <NButton @click="menuVisible = false">取消</NButton>
          <NButton type="primary" :loading="menuSaving" :disabled="menuLoading" @click="submitMenus">保存权限</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>
