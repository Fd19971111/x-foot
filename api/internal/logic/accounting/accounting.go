// =================================================================================
// 会计记账通用逻辑（accounting）
//
// 职责：所有涉及借贷记账、资金收付的业务，统一走本包，禁止各业务模块自行拼装凭证与流水。
// 一次调用即可得到完整双向记录：凭证头 + 借贷分录 + 资金流水 + 账户余额。
//
// 文件结构（按功能划分，从上到下阅读即可）：
//   一、常量定义         —— 借贷方向、资金方向、凭证状态、来源、表名
//   二、分录构造         —— Debit（借方）/ Credit（贷方）/ buildLeg（统一入口）
//   三、校验             —— 金额、科目、资金账户、借贷平衡、期间、红冲合法性
//   四、记账提交         —— Record（手工分录）/ RecordByRule（按映射规则）/ Reverse（红冲）
//   五、凭证生成         —— 凭证号、凭证头、分录、红冲标记
//   六、资金流水与余额   —— 流水记录、账户余额更新、账户余额查询
//   七、财务数据查询     —— GetVoucher（凭证详情）/ ListTransactions（资金流水）
//   八、内部辅助         —— 摘要模板渲染、方向推断、分页等
//
// 业务模块只需掌握三个入口：
//   Debit / Credit  —— 构造单边分录（含校验，不落库）
//   Record          —— 提交一笔双向分录
//   RecordByRule    —— 按配置好的业务类型自动出账
// 之后需要财务数据，统一走 GetVoucher / ListTransactions / GetAccountBalance。
// =================================================================================

package accounting

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"x-foot/internal/model"
	"x-foot/internal/model/entity"
	"x-foot/internal/service"
	"x-foot/internal/utility/biztype"
)

// ---------------------------------------------------------------------------------
// 一、常量定义
// ---------------------------------------------------------------------------------

const (
	// 借贷方向
	DirectionDebit  = "debit"  // 借
	DirectionCredit = "credit" // 贷

	// 资金方向
	FlowIn  = "in"  // 流入
	FlowOut = "out" // 流出

	// 凭证状态（与 finance_voucher.status 枚举一致：draft / posted / void）
	VoucherStatusPosted = "posted" // 已过账
	VoucherStatusVoid   = "void"   // 已作废，红冲后的原凭证置为该状态

	// 凭证来源
	SourceTypeManual  = "manual"  // 手工记账
	SourceTypeRule    = "rule"    // 按映射规则生成
	SourceTypeReverse = "reverse" // 红冲凭证

	// 默认凭证类型
	defaultVoucherType = "记"

	// 金额比较允许误差
	amountEpsilon = 0.005

	// 查询默认分页参数
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 200
)

// 涉及的数据表
const (
	tableCompany      = "finance_company"
	tableSubject      = "finance_subject"
	tableAccount      = "finance_account"
	tableTransaction  = "finance_transaction"
	tableVoucher      = "finance_voucher"
	tableVoucherEntry = "finance_voucher_entry"
	tablePostingRule  = "finance_posting_rule"
)

type sAccounting struct{}

func init() {
	service.RegisterAccounting(New())
}

// New 创建会计记账业务实例
func New() service.IAccounting {
	return &sAccounting{}
}

// ---------------------------------------------------------------------------------
// 二、分录构造：借方 / 贷方
// ---------------------------------------------------------------------------------

// Debit 构造一条借方分录：校验金额与科目、关联资金账户
func (s *sAccounting) Debit(ctx context.Context, in *model.DebitInput) (*model.Leg, error) {
	if in == nil {
		return nil, gerror.New("借方入参不能为空")
	}

	return s.buildLeg(ctx, &model.Leg{
		SubjectCode:   in.SubjectCode,
		Direction:     DirectionDebit,
		Amount:        in.Amount,
		Summary:       in.Summary,
		AccountId:     in.AccountId,
		FlowDirection: in.FlowDirection,
		AuxiliaryInfo: in.AuxiliaryInfo,
	})
}

