package model

// LoginInput 登录入参
type LoginInput struct {
	UserName string
	Password string
}

// LoginToken 登录令牌
type LoginToken struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

// UserInfo 当前登录用户信息
type UserInfo struct {
	UserId   string   `json:"userId"`
	UserName string   `json:"userName"`
	Roles    []string `json:"roles"`
	Buttons  []string `json:"buttons"`
}
