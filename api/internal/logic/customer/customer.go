// Package customer 客户管理业务逻辑：客户信息、客户资料与附件
//
// 文件划分：
//   customer.go —— 注册、表名常量、客户增删改查与内部通用辅助
//   attach.go   —— 附件上传、存储与删除
//   profile.go  —— 客户资料（关联的工程与合同）
package customer

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"x-foot/internal/model"
	"x-foot/internal/service"
)

// 表名常量
const (
	customerTable        = "site_customer"
	customerContactTable = "site_customer_contact"
	customerAttachTable  = "site_customer_attach"
	projectTable         = "biz_project"
	contractTable        = "biz_contract"
)

// 分页默认值
const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 200
)

type sCustomer struct{}

func init() {
	service.RegisterCustomer(New())
}

// New 创建客户管理业务实例
func New() service.ICustomer {
	return &sCustomer{}
}

// ---------------------------------------------------------------------------------
// 查询
// ---------------------------------------------------------------------------------

// ListCustomers 分页查询客户
//
// Keyword 匹配客户编号、姓名、公司与电话；列表行带联系人 / 附件数量。
func (s *sCustomer) ListCustomers(ctx context.Context, in *model.CustomerQuery) (*model.CustomerListResult, error) {
	page, pageSize := normalizePage(in.Page, in.PageSize)
	keyword := strings.TrimSpace(in.Keyword)

	// 统计与取数分别构造：Count() 会复用模型上已设置的 Fields，
	// 带别名的多字段会让 COUNT(...) 变成非法 SQL。
	newQuery := func() *gdb.Model {
		query := g.Model(customerTable).Ctx(ctx)

		if keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where(
				"customer_no LIKE ? OR customer_name LIKE ? OR company_name LIKE ? OR phone LIKE ?",
				like, like, like, like,
			)
		}

		return query
	}

	total, err := newQuery().Count()
	if err != nil {
		return nil, err
	}

	records, err := newQuery().Page(page, pageSize).OrderDesc("id").All()
	if err != nil {
		return nil, err
	}

	contactCount, err := countByField(ctx, customerContactTable, "customer_id")
	if err != nil {
		return nil, err
	}

	attachCount, err := countByField(ctx, customerAttachTable, "customer_id")
	if err != nil {
		return nil, err
	}

	list := make([]*model.CustomerItem, 0, records.Len())
	for _, record := range records {
		id := record["id"].Uint64()

		item := customerFromRecord(record)
		item.ContactCount = contactCount[id]
		item.AttachCount = attachCount[id]

		list = append(list, item)
	}

	return &model.CustomerListResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetCustomer 查询客户详情，含联系人与附件
func (s *sCustomer) GetCustomer(ctx context.Context, id uint64) (*model.CustomerItem, error) {
	if id == 0 {
		return nil, gerror.New("客户ID不能为空")
	}

	record, err := g.Model(customerTable).Ctx(ctx).Where("id", id).One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, gerror.Newf("客户不存在：%d", id)
	}

	item := customerFromRecord(record)

	contacts, err := loadContacts(ctx, id)
	if err != nil {
		return nil, err
	}
	item.Contacts = contacts
	item.ContactCount = len(contacts)

	attachments, err := loadAttachments(ctx, id)
	if err != nil {
		return nil, err
	}
	item.Attachments = attachments
	item.AttachCount = len(attachments)

	return item, nil
}

// ---------------------------------------------------------------------------------
// 保存与删除
// ---------------------------------------------------------------------------------

