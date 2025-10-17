package test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	goemail "github.com/JiuYu77/go-email"
)

func TestValidateFormat(t *testing.T) {
	testcases := []struct {
		in   string
		want bool
	}{
		{"soramaix@QQ.com", true},
		{"hjkjhk@645654.2121-6878.com.wcn", true},
		{"xxxxxxxxx@wwew-163.com.cn", true},

		{"441030517@QQ..com", false},
		{"119941779@qq,com", false},
		{"5579001QQ@.COM", false},
		{"1107531656@q?q?.com", false},
		{"654088115@@qq.com", false},
		{"495456580@qq@139.com", false},
		{"279985462@qq。com.cn", false},
		{"chen@foxmail.com)m", false},
		{"2990814514@?￡QQ.COM", false},
		{"xxxxxxxxx@___.com.cn", false},
		{"xxxxxxxxx@wwew_163sadasdf.com.cn", false},
	}
	for _, tc := range testcases {
		r, err := goemail.ValidateFormat(tc.in)
		if err != nil {
			t.Errorf("email: %q, error: %q", tc.in, err)
		}
		if r != tc.want {
			t.Errorf("email: %q, return: %v, want %v", tc.in, r, tc.want)
		}
	}
}

func Test_smtp_code_link(t *testing.T) {
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

	server := http.Server{Addr: ":" + port}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			return
		}

		// 确认链接
		em, err := verifier.VerifyConfirmationLink(token) // 验证
		if err != nil {
			t.Error(">>>>>verifier token:", err)
			w.Write([]byte("验证失败:" + err.Error()))
		} else {
			if em == email {
				fmt.Println(">>>>verifier token success:", em)
				w.Write([]byte("验证成功!"))
				go func() {
					fmt.Println("准备关闭服务器...")
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					if err := server.Shutdown(ctx); err != nil {
						fmt.Printf("服务器关闭错误: %v\n", err)
					}
				}()
			} else {
				fmt.Printf(">>>>verifier token failed: email:%v, em:%v\n", email, em)
				w.Write([]byte("验证失败!"))
			}
		}
	})

	server.ListenAndServe()
	fmt.Println("测试完成，服务器已关闭")
}
