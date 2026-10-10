<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns, FormInst, FormItemRule, FormRules } from 'naive-ui';
import { NInput, NSwitch } from 'naive-ui';
import { fetchCompanyInitOptions, submitCompanyInit } from '@/service/api';
import { useInitStore } from '@/store/modules/init';
import { useTabStore } from '@/store/modules/tab';
import { useRouterPush } from '@/hooks/common/router';

defineOptions({
  name: 'CompanyInit'
});

interface EditableSubject {
  code: string;
  name: string;
  category: Api.Company.SubjectItem['category'];
  direction: Api.Company.SubjectItem['direction'];
  auxiliaryType: string;
  isActive: boolean;
}

interface AccountRow {
  accountName: string;
  accountType: Api.Company.AccountItem['accountType'];
  subjectCode: string;
  accountNo: string;
  bankName: string;
  currency: string;
  balance: number | null;
}

interface OpeningBalanceRow {
  subjectCode: string;
  debitBalance: number | null;
  creditBalance: number | null;
}

interface FixedAssetRow {
  assetNo: string;
  assetName: string;
  assetCategory: string;
  quantity: number | null;
  originalValue: number | null;
  accumulatedDepreciation: number | null;
  netValue: number | null;
  usefulLifeMonths: number | null;
  startUseDate: string | null;
  remark: string;
}

interface PostingRuleRow extends Api.Company.PostingRuleItem {
  /** 对应科目是否启用，停用时该规则不会生成 */
  effective: boolean;
}

/** 固定资产 / 累计折旧 默认科目编码 */
const FIXED_ASSET_CODE = '1601';
const ACCUMULATED_DEPRECIATION_CODE = '1602';
/** 差额补平使用的科目（实收资本） */
const BALANCE_OFF_CODE = '4001';

const steps = [
  { key: 'company', title: '公司信息', desc: '填写公司基础资料与账套启用信息' },
  { key: 'subjects', title: '会计科目', desc: '系统已按会计准则预置科目，可预览与微调' },
  { key: 'accounts', title: '资金账户', desc: '录入现金、银行、第三方支付账户' },
  { key: 'opening', title: '期初余额', desc: '系统已自动汇总，核对并补充差额' },
  { key: 'assets', title: '固定资产', desc: '录入已存在的固定资产卡片（可选，可跳过）' },
  { key: 'rules', title: '默认映射', desc: '系统自动生成的记账规则，确认即可（可跳过）' }
];

const categoryMap: Record<string, string> = {
  asset: '资产',
  liability: '负债',
  equity: '所有者权益',
  cost: '成本',
  profit_loss: '损益'
};

const directionMap: Record<string, string> = { in: '收入', out: '支出' };
const entryDirectionMap: Record<string, string> = { debit: '借', credit: '贷' };

const initStore = useInitStore();
const tabStore = useTabStore();
const { routerPushByKey } = useRouterPush();

const companyFormRef = ref<FormInst | null>(null);
const accountsFormRef = ref<FormInst | null>(null);

const current = ref(0);
const loading = ref(false);
const submitting = ref(false);
const options = ref<Api.Company.InitOptions | null>(null);
const subjects = ref<EditableSubject[]>([]);
const openingBuilt = ref(false);

const form = reactive({
  companyName: '',
  taxNo: '',
  accountingStandard: '',
  baseCurrency: 'CNY',
  startPeriod: null as string | null,
  accounts: [] as AccountRow[],
  openingBalances: [] as OpeningBalanceRow[],
  fixedAssets: [] as FixedAssetRow[]
});

const companyRules: FormRules = {
  companyName: { required: true, message: '请输入公司名称', trigger: ['input', 'blur'] },
  startPeriod: { required: true, message: '请选择启用期间', trigger: ['change', 'blur'] }
};

const accountNameRule: FormItemRule = { required: true, message: '请输入账户名称', trigger: ['input', 'blur'] };
const accountSubjectRule: FormItemRule = { required: true, message: '请选择对应科目', trigger: ['change', 'blur'] };
const openingSubjectRule: FormItemRule = { required: true, message: '请选择会计科目', trigger: ['change', 'blur'] };

/* ------------------------------ 科目相关 ------------------------------ */

const allSubjectMap = computed(() => {
  const map = new Map<string, EditableSubject>();
  subjects.value.forEach(item => map.set(item.code, item));
  return map;
});