// SaveCustomer 新增或编辑客户，返回客户ID
//
// 只维护「客户信息」范围内的字段；工程与合同属于客户资料，不在这里修改。
// 联系人按客户整体覆盖保存。
func (s *sCustomer) SaveCustomer(ctx context.Context, in *model.CustomerSaveInput) (uint64, error) {
	in.CustomerName = strings.TrimSpace(in.CustomerName)
	in.CompanyName = strings.TrimSpace(in.CompanyName)

	// customer_name 为 NOT NULL，姓名留空时退化为公司名，保证两列至少有一列有值
	if in.CustomerName == "" {
		in.CustomerName = in.CompanyName
	}
	if in.CustomerName == "" {
		return 0, gerror.New("姓名与公司至少填写一项")
	}

	customerType := strings.TrimSpace(in.CustomerType)
	switch customerType {
	case model.CustomerTypePersonal, model.CustomerTypeCompany, model.CustomerTypeGovernment:
	default:
		customerType = model.CustomerTypePersonal
	}

	// 新增时不传 isActive 视为启用；编辑时按传入值处理，允许停用
	isActive := in.IsActive
	if in.Id == 0 && isActive == 0 {
		isActive = 1
	}

	data := g.Map{
		"customer_name": in.CustomerName,
		"company_name":  in.CompanyName,
		"customer_type": customerType,
		"phone":         strings.TrimSpace(in.Phone),
		"id_card":       strings.TrimSpace(in.IdCard),
		"email":         strings.TrimSpace(in.Email),
		"address":       strings.TrimSpace(in.Address),
		"level":         strings.TrimSpace(in.Level),
		"source":        strings.TrimSpace(in.Source),
		"remark":        strings.TrimSpace(in.Remark),
		"is_active":     isActive,
	}

	var customerId uint64

	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if in.Id == 0 {
			no, err := nextCustomerNo(ctx)
			if err != nil {
				return err
			}

			data["customer_no"] = no
			data["created_at"] = gtime.Now()

			id, err := tx.Model(customerTable).Ctx(ctx).Data(data).InsertAndGetId()
			if err != nil {
				return err
			}

			customerId = uint64(id)
		} else {
			exists, err := tx.Model(customerTable).Ctx(ctx).Where("id", in.Id).Exist()
			if err != nil {
				return err
			}
			if !exists {
				return gerror.Newf("客户不存在：%d", in.Id)
			}

			customerId = in.Id

			if _, err := tx.Model(customerTable).Ctx(ctx).Where("id", in.Id).Data(data).Update(); err != nil {
				return err
			}
		}

		return replaceContacts(ctx, tx, customerId, in.Contacts)
	})
	if err != nil {
		return 0, err
	}

	return customerId, nil
}

// DeleteCustomer 删除客户
//
// 客户下还有工程或合同时拒绝删除，避免出现无主的业务数据；
// 删除时一并清理联系人与附件（含磁盘文件）。
func (s *sCustomer) DeleteCustomer(ctx context.Context, id uint64) error {
	if id == 0 {
		return gerror.New("客户ID不能为空")
	}

	exists, err := g.Model(customerTable).Ctx(ctx).Where("id", id).Exist()
	if err != nil {
		return err
	}
	if !exists {
		return gerror.Newf("客户不存在：%d", id)
	}

	projectCount, err := g.Model(projectTable).Ctx(ctx).Where("customer_id", id).Count()
	if err != nil {
		return err
	}
	if projectCount > 0 {
		return gerror.Newf("该客户下还有 %d 个工程，请先处理后再删除", projectCount)
	}

	contractCount, err := g.Model(contractTable).Ctx(ctx).Where("customer_id", id).Count()
	if err != nil {
		return err
	}
	if contractCount > 0 {
		return gerror.Newf("该客户下还有 %d 份合同，请先处理后再删除", contractCount)
	}

	attachments, err := loadAttachments(ctx, id)
	if err != nil {
		return err
	}

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(customerContactTable).Ctx(ctx).Where("customer_id", id).Delete(); err != nil {
			return err
		}

		if _, err := tx.Model(customerAttachTable).Ctx(ctx).Where("customer_id", id).Delete(); err != nil {
			return err
		}

		_, err := tx.Model(customerTable).Ctx(ctx).Where("id", id).Delete()

		return err
	})
	if err != nil {
		return err
	}

	// 数据库已删除，磁盘文件尽力清理，失败不影响结果
	for _, item := range attachments {
		removeStoredFile(item.FileUrl)
	}

	return nil
}

// ---------------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------------

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