// Credit 构造一条贷方分录：校验金额与科目、关联资金账户
func (s *sAccounting) Credit(ctx context.Context, in *model.CreditInput) (*model.Leg, error) {
	if in == nil {
		return nil, gerror.New("贷方入参不能为空")
	}

	return s.buildLeg(ctx, &model.Leg{
		SubjectCode:   in.SubjectCode,
		Direction:     DirectionCredit,
		Amount:        in.Amount,
		Summary:       in.Summary,
		AccountId:     in.AccountId,
		FlowDirection: in.FlowDirection,
		AuxiliaryInfo: in.AuxiliaryInfo,
	})
}

// buildLeg 分录统一构造入口：方向 -> 金额 -> 科目 -> 资金账户 -> 资金方向
func (s *sAccounting) buildLeg(ctx context.Context, leg *model.Leg) (*model.Leg, error) {
	if leg.Direction != DirectionDebit && leg.Direction != DirectionCredit {
		return nil, gerror.Newf("分录借贷方向不合法：%s", leg.Direction)
	}
	if err := s.checkAmount(leg.Amount); err != nil {
		return nil, err
	}
	if err := s.fillSubject(ctx, leg); err != nil {
		return nil, err
	}
	if err := s.fillAccount(ctx, leg); err != nil {
		return nil, err
	}
	if leg.FlowDirection == "" && leg.AccountId > 0 {
		leg.FlowDirection = inferFlowDirection(leg.Direction)
	}

	return leg, nil
}

// ---------------------------------------------------------------------------------
// 三、校验
// ---------------------------------------------------------------------------------

// checkAmount 金额校验：必须为大于 0 的有效数字
func (s *sAccounting) checkAmount(amount float64) error {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return gerror.New("金额格式不正确")
	}
	if amount <= 0 {
		return gerror.New("金额必须大于 0")
	}

	return nil
}

// fillSubject 解析并校验会计科目：必须存在、启用、且为明细科目
func (s *sAccounting) fillSubject(ctx context.Context, leg *model.Leg) error {
	leg.SubjectCode = strings.TrimSpace(leg.SubjectCode)
	if leg.SubjectId == 0 && leg.SubjectCode == "" {
		return gerror.New("会计科目不能为空")
	}

	model := g.Model(tableSubject).Where("is_active", 1)
	if leg.SubjectId > 0 {
		model = model.Where("id", leg.SubjectId)
	} else {
		model = model.Where("subject_code", leg.SubjectCode)
	}

	record, err := model.One()
	if err != nil {
		return err
	}
	if record.IsEmpty() {
		return gerror.Newf("会计科目不存在或已停用：%s", subjectDesc(leg))
	}
	if record["is_leaf"].Int() != 1 {
		return gerror.Newf("科目【%s】不是明细科目，不能直接记账", record["subject_name"].String())
	}

	leg.SubjectId = record["id"].Uint64()
	leg.SubjectCode = record["subject_code"].String()
	leg.SubjectName = record["subject_name"].String()

	return nil
}

// fillAccount 解析并校验资金账户：必须存在、启用，且其对应科目与分录科目一致
func (s *sAccounting) fillAccount(ctx context.Context, leg *model.Leg) error {
	if leg.AccountId == 0 {
		return nil
	}

	record, err := g.Model(tableAccount).Where("id", leg.AccountId).One()
	if err != nil {
		return err
	}
	if record.IsEmpty() {
		return gerror.Newf("资金账户不存在：%d", leg.AccountId)
	}
	if record["is_active"].Int() != 1 {
		return gerror.Newf("资金账户【%s】已停用", record["account_name"].String())
	}

	if subjectId := record["subject_id"].Uint64(); subjectId > 0 && subjectId != leg.SubjectId {
		return gerror.Newf(
			"资金账户【%s】对应的会计科目与分录科目不一致",
			record["account_name"].String(),
		)
	}

	leg.AccountName = record["account_name"].String()

	return nil
}