const activeSubjects = computed(() => subjects.value.filter(item => item.isActive));

const subjectOptions = computed(() =>
  activeSubjects.value.map(item => ({ label: `${item.code} ${item.name}`, value: item.code }))
);

const standardOptions = computed(() =>
  (options.value?.accountingStandards || []).map(item => ({ label: item, value: item }))
);

const accountTypeOptions = [
  { label: '银行账户', value: 'bank' },
  { label: '现金账户', value: 'cash' },
  { label: '第三方支付', value: 'third_party' }
];

const subjectColumns = computed<DataTableColumns<EditableSubject>>(() => [
  { key: 'code', title: '科目编码', width: 110 },
  {
    key: 'name',
    title: '科目名称',
    minWidth: 220,
    render: row =>
      h(NInput, {
        value: row.name,
        size: 'small',
        placeholder: '请输入科目名称',
        'onUpdate:value': (value: string) => {
          row.name = value;
        }
      })
  },
  { key: 'category', title: '类别', width: 110, render: row => categoryMap[row.category] || row.category },
  { key: 'direction', title: '余额方向', width: 100, render: row => entryDirectionMap[row.direction] || row.direction },
  {
    key: 'isActive',
    title: '启用',
    width: 90,
    render: row =>
      h(NSwitch, {
        value: row.isActive,
        size: 'small',
        'onUpdate:value': (value: boolean) => {
          row.isActive = value;
        }
      })
  }
]);

/* ------------------------------ 默认映射 ------------------------------ */

const postingRules = computed(() =>
  (options.value?.defaultPostingRules || []).map(rule => {
    const subject = allSubjectMap.value.get(rule.subjectCode);
    return {
      ...rule,
      subjectName: subject?.name || rule.subjectName,
      effective: Boolean(subject?.isActive)
    };
  })
);

const bizTypeCount = computed(() => new Set(postingRules.value.map(item => item.bizType)).size);

const postingRuleColumns = computed<DataTableColumns<PostingRuleRow>>(() => [
  {
    key: 'bizType',
    title: '业务类型',
    width: 150,
    render: row => (row.bizTypeName ? `${row.bizTypeName} · ${row.bizType}` : row.bizType)
  },
  { key: 'direction', title: '资金方向', width: 100, render: row => directionMap[row.direction] || row.direction },
  { key: 'entryNo', title: '分录', width: 70 },
  { key: 'subject', title: '会计科目', minWidth: 180, render: row => `${row.subjectCode} ${row.subjectName}` },
  {
    key: 'entryDirection',
    title: '借贷方向',
    width: 100,
    render: row => entryDirectionMap[row.entryDirection] || row.entryDirection
  },
  { key: 'summaryTemplate', title: '摘要模板', minWidth: 160 },
  {
    key: 'effective',
    title: '状态',
    width: 90,
    render: row => (row.effective ? '生效' : '将跳过')
  }
]);

/* ------------------------------ 校验状态 ------------------------------ */

const companyValid = computed(() => Boolean(form.companyName.trim()) && Boolean(form.startPeriod));

const subjectsValid = computed(() => activeSubjects.value.length > 0);

const validAccounts = computed(() =>
  form.accounts.filter(item => Boolean(item.accountName.trim()) && Boolean(item.subjectCode))
);

const accountsValid = computed(() => validAccounts.value.length > 0);

const balanceTotal = computed(() => {
  let debit = 0;
  let credit = 0;
  form.openingBalances.forEach(row => {
    debit += Number(row.debitBalance) || 0;
    credit += Number(row.creditBalance) || 0;
  });
  return { debit, credit };
});

const balanceDiff = computed(() => balanceTotal.value.debit - balanceTotal.value.credit);

const balanceBalanced = computed(() => Math.abs(balanceDiff.value) < 0.001);

/** 各步骤是否已满足自身要求（用于解锁后续步骤） */
const stepValid = computed(() => [
  companyValid.value,
  subjectsValid.value,
  accountsValid.value,
  true, // 期初余额：平衡性在完成初始化时统一校验
  true, // 固定资产：可选
  true // 默认映射：可选
]);

const progressPercent = computed(() => Math.round(((current.value + 1) / steps.length) * 100));

function isStepUnlocked(index: number) {
  for (let i = 0; i < index; i += 1) {
    if (!stepValid.value[i]) return false;
  }
  return true;
}

