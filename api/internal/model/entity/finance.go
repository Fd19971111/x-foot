// Package entity 财务相关数据表实体定义
package entity

import "github.com/gogf/gf/v2/os/gtime"

// FinanceCompany 公司/账套表 finance_company
type FinanceCompany struct {
	Id                 uint64      `json:"id"`
	CompanyName        string      `json:"companyName"`
	TaxNo              string      `json:"taxNo"`
	AccountingStandard string      `json:"accountingStandard"`
	BaseCurrency       string      `json:"baseCurrency"`
	StartPeriod        string      `json:"startPeriod"`
	IsInitialized      int         `json:"isInitialized"`
	InitializedAt      *gtime.Time `json:"initializedAt"`
	CreatedAt          *gtime.Time `json:"createdAt"`
}

// FinanceSubject 会计科目表 finance_subject
type FinanceSubject struct {
	Id            uint64      `json:"id"`
	SubjectCode   string      `json:"subjectCode"`
	SubjectName   string      `json:"subjectName"`
	ParentId      uint64      `json:"parentId"`
	Level         int         `json:"level"`
	Category      string      `json:"category"`
	Direction     string      `json:"direction"`
	IsLeaf        int         `json:"isLeaf"`
	IsActive      int         `json:"isActive"`
	AuxiliaryType string      `json:"auxiliaryType"`
	Remark        string      `json:"remark"`
	CreatedAt     *gtime.Time `json:"createdAt"`
}

// FinanceAccount 资金账户表 finance_account
type FinanceAccount struct {
	Id          uint64      `json:"id"`
	AccountName string      `json:"accountName"`
	AccountType string      `json:"accountType"`
	SubjectId   uint64      `json:"subjectId"`
	AccountNo   string      `json:"accountNo"`
	BankName    string      `json:"bankName"`
	Currency    string      `json:"currency"`
	Balance     float64     `json:"balance"`
	IsActive    int         `json:"isActive"`
	CreatedAt   *gtime.Time `json:"createdAt"`
}

// FinanceOpeningBalance 期初余额表 finance_opening_balance
type FinanceOpeningBalance struct {
	Id            uint64      `json:"id"`
	SubjectId     uint64      `json:"subjectId"`
	Period        string      `json:"period"`
	DebitBalance  float64     `json:"debitBalance"`
	CreditBalance float64     `json:"creditBalance"`
	CreatedAt     *gtime.Time `json:"createdAt"`
}

// FinanceFixedAsset 固定资产卡片表 finance_fixed_asset
type FinanceFixedAsset struct {
	Id                      uint64      `json:"id"`
	AssetNo                 string      `json:"assetNo"`
	AssetName               string      `json:"assetName"`
	AssetCategory           string      `json:"assetCategory"`
	Spec                    string      `json:"spec"`
	Unit                    string      `json:"unit"`
	Quantity                float64     `json:"quantity"`
	OriginalValue           float64     `json:"originalValue"`
	AccumulatedDepreciation float64     `json:"accumulatedDepreciation"`
	NetValue                float64     `json:"netValue"`
	ResidualRate            float64     `json:"residualRate"`
	UsefulLifeMonths        int         `json:"usefulLifeMonths"`
	DepreciationMethod      string      `json:"depreciationMethod"`
	StartUseDate            *gtime.Time `json:"startUseDate"`
	Status                  string      `json:"status"`
	Remark                  string      `json:"remark"`
	CreatedAt               *gtime.Time `json:"createdAt"`
}

// FinancePostingRule 凭证生成映射规则表 finance_posting_rule
type FinancePostingRule struct {
	Id              uint64      `json:"id"`
	BizType         string      `json:"bizType"`
	Direction       string      `json:"direction"`
	EntryNo         int         `json:"entryNo"`
	SubjectId       uint64      `json:"subjectId"`
	EntryDirection  string      `json:"entryDirection"`
	AmountSource    string      `json:"amountSource"`
	SummaryTemplate string      `json:"summaryTemplate"`
	IsActive        int         `json:"isActive"`
	CreatedAt       *gtime.Time `json:"createdAt"`
}

// FinanceTransaction 资金流水表 finance_transaction
type FinanceTransaction struct {
	Id           uint64      `json:"id"`
	AccountId    uint64      `json:"accountId"`
	Direction    string      `json:"direction"`
	Amount       float64     `json:"amount"`
	BalanceAfter float64     `json:"balanceAfter"`
	BizType      string      `json:"bizType"`
	BizNo        string      `json:"bizNo"`
	Remark       string      `json:"remark"`
	OccurredAt   *gtime.Time `json:"occurredAt"`
	CreatedAt    *gtime.Time `json:"createdAt"`
}

// FinanceVoucher 凭证头表 finance_voucher
type FinanceVoucher struct {
	Id              uint64      `json:"id"`
	VoucherNo       string      `json:"voucherNo"`
	VoucherDate     *gtime.Time `json:"voucherDate"`
	Period          string      `json:"period"`
	VoucherType     string      `json:"voucherType"`
	AttachmentCount int         `json:"attachmentCount"`
	Summary         string      `json:"summary"`
	Status          string      `json:"status"`
	SourceType      string      `json:"sourceType"`
	SourceId        uint64      `json:"sourceId"`
	CreatedBy       uint64      `json:"createdBy"`
	PostedBy        uint64      `json:"postedBy"`
	PostedAt        *gtime.Time `json:"postedAt"`
	CreatedAt       *gtime.Time `json:"createdAt"`
}

// FinanceVoucherEntry 凭证分录表 finance_voucher_entry
type FinanceVoucherEntry struct {
	Id            uint64      `json:"id"`
	VoucherId     uint64      `json:"voucherId"`
	EntryNo       int         `json:"entryNo"`
	Summary       string      `json:"summary"`
	SubjectId     uint64      `json:"subjectId"`
	Direction     string      `json:"direction"`
	Amount        float64     `json:"amount"`
	AuxiliaryInfo string      `json:"auxiliaryInfo"`
	CreatedAt     *gtime.Time `json:"createdAt"`
}
