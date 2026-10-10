package service

import (
	"context"

	"x-foot/internal/model"
)

// ICompany 公司初始化业务接口
type ICompany interface {
	// InitStatus 查询公司是否已完成初始化
	InitStatus(ctx context.Context) (*model.InitStatus, error)
	// InitOptions 获取初始化页面所需的默认科目、默认映射、默认账户
	InitOptions(ctx context.Context) (*model.InitOptions, error)
	// Init 提交公司初始配置
	Init(ctx context.Context, in *model.InitInput) error
}

var localCompany ICompany

// Company 获取公司初始化业务实例
func Company() ICompany {
	if localCompany == nil {
		panic("implement not found for interface ICompany, forgot register?")
	}
	return localCompany
}

// RegisterCompany 注册公司初始化业务实现
func RegisterCompany(i ICompany) {
	localCompany = i
}
