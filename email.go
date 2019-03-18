package utils

import (
	"fmt"
	"github.com/scorredoira/email"
	"net/mail"
	"net/smtp"
)

type emailInfo struct {
	User         string   // 发送者邮箱
	Password     string   // 发送者密码
	SmtpHost     string   // smt服务器地址
	SmtpPort     int      // smt服务器地址
	toUser       []string // 接受者邮箱
	files        []string // 文件名称
	body         string
	title        string
	userNickName string
}

func (e *emailInfo) ToUser(Users ...string) *emailInfo {
	for _, user := range Users {
		e.toUser = append(e.toUser, user)
	}
	return e
}

func (e *emailInfo) Files(filenames ...string) *emailInfo {
	for _, filename := range filenames {
		e.files = append(e.files, filename)
	}
	return e
}

func (e *emailInfo) Body(body string) *emailInfo {
	e.body = body
	return e
}

func (e *emailInfo) Title(title string) *emailInfo {
	e.title = title
	return e
}

func (e *emailInfo) NickName(name string) *emailInfo {
	e.userNickName = name
	return e
}

func (e *emailInfo) SendAll() error {
	m := email.NewHTMLMessage(e.title, e.body)
	m.From = mail.Address{Name: If(e.userNickName == "", e.User, e.userNickName).(string), Address: e.User}
	m.To = e.toUser
	if len(e.files) != 0 {
		for _, file := range e.files {
			m.Attach(file)
		}
	}
	auth := smtp.PlainAuth("", e.User, e.Password, e.SmtpHost)
	return email.Send(fmt.Sprintf("%s:%d", e.SmtpHost, e.SmtpPort), auth, m)
}

func NewEmail(user, password, smtpHost string, smtpPort int) *emailInfo {
	return &emailInfo{User: user, Password: password, SmtpHost: smtpHost, SmtpPort: smtpPort}
}
