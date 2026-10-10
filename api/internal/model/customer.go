// Package model 客户管理模块（客户信息 / 客户资料）的交互结构体
package model

import "github.com/gogf/gf/v2/os/gtime"

// ---------------------------------------------------------------------------------
// 枚举
// ---------------------------------------------------------------------------------

// 客户类型，与 site_customer.customer_type 枚举一致
const (
	CustomerTypePersonal   = "personal"
	CustomerTypeCompany    = "company"
	CustomerTypeGovernment = "government"
)

// 附件类型，与 site_customer_attach.attach_type 注释一致
const (
	AttachTypeLicense       = 1 // 营业执照
	AttachTypeIdCard        = 2 // 身份证
	AttachTypeQualification = 3 // 资质证书
	AttachTypeContract      = 4 // 合同
	AttachTypeOther         = 5 // 其他
)

// ---------------------------------------------------------------------------------
// 客户信息
// ---------------------------------------------------------------------------------

// CustomerQuery 客户列表查询条件
type CustomerQuery struct {
	// Keyword 客户编号 / 姓名 / 公司 / 电话 模糊匹配
	Keyword string `json:"keyword"`
	Page    int    `json:"page"`
	PageSize int   `json:"pageSize"`
}

// CustomerContactItem 联系人
type CustomerContactItem struct {
	Id        uint64 `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Tel       string `json:"tel"`
	Email     string `json:"email"`
	Wechat    string `json:"wechat"`
	Position  string `json:"position"`
	IsPrimary int    `json:"isPrimary"`
	Sort      int    `json:"sort"`
	Remark    string `json:"remark"`
}

// CustomerContactSaveInput 联系人保存入参
//
// 联系人按客户整体覆盖保存，不单独维护。
type CustomerContactSaveInput struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Tel       string `json:"tel"`
	Email     string `json:"email"`
	Wechat    string `json:"wechat"`
	Position  string `json:"position"`
	IsPrimary int    `json:"isPrimary"`
	Sort      int    `json:"sort"`
	Remark    string `json:"remark"`
}

// CustomerAttachItem 附件
type CustomerAttachItem struct {
	Id         uint64      `json:"id"`
	FileName   string      `json:"fileName"`
	FileUrl    string      `json:"fileUrl"`
	FileExt    string      `json:"fileExt"`
	FileSize   int64       `json:"fileSize"`
	AttachType int         `json:"attachType"`
	CreatedAt  *gtime.Time `json:"createdAt"`
}

// CustomerItem 客户列表行 / 详情
type CustomerItem struct {
	Id           uint64      `json:"id"`
	CustomerNo   string      `json:"customerNo"`
	CustomerName string      `json:"customerName"`
	CompanyName  string      `json:"companyName"`
	CustomerType string      `json:"customerType"`
	Phone        string      `json:"phone"`
	IdCard       string      `json:"idCard"`
	Email        string      `json:"email"`
	Address      string      `json:"address"`
	Level        string      `json:"level"`
	Source       string      `json:"source"`
	IsActive     int         `json:"isActive"`
	Remark       string      `json:"remark"`
	CreatedAt    *gtime.Time `json:"createdAt"`
	// ContactCount / AttachCount 列表页展示用
	ContactCount int `json:"contactCount"`
	AttachCount  int `json:"attachCount"`
	// Contacts / Attachments 仅详情接口返回
	Contacts    []*CustomerContactItem `json:"contacts,omitempty"`
	Attachments []*CustomerAttachItem  `json:"attachments,omitempty"`
}

// CustomerListResult 客户分页列表
type CustomerListResult struct {
	List     []*CustomerItem `json:"list"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// CustomerSaveInput 新增 / 编辑客户
//
// 只包含「客户信息」页面上的字段：姓名、公司、电话、身份证、地址、联系人、附件，
// 不含工程与合同（那属于客户资料）。
type CustomerSaveInput struct {
	Id           uint64 `json:"id"`
	CustomerName string `json:"customerName"`
	CompanyName  string `json:"companyName"`
	CustomerType string `json:"customerType"`
	Phone        string `json:"phone"`
	IdCard       string `json:"idCard"`
	Email        string `json:"email"`
	Address      string `json:"address"`
	Level        string `json:"level"`
	Source       string `json:"source"`
	// IsActive 1=启用，0=停用
	IsActive int `json:"isActive"`
	Remark   string `json:"remark"`
	// Contacts 全量覆盖式保存
	Contacts []*CustomerContactSaveInput `json:"contacts"`
}

// ---------------------------------------------------------------------------------
// 附件上传
// ---------------------------------------------------------------------------------

// AttachmentSaveInput 附件保存入参
//
// Content 为文件字节，由控制器从 multipart 请求中读出后传入，
// 这样业务层不依赖具体的 HTTP 实现。
type AttachmentSaveInput struct {
	CustomerId uint64 `json:"customerId"`
	AttachType int    `json:"attachType"`
	FileName   string `json:"fileName"`
	Content    []byte `json:"-"`
}

// ---------------------------------------------------------------------------------
// 客户资料（关联的工程与合同）
// ---------------------------------------------------------------------------------

// CustomerBrief 客户简要信息，用于客户资料页头部
type CustomerBrief struct {
	Id           uint64 `json:"id"`
	CustomerNo   string `json:"customerNo"`
	CustomerName string `json:"customerName"`
	CompanyName  string `json:"companyName"`
	Phone        string `json:"phone"`
}

// ProjectItem 关联工程
type ProjectItem struct {
	Id             uint64  `json:"id"`
	ProjectNo      string  `json:"projectNo"`
	ProjectName    string  `json:"projectName"`
	Address        string  `json:"address"`
	StartDate      string  `json:"startDate"`
	EndDate        string  `json:"endDate"`
	Status         string  `json:"status"`
	ContractAmount float64 `json:"contractAmount"`
	Remark         string  `json:"remark"`
}

// ContractItem 关联合同
type ContractItem struct {
	Id           uint64  `json:"id"`
	ContractNo   string  `json:"contractNo"`
	ContractName string  `json:"contractName"`
	ContractType int     `json:"contractType"`
	ProjectId    uint64  `json:"projectId"`
	ProjectName  string  `json:"projectName"`
	PartyA       string  `json:"partyA"`
	PartyB       string  `json:"partyB"`
	Amount       float64 `json:"amount"`
	TaxRate      float64 `json:"taxRate"`
	TaxAmount    float64 `json:"taxAmount"`
	SignDate     string  `json:"signDate"`
	StartDate    string  `json:"startDate"`
	EndDate      string  `json:"endDate"`
	Duration     int     `json:"duration"`
	Status       int     `json:"status"`
	Remark       string  `json:"remark"`
}

// CustomerProfileResult 客户资料：该客户关联的工程与合同
type CustomerProfileResult struct {
	Customer *CustomerBrief `json:"customer"`
	Projects []*ProjectItem `json:"projects"`
	Contracts []*ContractItem `json:"contracts"`
	// 汇总信息，便于页面头部展示
	ProjectCount        int     `json:"projectCount"`
	ContractCount       int     `json:"contractCount"`
	ContractAmountTotal float64 `json:"contractAmountTotal"`
}
