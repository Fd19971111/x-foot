package company

import (
	"context"

	"x-foot/api/company/v1"
	"x-foot/internal/service"
)

// InitStatus 查询公司初始化状态
func (c *ControllerV1) InitStatus(ctx context.Context, req *v1.InitStatusReq) (res *v1.InitStatusRes, err error) {
	status, err := service.Company().InitStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.InitStatusRes{InitStatus: *status}, nil
}

// InitOptions 获取初始化默认数据
func (c *ControllerV1) InitOptions(ctx context.Context, req *v1.InitOptionsReq) (res *v1.InitOptionsRes, err error) {
	options, err := service.Company().InitOptions(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.InitOptionsRes{InitOptions: *options}, nil
}

// SubmitInit 提交公司初始配置
// 注意：不能命名为 Init，GoFrame 会把名为 Init 的方法当作框架初始化钩子 func(*ghttp.Request)
func (c *ControllerV1) SubmitInit(ctx context.Context, req *v1.InitReq) (res *v1.InitRes, err error) {
	if err = service.Company().Init(ctx, &req.InitInput); err != nil {
		return nil, err
	}
	return &v1.InitRes{}, nil
}
