package test

import (
	"time"

	goemail "github.com/JiuYu77/go-email"
)

var config = goemail.Config{
	SMTPConfig: goemail.SMTPConfig{
		Host:     "smtp.123.com",
		Port:     465, // 25 465 587
		Username: "example@123.com",
		Password: "1233345",
		From:     "example@123.com",
	},
	CodeExpiry:   5 * time.Minute,
	CacheCleanup: 10 * time.Minute,
}

var (
	email  = "example@123.com"
	email2 = "example1@123.com"
	smtp   = goemail.NewSMTP(
		config.Host, config.Port, config.Username, config.Password, config.From,
	)

	port = "8080"
	ip   = "http://172.25.157.206"
)
