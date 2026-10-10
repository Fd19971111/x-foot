package customer

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"x-foot/internal/model"
)

// CustomerProfile 客户资料：该客户关联的工程与合同
//
// 只返回工程与合同信息，不含客户身份信息（姓名 / 身份证等属于客户信息）。
func (s *sCustomer) CustomerProfile(ctx context.Context, customerId uint64) (*model.CustomerProfileResult, error) {
	if customerId == 0 {
		return nil, gerror.New("客户ID不能为空")
	}

	record, err := g.Model(customerTable).Ctx(ctx).Where("id", customerId).One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, gerror.Newf("客户不存在：%d", customerId)
	}

	result := &model.CustomerProfileResult{
		Customer: &model.CustomerBrief{
			Id:           customerId,
			CustomerNo:   record["customer_no"].String(),
			CustomerName: record["customer_name"].String(),
			CompanyName:  record["company_name"].String(),
			Phone:        record["phone"].String(),
		},
		Projects:  []*model.ProjectItem{},
		Contracts: []*model.ContractItem{},
	}

	projectRows, err := g.Model(projectTable).Ctx(ctx).
		Where("customer_id", customerId).
		OrderDesc("id").
		All()
	if err != nil {
		return nil, err
	}

	for _, row := range projectRows {
		result.Projects = append(result.Projects, &model.ProjectItem{
			Id:             row["id"].Uint64(),
			ProjectNo:      row["project_no"].String(),
			ProjectName:    row["project_name"].String(),
			Address:        row["address"].String(),
			StartDate:      dateOnly(row["start_date"]),
			EndDate:        dateOnly(row["end_date"]),
			Status:         row["status"].String(),
			ContractAmount: row["contract_amount"].Float64(),
			Remark:         row["remark"].String(),
		})
	}

	contractRows, err := g.Model(contractTable+" c").Ctx(ctx).
		LeftJoin(projectTable+" p", "c.project_id = p.id").
		Fields("c.*, p.project_name project_name").
		Where("c.customer_id", customerId).
		OrderDesc("c.id").
		All()
	if err != nil {
		return nil, err
	}

	total := 0.0

	for _, row := range contractRows {
		amount := row["amount"].Float64()

		result.Contracts = append(result.Contracts, &model.ContractItem{
			Id:           row["id"].Uint64(),
			ContractNo:   row["contract_no"].String(),
			ContractName: row["contract_name"].String(),
			ContractType: row["contract_type"].Int(),
			ProjectId:    row["project_id"].Uint64(),
			ProjectName:  row["project_name"].String(),
			PartyA:       row["party_a"].String(),
			PartyB:       row["party_b"].String(),
			Amount:       amount,
			TaxRate:      row["tax_rate"].Float64(),
			TaxAmount:    row["tax_amount"].Float64(),
			SignDate:     dateOnly(row["sign_date"]),
			StartDate:    dateOnly(row["start_date"]),
			EndDate:      dateOnly(row["end_date"]),
			Duration:     row["duration"].Int(),
			Status:       row["status"].Int(),
			Remark:       row["remark"].String(),
		})

		total += amount
	}

	result.ProjectCount = len(result.Projects)
	result.ContractCount = len(result.Contracts)
	result.ContractAmountTotal = total

	return result, nil
}
