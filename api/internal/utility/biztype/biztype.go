// Package biztype 定义记账业务类型编码及其显示名称。
//
// finance_posting_rule.biz_type 只存这里定义的英文编码，
// 业务代码（logic/accounting）一律按编码调用记账，避免中文取值带来的歧义；
// 需要在界面上展示时统一通过 Name() 取中文显示名。
package biztype

// 业务类型编码
const (
	SaleReceipt      = "sale_receipt"      // 销售收款
	PurchasePayment  = "purchase_payment"  // 采购付款
	ExpenseReimburse = "expense_reimburse" // 费用报销
	SalaryPay        = "salary_pay"        // 工资发放
	CapitalInvest    = "capital_invest"    // 收到投资
	BorrowIn         = "borrow_in"         // 取得借款
	RepayBorrow      = "repay_borrow"      // 归还借款
)

// names 业务类型编码 -> 中文显示名称
var names = map[string]string{
	SaleReceipt:      "销售收款",
	PurchasePayment:  "采购付款",
	ExpenseReimburse: "费用报销",
	SalaryPay:        "工资发放",
	CapitalInvest:    "收到投资",
	BorrowIn:         "取得借款",
	RepayBorrow:      "归还借款",
}

// ordered 展示顺序，保证列表稳定
var ordered = []string{
	SaleReceipt,
	PurchasePayment,
	ExpenseReimburse,
	SalaryPay,
	CapitalInvest,
	BorrowIn,
	RepayBorrow,
}

// Name 返回业务类型编码对应的中文显示名称；未定义的编码原样返回
func Name(code string) string {
	if name, ok := names[code]; ok {
		return name
	}
	return code
}

// Codes 返回全部业务类型编码
func Codes() []string {
	return ordered
}
