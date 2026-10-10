package model

// InitStatus 公司初始化状态
type InitStatus struct {
	Initialized bool   `json:"initialized"`
	CompanyId   uint64 `json:"companyId"`
	CompanyName string `json:"companyName"`
}

// SubjectItem 会计科目项
type SubjectItem struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	Category      string `json:"category"`
	Direction     string `json:"direction"`
	AuxiliaryType string `json:"auxiliaryType"`
	// IsActive 是否启用：1=启用，0=停用（提交初始化时以该字段为准）
	IsActive int `json:"isActive"`
}

// PostingRuleItem 凭证生成映射规则项
type PostingRuleItem struct {
	// BizType 业务类型编码（英文，对应 finance_posting_rule.biz_type）
	BizType string `json:"bizType"`
	// BizTypeName 业务类型中文显示名，仅用于界面展示
	BizTypeName     string `json:"bizTypeName"`
	Direction       string `json:"direction"`
	EntryNo         int    `json:"entryNo"`
	SubjectCode     string `json:"subjectCode"`
	SubjectName     string `json:"subjectName"`
	EntryDirection  string `json:"entryDirection"`
	AmountSource    string `json:"amountSource"`
	SummaryTemplate string `json:"summaryTemplate"`
}

// AccountItem 资金账户项
type AccountItem struct {
	AccountName string  `json:"accountName"`
	AccountType string  `json:"accountType"`
	SubjectCode string  `json:"subjectCode"`
	AccountNo   string  `json:"accountNo"`
	BankName    string  `json:"bankName"`
	Currency    string  `json:"currency"`
	Balance     float64 `json:"balance"`
}

// OpeningBalanceItem 期初余额项
type OpeningBalanceItem struct {
	SubjectCode   string  `json:"subjectCode"`
	Period        string  `json:"period"`
	DebitBalance  float64 `json:"debitBalance"`
	CreditBalance float64 `json:"creditBalance"`
}

// FixedAssetItem 固定资产项
type FixedAssetItem struct {
	AssetNo                 string  `json:"assetNo"`
	AssetName               string  `json:"assetName"`
	AssetCategory           string  `json:"assetCategory"`
	Spec                    string  `json:"spec"`
	Unit                    string  `json:"unit"`
	Quantity                float64 `json:"quantity"`
	OriginalValue           float64 `json:"originalValue"`
	AccumulatedDepreciation float64 `json:"accumulatedDepreciation"`
	NetValue                float64 `json:"netValue"`
	ResidualRate            float64 `json:"residualRate"`
	UsefulLifeMonths        int     `json:"usefulLifeMonths"`
	StartUseDate            string  `json:"startUseDate"`
	Remark                  string  `json:"remark"`
}

// InitOptions 初始化页所需的默认数据
type InitOptions struct {
	AccountingStandards  []string          `json:"accountingStandards"`
	SuggestedStartPeriod string            `json:"suggestedStartPeriod"`
	DefaultSubjects      []SubjectItem     `json:"defaultSubjects"`
	DefaultPostingRules  []PostingRuleItem `json:"defaultPostingRules"`
	DefaultAccounts      []AccountItem     `json:"defaultAccounts"`
}

// InitInput 初始化提交数据
type InitInput struct {
	CompanyName        string `json:"companyName"`
	TaxNo              string `json:"taxNo"`
	AccountingStandard string `json:"accountingStandard"`
	BaseCurrency       string `json:"baseCurrency"`
	StartPeriod        string `json:"startPeriod"`
	// Subjects 会计科目清单（含微调结果）；为空时使用系统内置默认科目表
	Subjects        []SubjectItem        `json:"subjects"`
	Accounts        []AccountItem        `json:"accounts"`
	OpeningBalances []OpeningBalanceItem `json:"openingBalances"`
	FixedAssets     []FixedAssetItem     `json:"fixedAssets"`
}
