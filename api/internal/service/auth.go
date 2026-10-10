package service

import (
	"context"

	"x-foot/internal/model"
)

// IAuth 认证业务接口
type IAuth interface {
	// Login 账号密码登录，返回访问令牌与刷新令牌
	Login(ctx context.Context, in *model.LoginInput) (*model.LoginToken, error)
	// GetUserInfo 根据用户ID获取登录用户信息
	GetUserInfo(ctx context.Context, userId uint64) (*model.UserInfo, error)
	// RefreshToken 使用刷新令牌换取新的令牌
	RefreshToken(ctx context.Context, refreshToken string) (*model.LoginToken, error)
	// SeedDefaultUsers 当系统没有任何用户时，初始化默认账号
	SeedDefaultUsers(ctx context.Context) error
}

var localAuth IAuth

// Auth 获取认证业务实例
func Auth() IAuth {
	if localAuth == nil {
		panic("implement not found for interface IAuth, forgot register?")
	}
	return localAuth
}

// RegisterAuth 注册认证业务实现
func RegisterAuth(i IAuth) {
	localAuth = i
}
