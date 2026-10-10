<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, FormInst, FormRules, TreeOption } from 'naive-ui';
import { NButton, NSpace, NTag } from 'naive-ui';
import {
  bindUserEmployee,
  fetchDeleteUser,
  fetchDepartments,
  fetchEmployeeOptions,
  fetchUsers,
  resetUserPassword,
  saveUser
} from '@/service/api';
import {
  roleLabel,
  roleOptions,
  toDeptTreeOptions,
  userStatusLabel,
  userStatusOptions,
  userStatusTagType
} from '../shared';

defineOptions({
  name: 'SettingUser'
});

const loading = ref(false);
const submitting = ref(false);

const list = ref<Api.System.UserItem[]>([]);
const total = ref(0);
const deptTree = ref<Api.System.DepartmentItem[]>([]);
const employeeOptions = ref<Api.System.EmployeeOption[]>([]);

const query = reactive({
  keyword: '',
  page: 1,
  pageSize: 20
});

const deptOptions = computed<TreeOption[]>(() => toDeptTreeOptions(deptTree.value));

/** 绑定员工下拉：额外提供一个「解绑」选项 */
const bindOptions = computed(() => [
  { label: '不绑定（解绑）', value: 0 },
  ...employeeOptions.value.map(item => ({ label: item.label, value: item.id, disabled: item.disabled }))
]);

function formatTime(value: string | null) {
  if (!value) return '—';

  return String(value).replace('T', ' ').slice(0, 19);
}

/* ------------------------------ 新增 / 编辑 ------------------------------ */

const modalVisible = ref(false);
const formRef = ref<FormInst | null>(null);
const editingId = ref(0);

const form = reactive({
  username: '',
  password: '',
  realName: '',
  phone: '',
  email: '',
  role: 'staff',
  departmentId: null as number | null,
  status: 'active'
});

const modalTitle = computed(() => (editingId.value ? '编辑账号' : '新增账号'));

const rules = computed<FormRules>(() => ({
  username: { required: true, message: '请输入登录用户名', trigger: ['input', 'blur'] },
  password: editingId.value
    ? { min: 6, message: '密码至少 6 位', trigger: ['input', 'blur'] }
    : {
        required: true,
        min: 6,
        message: '请输入至少 6 位的初始密码',
        trigger: ['input', 'blur']
      }
}));

/* ------------------------------ 绑定员工 ------------------------------ */

const bindVisible = ref(false);
const bindSaving = ref(false);
const bindUser = ref<Api.System.UserItem | null>(null);
const bindEmployeeId = ref(0);

/* ------------------------------ 重置密码 ------------------------------ */

const pwdVisible = ref(false);
const pwdSaving = ref(false);
const pwdUser = ref<Api.System.UserItem | null>(null);
const pwdForm = reactive({ password: '', confirm: '' });
const pwdRules: FormRules = {
  password: { required: true, min: 6, message: '请输入至少 6 位的新密码', trigger: ['input', 'blur'] },
  confirm: {
    required: true,
    trigger: ['input', 'blur'],
    validator: (_rule, value: string) => {
      if (!value) return new Error('请再次输入新密码');
      if (value !== pwdForm.password) return new Error('两次输入的密码不一致');

      return true;
    }
  }
};

async function loadDepartments() {
  const { data, error } = await fetchDepartments();

  if (error || !data) return;

  deptTree.value = data.list || [];
}

