package service

import (
	"context"

	"x-foot/internal/model"
)

// IAccounting 会计记账通用业务接口
//
// 所有涉及资金收付、凭证生成的业务统一调用这里，不要在各业务模块自行拼装凭证与流水。
type IAccounting interface {
	// Debit 构造一条借方分录（完成科目、资金账户校验与规范化，不落库）
	Debit(ctx context.Context, in *model.DebitInput) (*model.Leg, error)
	// Credit 构造一条贷方分录（完成科目、资金账户校验与规范化，不落库）
	Credit(ctx context.Context, in *model.CreditInput) (*model.Leg, error)
	// Record 提交一笔完整双向记账：校验借贷平衡 -> 生成凭证与分录 -> 记资金流水 -> 更新账户余额
	Record(ctx context.Context, in *model.RecordInput) (*model.RecordResult, error)
	// RecordByRule 按 finance_posting_rule 映射规则自动生成并提交一笔完整双向记账
	RecordByRule(ctx context.Context, in *model.RuleRecordInput) (*model.RecordResult, error)
	// Reverse 红冲一张已过账凭证：生成反向凭证，回退资金流水与账户余额，并标记原凭证已红冲
	Reverse(ctx context.Context, in *model.ReverseInput) (*model.RecordResult, error)
	// GetAccountBalance 查询资金账户当前余额
	GetAccountBalance(ctx context.Context, accountId uint64) (*model.AccountBalance, error)
	// GetVoucher 查询凭证详情（凭证头 + 借贷分录 + 资金流水）
	GetVoucher(ctx context.Context, in *model.VoucherQuery) (*model.VoucherDetail, error)
	// ListTransactions 分页查询资金流水
	ListTransactions(ctx context.Context, in *model.TransactionQuery) (*model.TransactionListResult, error)
}

var localAccounting IAccounting

// Accounting 获取会计记账业务实例
func Accounting() IAccounting {
	if localAccounting == nil {
		panic("implement not found for interface IAccounting, forgot register?")
	}
	return localAccounting
}

// RegisterAccounting 注册会计记账业务实现
func RegisterAccounting(i IAccounting) {
	localAccounting = i
}
