package smtp

import (
	"context"
	"io"
	"log"
	"mime"
	"mime/quotedprintable"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

type SMTPConfig struct {
	Host     string `json:"smtp_host"`     // SMTP 服务器主机名
	Port     int    `json:"smtp_port"`     // SMTP 服务器端口号
	Username string `json:"smtp_username"` // SMTP 服务器用户名
	Password string `json:"smtp_password"` // SMTP 服务器密码
	From     string `json:"from"`          // 发件人邮箱地址
}

type header = textproto.MIMEHeader
type copier func(io.Writer) error
type Encoding string

// A MessageSetting can be used as an argument in NewMessage to configure an
// email.
type MessageSetting func(m *Message)

const (
	// QuotedPrintable represents the quoted-printable encoding as defined in
	// RFC 2045.
	QuotedPrintable Encoding = "quoted-printable"
	// Base64 represents the base64 encoding as defined in RFC 2045.
	Base64 Encoding = "base64"
	// Unencoded can be used to avoid encoding the body of an email. The headers
	// will still be encoded using quoted-printable encoding.
	Unencoded Encoding = "8bit"
)

type mimeEncoder = mime.WordEncoder

const (
	bEncoding = mime.BEncoding
	qEncoding = mime.QEncoding
)

type part struct {
	contentType string
	copier      copier
	encoding    Encoding
}

// A PartSetting can be used as an argument in Message.SetBody,
// Message.AddAlternative or Message.AddAlternativeWriter to configure the part
// added to a message.
type PartSetting func(*part)

func Part(contentType string, f copier, encoding Encoding, settings []PartSetting) *part {
	part := &part{
		contentType: contentType,
		copier:      f,
		encoding:    encoding,
	}
	for _, s := range settings {
		s(part)
	}
	return part
}
func Copier(s string) func(io.Writer) error {
	return func(w io.Writer) error {
		_, err := io.WriteString(w, s) // 闭包捕获了 s
		return err
	}
}

// SetPartEncoding sets the encoding of the part added to the message. By
// default, parts use the same encoding than the message.
func SetPartEncoding(e Encoding) PartSetting {
	return PartSetting(func(p *part) {
		p.encoding = e
	})
}

type file struct {
	Name     string
	Header   header
	CopyFunc copier
}

func (f *file) setHeader(field, value string) {
	f.Header[field] = []string{value}
}

// A FileSetting can be used as an argument in Message.Attach or Message.Embed.
type FileSetting func(*file)

// SetHeader is a file setting to set the MIME header of the message part that
// contains the file content.
//
// Mandatory headers are automatically added if they are not set when sending
// the email.
func SetHeader(h map[string][]string) FileSetting {
	return func(f *file) {
		for k, v := range h {
			f.Header[k] = v
		}
	}
}

// Rename is a file setting to set the name of the attachment if the name is
// different than the filename on disk.
func Rename(name string) FileSetting {
	return func(f *file) {
		f.Name = name
	}
}

// SetCopyFunc is a file setting to replace the function that runs when the
// message is sent. It should copy the content of the file to the io.Writer.
//
// The default copy function opens the file with the given filename, and copy
// its content to the io.Writer.
func SetCopyFunc(f func(io.Writer) error) FileSetting {
	return func(fi *file) {
		fi.CopyFunc = f
	}
}

var (
	lastIndexByte = strings.LastIndexByte
	newQPWriter   = quotedprintable.NewWriter
)

// As required by RFC 2045, 6.7. (page 21) for quoted-printable, and
// RFC 2045, 6.8. (page 25) for base64.
const maxLineLen = 76

// base64LineWriter limits text encoded in base64 to 76 characters per line
type base64LineWriter struct {
	w       io.Writer
	lineLen int
}

func newBase64LineWriter(w io.Writer) *base64LineWriter {
	return &base64LineWriter{w: w}
}

func (w *base64LineWriter) Write(p []byte) (int, error) {
	n := 0
	for len(p)+w.lineLen > maxLineLen {
		w.w.Write(p[:maxLineLen-w.lineLen])
		w.w.Write([]byte("\r\n"))
		p = p[maxLineLen-w.lineLen:]
		n += maxLineLen - w.lineLen
		w.lineLen = 0
	}

	w.w.Write(p)
	w.lineLen += len(p)

	return n + len(p), nil
}

// ConnectionMonitor 用于监控 *smtp.Client 连接状态，保持连接活跃.
type ConnectionMonitor struct {
	Sender       *SMTPSender
	isMonitoring bool
	mtx          sync.Mutex
}

func (cm *ConnectionMonitor) IsMonitoring() bool {
	cm.mtx.Lock()
	defer cm.mtx.Unlock()

	return cm.isMonitoring
}

// 定期检查连接状态、重连.
// 可用于长时间处理邮件时保持连接.
func (cm *ConnectionMonitor) MonitorConnection(ctx context.Context, d time.Duration) {
	cm.mtx.Lock()
	defer cm.mtx.Unlock()

	if cm.isMonitoring {
		log.Println("MonitorConnection is already running")
		return
	}

	go func() {
		cm.isMonitoring = true
		ticker := time.NewTicker(d)
		defer ticker.Stop()

		retryCount := 0
		const maxRetries int = 5

		for {
			select {
			case <-ctx.Done():
				cm.isMonitoring = false
				return
			case <-ticker.C:
				if err := cm.Sender.Noop(); err != nil { // 触发重连逻辑
					log.Printf("SMTP connection check failed: %v", err)

					// 指数退避重试
					backoff := min(time.Duration(retryCount)*time.Second, 30*time.Second)
					time.Sleep(backoff)

					// 尝试重连
					if err := cm.Sender.Dial(); err != nil {
						log.Printf("SMTP reconnection failed: %v", err)
						retryCount++
						if retryCount >= maxRetries {
							log.Println("Max reconnection retries exceeded, stopping monitor")
							cm.isMonitoring = false
							return
						}
					} else {
						log.Println("SMTP reconnection successful")
						retryCount = 0
					}
				} else {
					retryCount = 0
				}
			}
		}
	}()
}
