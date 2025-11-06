package goemail

import (
	"time"

	"github.com/JiuYu77/go-email/cache"
	logx "github.com/JiuYu77/go-email/log"
	"github.com/JiuYu77/go-email/smtp"
	"github.com/JiuYu77/go-email/utils"
	"github.com/JiuYu77/go-email/verifier"
)

type (
	// SMTP
	SMTP           = smtp.SMTP
	SMTPConfig     = smtp.SMTPConfig
	SMTPSender     = smtp.SMTPSender
	Message        = smtp.Message
	MessageSetting = smtp.MessageSetting
	PartSetting    = smtp.PartSetting
	FileSetting    = smtp.FileSetting
	Encoding       = smtp.Encoding
	Copier         = smtp.Copier
	Header         = smtp.Header
	// verifier
	Config           = verifier.Config
	Verifier         = verifier.Verifier
	VerificationCode = verifier.VerificationCode
	Buildler         = verifier.Buildler
	// cache
	CacheValue          = cache.CacheValue
	Cache[T CacheValue] = cache.Cache[T]
)

// SMTP
func NewSMTP(host string, port int, username, password, from string) *SMTP {
	return smtp.NewSMTP(host, port, username, password, from)
}

// SMTPSender
func NewSMTPSender(s *SMTP) *SMTPSender {
	return smtp.NewSMTPSender(s)
}
func NewSMTPSender1(client *smtp.Client, s *SMTP) *SMTPSender {
	return smtp.NewSMTPSender1(client, s)
}

// Message
func NewMessage(settings ...MessageSetting) *Message {
	return smtp.NewMessage(settings...)
}
func SetCharset(charset string) MessageSetting {
	return smtp.SetCharset(charset)
}
func SetEncoding(encoding Encoding) MessageSetting {
	return smtp.SetEncoding(encoding)
}

// file settings
func SetCopyFunc(copier Copier) FileSetting {
	return smtp.SetCopyFunc(copier)
}
func SetHeader(h map[string][]string) FileSetting {
	return smtp.SetHeader(h)
}
func Rename(filename string) FileSetting {
	return smtp.Rename(filename)
}

// verifier
func ValidateFormat(email string) (bool, error) {
	return verifier.ValidateFormat(email)
}
func NewVerifier(cfg *Config) *verifier.Verifier {
	return verifier.NewVerifier(cfg)
}
func GenerateSecureCode(length int, charset string) (string, error) {
	return verifier.GenerateSecureCode(length, charset)
}
func GenerateNumberCode(length int) (string, error) {
	return verifier.GenerateNumberCode(length)
}
func GenerateRandomToken(length int) (string, error) {
	return verifier.GenerateRandomToken(length)
}

// cache
func NewCache(cleanupInterval time.Duration) *Cache[CacheValue] {
	return cache.NewCache[CacheValue](cleanupInterval)
}

// log
func SetMode(mode logx.LogLevel) {
	utils.SetMode(mode)
}

const (
	// verifier
	Numbers      = verifier.Numbers
	UpperLetters = verifier.UpperLetters
	LowerLetters = verifier.LowerLetters
	AlphaNumeric = verifier.AlphaNumeric
	// log
	DebugMode   = logx.DebugLevel
	ReleaseMode = logx.PanicLevel
)
