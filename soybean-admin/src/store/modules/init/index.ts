import { ref } from 'vue';
import { defineStore } from 'pinia';
import { fetchCompanyInitStatus } from '@/service/api';
import { SetupStoreId } from '@/enum';

/**
 * 公司初始化状态 store
 *
 * 单公司：未初始化时强制进入初始化配置页，初始化完成后放行进入首页
 */
export const useInitStore = defineStore(SetupStoreId.Init, () => {
  /** 公司是否已完成初始化 */
  const initialized = ref(false);

  /** 是否已向服务端确认过状态 */
  const checked = ref(false);

  /** 进行中的请求，避免并发重复调用 */
  let pending: Promise<boolean> | null = null;

  /**
   * 查询（并缓存）初始化状态
   *
   * @param force 是否强制重新查询
   */
  async function check(force = false): Promise<boolean> {
    if (checked.value && !force) {
      return initialized.value;
    }

    if (pending) {
      return pending;
    }

    pending = (async () => {
      const { data, error } = await fetchCompanyInitStatus();

      if (!error && data) {
        initialized.value = Boolean(data.initialized);
      } else {
        // 接口异常时按“已初始化”处理，避免因后端不可用而无法进入系统
        initialized.value = true;
      }

      checked.value = true;
      pending = null;

      return initialized.value;
    })();

    return pending;
  }

  /** 初始化完成后标记状态，避免守卫再次拦截 */
  function markInitialized() {
    initialized.value = true;
    checked.value = true;
  }

  /** 重置状态 */
  function reset() {
    initialized.value = false;
    checked.value = false;
    pending = null;
  }

  return {
    initialized,
    checked,
    check,
    markInitialized,
    reset
  };
});
