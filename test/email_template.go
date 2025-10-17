package test

import (
	"errors"
	"fmt"
)

func BuildVerificationCode(to string, code string) ([]byte, error) {
	if to == "" {
		return nil, errors.New("收件人不能为空")
	}

	from := config.From
	subject := "验证码"

	body := fmt.Sprintf("您的验证码为: %s, 有效期为 %v 分钟", code, config.CodeExpiry.Minutes())
	msg := fmt.Sprintf("Subject: %s\r\nFrom: %s\r\nTo: %s\r\n\r\n%s", subject, from, to, body)

	return []byte(msg), nil
}

func BuildConfirmationLink(to string, token string) ([]byte, error) {
	if to == "" {
		return nil, errors.New("收件人不能为空")
	}

	from := config.From
	subject := "确认链接"
	confirmationLink := fmt.Sprintf("%s?token=%s", ip+":"+port, token) // 构建确认链接

	body := fmt.Sprintf(
		`请点击以下链接确认您的邮箱，如链接无法点击请复制到浏览器打开: <a href="%s" target="_blank">%s</a>`,
		confirmationLink, confirmationLink,

	// `请点击以下链接确认您的邮箱，如链接无法点击请复制到浏览器打开: %s`,
	// confirmationLink,
	)

	msg := fmt.Sprintf("Subject: %s\r\n", subject) +
		fmt.Sprintf("From: %s\r\n", from) +
		fmt.Sprintf("To: %s\r\n", to) +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n\r\n" +
		body

	return []byte(msg), nil
}
