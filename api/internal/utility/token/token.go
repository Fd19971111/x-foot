// Package token 提供无状态的签名令牌生成与校验
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

const (
	defaultSecret = "x-foot-default-secret-please-change"
	defaultTTL    = 2 * time.Hour
)

// ErrInvalidToken 令牌无效或已过期
var ErrInvalidToken = errors.New("invalid token")

// secret 读取配置中的签名密钥
func secret() string {
	value := g.Cfg().MustGet(gctx.New(), "auth.secret", defaultSecret).String()
	if value == "" {
		return defaultSecret
	}
	return value
}

// TTL 默认令牌有效期
func TTL() time.Duration {
	seconds := g.Cfg().MustGet(gctx.New(), "auth.ttl", int(defaultTTL.Seconds())).Int()
	if seconds <= 0 {
		return defaultTTL
	}
	return time.Duration(seconds) * time.Second
}

// Generate 生成令牌，payload 为 userId|过期毫秒时间戳，签名使用 HMAC-SHA256
func Generate(userId uint64, ttl time.Duration) string {
	if ttl <= 0 {
		ttl = TTL()
	}
	payload := fmt.Sprintf("%d|%d", userId, time.Now().Add(ttl).UnixMilli())
	signature := sign(payload)

	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + signature
}

// Parse 校验令牌并返回 userId
func Parse(token string) (userId uint64, err error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, ErrInvalidToken
	}

	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, ErrInvalidToken
	}

	payload := string(raw)
	if !hmac.Equal([]byte(sign(payload)), []byte(parts[1])) {
		return 0, ErrInvalidToken
	}

	var expireAt int64
	if _, err = fmt.Sscanf(payload, "%d|%d", &userId, &expireAt); err != nil {
		return 0, ErrInvalidToken
	}
	if time.Now().UnixMilli() > expireAt {
		return 0, ErrInvalidToken
	}

	return userId, nil
}

func sign(payload string) string {
	h := hmac.New(sha256.New, []byte(secret()))
	h.Write([]byte(payload))

	return hex.EncodeToString(h.Sum(nil))
}