/* ------------------------------ 数据加载 ------------------------------ */

async function loadOptions() {
  loading.value = true;
  const { data, error } = await fetchCompanyInitOptions();
  loading.value = false;

  if (error || !data) {
    return;
  }

  options.value = data;
  form.accountingStandard = data.accountingStandards[0] || '';
  form.startPeriod = data.suggestedStartPeriod;
  subjects.value = data.defaultSubjects.map(item => ({
    code: item.code,
    name: item.name,
    category: item.category,
    direction: item.direction,
    auxiliaryType: item.auxiliaryType,
    isActive: item.isActive !== 0
  }));
  form.accounts = data.defaultAccounts.map(item => ({
    accountName: item.accountName,
    accountType: item.accountType,
    subjectCode: item.subjectCode,
    accountNo: item.accountNo,
    bankName: item.bankName,
    currency: item.currency || 'CNY',
    balance: null
  }));
}

function enableAllSubjects() {
  subjects.value.forEach(item => {
    item.isActive = true;
  });
}

/* ------------------------------ 期初余额汇总 ------------------------------ */

/** 根据资金账户与固定资产自动汇总期初余额 */
function buildOpeningBalances() {
  const map = new Map<string, { debit: number; credit: number }>();
  const add = (code: string, debit: number, credit: number) => {
    if (!code) return;
    const item = map.get(code) || { debit: 0, credit: 0 };
    item.debit += debit;
    item.credit += credit;
    map.set(code, item);
  };

  // 资金账户余额
  form.accounts.forEach(account => {
    const amount = Number(account.balance) || 0;
    if (!account.subjectCode || !amount) return;
    const subject = allSubjectMap.value.get(account.subjectCode);
    if (subject?.direction === 'credit') add(account.subjectCode, 0, amount);
    else add(account.subjectCode, amount, 0);
  });

  // 固定资产原值与累计折旧
  form.fixedAssets.forEach(asset => {
    const original = Number(asset.originalValue) || 0;
    const depreciation = Number(asset.accumulatedDepreciation) || 0;
    if (original) add(FIXED_ASSET_CODE, original, 0);
    if (depreciation) add(ACCUMULATED_DEPRECIATION_CODE, 0, depreciation);
  });

  form.openingBalances = Array.from(map.entries()).map(([code, value]) => ({
    subjectCode: code,
    debitBalance: value.debit || null,
    creditBalance: value.credit || null
  }));
  openingBuilt.value = true;
}

function rebuildOpeningBalances() {
  buildOpeningBalances();
  window.$message?.info('已按资金账户与固定资产重新汇总期初余额');
}

function addOpeningBalance() {
  form.openingBalances.push({ subjectCode: '', debitBalance: null, creditBalance: null });
}

function removeOpeningBalance(index: number) {
  form.openingBalances.splice(index, 1);
}

/** 将借贷差额补平到实收资本科目 */
function balanceOff() {
  const diff = balanceDiff.value;
  if (Math.abs(diff) < 0.001) return;

  const exist = form.openingBalances.find(row => row.subjectCode === BALANCE_OFF_CODE);
  if (exist) {
    if (diff > 0) exist.creditBalance = (Number(exist.creditBalance) || 0) + diff;
    else exist.debitBalance = (Number(exist.debitBalance) || 0) + Math.abs(diff);
    return;
  }

  form.openingBalances.push({
    subjectCode: BALANCE_OFF_CODE,
    debitBalance: diff < 0 ? Math.abs(diff) : null,
    creditBalance: diff > 0 ? diff : null
  });
}

/* ------------------------------ 明细行操作 ------------------------------ */

function addAccount() {
  form.accounts.push({
    accountName: '',
    accountType: 'bank',
    subjectCode: '',
    accountNo: '',
    bankName: '',
    currency: form.baseCurrency || 'CNY',
    balance: null
  });
}

function removeAccount(index: number) {
  form.accounts.splice(index, 1);
}

function addFixedAsset() {
  form.fixedAssets.push({
    assetNo: '',
    assetName: '',
    assetCategory: '',
    quantity: 1,
    originalValue: null,
    accumulatedDepreciation: null,
    netValue: null,
    usefulLifeMonths: null,
    startUseDate: null,
    remark: ''
  });
}

function removeFixedAsset(index: number) {
  form.fixedAssets.splice(index, 1);
}

/* ------------------------------ 步骤流转 ------------------------------ */

