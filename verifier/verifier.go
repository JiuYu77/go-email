package verifier

import (
	"errors"
	"sync"
	"time"

	cache "github.com/JiuYu77/go-email/cache"
	smtp "github.com/JiuYu77/go-email/smtp"
)

type Verifier struct {
	config *Config
	cache  *cache.Cache[*VerificationCode]
	sender *smtp.SMTPSender
	mtx    sync.Mutex
}

func NewVerifier(config *Config) *Verifier {
	if config.CodeExpiry <= 0 {
		config.CodeExpiry = 5 * time.Minute
	}

	if config.CacheCleanup <= 0 {
		config.CacheCleanup = 10 * time.Minute
	}

	if config.CodeLength <= 0 {
		config.CodeLength = 6 // default
	}
	if config.TokenLength <= 0 {
		config.TokenLength = 32 // default
	}

	verifier := &Verifier{
		config: config,
		cache:  cache.NewCache[*VerificationCode](config.CacheCleanup),
		sender: smtp.NewSMTPSender(
			smtp.NewSMTP(
				config.SMTPConfig.Host, config.SMTPConfig.Port,
				config.SMTPConfig.Username, config.SMTPConfig.Password, config.SMTPConfig.From,
			),
		),
	}
	return verifier
}

func (v *Verifier) sendEmail(email []string, msg []byte) error {
	v.mtx.Lock()
	defer v.mtx.Unlock()

	if err := v.sender.Dial(); err != nil {
		return err
	}
	return v.sender.Send1(email, msg)
}

// 构建并发送邮件
func (v *Verifier) sendVerificationEmail(email string, data string, buildEmail Buildler) error {
	msg, err := buildEmail(email, data)
	if err != nil {
		return err
	}
	return v.sendEmail([]string{email}, msg)
}

func (v *Verifier) ValidateFormat(email string) (bool, error) {
	matched, err := ValidateFormat(email)
	if err != nil {
		return matched, err
	}
	return matched, nil
}

// SendVerificationCode 发送验证码.
//
// # Args
//   - email {string} 接收者邮箱
//   - buildEmail {Buildler} 邮件构建器
//
// # Returns
//   - code {string} 验证码，数字组成的字符串
//   - err {error} 错误信息；nil 表示成功
func (v *Verifier) SendVerificationCode(email string, buildEmail Buildler) (code string, err error) {
	if matched, _ := v.ValidateFormat(email); !matched {
		return "", errors.New("email format is invalid")
	}

	// 生成验证码
	code, err = GenerateNumberCode(v.config.CodeLength)
	if err != nil {
		return "", err
	}

	// 存储验证码
	verCode := &VerificationCode{
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(v.config.CodeExpiry),
		Used:      false,
	}
	v.cache.Set(email, verCode)

	// 发送邮件: 发送验证码到邮箱
	if err := v.sendVerificationEmail(email, code, buildEmail); err != nil {
		v.cache.Delete(email)
		return "", errors.New("send email failed")
	}

	return code, nil
}

// VerifyCode 验证验证码.
//
// # Args
//
//	email {string} 接收者邮箱
//	code {string} 邮件构建器
//
// # Returns
//
//	一个错误信息，nil 表示验证成功.
func (v *Verifier) VerifyCode(email string, code string) error {
	cached, ok := v.cache.Get(email)
	if !ok {
		return errors.New("verification code not found or expired")
	}

	verificationCode, ok := cached.(*VerificationCode)
	if !ok {
		return errors.New("verification code type is invalid")
	}

	if time.Now().After(verificationCode.ExpiresAt) {
		v.cache.Delete(email)
		return errors.New("verification code has expired")
	}

	if verificationCode.Used {
		return errors.New("verification code already used")
	}

	if verificationCode.Code != code {
		return errors.New("verification code is invalid")
	}

	// 标记为已使用
	verificationCode.Used = true
	v.cache.Set(email, verificationCode)

	return nil // 验证成功
}

// SendConfirmationLink 发送确认链接。
func (v *Verifier) SendConfirmationLink(email string, buildEmail Buildler) (string, error) {
	token, err := GenerateRandomToken(v.config.TokenLength)
	if err != nil {
		return "", err
	}

	// 存储token
	code := &VerificationCode{
		Email:     email,
		Code:      token,
		ExpiresAt: time.Now().Add(v.config.CodeExpiry),
		Used:      false,
	}
	v.cache.Set(token, code)

	if err := v.sendVerificationEmail(email, token, buildEmail); err != nil {
		v.cache.Delete(token)
		return "", errors.New("send email failed")
	}

	return token, nil
}

// VerifyConfirmationLink 验证确认链接。
func (v *Verifier) VerifyConfirmationLink(token string) (string, error) {
	// 从缓存中获取确认链接
	cached, ok := v.cache.Get(token)
	if !ok {
		return "", errors.New("invalid or expired confirmation token")
	}

	code, ok := cached.(*VerificationCode)
	if !ok {
		return "", errors.New("confirmation link type is invalid")
	}

	if time.Now().After(code.ExpiresAt) {
		return "", errors.New("confirmation link expired")
	}

	if code.Used {
		return "", errors.New("confirmation link already used")
	}

	if code.Code != token {
		return "", errors.New("confirmation link is invalid")
	}

	// 标记为已确认
	code.Used = true
	v.cache.Set(token, code)

	return code.Email, nil // 验证成功
}

// 综合验证