// checkBalance 借贷平衡校验
func (s *sAccounting) checkBalance(debitTotal, creditTotal float64) error {
	if debitTotal <= 0 || creditTotal <= 0 {
		return gerror.New("记账必须同时包含借方与贷方分录，且金额大于 0")
	}
	if math.Abs(debitTotal-creditTotal) > amountEpsilon {
		return gerror.Newf("借贷不平衡：借方合计 %.2f，贷方合计 %.2f", debitTotal, creditTotal)
	}

	return nil
}

// resolvePeriod 解析记账期间并校验账套状态、启用期间
func (s *sAccounting) resolvePeriod(ctx context.Context, date *gtime.Time) (string, error) {
	company, err := g.Model(tableCompany).OrderAsc("id").One()
	if err != nil {
		return "", err
	}
	if company.IsEmpty() || company["is_initialized"].Int() != 1 {
		return "", gerror.New("公司尚未完成初始化，无法记账")
	}

	period := date.Format("Y-m")
	if startPeriod := company["start_period"].String(); startPeriod != "" && period < startPeriod {
		return "", gerror.Newf("记账期间 %s 不能早于账套启用期间 %s", period, startPeriod)
	}

	return period, nil
}

// ---------------------------------------------------------------------------------
// 四、记账提交
// ---------------------------------------------------------------------------------

// Record 提交一笔完整双向记账
//
// 流程：分录校验 -> 借贷平衡 -> 账套期间校验 -> 生成凭证 -> 记资金流水 -> 更新账户余额
func (s *sAccounting) Record(ctx context.Context, in *model.RecordInput) (result *model.RecordResult, err error) {
	if in == nil || len(in.Legs) == 0 {
		return nil, gerror.New("记账分录不能为空")
	}

	// 1. 分录校验与汇总
	legs, totalAmount, err := s.prepareLegs(ctx, in.Legs)
	if err != nil {
		return nil, err
	}

	// 2. 凭证日期与期间
	voucherDate := in.VoucherDate
	if voucherDate == nil {
		voucherDate = gtime.Now()
	}
	period, err := s.resolvePeriod(ctx, voucherDate)
	if err != nil {
		return nil, err
	}

	voucherType := in.VoucherType
	if voucherType == "" {
		voucherType = defaultVoucherType
	}

	// 3. 事务提交
	result = &model.RecordResult{TotalAmount: totalAmount}
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		voucherNo, err := s.nextVoucherNo(ctx, tx, period, voucherType)
		if err != nil {
			return err
		}

		voucher, err := s.insertVoucher(ctx, tx, in, period, voucherType, voucherNo, voucherDate, totalAmount)
		if err != nil {
			return err
		}

		entries, err := s.insertEntries(ctx, tx, voucher.Id, legs)
		if err != nil {
			return err
		}

		for _, leg := range legs {
			transaction, balance, err := s.recordFlow(ctx, tx, leg, in, voucherNo, voucherDate)
			if err != nil {
				return err
			}
			if transaction != nil {
				result.Transactions = append(result.Transactions, transaction)
			}
			if balance != nil {
				result.Accounts = append(result.Accounts, balance)
			}
		}

		result.Voucher = voucher
		result.Entries = entries

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// RecordByRule 按凭证映射规则自动生成并提交一笔完整双向记账
//
// 业务侧只需提供业务类型、资金方向、金额与资金账户，借贷科目由 finance_posting_rule 决定。
func (s *sAccounting) RecordByRule(ctx context.Context, in *model.RuleRecordInput) (*model.RecordResult, error) {
	if in == nil {
		return nil, gerror.New("入参不能为空")
	}
	if strings.TrimSpace(in.BizType) == "" {
		return nil, gerror.New("业务类型不能为空")
	}
	if err := s.checkAmount(in.Amount); err != nil {
		return nil, err
	}
	if in.Direction != FlowIn && in.Direction != FlowOut {
		return nil, gerror.New("资金方向必须为 in（收入）或 out（支出）")
	}

	// 1. 读取映射规则
	rules, err := g.Model(tablePostingRule).
		Where("biz_type", in.BizType).
		Where("direction", in.Direction).
		Where("is_active", 1).
		OrderAsc("entry_no").
		All()
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, gerror.Newf("未找到业务类型【%s】的凭证映射规则，请先配置记账规则", biztype.Name(in.BizType))
	}

	// 2. 资金账户与其对应科目自动关联
	var (
		accountSubjectId uint64
		accountName      string
	)
	if in.AccountId > 0 {
		account, err := g.Model(tableAccount).Where("id", in.AccountId).One()
		if err != nil {
			return nil, err
		}
		if account.IsEmpty() {
			return nil, gerror.Newf("资金账户不存在：%d", in.AccountId)
		}
		accountSubjectId = account["subject_id"].Uint64()
		accountName = account["account_name"].String()
	}

	// 3. 规则 -> 分录
	legs := make([]*model.Leg, 0, len(rules))
	for _, rule := range rules {
		subjectId := rule["subject_id"].Uint64()
		leg := &model.Leg{
			SubjectId: subjectId,
			Direction: rule["entry_direction"].String(),
			Amount:    in.Amount,
			Summary:   renderSummary(rule["summary_template"].String(), in.Params, in.Summary),
		}
		if in.AccountId > 0 && subjectId == accountSubjectId {
			leg.AccountId = in.AccountId
			leg.AccountName = accountName
			leg.FlowDirection = in.Direction
		}
		legs = append(legs, leg)
	}

	// 4. 复用统一记账入口
	return s.Record(ctx, &model.RecordInput{
		BizType:     in.BizType,
		BizNo:       in.BizNo,
		Summary:     in.Summary,
		VoucherDate: in.VoucherDate,
		OperatorId:  in.OperatorId,
		SourceType:  SourceTypeRule,
		Legs:        legs,
	})
}

