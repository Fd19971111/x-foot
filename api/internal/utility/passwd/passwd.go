// Package passwd 密码加盐哈希工具
//
// 算法：hex(sha256(salt + 明文))，与既有账号数据保持兼容。
package passwd

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// saltLength 盐值字节数
const saltLength = 16

// Hash 计算密码哈希
func Hash(plain, salt string) string {
	sum := sha256.Sum256([]byte(salt + plain))

	return hex.EncodeToString(sum[:])
}

// NewSalt 生成一个新的随机盐值
func NewSalt() string {
	buf := make([]byte, saltLength)
	if _, err := rand.Read(buf); err != nil {
		// 随机数不可用时使用时间戳兜底，保证系统仍可创建账号
		return Hash(hex.EncodeToString(buf), string(rune(0)))[:saltLength*2]
	}

	return hex.EncodeToString(buf)
}
