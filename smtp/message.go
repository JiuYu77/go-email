package smtp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"os"
	"path/filepath"
	"time"
)

type Message struct {
	header      header  // 邮件头信息
	parts       []*part // 正文部分
	attachments []*file // 附件部分
	embedded    []*file // 内联资源部分
	charset     string  // 字符集

	encoding      Encoding
	headerEncoder mimeEncoder
	buf           bytes.Buffer
}

func NewMessage(settings ...MessageSetting) *Message {
	m := &Message{
		header:   make(header),
		charset:  "UTF-8",
		encoding: QuotedPrintable,
	}

	m.applySettings(settings)

	if m.encoding == Base64 {
		m.headerEncoder = bEncoding
	} else {
		m.headerEncoder = qEncoding
	}

	return m
}

// Reset resets the message so it can be reused. The message keeps its previous
// settings so it is in the same state that after a call to NewMessage.
func (m *Message) Reset() {
	for k := range m.header {
		delete(m.header, k)
	}
	m.parts = nil
	m.attachments = nil
	m.embedded = nil
}
func (m *Message) applySettings(settings []MessageSetting) {
	for _, s := range settings {
		s(m)
	}
}

// SetCharset is a message setting to set the charset of the email.
func SetCharset(charset string) MessageSetting {
	return func(m *Message) {
		m.charset = charset
	}
}

// SetEncoding is a message setting to set the encoding of the email.
func SetEncoding(enc Encoding) MessageSetting {
	return func(m *Message) {
		m.encoding = enc
	}
}

// SetFrom 设置发件人
func (m *Message) SetFrom(from string, name string) {
	m.setHeader("From", from, name)
}
func parseAddress(address string) (string, error) {
	addr, err := mail.ParseAddress(address)
	if err != nil {
		return "", fmt.Errorf("[goemail] invalid address %q: %v", address, err)
	}
	return addr.Address, nil
}

// 获取发件人
func (m *Message) getFrom() (string, error) {
	from := m.header["From"]
	if len(from) == 0 {
		return "", errors.New(`[geomail] invalid message, "From" field is absent`)
	}
	return parseAddress(from[0])
}

// SetTo 设置收件人
func (m *Message) SetTo(to []string) {
	m.setHeader("To", to...)
}

func addAddress(list []string, addr string) []string {
	for _, a := range list {
		if addr == a {
			return list
		}
	}

	return append(list, addr)
}

// 获取收件人
func (m *Message) getRecipients() ([]string, error) {
	n := 0
	for _, field := range []string{"To", "Cc", "Bcc"} {
		if addresses, ok := m.header[field]; ok {
			n += len(addresses)
		}
	}
	list := make([]string, 0, n)

	for _, field := range []string{"To", "Cc", "Bcc"} {
		if addresses, ok := m.header[field]; ok {
			for _, a := range addresses {
				addr, err := parseAddress(a)
				if err != nil {
					return nil, err
				}
				list = addAddress(list, addr)
			}
		}
	}

	return list, nil
}

// SetSubject 设置邮件主题
func (m *Message) SetSubject(subject string) {
	m.setHeader("Subject", subject)
}

// SetDateHeader sets a date to the given header field.
func (m *Message) SetDateHeader(field string, date time.Time) {
	m.header[field] = []string{m.FormatDate(date)}
}

func (m *Message) SetHeader(key string, value ...string) {
	m.setHeader(key, value...)
}

// GetHeader gets a header field.
func (m *Message) GetHeader(field string) []string {
	return m.header[field]
}

// hasSpecials检查字符串是否包含任何特殊字符。
func hasSpecials(text string) bool {
	for i := 0; i < len(text); i++ {
		switch c := text[i]; c {
		case '(', ')', '<', '>', '[', ']', ':', ';', '@', '\\', ',', '.', '"':
			return true
		}
	}
	return false
}

