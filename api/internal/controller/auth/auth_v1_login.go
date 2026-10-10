package auth

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"x-foot/api/auth/v1"
	"x-foot/internal/model"
	"x-foot/internal/service"
)

// Login 账号密码登录
func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	data, err := service.Auth().Login(ctx, &model.LoginInput{
		UserName: req.UserName,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &v1.LoginRes{LoginToken: *data}, nil
}

// GetUserInfo 获取当前登录用户信息
func (c *ControllerV1) GetUserInfo(ctx context.Context, req *v1.GetUserInfoReq) (res *v1.GetUserInfoRes, err error) {
	userId := g.RequestFromCtx(ctx).GetCtxVar("userId").Uint64()

	data, err := service.Auth().GetUserInfo(ctx, userId)
	if err != nil {
		return nil, err
	}

	return &v1.GetUserInfoRes{UserInfo: *data}, nil
}

// RefreshToken 刷新登录令牌
func (c *ControllerV1) RefreshToken(ctx context.Context, req *v1.RefreshTokenReq) (res *v1.RefreshTokenRes, err error) {
	data, err := service.Auth().RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &v1.RefreshTokenRes{LoginToken: *data}, nil
}
