package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"x-foot/internal/model"
)

// ListCustomersReq 客户分页列表
type ListCustomersReq struct {
	g.Meta `path:"/api/customer/customers" method:"get" tags:"Customer" summary:"客户分页列表"`
	model.CustomerQuery
}

// ListCustomersRes 客户分页列表
type ListCustomersRes struct {
	model.CustomerListResult
}

// GetCustomerReq 客户详情
type GetCustomerReq struct {
	g.Meta `path:"/api/customer/customers/:id" method:"get" tags:"Customer" summary:"客户详情"`
	Id     uint64 `json:"id"`
}

// GetCustomerRes 客户详情
type GetCustomerRes struct {
	*model.CustomerItem
}

// SaveCustomerReq 新增或编辑客户
type SaveCustomerReq struct {
	g.Meta `path:"/api/customer/customers" method:"post" tags:"Customer" summary:"新增或编辑客户"`
	model.CustomerSaveInput
}

// SaveCustomerRes 新增或编辑客户
type SaveCustomerRes struct {
	Id uint64 `json:"id"`
}

// DeleteCustomerReq 删除客户
type DeleteCustomerReq struct {
	g.Meta `path:"/api/customer/customers/:id" method:"delete" tags:"Customer" summary:"删除客户"`
	Id     uint64 `json:"id"`
}

// DeleteCustomerRes 删除客户
type DeleteCustomerRes struct {
	g.Meta `mime:"application/json"`
}

// GetCustomerProfileReq 客户资料（关联的工程与合同）
type GetCustomerProfileReq struct {
	g.Meta `path:"/api/customer/customers/:id/profile" method:"get" tags:"Customer" summary:"客户资料：关联工程与合同"`
	Id     uint64 `json:"id"`
}

// GetCustomerProfileRes 客户资料
type GetCustomerProfileRes struct {
	model.CustomerProfileResult
}

// UploadAttachmentReq 上传客户附件
//
// 请求为 multipart/form-data，文件字段名为 file，
// 另有 customerId、attachType 两个表单字段。
type UploadAttachmentReq struct {
	g.Meta     `path:"/api/customer/attachments" method:"post" tags:"Customer" summary:"上传客户附件"`
	CustomerId uint64 `json:"customerId" v:"required#请先保存客户，再上传附件"`
	AttachType int    `json:"attachType"`
}

// UploadAttachmentRes 上传客户附件
type UploadAttachmentRes struct {
	model.CustomerAttachItem
}

// DeleteAttachmentReq 删除客户附件
type DeleteAttachmentReq struct {
	g.Meta `path:"/api/customer/attachments/:id" method:"delete" tags:"Customer" summary:"删除客户附件"`
	Id     uint64 `json:"id"`
}

// DeleteAttachmentRes 删除客户附件
type DeleteAttachmentRes struct {
	g.Meta `mime:"application/json"`
}
