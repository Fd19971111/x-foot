package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"x-foot/internal/model/entity"
)

// Leg 单条记账分录（借方或贷方）
type Leg struct {
	// 会计科目
	SubjectCode string
	SubjectId   uint64
	SubjectName string
	// 借贷方向：debit=借，credit=贷
	Direction string
	// 金额（正数）
	Amount float64
	// 分录摘要，为空时取科目名称
	Summary string
	// 资金账户（可选）：填写后会同步生成资金流水并更新账户余额
	AccountId   uint64
	AccountName string
	// 资金方向：in=流入，out=流出；为空时按借贷方向自动推断
	FlowDirection string
	// 辅助核算信息（JSON 字符串），如 {"customer_id":1}
	AuxiliaryInfo string
}

// DebitInput 借方入参
type DebitInput struct {
	SubjectCode   string
	Amount        float64
	Summary       string
	AccountId     uint64
	FlowDirection string
	AuxiliaryInfo string
}

// CreditInput 贷方入参
type CreditInput struct {
	SubjectCode   string
	Amount        float64
	Summary       string
	AccountId     uint64
	FlowDirection string
	AuxiliaryInfo string
}

// RecordInput 一笔完整双向记账入参
type RecordInput struct {
	// 业务类型，对应 finance_posting_rule.biz_type
	BizType string
	// 业务单号，便于追溯
	BizNo string
	// 凭证摘要
	Summary string
	// 凭证日期，默认当前时间
	VoucherDate *gtime.Time
	// 凭证类型：记/收/付/转，默认「记」
	VoucherType string
	// 制单人
	OperatorId uint64
	// 附件张数
	AttachmentCount int
	// 来源类型与来源单据ID，用于追溯
	SourceType string
	SourceId   uint64
	// 借贷分录
	Legs []*Leg
}

// RuleRecordInput 按凭证映射规则自动生成入参
type RuleRecordInput struct {
	// 业务类型，需已在 finance_posting_rule 中配置
	BizType string
	// 资金方向：in=收入，out=支出
	Direction string
	// 发生金额
	Amount float64
	// 资金账户（可选），与映射规则中对应的资金科目自动关联
	AccountId   uint64
	BizNo       string
	Summary     string
	VoucherDate *gtime.Time
	OperatorId  uint64
	// 摘要模板占位符，如 {"客户": "A公司"}，用于渲染 "收到{客户}货款"
	Params map[string]string
}

// ReverseInput 红冲（冲销）入参
type ReverseInput struct {
	// 被红冲的凭证ID（必填）
	VoucherId uint64
	// 红冲凭证日期，默认当前时间
	VoucherDate *gtime.Time
	// 制单人
	OperatorId uint64
	// 红冲原因，会拼进红冲凭证摘要
	Reason string
	// 科目对应多个资金账户时，显式指定资金回退到哪个账户
	AccountId uint64
}

// VoucherQuery 凭证查询入参
type VoucherQuery struct {
	// 凭证ID与凭证号二选一
	VoucherId uint64
	VoucherNo string
	// 业务单号，传入时一并返回该单据的资金流水
	BizNo string
}

// TransactionQuery 资金流水查询入参
type TransactionQuery struct {
	// 资金账户ID
	AccountId uint64
	// 记账期间，格式 2026-10
	Period string
	// 业务类型，见 biztype 包
	BizType string
	// 业务单号
	BizNo string
	// 分页参数，默认第 1 页、每页 20 条
	Page     int
	PageSize int
}

// VoucherDetail 凭证详情：凭证头 + 借贷分录 + 资金流水
type VoucherDetail struct {
	Voucher      *entity.FinanceVoucher        `json:"voucher"`
	Entries      []*entity.FinanceVoucherEntry `json:"entries"`
	Transactions []*entity.FinanceTransaction  `json:"transactions"`
}

// TransactionListResult 资金流水分页查询结果
type TransactionListResult struct {
	List     []*entity.FinanceTransaction `json:"list"`
	Total    int                          `json:"total"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"pageSize"`
}

// AccountBalance 账户余额快照
type AccountBalance struct {
	AccountId   uint64
	AccountName string
	Balance     float64
}

// RecordResult 完整财务记录（凭证 + 分录 + 资金流水 + 账户余额）
type RecordResult struct {
	Voucher      *entity.FinanceVoucher
	Entries      []*entity.FinanceVoucherEntry
	Transactions []*entity.FinanceTransaction
	Accounts     []*AccountBalance
	TotalAmount  float64
}