async function validateStep(index: number): Promise<boolean> {
  const key = steps[index].key;

  if (key === 'company') {
    try {
      await companyFormRef.value?.validate();
    } catch {
      window.$message?.warning('请完善公司信息');
      return false;
    }
    return true;
  }

  if (key === 'subjects') {
    if (!subjectsValid.value) {
      window.$message?.warning('请至少启用一个会计科目');
      return false;
    }
    return true;
  }

  if (key === 'accounts') {
    if (!form.accounts.length) {
      window.$message?.warning('请至少添加一个资金账户');
      return false;
    }
    try {
      await accountsFormRef.value?.validate();
    } catch {
      window.$message?.warning('请完善资金账户信息');
      return false;
    }
    if (!accountsValid.value) {
      window.$message?.warning('请至少填写一个完整的资金账户');
      return false;
    }
    return true;
  }

  if (key === 'opening') {
    if (form.openingBalances.some(row => !row.subjectCode)) {
      window.$message?.warning('请为每一条期初余额选择会计科目');
      return false;
    }
    return true;
  }

  return true;
}

function goStep(index: number) {
  if (index < 0 || index >= steps.length) return;

  if (!isStepUnlocked(index)) {
    window.$message?.warning('请先完成前面的必填步骤');
    return;
  }

  current.value = index;

  if (steps[index].key === 'opening' && !openingBuilt.value) {
    buildOpeningBalances();
  }
}

async function handleNext() {
  if (!(await validateStep(current.value))) return;

  const next = current.value + 1;
  if (next >= steps.length) return;

  goStep(next);
}

async function handleSubmit() {
  if (!companyValid.value) {
    goStep(0);
    window.$message?.warning('请完善公司信息');
    return;
  }
  if (!subjectsValid.value) {
    goStep(1);
    window.$message?.warning('请至少启用一个会计科目');
    return;
  }
  if (!accountsValid.value) {
    goStep(2);
    window.$message?.warning('请至少填写一个完整的资金账户');
    return;
  }
  if (!balanceBalanced.value) {
    goStep(3);
    window.$message?.warning('期初余额借贷不平衡，请调整后再提交');
    return;
  }

  const startPeriod = form.startPeriod as string;

  const payload: Api.Company.InitPayload = {
    companyName: form.companyName.trim(),
    taxNo: form.taxNo,
    accountingStandard: form.accountingStandard,
    baseCurrency: form.baseCurrency || 'CNY',
    startPeriod,
    subjects: subjects.value.map(item => ({
      code: item.code,
      name: item.name,
      category: item.category,
      direction: item.direction,
      auxiliaryType: item.auxiliaryType,
      isActive: item.isActive ? 1 : 0
    })),
    accounts: form.accounts
      .filter(item => item.accountName.trim())
      .map(item => ({
        accountName: item.accountName.trim(),
        accountType: item.accountType,
        subjectCode: item.subjectCode,
        accountNo: item.accountNo,
        bankName: item.bankName,
        currency: item.currency || 'CNY',
        balance: Number(item.balance) || 0
      })),
    openingBalances: form.openingBalances
      .filter(item => item.subjectCode)
      .map(item => ({
        subjectCode: item.subjectCode,
        period: startPeriod,
        debitBalance: Number(item.debitBalance) || 0,
        creditBalance: Number(item.creditBalance) || 0
      })),
    fixedAssets: form.fixedAssets
      .filter(item => item.assetName.trim())
      .map(item => ({
        assetNo: item.assetNo.trim(),
        assetName: item.assetName.trim(),
        assetCategory: item.assetCategory,
        spec: '',
        unit: '',
        quantity: Number(item.quantity) || 1,
        originalValue: Number(item.originalValue) || 0,
        accumulatedDepreciation: Number(item.accumulatedDepreciation) || 0,
        netValue: Number(item.netValue) || 0,
        residualRate: 0,
        usefulLifeMonths: Number(item.usefulLifeMonths) || 0,
        startUseDate: item.startUseDate || '',
        remark: item.remark
      }))
  };

  submitting.value = true;
  const { error } = await submitCompanyInit(payload);
  submitting.value = false;

  if (error) {
    return;
  }

  initStore.markInitialized();
  window.$message?.success('初始化完成，正在进入系统');

  await routerPushByKey('root');
  // 初始化页签不再需要，离开后移除
  await tabStore.removeTabByRouteName('init');
}