// customerFromRecord 把结果集行转换为客户结构体
func customerFromRecord(record gdb.Record) *model.CustomerItem {
	return &model.CustomerItem{
		Id:           record["id"].Uint64(),
		CustomerNo:   record["customer_no"].String(),
		CustomerName: record["customer_name"].String(),
		CompanyName:  record["company_name"].String(),
		CustomerType: record["customer_type"].String(),
		Phone:        record["phone"].String(),
		IdCard:       record["id_card"].String(),
		Email:        record["email"].String(),
		Address:      record["address"].String(),
		Level:        record["level"].String(),
		Source:       record["source"].String(),
		IsActive:     record["is_active"].Int(),
		Remark:       record["remark"].String(),
		CreatedAt:    gtime.New(record["created_at"].String()),
	}
}

// loadContacts 读取客户联系人，按排序升序
func loadContacts(ctx context.Context, customerId uint64) ([]*model.CustomerContactItem, error) {
	records, err := g.Model(customerContactTable).Ctx(ctx).
		Where("customer_id", customerId).
		OrderAsc("sort").
		OrderAsc("id").
		All()
	if err != nil {
		return nil, err
	}

	contacts := make([]*model.CustomerContactItem, 0, records.Len())
	for _, record := range records {
		contacts = append(contacts, &model.CustomerContactItem{
			Id:        record["id"].Uint64(),
			Name:      record["name"].String(),
			Phone:     record["phone"].String(),
			Tel:       record["tel"].String(),
			Email:     record["email"].String(),
			Wechat:    record["wechat"].String(),
			Position:  record["position"].String(),
			IsPrimary: record["is_primary"].Int(),
			Sort:      record["sort"].Int(),
			Remark:    record["remark"].String(),
		})
	}

	return contacts, nil
}

// replaceContacts 全量覆盖客户联系人
func replaceContacts(ctx context.Context, tx gdb.TX, customerId uint64, contacts []*model.CustomerContactSaveInput) error {
	if _, err := tx.Model(customerContactTable).Ctx(ctx).Where("customer_id", customerId).Delete(); err != nil {
		return err
	}

	now := gtime.Now()

	for index, contact := range contacts {
		if contact == nil {
			continue
		}

		name := strings.TrimSpace(contact.Name)
		if name == "" {
			continue
		}

		sortOrder := contact.Sort
		if sortOrder == 0 {
			sortOrder = index + 1
		}

		if _, err := tx.Model(customerContactTable).Ctx(ctx).Data(g.Map{
			"customer_id": customerId,
			"name":        name,
			"phone":       strings.TrimSpace(contact.Phone),
			"tel":         strings.TrimSpace(contact.Tel),
			"email":       strings.TrimSpace(contact.Email),
			"wechat":      strings.TrimSpace(contact.Wechat),
			"position":    strings.TrimSpace(contact.Position),
			"is_primary":  contact.IsPrimary,
			"sort":        sortOrder,
			"status":      1,
			"remark":      strings.TrimSpace(contact.Remark),
			"create_time": now,
			"update_time": now,
		}).Insert(); err != nil {
			return err
		}
	}

	return nil
}

// nextCustomerNo 生成客户编号，形如 C00001
func nextCustomerNo(ctx context.Context) (string, error) {
	record, err := g.Model(customerTable).Ctx(ctx).Fields("IFNULL(MAX(id),0) max_id").One()
	if err != nil {
		return "", err
	}

	next := record["max_id"].Int() + 1

	for i := 0; i < 1000; i++ {
		no := fmt.Sprintf("C%05d", next+i)

		exists, err := g.Model(customerTable).Ctx(ctx).Where("customer_no", no).Exist()
		if err != nil {
			return "", err
		}
		if !exists {
			return no, nil
		}
	}

	return "", gerror.New("生成客户编号失败，请重试")
}

// countByField 按指定字段分组计数，返回 值 -> 数量
func countByField(ctx context.Context, table, field string) (map[uint64]int, error) {
	records, err := g.Model(table).Ctx(ctx).Fields(field, "count(*) total").Group(field).All()
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]int, records.Len())
	for _, record := range records {
		result[record[field].Uint64()] = record["total"].Int()
	}

	return result, nil
}

// dateOnly 把 date / datetime 值规范成 YYYY-MM-DD，空值与零值返回空串
func dateOnly(value *gvar.Var) string {
	text := strings.TrimSpace(value.String())
	if text == "" || strings.HasPrefix(text, "0000-00-00") {
		return ""
	}
	if len(text) > 10 {
		return text[:10]
	}

	return text
}
