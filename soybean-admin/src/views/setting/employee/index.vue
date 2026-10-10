<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, FormInst, FormRules, TreeOption } from 'naive-ui';
import { NButton, NSpace, NTag } from 'naive-ui';
import {
  fetchDeleteEmployee,
  fetchDepartments,
  fetchEmployeeMenus,
  fetchEmployees,
  saveEmployee,
  saveEmployeeMenus
} from '@/service/api';
import { activeTagType, collectCheckedMenuIds, toDeptTreeOptions, toMenuTreeOptions } from '../shared';

defineOptions({
  name: 'SettingEmployee'
});

const loading = ref(false);
const submitting = ref(false);

const list = ref<Api.System.EmployeeItem[]>([]);
const total = ref(0);
const deptTree = ref<Api.System.DepartmentItem[]>([]);

const query = reactive({
  deptId: null as number | null,
  keyword: '',
  page: 1,
  pageSize: 20
});

const deptOptions = computed<TreeOption[]>(() => toDeptTreeOptions(deptTree.value));

/* ------------------------------ 新增 / 编辑 ------------------------------ */

const modalVisible = ref(false);
const formRef = ref<FormInst | null>(null);
const editingId = ref(0);

const form = reactive({
  empCode: '',
  empName: '',
  deptId: null as number | null,
  position: '',
  phone: '',
  email: '',
  hireDate: null as string | null,
  isActive: true
});

const rules: FormRules = {
  empCode: { required: true, message: '请输入员工编号', trigger: ['input', 'blur'] },
  empName: { required: true, message: '请输入员工姓名', trigger: ['input', 'blur'] },
  deptId: { required: true, type: 'number', message: '请选择所属部门', trigger: ['change', 'blur'] }
};

const modalTitle = computed(() => (editingId.value ? '编辑员工' : '新增员工'));

/* ------------------------------ 菜单权限 ------------------------------ */

const menuVisible = ref(false);
const menuLoading = ref(false);
const menuSaving = ref(false);
const menuEmployee = ref<Api.System.EmployeeItem | null>(null);
const menuOptions = ref<TreeOption[]>([]);
const checkedMenuKeys = ref<number[]>([]);
const allMenuKeys = ref<number[]>([]);

function flattenKeys(options: TreeOption[]): number[] {
  return options.reduce<number[]>((keys, option) => {
    keys.push(option.key as number);
    if (option.children?.length) keys.push(...flattenKeys(option.children));

    return keys;
  }, []);
}

async function loadDepartments() {
  const { data, error } = await fetchDepartments();

  if (error || !data) return;

  deptTree.value = data.list || [];
}

async function loadEmployees() {
  loading.value = true;
  const { data, error } = await fetchEmployees({
    deptId: query.deptId || undefined,
    keyword: query.keyword.trim() || undefined,
    page: query.page,
    pageSize: query.pageSize
  });
  loading.value = false;

  if (error || !data) return;

  list.value = data.list || [];
  total.value = data.total || 0;
}

function handleSearch() {
  query.page = 1;
  loadEmployees();
}

function handleReset() {
  query.deptId = null;
  query.keyword = '';
  query.page = 1;
  loadEmployees();
}

function resetForm() {
  form.empCode = '';
  form.empName = '';
  form.deptId = null;
  form.position = '';
  form.phone = '';
  form.email = '';
  form.hireDate = null;
  form.isActive = true;
}

function openCreate() {
  editingId.value = 0;
  resetForm();
  form.deptId = query.deptId;
  modalVisible.value = true;
}

function openEdit(row: Api.System.EmployeeItem) {
  editingId.value = row.id;
  form.empCode = row.empCode;
  form.empName = row.empName;
  form.deptId = row.deptId || null;
  form.position = row.position;
  form.phone = row.phone;
  form.email = row.email;
  form.hireDate = row.hireDate ? String(row.hireDate).slice(0, 10) : null;
  form.isActive = row.isActive === 1;
  modalVisible.value = true;
}

async function submitForm() {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  const payload: Api.System.EmployeeSaveInput = {
    id: editingId.value,
    empCode: form.empCode.trim(),
    empName: form.empName.trim(),
    deptId: form.deptId || 0,
    position: form.position,
    phone: form.phone,
    email: form.email,
    hireDate: form.hireDate || '',
    isActive: form.isActive ? 1 : 0
  };

  submitting.value = true;
  const { error } = await saveEmployee(payload);
  submitting.value = false;

  if (error) return;

  window.$message?.success(editingId.value ? '修改成功' : '新增成功');
  modalVisible.value = false;
  await loadEmployees();
}