// FormatAddress formats an address and a name as a valid RFC 5322 address.
func (m *Message) FormatAddress(address, name string) string {
	if name == "" {
		return address
	}

	enc := m.encodeString(name)
	if enc == name {
		m.buf.WriteByte('"')
		for i := 0; i < len(name); i++ {
			b := name[i]
			if b == '\\' || b == '"' {
				m.buf.WriteByte('\\')
			}
			m.buf.WriteByte(b)
		}
		m.buf.WriteByte('"')
	} else if hasSpecials(name) {
		m.buf.WriteString(bEncoding.Encode(m.charset, name))
	} else {
		m.buf.WriteString(enc)
	}
	m.buf.WriteString(" <")
	m.buf.WriteString(address)
	m.buf.WriteByte('>')

	addr := m.buf.String()
	m.buf.Reset()
	return addr
}

func (m *Message) setHeaderAdderss(key, addr string, name string) {
	m.header[key] = []string{m.FormatAddress(addr, name)}
}

// setHeader 设置邮件头
func (m *Message) setHeader(key string, value ...string) {
	switch key {
	case "From":
		m.setHeaderAdderss(key, value[0], value[1])
	default:
		m.encodeHeader(value)
		m.header[key] = value
	}
}

// SetBody 设置邮件正文
func (m *Message) SetBody(contentType, body string, settings ...PartSetting) {
	m.parts = []*part{Part(contentType, Copier(body), m.encoding, settings)}
}

func (m *Message) encodeString(value string) string {
	return m.headerEncoder.Encode(m.charset, value)
}

func (m *Message) encodeHeader(values []string) {
	for i := range values {
		values[i] = m.encodeString(values[i])
	}
}

func (m *Message) FormatDate(t time.Time) string {
	return t.Format(time.RFC1123Z)
}

func (m *Message) hasMixedPart() bool {
	return (len(m.parts) > 0 && len(m.attachments) > 0) || len(m.attachments) > 1
}
func (m *Message) hasRelatedPart() bool {
	return (len(m.parts) > 0 && len(m.embedded) > 0) || len(m.embedded) > 1
}
func (m *Message) hasAlternativePart() bool {
	return len(m.parts) > 1
}

func (m *Message) appendFile(list []*file, name string, settings []FileSetting) ([]*file, error) {
	_, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file is not exist: %s", name)
		}
		return nil, fmt.Errorf("unable to access the file %s: %w", name, err)
	}

	f := &file{
		Name:   filepath.Base(name),
		Header: make(map[string][]string),
		CopyFunc: func(w io.Writer) error {
			h, err := os.Open(name)
			if err != nil {
				return err
			}
			if _, err := io.Copy(w, h); err != nil {
				h.Close()
				return err
			}
			return h.Close()
		},
	}

	for _, s := range settings {
		s(f)
	}

	if list == nil {
		return []*file{f}, nil
	}

	return append(list, f), nil
}

// Attach attaches the files to the email.
func (m *Message) Attach(filename string, settings ...FileSetting) error {
	attachments, err := m.appendFile(m.attachments, filename, settings)
	if err != nil {
		return err
	}
	m.attachments = attachments
	return nil
}

// Embed embeds the images to the email.
func (m *Message) Embed(filename string, settings ...FileSetting) error {
	embedded, err := m.appendFile(m.embedded, filename, settings)
	if err != nil {
		return err
	}
	m.embedded = embedded
	return nil
}

// AddAlternative adds an alternative part to the message.
//
// It is commonly used to send HTML emails that default to the plain text
// version for backward compatibility. AddAlternative appends the new part to
// the end of the message. So the plain text part should be added before the
// HTML part. See http://en.wikipedia.org/wiki/MIME#Alternative
func (m *Message) AddAlternative(contentType, body string, settings ...PartSetting) {
	m.AddAlternativeWriter(contentType, Copier(body), settings...)
}

// AddAlternativeWriter adds an alternative part to the message. It can be
// useful with the text/template or html/template packages.
func (m *Message) AddAlternativeWriter(contentType string, f func(io.Writer) error, settings ...PartSetting) {
	m.parts = append(m.parts, Part(contentType, f, m.encoding, settings))
}

// WriteTo 写入邮件到 io.Writer
// @return n int64 写入的字节数
// @return err error 写入错误, nil 表示写入成功
func (m *Message) WriteTo(w io.Writer) (n int64, err error) {
	msgWriter := MessageWriter{
		w: w,
	}
	msgWriter.writeMessage(m)

	return msgWriter.n, msgWriter.err
}
