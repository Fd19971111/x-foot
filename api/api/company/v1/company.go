package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"x-foot/internal/model"
)

// InitStatusReq 查询公司初始化状态
type InitStatusReq struct {
	g.Meta `path:"/api/company/initStatus" method:"get" tags:"Company" summary:"查询公司初始化状态"`
}

// InitStatusRes 查询公司初始化状态返回
type InitStatusRes struct {
	model.InitStatus
}

// InitOptionsReq 获取初始化默认数据
type InitOptionsReq struct {
	g.Meta `path:"/api/company/initOptions" method:"get" tags:"Company" summary:"获取初始化默认数据"`
}

// InitOptionsRes 获取初始化默认数据返回
type InitOptionsRes struct {
	model.InitOptions
}

// InitReq 提交公司初始配置
type InitReq struct {
	g.Meta `path:"/api/company/init" method:"post" tags:"Company" summary:"提交公司初始配置"`
	model.InitInput
}

// InitRes 提交公司初始配置返回
type InitRes struct {
	g.Meta `mime:"application/json"`
}
