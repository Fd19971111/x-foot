package entity

import "github.com/gogf/gf/v2/os/gtime"

// SiteCustomer 客户主表 site_customer
//
// CustomerName 承载「姓名」，CompanyName 承载「公司」，IdCard 承载「身份证号」。
type SiteCustomer struct {
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
	SalesId      uint64      `json:"salesId"`
	DealerId     uint64      `json:"dealerId"`
	IsActive     int         `json:"isActive"`
	Remark       string      `json:"remark"`
	CreatedAt    *gtime.Time `json:"createdAt"`
}

// SiteCustomerContact 客户联系人表 site_customer_contact
type SiteCustomerContact struct {
	Id         uint64      `json:"id"`
	CustomerId uint64      `json:"customerId"`
	Name       string      `json:"name"`
	Phone      string      `json:"phone"`
	Tel        string      `json:"tel"`
	Email      string      `json:"email"`
	Wechat     string      `json:"wechat"`
	Position   string      `json:"position"`
	IsPrimary  int         `json:"isPrimary"`
	Sort       int         `json:"sort"`
	Status     int         `json:"status"`
	Remark     string      `json:"remark"`
	CreateTime *gtime.Time `json:"createTime"`
	UpdateTime *gtime.Time `json:"updateTime"`
}

// SiteCustomerAttach 客户附件表 site_customer_attach
type SiteCustomerAttach struct {
	Id         uint64      `json:"id"`
	CustomerId uint64      `json:"customerId"`
	AttachType int         `json:"attachType"`
	FileName   string      `json:"fileName"`
	FilePath   string      `json:"filePath"`
	FileUrl    string      `json:"fileUrl"`
	FileExt    string      `json:"fileExt"`
	FileSize   int64       `json:"fileSize"`
	Remark     string      `json:"remark"`
	CreateTime *gtime.Time `json:"createTime"`
}
