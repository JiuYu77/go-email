package smtp

import (
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

type MessageWriter struct {
	w io.Writer
	n int64

	writers    [3]*multipart.Writer
	partWriter io.Writer
	depth      uint8
	err        error
}

// createPart 创建 multipart.Writer 部分
func (w *MessageWriter) createPart(h Header) {
	w.partWriter, w.err = w.writers[w.depth-1].CreatePart(h)
}

func (w *MessageWriter) openMultipart(mimeType string) {
	mw := multipart.NewWriter(w)
	contentType := "multipart/" + mimeType + ";\r\n boundary=" + mw.Boundary()
	w.writers[w.depth] = mw

	if w.depth == 0 {
		w.writeHeader("Content-Type", contentType)
		w.writeString("\r\n")
	} else {
		w.createPart(map[string][]string{
			"Content-Type": {contentType},
		})
	}
	w.depth++
}

func (w *MessageWriter) closeMultipart() {
	if w.depth > 0 {
		w.writers[w.depth-1].Close()
		w.depth--
	}
}

// io.Writer 接口实现
// 写入数据到 io.Writer
func (w *MessageWriter) Write(p []byte) (n int, err error) {
	if w.err != nil {
		return 0, errors.New("goemail: cannot write as writer is in error")
	}

	n, w.err = w.w.Write(p)
	w.n += int64(n)
	return n, w.err
}

// writeString 写入字符串到 io.Writer
func (w *MessageWriter) writeString(s string) {
	n, _ := io.WriteString(w.w, s)
	w.n += int64(n)
}

func (w *MessageWriter) writeLine(s string, charsLeft int) string {
	// If there is already a newline before the limit. Write the line.
	if i := strings.IndexByte(s, '\n'); i != -1 && i < charsLeft {
		w.writeString(s[:i+1])
		return s[i+1:]
	}

	for i := charsLeft - 1; i >= 0; i-- {
		if s[i] == ' ' {
			w.writeString(s[:i])
			w.writeString("\r\n ")
			return s[i+1:]
		}
	}

	// We could not insert a newline cleanly so look for a space or a newline
	// even if it is after the limit.
	for i := 75; i < len(s); i++ {
		if s[i] == ' ' {
			w.writeString(s[:i])
			w.writeString("\r\n ")
			return s[i+1:]
		}
		if s[i] == '\n' {
			w.writeString(s[:i+1])
			return s[i+1:]
		}
	}

	// Too bad, no space or newline in the whole string. Just write everything.
	w.writeString(s)
	return ""
}

func (w *MessageWriter) writeHeader(key string, value ...string) {
	w.writeString(key)
	if len(value) == 0 {
		w.writeString(":\r\n")
		return
	}
	w.writeString(": ")

	// Max header line length is 78 characters in RFC 5322 and 76 characters
	// in RFC 2047. So for the sake of simplicity we use the 76 characters
	// limit.
	charsLeft := 76 - len(key) - len(": ")

	for i, s := range value {
		// If the line is already too long, insert a newline right away.
		if charsLeft < 1 {
			if i == 0 {
				w.writeString("\r\n ")
			} else {
				w.writeString(",\r\n ")
			}
			charsLeft = 75
		} else if i != 0 {
			w.writeString(", ")
			charsLeft -= 2
		}

		// While the header content is too long, fold it by inserting a newline.
		for len(s) > charsLeft {
			s = w.writeLine(s, charsLeft)
			charsLeft = 75
		}
		w.writeString(s)
		if i := lastIndexByte(s, '\n'); i != -1 {
			charsLeft = 75 - (len(s) - i - 1)
		} else {
			charsLeft -= len(s)
		}
	}
	w.writeString("\r\n")
}

func (w *MessageWriter) writeHeaders(h Header) {
	if w.depth == 0 {
		for k, v := range h {
			if k != "Bcc" {
				w.writeHeader(k, v...)
			}
		}
	} else {
		w.createPart(h)
	}
}

func (w *MessageWriter) writeBody(f func(io.Writer) error, enc Encoding) {
	var subWriter io.Writer
	if w.depth == 0 {
		w.writeString("\r\n")
		subWriter = w.w
	} else {
		subWriter = w.partWriter
	}

	switch enc {
	case Base64:
		wc := base64.NewEncoder(base64.StdEncoding, newBase64LineWriter(subWriter))
		w.err = f(wc)
		wc.Close()
	case Unencoded:
		w.err = f(subWriter)
	default:
		wc := newQPWriter(subWriter)
		w.err = f(wc)
		wc.Close()
	}
}

func (w *MessageWriter) writePart(p *part, charset string) {
	w.writeHeaders(map[string][]string{
		"Content-Type":              {p.contentType + "; charset=" + charset},
		"Content-Transfer-Encoding": {string(p.encoding)},
	})
	w.writeBody(p.copier, p.encoding)
}

func (w *MessageWriter) addFiles(files []*file, isAttachment bool) {
	for _, f := range files {
		if _, ok := f.Header["Content-Type"]; !ok {
			mediaType := mime.TypeByExtension(filepath.Ext(f.Name))
			if mediaType == "" {
				mediaType = "application/octet-stream"
			}
			f.setHeader("Content-Type", mediaType+`; name="`+f.Name+`"`)
		}

		if _, ok := f.Header["Content-Transfer-Encoding"]; !ok {
			f.setHeader("Content-Transfer-Encoding", string(Base64))
		}

		if _, ok := f.Header["Content-Disposition"]; !ok {
			var disp string
			if isAttachment {
				disp = "attachment"
			} else {
				disp = "inline"
			}
			f.setHeader("Content-Disposition", disp+`; filename="`+f.Name+`"`)
		}

		if !isAttachment {
			if _, ok := f.Header["Content-ID"]; !ok {
				f.setHeader("Content-ID", "<"+f.Name+">")
			}
		}
		w.writeHeaders(f.Header)
		w.writeBody(f.CopyFunc, Base64)
	}
}

func (w *MessageWriter) writeMessage(m *Message) {
	if _, ok := m.header["Mime-Version"]; !ok {
		w.writeString("Mime-Version: 1.0\r\n")
	}
	// 写入头信息
	if _, ok := m.header["Date"]; !ok { // 若 Date 头信息不存在
		w.writeHeader("Date", m.FormatDate(time.Now()))
	}
	w.writeHeaders(m.header)

	// 写入正文
	if m.hasMixedPart() {
		w.openMultipart("mixed") // 若有 mixed 部分，开启 multipart/mixed 模式
	}

	if m.hasRelatedPart() {
		w.openMultipart("related") // 若有 related 部分，开启 multipart/related 模式
	}
	if m.hasAlternativePart() {
		w.openMultipart("alternative") // 若有 alternative 部分，开启 multipart/alternative 模式
	}

	for _, part := range m.parts {
		w.writePart(part, m.charset)
	}
	if m.hasAlternativePart() { // 若有 alternative 部分，关闭 multipart/alternative 模式
		w.closeMultipart()
	}

	// 写入嵌入式文件
	w.addFiles(m.embedded, false)
	if m.hasRelatedPart() {
		w.closeMultipart()
	}

	// 写入附件文件
	w.addFiles(m.attachments, true)
	if m.hasMixedPart() {
		w.closeMultipart()
	}
}
