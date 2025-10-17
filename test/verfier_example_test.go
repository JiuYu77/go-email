package test

import (
	"fmt"
	"testing"

	goemail "github.com/JiuYu77/go-email"
)

func TestVerifierExample(t *testing.T) {
	verifier := goemail.NewVerifier(&config)

	// 验证码
	code, err := verifier.SendVerificationCode(email, BuildVerificationCode)
	if err != nil {
		t.Error("code:", err)
	} else {
		fmt.Println(">>>>code:", code)
	}
	err = verifier.VerifyCode(email, code) // 验证
	if err != nil {
		fmt.Println(">>>>verifer code:", err)
	} else {
		fmt.Println(">>>>verifer code success")
	}

	// 确认链接
	token, err := verifier.SendConfirmationLink(email, BuildConfirmationLink)
	if err != nil {
		t.Error("token:", err)
	} else {
		fmt.Println(">>>>token:", token)
	}

	em, err := verifier.VerifyConfirmationLink(token) // 验证
	if err != nil {
		t.Error(">>>>>verifier token:", err)
	} else {
		if em == email {
			fmt.Println(">>>>verifier token success:", em)
		}
	}
}
