import { request } from '../request';

/** 查询公司初始化状态 */
export function fetchCompanyInitStatus() {
  return request<Api.Company.InitStatus>({ url: '/api/company/initStatus' });
}

/** 获取初始化默认数据（默认科目、默认映射、默认账户） */
export function fetchCompanyInitOptions() {
  return request<Api.Company.InitOptions>({ url: '/api/company/initOptions' });
}

/** 提交公司初始配置 */
export function submitCompanyInit(data: Api.Company.InitPayload) {
  return request<Record<string, never>>({
    url: '/api/company/init',
    method: 'post',
    data
  });
}
