package company

import (
	"context"
	"fmt"
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

const (
	companyTable        = "finance_company"
	subjectTable        = "finance_subject"
	accountTable        = "finance_account"
	openingBalanceTable = "finance_opening_balance"
	fixedAssetTable     = "finance_fixed_asset"
	postingRuleTable    = "finance_posting_rule"
)

type sCompany struct{}

func init() {
	service.RegisterCompany(New())
}

// New 创建公司初始化业务实例
func New() service.ICompany {
	return &sCompany{}
}

// InitStatus 查询公司是否已完成初始化（单公司：取第一条账套记录）
func (s *sCompany) InitStatus(ctx context.Context) (*model.InitStatus, error) {
	one, err := g.Model(companyTable).OrderAsc("id").One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return &model.InitStatus{Initialized: false}, nil
	}

	var company entity.FinanceCompany
	if err = one.Struct(&company); err != nil {
		return nil, err
	}

	return &model.InitStatus{
		Initialized: company.IsInitialized == 1,
		CompanyId:   company.Id,
		CompanyName: company.CompanyName,
	}, nil
}

// InitOptions 获取初始化页面所需的默认数据
func (s *sCompany) InitOptions(ctx context.Context) (*model.InitOptions, error) {
	nameMap := make(map[string]string, len(defaultSubjects))
	subjects := make([]model.SubjectItem, 0, len(defaultSubjects))
	for _, sub := range defaultSubjects {
		item := sub
		item.IsActive = 1
		subjects = append(subjects, item)
		nameMap[item.Code] = item.Name
	}

	rules := make([]model.PostingRuleItem, 0, len(defaultPostingRules))
	for _, rule := range defaultPostingRules {
		item := rule
		item.SubjectName = nameMap[item.SubjectCode]
		item.BizTypeName = biztype.Name(item.BizType)
		if item.AmountSource == "" {
			item.AmountSource = "amount"
		}
		rules = append(rules, item)
	}

	return &model.InitOptions{
		AccountingStandards:  accountingStandards,
		SuggestedStartPeriod: gtime.Now().Format("Y-m"),
		DefaultSubjects:      subjects,
		DefaultPostingRules:  rules,
		DefaultAccounts:      defaultAccounts,
	}, nil
}