// Reverse 红冲（冲销）一张已过账凭证
//
// 按原凭证分录生成一组方向相反、金额相同的分录，生成新的红冲凭证，
// 并回退资金流水与账户余额；原凭证标记为已红冲，避免重复冲销。
func (s *sAccounting) Reverse(ctx context.Context, in *model.ReverseInput) (*model.RecordResult, error) {
	if in == nil || in.VoucherId == 0 {
		return nil, gerror.New("被红冲的凭证ID不能为空")
	}

	// 1. 读取原凭证与分录
	voucher, entries, err := s.getReversibleVoucher(ctx, in.VoucherId)
	if err != nil {
		return nil, err
	}

	// 2. 生成反向分录：借贷方向取反，金额不变
	summary := strings.TrimSpace(fmt.Sprintf("红冲 %s %s", voucher.VoucherNo, in.Reason))
	legs := make([]*model.Leg, 0, len(entries))
	for _, entry := range entries {
		legs = append(legs, &model.Leg{
			SubjectId:     entry.SubjectId,
			Direction:     oppositeDirection(entry.Direction),
			Amount:        entry.Amount,
			Summary:       summary,
			AuxiliaryInfo: entry.AuxiliaryInfo,
		})
	}

	// 3. 回填资金账户并写入反向流水：科目命中唯一账户时自动匹配，多个时按 AccountId 显式指定
	//    借贷方向取反后资金方向也自动取反，因此余额会被原路退回。
	if err := s.fillReverseAccounts(ctx, entries, legs, in.AccountId); err != nil {
		return nil, err
	}

	// 4. 复用统一记账入口
	result, err := s.Record(ctx, &model.RecordInput{
		BizType:     SourceTypeReverse,
		Summary:     summary,
		VoucherDate: in.VoucherDate,
		OperatorId:  in.OperatorId,
		SourceType:  SourceTypeReverse,
		SourceId:    in.VoucherId,
		Legs:        legs,
	})
	if err != nil {
		return nil, err
	}

	// 5. 标记原凭证已红冲
	if err := s.markVoucherReversed(ctx, in.VoucherId); err != nil {
		return nil, err
	}

	return result, nil
}