async function loadUsers() {
  loading.value = true;
  const { data, error } = await fetchUsers({
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
  loadUsers();
}

function handleReset() {
  query.keyword = '';
  query.page = 1;
  loadUsers();
}

function resetForm() {
  form.username = '';
  form.password = '';
  form.realName = '';
  form.phone = '';
  form.email = '';
  form.role = 'staff';
  form.departmentId = null;
  form.status = 'active';
}

function openCreate() {
  editingId.value = 0;
  resetForm();
  modalVisible.value = true;
}

function openEdit(row: Api.System.UserItem) {
  editingId.value = row.id;
  form.username = row.username;
  form.password = '';
  form.realName = row.realName;
  form.phone = row.phone;
  form.email = row.email;
  form.role = row.role || 'staff';
  form.departmentId = row.departmentId || null;
  form.status = row.status || 'active';
  modalVisible.value = true;
}

async function submitForm() {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  const payload: Api.System.UserSaveInput = {
    id: editingId.value,
    username: form.username.trim(),
    password: form.password,
    realName: form.realName,
    phone: form.phone,
    email: form.email,
    role: form.role,
    departmentId: form.departmentId || 0,
    status: form.status
  };

  submitting.value = true;
  const { error } = await saveUser(payload);
  submitting.value = false;

  if (error) return;

  window.$message?.success(editingId.value ? '修改成功' : '新增成功');
  modalVisible.value = false;
  await loadUsers();
}

function handleDelete(row: Api.System.UserItem) {
  window.$dialog?.warning({
    title: '删除账号',
    content: `确认删除登录账号「${row.username}」吗？与之绑定的员工档案会保留，但会解除绑定关系。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { error } = await fetchDeleteUser(row.id);
      if (error) return false;

      window.$message?.success('删除成功');
      await loadUsers();

      return true;
    }
  });
}

/* ------------------------------ 绑定员工操作 ------------------------------ */

async function openBindModal(row: Api.System.UserItem) {
  bindUser.value = row;
  bindEmployeeId.value = row.employeeId || 0;
  bindVisible.value = true;

  const { data, error } = await fetchEmployeeOptions();

  if (error || !data) return;

  employeeOptions.value = data.list || [];
}

async function submitBind() {
  if (!bindUser.value) return;

  bindSaving.value = true;
  const { error } = await bindUserEmployee(bindUser.value.id, bindEmployeeId.value);
  bindSaving.value = false;

  if (error) return;

  window.$message?.success(bindEmployeeId.value ? '绑定成功' : '已解绑');
  bindVisible.value = false;
  await loadUsers();
}

/* ------------------------------ 重置密码操作 ------------------------------ */

function openPwdModal(row: Api.System.UserItem) {
  pwdUser.value = row;
  pwdForm.password = '';
  pwdForm.confirm = '';
  pwdVisible.value = true;
}

async function submitPwd() {
  pwdSaving.value = true;
  const { error } = await resetUserPassword(pwdUser.value?.id || 0, pwdForm.password);
  pwdSaving.value = false;

  if (error) return;

  window.$message?.success('密码已重置');
  pwdVisible.value = false;
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
    loadUsers();
  },
  onUpdatePageSize: (pageSize: number) => {
    query.pageSize = pageSize;
    query.page = 1;
    loadUsers();
  }
}));

const columns: DataTableColumns<Api.System.UserItem> = [
  { key: 'username', title: '登录用户名', width: 140 },
  { key: 'realName', title: '姓名', minWidth: 110, render: row => row.realName || '—' },
  {
    key: 'role',
    title: '角色',
    width: 120,
    render: row => h(NTag, { size: 'small', bordered: false, type: 'primary' }, { default: () => roleLabel(row.role) })
  },
  { key: 'departmentName', title: '所属部门', minWidth: 130, render: row => row.departmentName || '—' },
  { key: 'employeeName', title: '绑定员工', minWidth: 120, render: row => row.employeeName || '未绑定' },
  {
    key: 'status',
    title: '状态',
    width: 90,
    render: row =>
      h(
        NTag,
        { size: 'small', bordered: false, type: userStatusTagType(row.status) },
        { default: () => userStatusLabel(row.status) }
      )
  },
  { key: 'lastLoginAt', title: '最近登录', width: 170, render: row => formatTime(row.lastLoginAt) },
  {
    key: 'action',
    title: '操作',
    width: 280,
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
              { size: 'tiny', quaternary: true, type: 'info', onClick: () => openBindModal(row) },
              { default: () => '绑定员工' }
            ),
            h(
              NButton,
              { size: 'tiny', quaternary: true, type: 'warning', onClick: () => openPwdModal(row) },
              { default: () => '重置密码' }
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
  await loadUsers();
});

/** 重置密码弹窗内部使用的表单实例 */
const pwdFormRef = ref<FormInst | null>(null);

async function handlePwdConfirm() {
  try {
    await pwdFormRef.value?.validate();
  } catch {
    return;
  }

  await submitPwd();
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <NCard title="账号管理" :bordered="false" size="small" class="card-wrapper">
      <template #header-extra>
        <NButton size="small" type="primary" @click="openCreate">
          <template #icon>
            <icon-ic-round-plus class="text-icon" />
          </template>
          新增账号
        </NButton>
      </template>

      <NSpace class="mb-12px" :size="12" align="center" wrap>
        <NInput
          v-model:value="query.keyword"
          placeholder="用户名 / 姓名 / 手机号"
          class="w-240px"
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
        :scroll-x="1300"
        size="small"
      />
    </NCard>

    <!-- 新增 / 编辑账号 -->
    <NModal v-model:show="modalVisible" preset="card" :title="modalTitle" class="w-640px">
      <NForm ref="formRef" :model="form" :rules="rules" label-placement="top">
        <NGrid :x-gap="16" :y-gap="0" responsive="screen" item-responsive>
          <NGridItem span="24 s:12">
            <NFormItem label="登录用户名" path="username">
              <NInput v-model:value="form.username" placeholder="登录时使用，需唯一" clearable />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem :label="editingId ? '密码（留空不修改）' : '初始密码'" path="password">
              <NInput
                v-model:value="form.password"
                type="password"
                show-password-on="click"
                :placeholder="editingId ? '留空表示不修改密码' : '至少 6 位'"
                clearable
              />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="姓名">
              <NInput v-model:value="form.realName" placeholder="选填" clearable />
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
            <NFormItem label="角色">
              <NSelect v-model:value="form.role" :options="roleOptions" />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="所属部门">
              <NTreeSelect
                v-model:value="form.departmentId"
                :options="deptOptions"
                placeholder="选填；绑定员工后会自动同步"
                clearable
                filterable
                default-expand-all
              />
            </NFormItem>
          </NGridItem>
          <NGridItem span="24 s:12">
            <NFormItem label="账号状态">
              <NSelect v-model:value="form.status" :options="userStatusOptions" />
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

    <!-- 绑定员工 -->
    <NModal v-model:show="bindVisible" preset="card" class="w-520px" :title="`绑定员工 · ${bindUser?.username || ''}`">
      <NAlert type="info" :bordered="false" class="mb-12px">
        账号与员工是一对一的。绑定后该账号的可用菜单会按「员工权限 ∩ 部门权限」计算。
      </NAlert>
      <NSelect v-model:value="bindEmployeeId" :options="bindOptions" filterable placeholder="请选择员工" />

      <template #footer>
        <NSpace justify="end" :size="12">
          <NButton @click="bindVisible = false">取消</NButton>
          <NButton type="primary" :loading="bindSaving" @click="submitBind">保存</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 重置密码 -->
    <NModal v-model:show="pwdVisible" preset="card" class="w-480px" :title="`重置密码 · ${pwdUser?.username || ''}`">
      <NForm ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-placement="top">
        <NFormItem label="新密码" path="password">
          <NInput v-model:value="pwdForm.password" type="password" show-password-on="click" placeholder="至少 6 位" />
        </NFormItem>
        <NFormItem label="确认新密码" path="confirm">
          <NInput
            v-model:value="pwdForm.confirm"
            type="password"
            show-password-on="click"
            placeholder="请再次输入新密码"
          />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end" :size="12">
          <NButton @click="pwdVisible = false">取消</NButton>
          <NButton type="primary" :loading="pwdSaving" @click="handlePwdConfirm">确认重置</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>
