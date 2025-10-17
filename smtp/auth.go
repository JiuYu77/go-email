package smtp

import (
	"bytes"
	"errors"
	"fmt"
	"net/smtp"
)

type SMTPAuthType = string

const (
	SMTPAuthCramMD5 SMTPAuthType = "CRAM-MD5"
	SMTPAuthXOAuth2 SMTPAuthType = "XOAUTH2"
	SMTPAuthLogin   SMTPAuthType = "LOGIN"
	SMTPAuthPlain   SMTPAuthType = "PLAIN"
)

// loginAuth is an smtp.Auth that implements the LOGIN authentication mechanism.
type loginAuth struct {
	username string
	password string
	host     string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS {
		advertised := false
		for _, mechanism := range server.Auth {
			if mechanism == "LOGIN" {
				advertised = true
				break
			}
		}
		if !advertised {
			return "", nil, errors.New("gomail: unencrypted connection")
		}
	}
	if server.Name != a.host {
		return "", nil, errors.New("gomail: wrong host name")
	}
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}

	switch {
	case bytes.Equal(fromServer, []byte("Username:")):
		return []byte(a.username), nil
	case bytes.Equal(fromServer, []byte("Password:")):
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("gomail: unexpected server challenge: %s", fromServer)
	}
}

// xoauth2Auth is an smtp.Auth that implements the XOAUTH2 authentication mechanism.
// xoauth2Auth 是 OAuth2.0 认证机制的实现.
type xoauth2Auth struct {
	username string
	token    string
}

func (a *xoauth2Auth) Start(server *smtp.ServerInfo) (proto string, toServer []byte, err error) {
	proto = "XOAUTH2"
	toServer = []byte("user=" + a.username + "\x01auth=Bearer " + a.token + "\x01\x01")
	err = nil
	return
}
func (a *xoauth2Auth) Next(fromServer []byte, more bool) (toServer []byte, err error) {
	if more {
		// 处理服务器挑战（通常OAuth2不需要）
		return []byte(""), fmt.Errorf("unexpected server challenge")
	}
	return nil, nil
}
