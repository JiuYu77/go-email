# go-email

这是一个**Golang邮箱模块**。

本项目基于[**github.com/go-gomail/gomail**](https://github.com/go-gomail/gomail) 。

学习了[**github.com/wneessen/go-mail**](https://github.com/wneessen/go-mail)。

## 安装 Install
```shell
go get github.com/JiuYu77/go-email
```

## 导包
```go
import goemail "github.com/JiuYu77/go-email"
```

## smtp包 package

### 简介

此部分源码，请查看[代码目录](smtp/)。

**SMTP**
* [X] `from` 属性（发件人邮箱），默认为登录SMTP的邮箱，可直接用来发邮件
* [X] 有些 SMTP 服务器 587端口 也是隐式TLS加密，这时需要设置 `SSL` 为 true, 否则会连接失败
* [X] 可发送 `Message`对象、简单的 `[]byte` 邮件内容

**Message**
* [X] 代码重构，如将部分数据类型提取到`types.go`
* [X] 修复发现的`bug`，如文件存在性检查

**MessageWriter**
* [X] 发送邮件内容，Message 的 `WriteTo` 方法中使用

**SMTPSender**
* [X] SMTP 邮件发送器
* [X] 可发送 Message 类型
* [X] 也可直接发送 []byte 消息

### 示例 Example

更多内容请查看：[**邮件发送示例**](test/send_example_test.go)

**示例：**
```go
package main

import goemail "github.com/JiuYu77/go-email"

func main() {
    msg := goemail.NewMessage()
    msg.SetFrom("example@123.com", "Sora")
    msg.SetTo([]string{"example@123.com"})
    msg.SetSubject("Hello")
    // msg.SetBody("text/plain", "This is an email.")
    msg.SetBody("text/html", "<p>This is an email.</p>")

    msg.Attach("a.zip") // 附件

    // 内嵌文件
    err := msg.Embed("/home/jyu/Desktop/截图_选择区域_20251013145057.png")
    if err == nil {
        // 使用默认的 Content-ID，即文件名
        msg.AddAlternative("text/html", `<img src="cid:截图_选择区域_20251013145057.png" alt="Logo">`) 
    }

    err = msg.Embed("/home/jyu/Desktop/截图_选择区域_20251013145057.png", goemail.SetHeader(map[string][]string{"Content-ID": {"<test-content-id>"}}))
    if err == nil {
        // 使用指定的 Content-ID
        msg.AddAlternative("text/html", "<img src=\"cid:test-content-id\" alt=\"Logo\">") 
    }

    smtp := goemail.NewSMTP("smtp.example.com", 465, "example@123.com", "123456", "")

    smtp.DialAndSend(true, msg)

    // byte msg
    smtp.DialAndSend1([]string{"example@123.com"}, []byte("Subject: Hi\r\nFrom: example@123.com\r\nTo: example@123.com\r\n\r\nHello Golang!"))
}
```

## 邮箱验证工具

### 简介

此部分源码，请查看[代码目录](verifier/)。

* 验证功能
  * [X] 验证邮箱格式合法性
  * [X] 安全码生成功能，可用于发送验证码，确认链接
  * [X] 安全码缓存功能，过期自动删除

### 示例 Example

更多内容请查看：[**邮件验证工具示例**](test/verfier_example_test.go)

**示例：**
```go
package main

import (
	"fmt"

	goemail "github.com/JiuYu77/go-email"
)

func main() {
    var config = goemail.Config{
        SMTPConfig: goemail.SMTPConfig{
            Host:     "smtp.123.com",
            Port:     465, // 25 465 587
            Username: "example@123.com",
            Password: "123456",
            From:     "example@123.com",
        },
        CodeExpiry:   5 * time.Minute,
        CacheCleanup: 10 * time.Minute,
    }
    email := "example@234.com"

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
```

## logx

### 简介

此部分源码，请查看[代码目录](log/)。

- [**NewLogger**](log/logger.go)
- [**NewJsonLogger**](log/json.go) JSON日志（JSON格式的日志）
- [**NewJsonRotation**](log/rotation.go) JSON日志，日志轮转（基于文件大小）

### 示例 Example

更多内容请查看：[**logx测试**](test/logx_test.go)