// fillReverseAccounts 为红冲分录回填资金账户
//
// 按科目匹配启用中的资金账户：唯一命中时直接使用；
// 同一科目挂了多个账户（如两家银行同科目 1002）时必须由调用方通过 specifyAccountId 指定。
func (s *sAccounting) fillReverseAccounts(
	ctx context.Context,
	entries []*entity.FinanceVoucherEntry,
	legs []*model.Leg,
	specifyAccountId uint64,
) error {
	records, err := g.Model(tableAccount).Where("is_active", 1).All()
	if err != nil {
		return err
	}

	grouped := make(map[uint64][]gdb.Record)
	for _, record := range records {
		subjectId := record["subject_id"].Uint64()
		if subjectId == 0 {
			continue
		}
		grouped[subjectId] = append(grouped[subjectId], record)
	}

	for index, entry := range entries {
		matched := grouped[entry.SubjectId]
		if len(matched) == 0 {
			continue
		}

		target := matched[0]
		if len(matched) > 1 {
			if specifyAccountId == 0 {
				return gerror.Newf(
					"科目存在多个资金账户，红冲时请通过 AccountId 明确指定回退账户",
				)
			}

			target = nil
			for _, item := range matched {
				if item["id"].Uint64() == specifyAccountId {
					target = item
					break
				}
			}
			if target == nil {
				return gerror.Newf("指定的资金账户【%d】与该凭证科目不匹配", specifyAccountId)
			}
		}

		legs[index].AccountId = target["id"].Uint64()
		legs[index].AccountName = target["account_name"].String()
	}

	return nil
}

// getReversibleVoucher 读取凭证及其分录，并校验是否允许红冲
func (s *sAccounting) getReversibleVoucher(
	ctx context.Context,
	voucherId uint64,
) (*entity.FinanceVoucher, []*entity.FinanceVoucherEntry, error) {
	record, err := g.Model(tableVoucher).Where("id", voucherId).One()
	if err != nil {
		return nil, nil, err
	}
	if record.IsEmpty() {
		return nil, nil, gerror.Newf("凭证不存在：%d", voucherId)
	}

	voucher := &entity.FinanceVoucher{}
	if err := record.Struct(voucher); err != nil {
		return nil, nil, err
	}
	if voucher.Status != VoucherStatusPosted {
		return nil, nil, gerror.Newf("凭证【%s】状态为 %s，不可红冲", voucher.VoucherNo, voucher.Status)
	}

	entryRecords, err := g.Model(tableVoucherEntry).
		Where("voucher_id", voucherId).
		OrderAsc("entry_no").
		All()
	if err != nil {
		return nil, nil, err
	}
	if entryRecords.IsEmpty() {
		return nil, nil, gerror.Newf("凭证【%s】没有分录，无法红冲", voucher.VoucherNo)
	}

	entries := make([]*entity.FinanceVoucherEntry, 0, entryRecords.Len())
	if err := entryRecords.Structs(&entries); err != nil {
		return nil, nil, err
	}

	return voucher, entries, nil
}

// prepareLegs 逐条校验分录并汇总借贷金额
func (s *sAccounting) prepareLegs(ctx context.Context, legs []*model.Leg) ([]*model.Leg, float64, error) {
	prepared := make([]*model.Leg, 0, len(legs))

	var debitTotal, creditTotal float64
	for _, leg := range legs {
		if leg == nil {
			continue
		}

		item := *leg // 复制一份，避免修改调用方对象
		if _, err := s.buildLeg(ctx, &item); err != nil {
			return nil, 0, err
		}

		if item.Direction == DirectionDebit {
			debitTotal += item.Amount
		} else {
			creditTotal += item.Amount
		}
		prepared = append(prepared, &item)
	}

	if len(prepared) < 2 {
		return nil, 0, gerror.New("至少需要一借一贷两条分录")
	}
	if err := s.checkBalance(debitTotal, creditTotal); err != nil {
		return nil, 0, err
	}

	return prepared, debitTotal, nil
}

// ---------------------------------------------------------------------------------
// 五、凭证生成
// ---------------------------------------------------------------------------------