// Init 提交公司初始配置：预置默认科目、写入公司/账户/期初余额/固定资产、生成默认映射
func (s *sCompany) Init(ctx context.Context, in *model.InitInput) error {
	if in == nil {
		return gerror.New("初始化参数不能为空")
	}

	in.CompanyName = strings.TrimSpace(in.CompanyName)
	if in.CompanyName == "" {
		return gerror.New("公司名称不能为空")
	}
	if strings.TrimSpace(in.StartPeriod) == "" {
		in.StartPeriod = gtime.Now().Format("Y-m")
	}
	if strings.TrimSpace(in.BaseCurrency) == "" {
		in.BaseCurrency = "CNY"
	}

	status, err := s.InitStatus(ctx)
	if err != nil {
		return err
	}
	if status.Initialized {
		return gerror.New("该公司已完成初始化，请勿重复提交")
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 预置会计科目（映射规则与期初余额都依赖科目）
		subjectIds, err := seedSubjects(ctx, tx, in.Subjects)
		if err != nil {
			return err
		}

		// 2. 写入或更新公司账套
		companyData := g.Map{
			"company_name":        in.CompanyName,
			"tax_no":              in.TaxNo,
			"accounting_standard": in.AccountingStandard,
			"base_currency":       in.BaseCurrency,
			"start_period":        in.StartPeriod,
			"is_initialized":      1,
			"initialized_at":      gtime.Now(),
		}
		if status.CompanyId > 0 {
			if _, err = tx.Model(companyTable).Where("id", status.CompanyId).Data(companyData).Update(); err != nil {
				return err
			}
		} else {
			companyData["created_at"] = gtime.Now()
			if _, err = tx.Model(companyTable).Data(companyData).Insert(); err != nil {
				return err
			}
		}

		// 3. 资金账户（含账户期初余额）
		for _, acc := range in.Accounts {
			if strings.TrimSpace(acc.AccountName) == "" {
				continue
			}
			accountType := acc.AccountType
			if accountType == "" {
				accountType = "bank"
			}
			currency := acc.Currency
			if currency == "" {
				currency = in.BaseCurrency
			}
			if _, err = tx.Model(accountTable).Data(g.Map{
				"account_name": acc.AccountName,
				"account_type": accountType,
				"subject_id":   subjectIds[acc.SubjectCode],
				"account_no":   acc.AccountNo,
				"bank_name":    acc.BankName,
				"currency":     currency,
				"balance":      acc.Balance,
				"is_active":    1,
			}).Insert(); err != nil {
				return err
			}
		}

		// 4. 期初余额
		for _, ob := range in.OpeningBalances {
			subjectId := subjectIds[strings.TrimSpace(ob.SubjectCode)]
			if subjectId == 0 {
				continue
			}
			if ob.DebitBalance == 0 && ob.CreditBalance == 0 {
				continue
			}
			period := ob.Period
			if period == "" {
				period = in.StartPeriod
			}
			if _, err = tx.Model(openingBalanceTable).Data(g.Map{
				"subject_id":     subjectId,
				"period":         period,
				"debit_balance":  ob.DebitBalance,
				"credit_balance": ob.CreditBalance,
			}).Insert(); err != nil {
				return err
			}
		}

		// 5. 固定资产卡片
		for idx, fa := range in.FixedAssets {
			if strings.TrimSpace(fa.AssetName) == "" {
				continue
			}
			assetNo := strings.TrimSpace(fa.AssetNo)
			if assetNo == "" {
				assetNo = fmt.Sprintf("FA%s%03d", gtime.Now().Format("YmdHis"), idx+1)
			}
			netValue := fa.NetValue
			if netValue == 0 {
				netValue = fa.OriginalValue - fa.AccumulatedDepreciation
			}
			quantity := fa.Quantity
			if quantity == 0 {
				quantity = 1
			}
			remark := fa.Remark
			if remark == "" {
				remark = "初始化录入"
			}
			if _, err = tx.Model(fixedAssetTable).Data(g.Map{
				"asset_no":                 assetNo,
				"asset_name":               fa.AssetName,
				"asset_category":           fa.AssetCategory,
				"spec":                     fa.Spec,
				"unit":                     fa.Unit,
				"quantity":                 quantity,
				"original_value":           fa.OriginalValue,
				"accumulated_depreciation": fa.AccumulatedDepreciation,
				"net_value":                netValue,
				"residual_rate":            fa.ResidualRate,
				"useful_life_months":       fa.UsefulLifeMonths,
				"depreciation_method":      "straight_line",
				"start_use_date":           gtime.New(fa.StartUseDate),
				"status":                   "in_use",
				"remark":                   remark,
			}).Insert(); err != nil {
				return err
			}
		}

		// 6. 生成默认凭证映射规则
		for _, rule := range defaultPostingRules {
			subjectId := subjectIds[rule.SubjectCode]
			if subjectId == 0 {
				continue
			}
			amountSource := rule.AmountSource
			if amountSource == "" {
				amountSource = "amount"
			}
			if _, err = tx.Model(postingRuleTable).Data(g.Map{
				"biz_type":         rule.BizType,
				"direction":        rule.Direction,
				"entry_no":         rule.EntryNo,
				"subject_id":       subjectId,
				"entry_direction":  rule.EntryDirection,
				"amount_source":    amountSource,
				"summary_template": rule.SummaryTemplate,
				"is_active":        1,
			}).Insert(); err != nil {
				return err
			}
		}

		return nil
	})
}

// seedSubjects 预置会计科目，返回 科目编码 -> 科目ID 映射（仅包含启用科目）
//
// subjects 为空时使用系统内置默认科目表（全部启用）；
// subjects 非空时以调用方传入的清单为准（IsActive 为 0 表示停用）。
func seedSubjects(ctx context.Context, tx gdb.TX, subjects []model.SubjectItem) (map[string]uint64, error) {
	useDefaults := len(subjects) == 0
	if useDefaults {
		subjects = defaultSubjects
	}

	subjectIds := make(map[string]uint64, len(subjects))

	existing, err := tx.Model(subjectTable).Fields("id", "subject_code", "is_active").All()
	if err != nil {
		return nil, err
	}
	for _, record := range existing {
		if record["is_active"].Int() == 1 {
			subjectIds[record["subject_code"].String()] = record["id"].Uint64()
		}
	}

	for _, sub := range subjects {
		isActive := 1
		if !useDefaults && sub.IsActive != 1 {
			isActive = 0
		}

		if _, ok := subjectIds[sub.Code]; ok {
			continue
		}

		id, err := tx.Model(subjectTable).Data(g.Map{
			"subject_code":   sub.Code,
			"subject_name":   sub.Name,
			"level":          1,
			"category":       sub.Category,
			"direction":      sub.Direction,
			"is_leaf":        1,
			"is_active":      isActive,
			"auxiliary_type": sub.AuxiliaryType,
		}).InsertAndGetId()
		if err != nil {
			return nil, err
		}
		if isActive == 1 {
			subjectIds[sub.Code] = gconv.Uint64(id)
		}
	}

	return subjectIds, nil
}