function handleDelete(row: Api.System.EmployeeItem) {
  window.$dialog?.warning({
    title: '删除员工',
    content: `确认删除员工「${row.empName}」吗？其菜单授权会被清除，绑定的登录账号将自动解绑。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { error } = await fetchDeleteEmployee(row.id);
      if (error) return false;

      window.$message?.success('删除成功');
      await loadEmployees();

      return true;
    }
  });
}

async function openMenuModal(row: Api.System.EmployeeItem) {
  menuEmployee.value = row;
  menuVisible.value = true;
  menuLoading.value = true;
  menuOptions.value = [];
  checkedMenuKeys.value = [];
  allMenuKeys.value = [];

  const { data, error } = await fetchEmployeeMenus(row.id);
  menuLoading.value = false;

  if (error || !data) return;

  menuOptions.value = toMenuTreeOptions(data.list || []);
  checkedMenuKeys.value = collectCheckedMenuIds(data.list || []);
  allMenuKeys.value = flattenKeys(menuOptions.value);
}

async function submitMenus() {
  if (!menuEmployee.value) return;

  menuSaving.value = true;
  const { error } = await saveEmployeeMenus(menuEmployee.value.id, { menuIds: checkedMenuKeys.value });
  menuSaving.value = false;

  if (error) return;

  window.$message?.success('菜单权限已保存');
  menuVisible.value = false;
  await loadEmployees();
}

/* ------------------------------ 表格 ------------------------------ */

const pagination = computed(() => ({
  page: query.page,
  pageSize: query.pageSize,
  itemCount: total.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  prefix: ({ itemCount }: { itemCount?: number }) => `共 ${itemCount ?? 0} 条`,
  onUpdatePage: (page: number) => {
    query.page = page;
    loadEmployees();
  },
  onUpdatePageSize: (pageSize: number) => {
    query.pageSize = pageSize;
    query.page = 1;
    loadEmployees();
  }
}));

const columns: DataTableColumns<Api.System.EmployeeItem> = [
  { key: 'empCode', title: '员工编号', width: 120 },
  { key: 'empName', title: '姓名', minWidth: 110 },
  { key: 'deptName', title: '所属部门', minWidth: 130, render: row => row.deptName || '—' },
  { key: 'position', title: '岗位', width: 120, render: row => row.position || '—' },
  { key: 'phone', title: '手机号', width: 130, render: row => row.phone || '—' },
  {
    key: 'username',
    title: '登录账号',
    width: 130,
    render: row =>
      row.username
        ? h(NTag, { size: 'small', bordered: false, type: 'success' }, { default: () => row.username })
        : h('span', { class: 'text-gray-400' }, '未绑定')
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
  {
    key: 'isActive',
    title: '状态',
    width: 90,
    render: row =>
      h(
        NTag,
        { size: 'small', bordered: false, type: activeTagType(row.isActive) },
        { default: () => (row.isActive === 1 ? '在职' : '停用') }
      )
  },
  {
    key: 'action',
    title: '操作',
    width: 200,
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
              { size: 'tiny', quaternary: true, type: 'info', onClick: () => openMenuModal(row) },
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
  await loadDepartments();
  await loadEmployees();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard title="员工管理" :bordered="false" size="small" class="card-wrapper">
      <template #header-extra>
        <NButton size="small" type="primary" @click="openCreate">
          <template #icon>
            <icon-ic-round-plus class="text-icon" />
          </template>
          新增员工
        </NButton>
      </template>

      <NSpace class="mb-12px" :size="12" align="center" wrap>
        <NTreeSelect
          v-model:value="query.deptId"
          :options="deptOptions"
          placeholder="按部门筛选（含子部门）"
          class="w-240px"
          clearable
          filterable
          default-expand-all
        />
        <NInput
          v-model:value="query.keyword"
          placeholder="编号 / 姓名 / 手机号"
          class="w-220px"
          clearable
          @keyup.enter="handleSearch"
        />
        <NButton size="small" type="primary" @click="handleSearch">搜索</NButton>
        <NButton size="small" @click="handleReset">重置</NButton>
      </NSpace>

      <NDataTable
        remote
        :columns="columns"
        :data="list"
        :loading="loading"
        :pagination="pagination"
        :row-key="row => row.id"
        :bordered="false"
        :single-line="false"
        :scroll-x="1200"
        size="small"
      />
    </NCard>

    <!-- 新增 / 编辑员工 -->
    <NModal v-model:show="modalVisible" preset="card" :title="modalTitle" class="w-640px">
      <NForm ref="formRef" :model="form" :rules="rules" label-placement="top">
        <NGrid :x-gap="16" :y-gap="0" responsive="screen" item-responsive>
          <NGridItem span="24 s:12">
            <NFormItem label="员工编号" path="empCode">
              <NInput v-model:value="form.empCode" placeholder="如：E001，需唯一" clearable />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="姓名" path="empName">
              <NInput v-model:value="form.empName" placeholder="请输入员工姓名" clearable />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="所属部门" path="deptId">
              <NTreeSelect
                v-model:value="form.deptId"
                :options="deptOptions"
                placeholder="请选择所属部门"
                filterable
                default-expand-all
              />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="岗位">
              <NInput v-model:value="form.position" placeholder="如：会计" clearable />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="手机号">
              <NInput v-model:value="form.phone" placeholder="选填" clearable />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="邮箱">
              <NInput v-model:value="form.email" placeholder="选填" clearable />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="入职日期">
              <NDatePicker
                v-model:formatted-value="form.hireDate"
                type="date"
                value-format="yyyy-MM-dd"
                class="w-full"
                clearable
              />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="在职状态">
              <NSwitch v-model:value="form.isActive">
                <template #checked>在职</template>
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

    <!-- 菜单权限（可选范围受限于所属部门） -->
    <NModal
      v-model:show="menuVisible"
      preset="card"
      class="w-560px"
      :title="`菜单权限 · ${menuEmployee?.empName || ''}`"
    >
      <NSpin :show="menuLoading">
        <NAlert type="warning" :bordered="false" class="mb-12px">
          可选范围受限于「{{ menuEmployee?.deptName || '所属部门' }}」已授权的菜单，超出部门范围的权限不会生效。
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
          <NEmpty v-if="!menuLoading && !menuOptions.length" size="small" description="所属部门尚未授权任何菜单" />
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
