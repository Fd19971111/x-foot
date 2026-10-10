package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"x-foot/internal/model"
	"x-foot/internal/service"
	"x-foot/internal/utility/token"
)

const (
	userTable = "site_user"
	// refreshTTL 刷新令牌有效期
	refreshTTL = 7 * 24 * time.Hour
)

type sAuth struct{}

func init() {
	service.RegisterAuth(New())
}

// New 创建认证业务实例
func New() service.IAuth {
	return &sAuth{}
}

// Login 账号密码登录
func (s *sAuth) Login(ctx context.Context, in *model.LoginInput) (*model.LoginToken, error) {
	if in == nil || strings.TrimSpace(in.UserName) == "" || in.Password == "" {
		return nil, gerror.New("请输入用户名和密码")
	}
	userName := strings.TrimSpace(in.UserName)

	record, err := g.Model(userTable).Where("username", userName).One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, gerror.New("用户名或密码错误")
	}

	if record["password"].String() != hashPassword(in.Password, record["salt"].String()) {
		_, _ = g.Model(userTable).Where("id", record["id"]).Increment("login_fail_count", 1)
		return nil, gerror.New("用户名或密码错误")
	}

	switch record["status"].String() {
	case "disabled", "locked":
		return nil, gerror.New("账号已被禁用或锁定，请联系管理员")
	}

	userId := record["id"].Uint64()
	ip := ""
	if r := g.RequestFromCtx(ctx); r != nil {
		ip = r.GetClientIp()
	}
	_, _ = g.Model(userTable).Where("id", userId).Data(g.Map{
		"last_login_at":    gtime.Now(),
		"last_login_ip":    ip,
		"login_fail_count": 0,
	}).Update()

	return s.issueToken(userId), nil
}

// GetUserInfo 获取登录用户信息
func (s *sAuth) GetUserInfo(ctx context.Context, userId uint64) (*model.UserInfo, error) {
	record, err := g.Model(userTable).Where("id", userId).One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, gerror.New("用户不存在或已被删除")
	}

	userName := record["real_name"].String()
	if userName == "" {
		userName = record["username"].String()
	}

	return &model.UserInfo{
		UserId:   gconv.String(userId),
		UserName: userName,
		Roles:    rolesOf(record["role"].String()),
		Buttons:  []string{},
	}, nil
}

// RefreshToken 使用刷新令牌换取新的令牌
func (s *sAuth) RefreshToken(ctx context.Context, refreshToken string) (*model.LoginToken, error) {
	userId, err := token.Parse(refreshToken)
	if err != nil {
		return nil, gerror.New("登录状态已过期，请重新登录")
	}

	record, err := g.Model(userTable).Where("id", userId).One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, gerror.New("用户不存在或已被删除")
	}

	return s.issueToken(userId), nil
}

// SeedDefaultUsers 系统无任何用户时初始化默认账号
func (s *sAuth) SeedDefaultUsers(ctx context.Context) error {
	count, err := g.Model(userTable).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	seeds := []struct {
		UserName string
		Password string
		RealName string
		Role     string
	}{
		{UserName: "Soybean", Password: "123456", RealName: "超级管理员", Role: "admin"},
		{UserName: "Admin", Password: "123456", RealName: "财务主管", Role: "finance"},
		{UserName: "User", Password: "123456", RealName: "普通员工", Role: "staff"},
	}

	for _, item := range seeds {
		salt := randomSalt()
		if _, err = g.Model(userTable).Data(g.Map{
			"username":   item.UserName,
			"password":   hashPassword(item.Password, salt),
			"salt":       salt,
			"real_name":  item.RealName,
			"role":       item.Role,
			"status":     "active",
			"created_at": gtime.Now(),
			"updated_at": gtime.Now(),
		}).Insert(); err != nil {
			return err
		}
	}

	g.Log().Warning(ctx, "系统当前无任何用户，已初始化默认账号 Soybean / Admin / User（初始密码均为 123456），请登录后尽快修改")

	return nil
}

func (s *sAuth) issueToken(userId uint64) *model.LoginToken {
	return &model.LoginToken{
		Token:        token.Generate(userId, token.TTL()),
		RefreshToken: token.Generate(userId, refreshTTL),
	}
}

// rolesOf 将数据库角色映射为前端角色标识
func rolesOf(role string) []string {
	switch role {
	case "admin":
		return []string{"R_SUPER"}
	case "finance":
		return []string{"R_FINANCE"}
	default:
		return []string{"R_STAFF"}
	}
}

func randomSalt() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(buf)
}

func hashPassword(plain, salt string) string {
	sum := sha256.Sum256([]byte(salt + plain))

	return hex.EncodeToString(sum[:])
}
