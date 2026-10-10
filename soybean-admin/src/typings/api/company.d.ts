declare namespace Api {
  /**
   * namespace Company
   *
   * backend api module: "company"
   */
  namespace Company {
    /** 公司初始化状态 */
    interface InitStatus {
      initialized: boolean;
      companyId: number;
      companyName: string;
    }

    /** 会计科目 */
    interface SubjectItem {
      code: string;
      name: string;
      category: 'asset' | 'liability' | 'equity' | 'cost' | 'profit_loss';
      direction: 'debit' | 'credit';
      auxiliaryType: string;
      /** 是否启用：1=启用，0=停用 */
      isActive: number;
    }

    /** 凭证生成映射规则 */
    interface PostingRuleItem {
      /** 业务类型编码（英文，编码统一以此为准） */
      bizType: string;
      /** 业务类型中文显示名 */
      bizTypeName: string;
      direction: 'in' | 'out';
      entryNo: number;
      subjectCode: string;
      subjectName: string;
      entryDirection: 'debit' | 'credit';
      amountSource: string;
      summaryTemplate: string;
    }

    /** 资金账户 */
    interface AccountItem {
      accountName: string;
      accountType: 'bank' | 'cash' | 'third_party';
      subjectCode: string;
      accountNo: string;
      bankName: string;
      currency: string;
      balance: number;
    }

    /** 期初余额 */
    interface OpeningBalanceItem {
      subjectCode: string;
      period: string;
      debitBalance: number;
      creditBalance: number;
    }

    /** 固定资产 */
    interface FixedAssetItem {
      assetNo: string;
      assetName: string;
      assetCategory: string;
      spec: string;
      unit: string;
      quantity: number;
      originalValue: number;
      accumulatedDepreciation: number;
      netValue: number;
      residualRate: number;
      usefulLifeMonths: number;
      startUseDate: string;
      remark: string;
    }

    /** 初始化页默认数据 */
    interface InitOptions {
      accountingStandards: string[];
      suggestedStartPeriod: string;
      defaultSubjects: SubjectItem[];
      defaultPostingRules: PostingRuleItem[];
      defaultAccounts: AccountItem[];
    }

    /** 初始化提交数据 */
    interface InitPayload {
      companyName: string;
      taxNo: string;
      accountingStandard: string;
      baseCurrency: string;
      startPeriod: string;
      /** 会计科目清单（含微调后的启用状态）；为空则由后端使用内置默认科目表 */
      subjects: SubjectItem[];
      accounts: AccountItem[];
      openingBalances: OpeningBalanceItem[];
      fixedAssets: FixedAssetItem[];
    }
  }
}
