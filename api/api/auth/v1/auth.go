package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"x-foot/internal/model"
)

// LoginReq 登录
type LoginReq struct {
	g.Meta   `path:"/auth/login" method:"post" tags:"Auth" summary:"账号密码登录"`
	UserName string `json:"userName" v:"required#请输入用户名" dc:"用户名"`
	Password string `json:"password" v:"required#请输入密码" dc:"密码"`
}

// LoginRes 登录返回
type LoginRes struct {
	model.LoginToken
}

// GetUserInfoReq 获取登录用户信息
type GetUserInfoReq struct {
	g.Meta `path:"/auth/getUserInfo" method:"get" tags:"Auth" summary:"获取当前登录用户信息"`
}

// GetUserInfoRes 获取登录用户信息返回
type GetUserInfoRes struct {
	model.UserInfo
}

// RefreshTokenReq 刷新令牌
type RefreshTokenReq struct {
	g.Meta       `path:"/auth/refreshToken" method:"post" tags:"Auth" summary:"刷新登录令牌"`
	RefreshToken string `json:"refreshToken" dc:"刷新令牌"`
}

// RefreshTokenRes 刷新令牌返回
type RefreshTokenRes struct {
	model.LoginToken
}
