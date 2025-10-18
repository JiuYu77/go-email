package test

import (
	"io"
	"path/filepath"
	"testing"

	goemail "github.com/JiuYu77/go-email"
)

func Test_smtp_sender(t *testing.T) {
	// goemail.SetMode(goemail.ReleaseMode) // 禁用调试信息

	msg := goemail.NewMessage()
	es := []string{email}
	es = append(es, msg.FormatAddress(email2, "123多个"))

	msg.SetFrom(config.From, "JiuYu77")
	msg.SetTo(es)
	msg.SetSubject("Test Subject")
	// msg.SetBody("text/plain", "Test_smtp_sender!") // 纯文本格式
	msg.SetBody("text/html", "Test_smtp_sender!")

	msg.Attach("../go.mod")

	msg.AddAlternative("text/html", `<img src="cid:截图_选择区域_20251013145057.png" alt="Logo">`) // 使用默认的 Content-ID，即文件名
	msg.Embed("/home/jyu/Desktop/截图_选择区域_20251013145057.png")
	// MockCopyFile 模拟文件复制函数，返回文件名和文件设置，它发送的是文件名，而不是文件内容，导致嵌入图片失败
	// msg.Embed(MockCopyFile("/home/jyu/Desktop/截图_选择区域_20251013145057.png"))

	msg.AddAlternative("text/html", "<img src=\"cid:test-content-id\" alt=\"Logo\">") // 使用指定的 Content-ID
	msg.Embed("/home/jyu/Desktop/截图_选择区域_20251013145057.png", goemail.SetHeader(map[string][]string{"Content-ID": {"<test-content-id>"}}))
	// MockCopyFileWithHeader 中 mockCopyFile 模拟文件复制函数，返回文件名和文件设置，它发送的是文件名，而不是文件内容，导致嵌入图片失败
	// msg.Embed(MockCopyFileWithHeader("/home/jyu/Desktop/截图_选择区域_20251013145057.png", map[string][]string{"Content-ID": {"<test-content-id>"}}))

	// smtp.SSL = true // 587 端口也可能使用隐式TLS加密, 如 126.com 邮箱

	err := smtp.DialAndSend(true, msg) // 1.
	if err != nil {
		t.Error("[SMTP test] send email failed:", err)
	}

	sender, err := smtp.Dial()
	if err != nil {
		t.Error("[Test_smtp_sender] [SMTP test] dial failed:", err)
	}

	msg.SetSubject("Test Sender")

	err = sender.Send(true, msg) // 2.
	if err != nil {
		t.Error("[Sender test] send email failed:", err)
	}
	msg2 := goemail.NewMessage()
	msg2.SetFrom(config.From, "JiuYu77")
	msg2.SetTo(es)
	msg2.SetSubject("Test Sender2")
	msg2.SetBody("text/plain", "Test_smtp_sender!")

	sender.Send(false, msg, msg2) // 3. 4.
	if err != nil {
		t.Error("[Sender test] multi-msg send email failed:", err)
	}
	sender.Quit()
}

// 模拟文件复制函数，返回文件名和文件设置
func MockCopyFile(name string) (string, goemail.FileSetting) {
	return name, goemail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write([]byte("Content of " + filepath.Base(name))) // 模拟写入文件内容, 写入了文件名
		return err
	})
}
func MockCopyFileWithHeader(name string, h map[string][]string) (string, goemail.FileSetting, goemail.FileSetting) {
	name, f := MockCopyFile(name)
	return name, f, goemail.SetHeader(h)
}
