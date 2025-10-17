package smtp

import (
	"errors"
	"fmt"
	"io"
	"net/smtp"
	"sync"
	"time"
)

// SMTPSender 一个邮件发送客户端。
type SMTPSender struct {
	client *smtp.Client
	smtp   *SMTP
}

func NewSMTPSender(smtp *SMTP) *SMTPSender {
	return &SMTPSender{smtp: smtp}
}
func NewSMTPSender1(client *smtp.Client, smtp *SMTP) *SMTPSender {
	return &SMTPSender{client: client, smtp: smtp}
}

// Quit 关闭连接
func (s *SMTPSender) Quit() error {
	if err := s.client.Quit(); err != nil {
		return err
	}
	s.client = nil
	return nil
}

// Noop 发送 NOOP 命令(No Operation, 无操作命令)，用于测试连接是否正常。
//
// Return
//   - {error} 错误信息。nil 表示 NOOP 命令成功，连接正常。
func (s SMTPSender) Noop() error {
	return s.client.Noop()
}

func (s *SMTPSender) Client() *smtp.Client {
	return s.client
}

func (s *SMTPSender) IsConnected() bool {
	if s.client != nil && s.Noop() == nil {
		return true // 已连接, 连接正常、未断开
	}
	return false
}

// Dial 连接SMTP server。未连接 或 连接断开，才会重新连接SMTP服务器。
func (s *SMTPSender) Dial() error {
	if s.IsConnected() {
		return nil
	}

	sender, err := s.smtp.Dial()
	if err != nil {
		return err
	}
	s.client = sender.client
	return nil
}

// Send 发送邮件
//
// Args
//   - whereFrom: 是否从消息中获取发件人, true 使用 smtp字段 中的发件人（smtp.from）,
//     false 从消息中获取发件人。
//   - msgs 邮件内容，*Message 列表。
func (s *SMTPSender) Send(whereFrom bool, msgs ...*Message) error {
	for i, m := range msgs {
		if err := s.send(whereFrom, m); err != nil {
			return fmt.Errorf("goemail: could not send email %d: %v", i+1, err)
		}
	}
	return nil
}

func (s *SMTPSender) send(whereFrom bool, m *Message) error {
	var from string

	switch whereFrom {
	case true:
		from = s.smtp.from // 使用配置中的发件人
	case false:
		email, err := m.getFrom() // 从消息中获取发件人，“服务商”可能会拒绝此方式
		from = email
		if err != nil {
			return fmt.Errorf("获取发件人失败: %v", err)
		}
	}

	to, err := m.getRecipients() // 获取收件人
	if err != nil {
		return fmt.Errorf("获取收件人失败: %v", err)
	}
	return s.SendEmail(from, to, m)
}

// SendEmail 发送邮件，可以将 msg 发送给多个收件人（群发）。
//
// Args
//   - from {string} 发件人邮箱
//   - to {[]string} 收件人邮箱切片
//   - msg {io.WriterTo} 邮件内容，需实现 io.WriterTo 接口
func (s *SMTPSender) SendEmail(from string, to []string, msg io.WriterTo) error {
	if s.client == nil {
		return errors.New("SMTP 客户端未连接")
	} else if s.Noop() != nil {
		return errors.New("NOOP 命令失败, 连接可能已断开")
	}

	// 设置发件人和收件人
	if err := s.client.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err := s.client.Rcpt(addr); err != nil {
			return err
		}
	}

	// 发送邮件内容
	w, err := s.client.Data()
	if err != nil {
		return err
	}

	if _, err := msg.WriteTo(w); err != nil {
		return err
	}
	return w.Close()
}

// Send1 发送邮件，使用 smtp.from 作为发件人
//
// Args
//   - to: 收件人列表/切片
//   - msg {[]byte}: 邮件内容
func (s *SMTPSender) Send1(to []string, msg []byte) error {
	return s.sendEmailBytes(s.smtp.from, to, msg)
}

