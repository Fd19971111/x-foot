package company

import (
	"x-foot/internal/model"
	"x-foot/internal/utility/biztype"
)

// accountingStandards 可选的会计准则
var accountingStandards = []string{
	"小企业会计准则",
	"企业会计准则",
}

// defaultSubjects 默认会计科目表（小企业会计准则常用科目）
var defaultSubjects = []model.SubjectItem{
	{Code: "1001", Name: "库存现金", Category: "asset", Direction: "debit"},
	{Code: "1002", Name: "银行存款", Category: "asset", Direction: "debit"},
	{Code: "1012", Name: "其他货币资金", Category: "asset", Direction: "debit"},
	{Code: "1122", Name: "应收账款", Category: "asset", Direction: "debit", AuxiliaryType: "customer"},
	{Code: "1123", Name: "预付账款", Category: "asset", Direction: "debit", AuxiliaryType: "supplier"},
	{Code: "1221", Name: "其他应收款", Category: "asset", Direction: "debit"},
	{Code: "1401", Name: "材料采购", Category: "asset", Direction: "debit"},
	{Code: "1403", Name: "原材料", Category: "asset", Direction: "debit"},
	{Code: "1405", Name: "库存商品", Category: "asset", Direction: "debit"},
	{Code: "1601", Name: "固定资产", Category: "asset", Direction: "debit"},
	{Code: "1602", Name: "累计折旧", Category: "asset", Direction: "credit"},
	{Code: "1701", Name: "无形资产", Category: "asset", Direction: "debit"},
	{Code: "2001", Name: "短期借款", Category: "liability", Direction: "credit"},
	{Code: "2202", Name: "应付账款", Category: "liability", Direction: "credit", AuxiliaryType: "supplier"},
	{Code: "2203", Name: "预收账款", Category: "liability", Direction: "credit", AuxiliaryType: "customer"},
	{Code: "2211", Name: "应付职工薪酬", Category: "liability", Direction: "credit"},
	{Code: "2221", Name: "应交税费", Category: "liability", Direction: "credit"},
	{Code: "2501", Name: "长期借款", Category: "liability", Direction: "credit"},
	{Code: "4001", Name: "实收资本", Category: "equity", Direction: "credit"},
	{Code: "4002", Name: "资本公积", Category: "equity", Direction: "credit"},
	{Code: "4103", Name: "本年利润", Category: "equity", Direction: "credit"},
	{Code: "4104", Name: "利润分配", Category: "equity", Direction: "credit"},
	{Code: "5001", Name: "生产成本", Category: "cost", Direction: "debit"},
	{Code: "6001", Name: "主营业务收入", Category: "profit_loss", Direction: "credit"},
	{Code: "6051", Name: "其他业务收入", Category: "profit_loss", Direction: "credit"},
	{Code: "6301", Name: "营业外收入", Category: "profit_loss", Direction: "credit"},
	{Code: "6401", Name: "主营业务成本", Category: "profit_loss", Direction: "debit"},
	{Code: "6402", Name: "其他业务成本", Category: "profit_loss", Direction: "debit"},
	{Code: "6601", Name: "销售费用", Category: "profit_loss", Direction: "debit"},
	{Code: "6602", Name: "管理费用", Category: "profit_loss", Direction: "debit"},
	{Code: "6603", Name: "财务费用", Category: "profit_loss", Direction: "debit"},
	{Code: "6711", Name: "营业外支出", Category: "profit_loss", Direction: "debit"},
}

// defaultPostingRules 默认凭证生成映射规则（业务类型编码 -> 科目借贷映射）
//
// BizType 必须使用 biztype 包中的英文编码，摘要模板占位符同样使用英文 key。
var defaultPostingRules = []model.PostingRuleItem{
	{BizType: biztype.SaleReceipt, Direction: "in", EntryNo: 1, SubjectCode: "1002", EntryDirection: "debit", SummaryTemplate: "收到{customer}货款"},
	{BizType: biztype.SaleReceipt, Direction: "in", EntryNo: 2, SubjectCode: "1122", EntryDirection: "credit", SummaryTemplate: "收到{customer}货款"},

	{BizType: biztype.PurchasePayment, Direction: "out", EntryNo: 1, SubjectCode: "2202", EntryDirection: "debit", SummaryTemplate: "支付{supplier}货款"},
	{BizType: biztype.PurchasePayment, Direction: "out", EntryNo: 2, SubjectCode: "1002", EntryDirection: "credit", SummaryTemplate: "支付{supplier}货款"},

	{BizType: biztype.ExpenseReimburse, Direction: "out", EntryNo: 1, SubjectCode: "6602", EntryDirection: "debit", SummaryTemplate: "报销费用"},
	{BizType: biztype.ExpenseReimburse, Direction: "out", EntryNo: 2, SubjectCode: "1002", EntryDirection: "credit", SummaryTemplate: "报销费用"},

	{BizType: biztype.SalaryPay, Direction: "out", EntryNo: 1, SubjectCode: "2211", EntryDirection: "debit", SummaryTemplate: "发放工资"},
	{BizType: biztype.SalaryPay, Direction: "out", EntryNo: 2, SubjectCode: "1002", EntryDirection: "credit", SummaryTemplate: "发放工资"},

	{BizType: biztype.CapitalInvest, Direction: "in", EntryNo: 1, SubjectCode: "1002", EntryDirection: "debit", SummaryTemplate: "收到投资款"},
	{BizType: biztype.CapitalInvest, Direction: "in", EntryNo: 2, SubjectCode: "4001", EntryDirection: "credit", SummaryTemplate: "收到投资款"},

	{BizType: biztype.BorrowIn, Direction: "in", EntryNo: 1, SubjectCode: "1002", EntryDirection: "debit", SummaryTemplate: "取得借款"},
	{BizType: biztype.BorrowIn, Direction: "in", EntryNo: 2, SubjectCode: "2001", EntryDirection: "credit", SummaryTemplate: "取得借款"},

	{BizType: biztype.RepayBorrow, Direction: "out", EntryNo: 1, SubjectCode: "2001", EntryDirection: "debit", SummaryTemplate: "归还借款"},
	{BizType: biztype.RepayBorrow, Direction: "out", EntryNo: 2, SubjectCode: "1002", EntryDirection: "credit", SummaryTemplate: "归还借款"},
}

// defaultAccounts 默认资金账户
var defaultAccounts = []model.AccountItem{
	{AccountName: "库存现金", AccountType: "cash", SubjectCode: "1001", Currency: "CNY"},
	{AccountName: "基本存款账户", AccountType: "bank", SubjectCode: "1002", Currency: "CNY"},
}
