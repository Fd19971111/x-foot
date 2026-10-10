package service

import (
	"context"

	"x-foot/internal/model"
)

// ICustomer 客户管理业务接口（客户信息 / 客户资料 / 附件）
type ICustomer interface {
	// ListCustomers 分页查询客户
	ListCustomers(ctx context.Context, in *model.CustomerQuery) (*model.CustomerListResult, error)
	// GetCustomer 查询客户详情（含联系人与附件）
	GetCustomer(ctx context.Context, id uint64) (*model.CustomerItem, error)
	// SaveCustomer 新增或编辑客户，返回客户ID
	SaveCustomer(ctx context.Context, in *model.CustomerSaveInput) (uint64, error)
	// DeleteCustomer 删除客户及其联系人与附件
	DeleteCustomer(ctx context.Context, id uint64) error

	// CustomerProfile 客户资料：该客户关联的工程与合同
	CustomerProfile(ctx context.Context, customerId uint64) (*model.CustomerProfileResult, error)

	// SaveAttachment 保存上传的附件
	SaveAttachment(ctx context.Context, in *model.AttachmentSaveInput) (*model.CustomerAttachItem, error)
	// DeleteAttachment 删除附件（同时清理磁盘文件）
	DeleteAttachment(ctx context.Context, id uint64) error
}

var localCustomer ICustomer

// Customer 获取客户管理业务实例
func Customer() ICustomer {
	if localCustomer == nil {
		panic("implement not found for interface ICustomer, forgot register?")
	}

	return localCustomer
}

// RegisterCustomer 注册客户管理业务实现
func RegisterCustomer(i ICustomer) {
	localCustomer = i
}
