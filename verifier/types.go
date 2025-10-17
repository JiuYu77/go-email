package verifier

import (
	"time"

	"github.com/JiuYu77/go-email/smtp"
)

// 验证码信息
type VerificationCode struct {
	Email     string    `json:"email"`      // 邮箱地址
	Code      string    `json:"code"`       // 验证码
	ExpiresAt time.Time `json:"expires_at"` // 验证码过期时间
	Used      bool      `json:"used"`       // 验证码是否已被使用
}

func (vc *VerificationCode) Expiry() time.Time {
	return vc.ExpiresAt
}

// 配置
type Config struct {
	smtp.SMTPConfig

	CodeExpiry   time.Duration `json:"code_expiry"`   // 验证码/token 过期时间
	CacheCleanup time.Duration `json:"cache_cleanup"` // 缓存清理时间间隔
	CodeLength   int           `json:"code_length"`   // 验证码长度
	TokenLength  int           `json:"token_length"`  // token长度
}

type Buildler func(email, data string) ([]byte, error)