onMounted(loadOptions);
</script>

<template>
  <div class="relative min-h-screen overflow-hidden bg-layout">
    <!-- 背景点缀 -->
    <div class="pointer-events-none absolute inset-0">
      <div class="decor decor-grid"></div>
      <div class="decor decor-a"></div>
      <div class="decor decor-b"></div>
    </div>
    <div class="relative p-24px">
      <div class="mx-auto w-full max-w-[1080px]">
        <div class="mb-24px flex flex-col items-center gap-8px">
          <SystemLogo class="h-40px w-40px" />
          <NText strong class="text-20px">系统初始配置</NText>
          <NText depth="3" class="text-13px">请按顺序完成以下 {{ steps.length }} 步，完成后即可进入系统</NText>
        </div>
        <NSpace vertical :size="16">
          <!-- 进度与步骤导航 -->
          <NCard :bordered="false" size="small" class="card-wrapper">
            <div class="mb-12px flex flex-wrap items-center justify-between gap-8px">
              <NText strong>第 {{ current + 1 }} 步 / 共 {{ steps.length }} 步 · {{ steps[current].title }}</NText>
              <NTag size="small" :bordered="false" type="primary">{{ progressPercent }}%</NTag>
            </div>
            <NProgress
              type="line"
              :percentage="progressPercent"
              :show-indicator="false"
              :height="6"
              :border-radius="3"
            />
            <div class="mt-16px">
              <NSteps :current="current + 1" size="small">
                <NStep v-for="(step, index) in steps" :key="step.key" :disabled="!isStepUnlocked(index)">
                  <template #title>
                    <span class="cursor-pointer" @click="goStep(index)">{{ step.title }}</span>
                  </template>
                </NStep>
              </NSteps>
            </div>
          </NCard>

          <!-- 步骤 1：公司信息 -->
          <NCard
            v-if="steps[current].key === 'company'"
            title="公司信息"
            :bordered="false"
            size="small"
            segmented
            class="card-wrapper"
          >
            <template #header-extra>
              <NText depth="3" class="text-12px">带 * 为必填项</NText>
            </template>
            <NForm ref="companyFormRef" :model="form" :rules="companyRules" label-placement="top">
              <NGrid :x-gap="16" :y-gap="0" responsive="screen" item-responsive>
                <NGridItem span="24 s:12 m:8">
                  <NFormItem label="公司名称" path="companyName">
                    <NInput v-model:value="form.companyName" placeholder="请输入公司全称" clearable />
                  </NFormItem>
                </NGridItem>
                <NGridItem span="24 s:12 m:8">
                  <NFormItem label="纳税人识别号" path="taxNo">
                    <NInput v-model:value="form.taxNo" placeholder="请输入纳税人识别号" clearable />
                  </NFormItem>
                </NGridItem>
                <NGridItem span="24 s:12 m:8">
                  <NFormItem label="会计准则" path="accountingStandard">
                    <NSelect
                      v-model:value="form.accountingStandard"
                      :options="standardOptions"
                      placeholder="请选择会计准则"
                    />
                  </NFormItem>
                </NGridItem>
                <NGridItem span="24 s:12 m:8">
                  <NFormItem label="记账本位币" path="baseCurrency">
                    <NInput v-model:value="form.baseCurrency" placeholder="CNY" />
                  </NFormItem>
                </NGridItem>
                <NGridItem span="24 s:12 m:8">
                  <NFormItem label="启用期间" path="startPeriod">
                    <NDatePicker
                      v-model:formatted-value="form.startPeriod"
                      type="month"
                      value-format="yyyy-MM"
                      class="w-full"
                      clearable
                    />
                  </NFormItem>
                </NGridItem>
              </NGrid>
            </NForm>
          </NCard>

          <!-- 步骤 2：会计科目 -->
          <NCard
            v-else-if="steps[current].key === 'subjects'"
            title="会计科目"
            :bordered="false"
            size="small"
            segmented
            class="card-wrapper"
          >
            <template #header-extra>
              <NSpace :size="12" align="center">
                <NTag size="small" :bordered="false" type="info">
                  已启用 {{ activeSubjects.length }} / {{ subjects.length }}
                </NTag>
                <NButton size="small" quaternary @click="enableAllSubjects">全部启用</NButton>
              </NSpace>
            </template>
            <NAlert type="info" :bordered="false" class="mb-12px">
              系统已根据「{{
                form.accountingStandard || '所选会计准则'
              }}」预置常用科目，可直接下一步；如需调整，可修改科目名称或停用不用的科目。
            </NAlert>
            <NDataTable
              :columns="subjectColumns"
              :data="subjects"
              :loading="loading"
              :row-key="row => row.code"
              :bordered="false"
              :max-height="420"
              size="small"
            />
          </NCard>

          <!-- 步骤 3：资金账户 -->
          <NCard
            v-else-if="steps[current].key === 'accounts'"
            title="资金账户"
            :bordered="false"
            size="small"
            segmented
            class="card-wrapper"
          >
            <template #header-extra>
              <NButton size="small" type="primary" ghost @click="addAccount">添加账户</NButton>
            </template>
            <NForm ref="accountsFormRef" :model="form" label-placement="top">
              <NSpace vertical :size="12">
                <div v-for="(row, index) in form.accounts" :key="index" class="rounded-6px bg-layout p-12px">
                  <div class="mb-8px flex items-center justify-between">
                    <NText depth="3" class="text-13px">账户 {{ index + 1 }}</NText>
                    <NButton size="tiny" quaternary type="error" @click="removeAccount(index)">移除</NButton>
                  </div>
                  <NGrid :x-gap="12" :y-gap="0" responsive="screen" item-responsive>
                    <NGridItem span="24 s:12 m:6">
                      <NFormItem label="账户名称" :path="`accounts.${index}.accountName`" :rule="accountNameRule">
                        <NInput v-model:value="row.accountName" placeholder="如：基本存款账户" />
                      </NFormItem>
                    </NGridItem>
                    <NGridItem span="24 s:12 m:4">
                      <NFormItem label="账户类型" :show-feedback="false">
                        <NSelect v-model:value="row.accountType" :options="accountTypeOptions" />
                      </NFormItem>
                    </NGridItem>
                    <NGridItem span="24 s:12 m:5">
                      <NFormItem label="对应科目" :path="`accounts.${index}.subjectCode`" :rule="accountSubjectRule">
                        <NSelect
                          v-model:value="row.subjectCode"
                          :options="subjectOptions"
                          filterable
                          placeholder="请选择科目"
                        />
                      </NFormItem>
                    </NGridItem>
                    <NGridItem span="24 s:12 m:5">
                      <NFormItem label="期初余额" :show-feedback="false">
                        <NInputNumber
                          v-model:value="row.balance"
                          class="w-full"
                          :precision="2"
                          :show-button="false"
                          placeholder="0.00"
                        />
                      </NFormItem>
                    </NGridItem>
                    <NGridItem span="24 s:12 m:8">
                      <NFormItem label="开户行" :show-feedback="false">
                        <NInput v-model:value="row.bankName" placeholder="非银行账户可留空" />
                      </NFormItem>
                    </NGridItem>
                    <NGridItem span="24 s:12 m:8">
                      <NFormItem label="账号" :show-feedback="false">
                        <NInput v-model:value="row.accountNo" placeholder="银行账号 / 第三方账号" />
                      </NFormItem>
                    </NGridItem>
                    <NGridItem span="24 s:12 m:4">
                      <NFormItem label="币种" :show-feedback="false">
                        <NInput v-model:value="row.currency" placeholder="CNY" />
                      </NFormItem>
                    </NGridItem>
                  </NGrid>
                </div>

                <NEmpty v-if="!form.accounts.length" size="small" description="请至少添加一个资金账户" />
              </NSpace>
            </NForm>
          </NCard>

          <!-- 步骤 4：期初余额 -->
          <NCard
            v-else-if="steps[current].key === 'opening'"
            title="期初余额"
            :bordered="false"
            size="small"
            segmented
            class="card-wrapper"
          >
            <template #header-extra>
              <NSpace :size="12" align="center">
                <NTag :type="balanceBalanced ? 'success' : 'error'" size="small" :bordered="false">
                  借 {{ balanceTotal.debit.toFixed(2) }} / 贷 {{ balanceTotal.credit.toFixed(2) }}
                </NTag>
                <NButton size="small" quaternary @click="rebuildOpeningBalances">重新汇总</NButton>
                <NButton size="small" type="primary" ghost @click="addOpeningBalance">添加科目</NButton>
              </NSpace>
            </template>

            <NAlert v-if="!balanceBalanced" type="warning" :bordered="false" class="mb-12px">
              <div class="flex flex-wrap items-center justify-between gap-8px">
                <span>
                  录入了资金账户与固定资产，借方与贷方差额 {{ Math.abs(balanceDiff).toFixed(2) }}，请补充对方科目
                </span>
                <NButton size="tiny" type="primary" @click="balanceOff">差额计入实收资本</NButton>
              </div>
            </NAlert>
            <NAlert v-else type="success" :bordered="false" class="mb-12px">借贷已平衡，可直接进入下一步</NAlert>

            <NForm :model="form" label-placement="top">
              <NSpace vertical :size="12">
                <div v-for="(row, index) in form.openingBalances" :key="index" class="rounded-6px bg-layout p-12px">
                  <div class="mb-8px flex items-center justify-between">
                    <NText depth="3" class="text-13px">科目余额 {{ index + 1 }}</NText>
                    <NButton size="tiny" quaternary type="error" @click="removeOpeningBalance(index)">移除</NButton>
                  </div>
                  <div class="flex flex-wrap gap-12px">
                    <NFormItem
                      label="会计科目"
                      :path="`openingBalances.${index}.subjectCode`"
                      :rule="openingSubjectRule"
                      class="min-w-[260px] flex-1"
                    >
                      <NSelect
                        v-model:value="row.subjectCode"
                        :options="subjectOptions"
                        filterable
                        placeholder="请选择科目"
                      />
                    </NFormItem>
                    <NFormItem label="借方余额" :show-feedback="false" class="w-[180px]">
                      <NInputNumber
                        v-model:value="row.debitBalance"
                        class="w-full"
                        :precision="2"
                        :show-button="false"
                        placeholder="0.00"
                      />
                    </NFormItem>
                    <NFormItem label="贷方余额" :show-feedback="false" class="w-[180px]">
                      <NInputNumber
                        v-model:value="row.creditBalance"
                        class="w-full"
                        :precision="2"
                        :show-button="false"
                        placeholder="0.00"
                      />
                    </NFormItem>
                  </div>
                </div>

                <NEmpty v-if="!form.openingBalances.length" size="small" description="暂无需录入的期初余额" />
              </NSpace>
            </NForm>
          </NCard>

          <!-- 步骤 5：固定资产 -->
          <NCard
            v-else-if="steps[current].key === 'assets'"
            title="固定资产"
            :bordered="false"
            size="small"
            segmented
            class="card-wrapper"
          >
            <template #header-extra>
              <NButton size="small" type="primary" ghost @click="addFixedAsset">添加资产</NButton>
            </template>
            <NAlert v-if="!form.fixedAssets.length" type="info" :bordered="false" class="mb-12px">
              如果公司没有固定资产，可直接进入下一步
            </NAlert>
            <NSpace vertical :size="12">
              <div v-for="(row, index) in form.fixedAssets" :key="index" class="rounded-6px bg-layout p-12px">
                <div class="mb-8px flex items-center justify-between">
                  <NText depth="3" class="text-13px">资产 {{ index + 1 }}</NText>
                  <NButton size="tiny" quaternary type="error" @click="removeFixedAsset(index)">移除</NButton>
                </div>
                <NGrid :x-gap="12" :y-gap="0" responsive="screen" item-responsive>
                  <NGridItem span="24 s:12 m:6">
                    <NFormItem label="资产名称" :show-feedback="false">
                      <NInput v-model:value="row.assetName" placeholder="如：办公电脑" />
                    </NFormItem>
                  </NGridItem>
                  <NGridItem span="24 s:12 m:4">
                    <NFormItem label="资产编号" :show-feedback="false">
                      <NInput v-model:value="row.assetNo" placeholder="留空自动生成" />
                    </NFormItem>
                  </NGridItem>
                  <NGridItem span="24 s:12 m:4">
                    <NFormItem label="资产类别" :show-feedback="false">
                      <NInput v-model:value="row.assetCategory" placeholder="如：电子设备" />
                    </NFormItem>
                  </NGridItem>
                  <NGridItem span="24 s:12 m:3">
                    <NFormItem label="数量" :show-feedback="false">
                      <NInputNumber v-model:value="row.quantity" class="w-full" :min="0" :show-button="false" />
                    </NFormItem>
                  </NGridItem>
                  <NGridItem span="24 s:12 m:3">
                    <NFormItem label="使用月份" :show-feedback="false">
                      <NInputNumber v-model:value="row.usefulLifeMonths" class="w-full" :min="0" :show-button="false" />
                    </NFormItem>
                  </NGridItem>
                  <NGridItem span="24 s:12 m:4">
                    <NFormItem label="原值" :show-feedback="false">
                      <NInputNumber
                        v-model:value="row.originalValue"
                        class="w-full"
                        :precision="2"
                        :show-button="false"
                      />
                    </NFormItem>
                  </NGridItem>
                  <NGridItem span="24 s:12 m:4">
                    <NFormItem label="累计折旧" :show-feedback="false">
                      <NInputNumber
                        v-model:value="row.accumulatedDepreciation"
                        class="w-full"
                        :precision="2"
                        :show-button="false"
                      />
                    </NFormItem>
                  </NGridItem>
                  <NGridItem span="24 s:12 m:4">
                    <NFormItem label="净值" :show-feedback="false">
                      <NInputNumber v-model:value="row.netValue" class="w-full" :precision="2" :show-button="false" />
                    </NFormItem>
                  </NGridItem>
                  <NGridItem span="24 s:12 m:6">
                    <NFormItem label="开始使用日期" :show-feedback="false">
                      <NDatePicker
                        v-model:formatted-value="row.startUseDate"
                        type="date"
                        value-format="yyyy-MM-dd"
                        class="w-full"
                        clearable
                      />
                    </NFormItem>
                  </NGridItem>
                  <NGridItem span="24 s:12 m:6">
                    <NFormItem label="备注" :show-feedback="false">
                      <NInput v-model:value="row.remark" placeholder="选填" />
                    </NFormItem>
                  </NGridItem>
                </NGrid>
              </div>
            </NSpace>
          </NCard>

          <!-- 步骤 6：默认映射 -->
          <NCard v-else title="默认凭证映射" :bordered="false" size="small" segmented class="card-wrapper">
            <template #header-extra>
              <NTag size="small" :bordered="false" type="primary">
                {{ postingRules.length }} 条规则 · {{ bizTypeCount }} 类业务
              </NTag>
            </template>
            <NAlert type="info" :bordered="false" class="mb-12px">
              以下为系统根据业务类型自动生成的记账规则，可直接完成初始化；标为「将跳过」的规则因对应科目已停用而不会生成。
            </NAlert>
            <NDataTable
              :columns="postingRuleColumns"
              :data="postingRules"
              :loading="loading"
              :row-key="row => `${row.bizType}-${row.direction}-${row.entryNo}`"
              :bordered="false"
              :max-height="420"
              size="small"
            />
          </NCard>

          <!-- 底部操作 -->
          <NCard :bordered="false" size="small" class="card-wrapper">
            <div class="flex flex-wrap items-center justify-between gap-12px">
              <NText depth="3" class="text-13px">{{ steps[current].desc }}</NText>
              <NSpace :size="12">
                <NButton :disabled="current === 0" @click="goStep(current - 1)">上一步</NButton>
                <NButton v-if="current < steps.length - 1" type="primary" @click="handleNext">下一步</NButton>
                <NButton v-else type="primary" :loading="submitting" @click="handleSubmit">
                  完成初始化并进入系统
                </NButton>
              </NSpace>
            </div>
          </NCard>
        </NSpace>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 背景点缀：仅做轻量氛围修饰，不参与交互也不影响内容可读性 */

.decor {
  position: absolute;
}

/* 淡点阵网格，自上而下渐隐 */
.decor-grid {
  inset: 0;
  background-image: radial-gradient(rgb(var(--primary-300-color) / 0.4) 1px, transparent 1px);
  background-size: 22px 22px;
  opacity: 0.3;
  mask-image: linear-gradient(to bottom, #000 0, transparent 320px);
}

/* 左上柔和光晕 */
.decor-a {
  top: -160px;
  left: -140px;
  width: 420px;
  height: 420px;
  border-radius: 9999px;
  background: radial-gradient(circle, rgb(var(--primary-400-color) / 0.16), transparent 70%);
}

/* 右侧柔和光晕 */
.decor-b {
  top: 140px;
  right: -160px;
  width: 380px;
  height: 380px;
  border-radius: 9999px;
  background: radial-gradient(circle, rgb(var(--primary-300-color) / 0.14), transparent 70%);
}
</style>
