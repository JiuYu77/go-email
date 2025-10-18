package goemail

import (
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
	// verifier
	Config           = verifier.Config
	VerificationCode = verifier.VerificationCode
	Buildler         = verifier.Buildler
	// cache
	CacheValue          = cache.CacheValue
	Cache[T CacheValue] = cache.Cache[T]
)

// function
var (
	// SMTP
	NewSMTP        = smtp.NewSMTP
	NewSMTPSender  = smtp.NewSMTPSender
	NewSMTPSender1 = smtp.NewSMTPSender1
	// Message
	NewMessage  = smtp.NewMessage
	SetCharset  = smtp.SetCharset
	SetEncoding = smtp.SetEncoding
	// file settings
	SetCopyFunc = smtp.SetCopyFunc
	SetHeader   = smtp.SetHeader
	Rename      = smtp.Rename
	// verifier
	ValidateFormat      = verifier.ValidateFormat
	NewVerifier         = verifier.NewVerifier
	GenerateSecureCode  = verifier.GenerateSecureCode
	GenerateNumberCode  = verifier.GenerateNumberCode
	GenerateRandomToken = verifier.GenerateRandomToken
	// cache
	NewCache = cache.NewCache[cache.CacheValue]
	// log
	SetMode = utils.SetMode
)

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
