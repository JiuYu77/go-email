package smtp

import (
	"crypto/tls"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/JiuYu77/go-email/utils"
)

type SMTP struct {
	// Host represents the host of the SMTP server.
	host string
	// Port represents the port of the SMTP server.
	port int // 25 465 587
	// Username is the username to use to authenticate to the SMTP server.
	username string
	// Password is the password to use to authenticate to the SMTP server.
	password string // 不同认证机制使用不同 密码、token
	// auth represents the authentication mechanism used to authenticate to the
	// SMTP server.
	auth smtp.Auth
	// from email is usually the same as username.
	from string
	// ssl defines whether an SSL connection is used. It should be false in
	// most cases since the authentication mechanism should use the STARTTLS
	// extension instead.
	//
	// SSL 设置是否使用 SSL 连接，SSL 或 TLS 协议 (实际只使用TLS协议)
	//
	// 465 端口使用隐式TLS加密，意味着连接一开始就建立在安全通道上。
	// 与587端口(使用STARTTLS)不同，587是先建立普通连接再升级到TLS加密连接，
	// 但有些 SMTP 服务器 587端口 也是隐式TLS加密，
	// 这时候需要设置 SSL 为 true, 否则会连接失败。
	//
	// SSL true 使用 SSL/TLS 连接, false 尝试使用 STARTTLS extension(扩展).
	SSL bool
	// tlsConfig represents the TLS configuration used for the TLS (when the
	// STARTTLS extension is used) or SSL connection.
	TLSConfig *tls.Config
	// LocalName is the hostname sent to the SMTP server with the HELO command.
	// By default, "localhost" is sent.
	LocalName string // 本地主机名
}

// NewSMTP 创建一个新的 SMTP 客户端
//
// # Args
//   - host: SMTP 服务器主机名
//   - port: SMTP 服务器端口号
//   - username: 登录 SMTP 服务器的用户名, 通常是邮箱地址
//   - password: 登录 SMTP 服务器的密码
//   - from: 发件人 email, 通常与 username 相同
//
// # Returns
//   - {*SMTP} SMTP对象指针
func NewSMTP(host string, port int, username string, password string, from string) *SMTP {
	if from == "" {
		from = username // default
	}

	return &SMTP{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		SSL:      port == 465, // SSL 或 TLS 协议 (实际只使用TLS协议)
		// SSL: true 使用 TLS 连接, false 尝试使用 STARTTLS extension(扩展).
	}
}

func (s *SMTP) tlsCfg() *tls.Config {
	if s.TLSConfig == nil {
		return &tls.Config{ServerName: s.host} // default
	}
	return s.TLSConfig
}

// Dial 连接SMTP服务器
//
// # Returns
//
//	{*SMTPSender} 发送邮件的结构体指针
//	{error} nil 表示成功
func (s *SMTP) Dial() (*SMTPSender, error) {
	utils.Logger.Debugln(utils.LogPrefix, "SMTP Dial() start.")
	conn, err := net.DialTimeout("tcp", addr(s.host, s.port), 10*time.Second)
	if err != nil {
		utils.Logger.Errorln(utils.LogPrefix, err)
		return nil, err
	}

	if s.SSL { // 465 端口使用隐式TLS加密, 587 端口也可能使用隐式TLS加密
		conn = tls.Client(conn, s.tlsCfg())
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		utils.Logger.Errorln(utils.LogPrefix, err)
		return nil, err
	}

	if s.LocalName != "" {
		if err := client.Hello(s.LocalName); err != nil {
			utils.Logger.Errorln(utils.LogPrefix, err)
			return nil, err
		}
	}

	if !s.SSL { // 非 SSL/TLS 连接, 端口号为 25、587, 尝试使用 STARTTLS 扩展
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(s.tlsCfg()); err != nil {
				client.Close()
				utils.Logger.Errorln(utils.LogPrefix, err)
				return nil, err
			}
		}
	}

	// 认证机制 (authentication mechanism)
	if s.auth == nil && s.username != "" {
		if ok, auth := client.Extension("AUTH"); ok {
			switch {
			case strings.Contains(auth, "CRAM-MD5"):
				s.auth = smtp.CRAMMD5Auth(s.username, s.password)
				utils.Logger.Debugln(utils.LogPrefix, "CRAM-MD5 auth")
			case strings.Contains(auth, "XOAUTH2") && !strings.Contains(auth, "PLAIN"):
				s.auth = &xoauth2Auth{
					username: s.username,
					token:    s.password,
				}
				utils.Logger.Debugln(utils.LogPrefix, "XOAUTH2 auth")
			case strings.Contains(auth, "LOGIN") && !strings.Contains(auth, "PLAIN"):
				s.auth = &loginAuth{
					username: s.username,
					password: s.password,
					host:     s.host,
				}
				utils.Logger.Debugln(utils.LogPrefix, "LOGIN auth")
			default:
				s.auth = smtp.PlainAuth("", s.username, s.password, s.host)
				utils.Logger.Debugln(utils.LogPrefix, "Plain auth")
			}
		}
	}

	if s.auth != nil { // 认证
		if err = client.Auth(s.auth); err != nil {
			client.Close()
			utils.Logger.Errorln(utils.LogPrefix, err)
			return nil, err
		}
	}
	return &SMTPSender{client: client, smtp: s}, nil
}

// DialAndSend 发送邮件，可以一次发送多封邮件。
//
// # Args
//   - whereFrom 是否从消息中获取发件人, true 使用配置中的发件人, false 从消息中获取发件人
//   - msgs 邮件内容，*Message 列表
func (s *SMTP) DialAndSend(whereFrom bool, msgs ...*Message) error {
	sender, err := s.Dial() // 每次都会重新连接 SMTP服务器
	if err != nil {
		return err
	}
	defer sender.Quit()
	return sender.Send(whereFrom, msgs...)
}

// DialAndSend1 发送邮件，可以群发
//
// # Args
//
//	to {[]string} 收件人列表
//	msg {[]byte} 邮件内容
func (s *SMTP) DialAndSend1(to []string, msg []byte) error {
	sender, err := s.Dial()
	if err != nil {
		return err
	}
	defer sender.Quit()
	return sender.Send1(to, msg)
}

// DialAndSend2 发送邮件，可以群发
//
// # Args
//   - from  发件人email
//   - to  收件人列表
//   - msg  邮件内容
func (s *SMTP) DialAndSend2(from string, to []string, msg []byte) error {
	sender, err := s.Dial()
	if err != nil {
		return err
	}
	defer sender.Quit()
	return sender.Send2(from, to, msg)
}