// nextVoucherNo 生成凭证号，格式：记-202610-0001
func (s *sAccounting) nextVoucherNo(ctx context.Context, tx gdb.TX, period, voucherType string) (string, error) {
	prefix := fmt.Sprintf("%s-%s-", voucherType, strings.ReplaceAll(period, "-", ""))

	// 注意：gf 的 Model.Max() 返回 float64，凭证号为字符串，这里必须用 Fields + Value 取原始值
	value, err := tx.Model(tableVoucher).
		Where("period", period).
		Where("voucher_type", voucherType).
		Fields("MAX(voucher_no)").
		Value()
	if err != nil {
		return "", err
	}

	seq := 1
	if last := gconv.String(value); last != "" {
		if idx := strings.LastIndex(last, "-"); idx >= 0 {
			if n, err := strconv.Atoi(last[idx+1:]); err == nil {
				seq = n + 1
			}
		}
	}

	return fmt.Sprintf("%s%04d", prefix, seq), nil
}

// insertVoucher 写入凭证头
func (s *sAccounting) insertVoucher(
	ctx context.Context,
	tx gdb.TX,
	in *model.RecordInput,
	period, voucherType, voucherNo string,
	voucherDate *gtime.Time,
	totalAmount float64,
) (*entity.FinanceVoucher, error) {
	summary := strings.TrimSpace(in.Summary)
	if summary == "" {
		summary = biztype.Name(in.BizType)
	}

	status := VoucherStatusPosted
	data := g.Map{
		"voucher_no":       voucherNo,
		"voucher_date":     voucherDate.Format("Y-m-d"),
		"period":           period,
		"voucher_type":     voucherType,
		"attachment_count": in.AttachmentCount,
		"summary":          summary,
		"status":           status,
		"source_type":      in.SourceType,
		"source_id":        in.SourceId,
		"posted_at":        gtime.Now(),
	}
	if in.OperatorId > 0 {
		data["created_by"] = in.OperatorId
		data["posted_by"] = in.OperatorId
	}

	id, err := tx.Model(tableVoucher).Data(data).InsertAndGetId()
	if err != nil {
		return nil, err
	}

	return &entity.FinanceVoucher{
		Id:          uint64(id),
		VoucherNo:   voucherNo,
		VoucherDate: voucherDate,
		Period:      period,
		VoucherType: voucherType,
		Summary:     summary,
		Status:      status,
		SourceType:  in.SourceType,
		SourceId:    in.SourceId,
		CreatedBy:   in.OperatorId,
		PostedBy:    in.OperatorId,
	}, nil
}

// insertEntries 写入凭证分录
func (s *sAccounting) insertEntries(
	ctx context.Context,
	tx gdb.TX,
	voucherId uint64,
	legs []*model.Leg,
) ([]*entity.FinanceVoucherEntry, error) {
	entries := make([]*entity.FinanceVoucherEntry, 0, len(legs))

	for index, leg := range legs {
		entryNo := index + 1
		summary := leg.Summary
		if summary == "" {
			summary = leg.SubjectName
		}

		data := g.Map{
			"voucher_id": voucherId,
			"entry_no":   entryNo,
			"summary":    summary,
			"subject_id": leg.SubjectId,
			"direction":  leg.Direction,
			"amount":     leg.Amount,
		}
		// auxiliary_info 为 JSON 列，空值不写入，避免非法 JSON
		if strings.TrimSpace(leg.AuxiliaryInfo) != "" {
			data["auxiliary_info"] = leg.AuxiliaryInfo
		}

		id, err := tx.Model(tableVoucherEntry).Data(data).InsertAndGetId()
		if err != nil {
			return nil, err
		}

		entries = append(entries, &entity.FinanceVoucherEntry{
			Id:            uint64(id),
			VoucherId:     voucherId,
			EntryNo:       entryNo,
			Summary:       summary,
			SubjectId:     leg.SubjectId,
			Direction:     leg.Direction,
			Amount:        leg.Amount,
			AuxiliaryInfo: leg.AuxiliaryInfo,
		})
	}

	return entries, nil
}

// markVoucherReversed 把原凭证标记为已红冲，避免重复冲销
func (s *sAccounting) markVoucherReversed(ctx context.Context, voucherId uint64) error {
	_, err := g.Model(tableVoucher).
		Where("id", voucherId).
		Data(g.Map{"status": VoucherStatusVoid}).
		Update()

	return err
}