// Send2 发送邮件，指定发件人邮箱
//
// Args
//   - from: 发件人
//   - to: 收件人列表
//   - msg {[]byte}: 邮件内容
func (s *SMTPSender) Send2(from string, to []string, msg []byte) error {
	return s.sendEmailBytes(from, to, msg)
}

// 发送邮件
// from: 发件人
// to: 收件人列表
// msg {[]byte}: 邮件内容
func (s *SMTPSender) sendEmailBytes(from string, to []string, msg []byte) error {
	if s.client == nil {
		return errors.New("SMTP 客户端未连接")
	} else if s.Noop() != nil {
		return errors.New("NOOP 命令失败, 连接可能已断开")
	}

	// 设置发件人和收件人
	if err := s.client.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err := s.client.Rcpt(addr); err != nil {
			return err
		}
	}

	// 发送邮件内容
	w, err := s.client.Data()
	if err != nil {
		return err
	}

	if _, err := w.Write(msg); err != nil {
		return err
	}

	return w.Close()
}

/* ####################################################################### */

// SMTPSenderPool 是一个 SMTP 发送池。
// 它维护一个连接池, 每个连接都是一个 Sender 实例。
// 每个连接都有一个 SMTP 客户端, 用于发送邮件。
type SMTPSenderPool struct {
	senders chan *SMTPSender
	config  *SMTPConfig
	maxSize int
	created int // 当前已创建的连接数
	mtx     sync.Mutex
}

func NewSMTPSenderPool(poolSize int, config *SMTPConfig) *SMTPSenderPool {
	pool := &SMTPSenderPool{
		senders: make(chan *SMTPSender, poolSize),
		config:  config,
		maxSize: poolSize,
	}

	// 预初始化连接
	smtp := NewSMTP(
		config.Host, config.Port, config.Username, config.Password, config.From,
	)
	for range poolSize {
		sender := NewSMTPSender(smtp)
		if sender.Dial() == nil {
			pool.senders <- sender
			pool.created++
		}
	}

	return pool
}

func (p *SMTPSenderPool) Get() (*SMTPSender, error) {
	select {
	case sender := <-p.senders:
		if sender.IsConnected() {
			return sender, nil
		} else {
			sender.Quit()
			return p.replaceSender()
		}
	default:
		// 池为空，检查是否还可以创建新连接
		p.mtx.Lock()
		defer p.mtx.Unlock()
		if p.created < p.maxSize {
			return p.createNewSender()
		}

		// 等待可用的连接
		select {
		case sender := <-p.senders:
			return sender, nil
		case <-time.After(30 * time.Second): // 超时
			return nil, errors.New("no available sender in pool")
		}
	}
}

func (p *SMTPSenderPool) Put(sender *SMTPSender) {
	// 检查连接是否仍然有效
	if !sender.IsConnected() {
		sender.Quit()
		p.mtx.Lock()
		p.created--
		p.mtx.Unlock()
		return
	}

	select {
	case p.senders <- sender:
		// 成功放回池中
	default:
		// 池已满，关闭连接
		sender.Quit()
		p.mtx.Lock()
		p.created--
		p.mtx.Unlock()
	}
}

func (p *SMTPSenderPool) replaceSender() (*SMTPSender, error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	sender, err := p.create()
	if err != nil {
		p.created--
		return nil, err
	}

	return sender, nil
}

func (p *SMTPSenderPool) createNewSender() (*SMTPSender, error) {
	sender, err := p.create()
	if err != nil {
		return nil, err
	}

	p.created++
	return sender, nil
}

func (p *SMTPSenderPool) create() (*SMTPSender, error) {
	smtp := NewSMTP(
		p.config.Host, p.config.Port, p.config.Username, p.config.Password, p.config.From,
	)
	sender := NewSMTPSender(smtp)
	if err := sender.Dial(); err != nil {
		return nil, err
	}
	return sender, nil
}

// Close 关闭连接池中的所有连接
func (p *SMTPSenderPool) Close() {
	close(p.senders)
	for sender := range p.senders {
		sender.Quit()
	}
}
