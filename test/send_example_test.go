package test

import (
	"testing"

	goemail "github.com/JiuYu77/go-email"
)

func TestExample(t *testing.T) {
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
		msg.AddAlternative("text/html", `<img src="cid:截图_选择区域_20251013145057.png" alt="Logo">`) // 使用默认的 Content-ID，即文件名
	}

	err = msg.Embed("/home/jyu/Desktop/截图_选择区域_20251013145057.png", goemail.SetHeader(map[string][]string{"Content-ID": {"<test-content-id>"}}))
	if err == nil {
		msg.AddAlternative("text/html", "<img src=\"cid:test-content-id\" alt=\"Logo\">") // 使用指定的 Content-ID
	}

	smtp := goemail.NewSMTP("smtp.example.com", 465, "example@123.com", "123456", "")

	smtp.DialAndSend(true, msg)
	// byte msg
	smtp.DialAndSend1([]string{"example@123.com"}, []byte("Subject: Hi\r\nFrom: example@123.com\r\nTo: example@123.com\r\n\r\nHello Golang!"))
}

func TestExample2(t *testing.T) {
	msg := goemail.NewMessage()
	msg.SetFrom("example@123.com", "Sora")
	msg.SetTo([]string{"example@123.com"})
	msg.SetSubject("Hello")
	msg.SetBody("text/plain", "This is an email!")

	msg2 := goemail.NewMessage()
	msg2.SetFrom("example@123.com", "Sora")
	msg2.SetTo([]string{"example@123.com"})
	msg2.SetSubject("Hello2")
	msg2.SetBody("text/plain", "This is a message!")

	smtp := goemail.NewSMTP("smtp.example.com", 465, "example@123.com", "123456", "")
	sender, _ := smtp.Dial() // connect to SMTP server
	sender.Send(true, msg, msg2)
}

func TestExample3(t *testing.T) {
	msg := goemail.NewMessage()
	msg.SetFrom("example@123.com", "Sora")
	msg.SetTo([]string{"example@123.com"})
	msg.SetSubject("Hello")
	msg.SetBody("text/plain", "This is an email!")

	msg2 := goemail.NewMessage()
	msg2.SetFrom("example@123.com", "Sora")
	msg2.SetTo([]string{"example@123.com"})
	msg2.SetSubject("Hello2")
	msg2.SetBody("text/plain", "This is a message!")

	sender := goemail.NewSMTPSender(
		goemail.NewSMTP("smtp.example.com", 465, "example@123.com", "123456", ""),
	)
	sender.Dial() // connect to SMTP server
	sender.Send(true, msg)
	sender.Send(true, msg2)
}