// ---------------------------------------------------------------------------------
// 六、资金流水与余额
// ---------------------------------------------------------------------------------

// recordFlow 记录资金流水并更新账户余额（仅当分录关联了资金账户时生效）
func (s *sAccounting) recordFlow(
	ctx context.Context,
	tx gdb.TX,
	leg *model.Leg,
	in *model.RecordInput,
	voucherNo string,
	occurredAt *gtime.Time,
) (*entity.FinanceTransaction, *model.AccountBalance, error) {
	if leg.AccountId == 0 {
		return nil, nil, nil
	}

	// 行锁读取账户，避免并发记账导致余额错乱
	record, err := tx.Model(tableAccount).Where("id", leg.AccountId).LockUpdate().One()
	if err != nil {
		return nil, nil, err
	}
	if record.IsEmpty() {
		return nil, nil, gerror.Newf("资金账户不存在：%d", leg.AccountId)
	}

	balanceBefore := record["balance"].Float64()
	balanceAfter := balanceBefore + leg.Amount
	if leg.FlowDirection == FlowOut {
		balanceAfter = balanceBefore - leg.Amount
	}

	if _, err = tx.Model(tableAccount).
		Where("id", leg.AccountId).
		Data(g.Map{"balance": balanceAfter}).
		Update(); err != nil {
		return nil, nil, err
	}

	remark := leg.Summary
	if remark == "" {
		remark = voucherNo
	}

	id, err := tx.Model(tableTransaction).Data(g.Map{
		"account_id":    leg.AccountId,
		"direction":     leg.FlowDirection,
		"amount":        leg.Amount,
		"balance_after": balanceAfter,
		"biz_type":      in.BizType,
		"biz_no":        in.BizNo,
		"remark":        remark,
		"occurred_at":   occurredAt,
	}).InsertAndGetId()
	if err != nil {
		return nil, nil, err
	}

	return &entity.FinanceTransaction{
			Id:           uint64(id),
			AccountId:    leg.AccountId,
			Direction:    leg.FlowDirection,
			Amount:       leg.Amount,
			BalanceAfter: balanceAfter,
			BizType:      in.BizType,
			BizNo:        in.BizNo,
			Remark:       remark,
			OccurredAt:   occurredAt,
		}, &model.AccountBalance{
			AccountId:   leg.AccountId,
			AccountName: record["account_name"].String(),
			Balance:     balanceAfter,
		}, nil
}

// GetAccountBalance 查询资金账户当前余额
func (s *sAccounting) GetAccountBalance(ctx context.Context, accountId uint64) (*model.AccountBalance, error) {
	if accountId == 0 {
		return nil, gerror.New("资金账户ID不能为空")
	}

	record, err := g.Model(tableAccount).Where("id", accountId).One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, gerror.Newf("资金账户不存在：%d", accountId)
	}

	return &model.AccountBalance{
		AccountId:   accountId,
		AccountName: record["account_name"].String(),
		Balance:     record["balance"].Float64(),
	}, nil
}

// ---------------------------------------------------------------------------------
// 七、财务数据查询
//
// 记账产生的财务数据统一从这里读取，业务模块不需要自己拼 SQL。
// ---------------------------------------------------------------------------------

// GetVoucher 查询凭证详情：凭证头 + 借贷分录 + 资金流水
//
// 凭证可通过 VoucherId 或 VoucherNo 定位；资金流水按 BizNo 关联，
// 若调用方未传 BizNo 则不返回流水（此时可用 ListTransactions 按账户/期间查）。
func (s *sAccounting) GetVoucher(ctx context.Context, in *model.VoucherQuery) (*model.VoucherDetail, error) {
	if in == nil {
		return nil, gerror.New("查询条件不能为空")
	}
	if in.VoucherId == 0 && strings.TrimSpace(in.VoucherNo) == "" {
		return nil, gerror.New("凭证ID与凭证号不能同时为空")
	}

	m := g.Model(tableVoucher)
	if in.VoucherId > 0 {
		m = m.Where("id", in.VoucherId)
	} else {
		m = m.Where("voucher_no", strings.TrimSpace(in.VoucherNo))
	}

	record, err := m.One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, gerror.New("凭证不存在")
	}

	voucher := &entity.FinanceVoucher{}
	if err := record.Struct(voucher); err != nil {
		return nil, err
	}

	entryRecords, err := g.Model(tableVoucherEntry).
		Where("voucher_id", voucher.Id).
		OrderAsc("entry_no").
		All()
	if err != nil {
		return nil, err
	}

	entries := make([]*entity.FinanceVoucherEntry, 0, entryRecords.Len())
	if err := entryRecords.Structs(&entries); err != nil {
		return nil, err
	}

	transactions := make([]*entity.FinanceTransaction, 0)
	if bizNo := strings.TrimSpace(in.BizNo); bizNo != "" {
		flowRecords, err := g.Model(tableTransaction).Where("biz_no", bizNo).OrderAsc("id").All()
		if err != nil {
			return nil, err
		}
		if err := flowRecords.Structs(&transactions); err != nil {
			return nil, err
		}
	}

	return &model.VoucherDetail{
		Voucher:      voucher,
		Entries:      entries,
		Transactions: transactions,
	}, nil
}

// ListTransactions 分页查询资金流水，可按资金账户、期间、业务类型、业务单号过滤
func (s *sAccounting) ListTransactions(
	ctx context.Context,
	in *model.TransactionQuery,
) (*model.TransactionListResult, error) {
	if in == nil {
		return nil, gerror.New("查询条件不能为空")
	}

	page, pageSize := normalizePage(in.Page, in.PageSize)

	m := g.Model(tableTransaction)
	if in.AccountId > 0 {
		m = m.Where("account_id", in.AccountId)
	}
	if bizType := strings.TrimSpace(in.BizType); bizType != "" {
		m = m.Where("biz_type", bizType)
	}
	if bizNo := strings.TrimSpace(in.BizNo); bizNo != "" {
		m = m.Where("biz_no", bizNo)
	}
	if period := strings.TrimSpace(in.Period); period != "" {
		start, err := gtime.StrToTime(period + "-01")
		if err != nil {
			return nil, gerror.Newf("期间格式不正确：%s，正确格式为 2026-10", period)
		}
		end := start.EndOfMonth()
		m = m.WhereGTE("occurred_at", start.Format("Y-m-d 00:00:00")).
			WhereLTE("occurred_at", end.Format("Y-m-d 23:59:59"))
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	records, err := m.Page(page, pageSize).OrderDesc("id").All()
	if err != nil {
		return nil, err
	}

	list := make([]*entity.FinanceTransaction, 0, records.Len())
	if err := records.Structs(&list); err != nil {
		return nil, err
	}

	return &model.TransactionListResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ---------------------------------------------------------------------------------
// 八、内部辅助
// ---------------------------------------------------------------------------------

// oppositeDirection 取反借贷方向
func oppositeDirection(direction string) string {
	if direction == DirectionDebit {
		return DirectionCredit
	}

	return DirectionDebit
}

// normalizePage 规范分页参数
func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	return page, pageSize
}

// inferFlowDirection 按借贷方向推断资金方向：借方=流入，贷方=流出
func inferFlowDirection(direction string) string {
	if direction == DirectionDebit {
		return FlowIn
	}

	return FlowOut
}

// renderSummary 渲染摘要模板，如 "收到{customer}货款" + {"customer": "A公司"}
func renderSummary(template string, params map[string]string, fallback string) string {
	result := strings.TrimSpace(template)
	for key, value := range params {
		result = strings.ReplaceAll(result, "{"+key+"}", value)
	}
	if result == "" {
		return fallback
	}

	return result
}

// subjectDesc 科目描述，用于错误提示
func subjectDesc(leg *model.Leg) string {
	if leg.SubjectCode != "" {
		return leg.SubjectCode
	}

	return gconv.String(leg.SubjectId)
}
