package mailing

import (
	"bytes"
	"html/template"
	"net/smtp"
	"os"
)

type Client struct {
	User     string
	Password string
	Host     string
	Auth     smtp.Auth
}

func NewClient() *Client {
	client := &Client{
		User:     os.Getenv("EMAIL_HOST_USER"),
		Password: os.Getenv("EMAIL_HOST_PASSWORD"),
		Host:     os.Getenv("EMAIL_HOST"),
		Auth:     nil,
	}
	client.Auth = smtp.PlainAuth("", client.User, client.Password, client.Host)
	return client
}

type Mail struct {
	Subject string
	Mime    string
	To      []string
	Content string
}

func NewMail(subject, t string, data interface{}) *Mail {
	mail := &Mail{
		Mime:    "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n",
		Subject: subject,
	}
	/*templateData := struct {
		Total     string
		AvgCredit string
		AvgDebit  string
	}{
		Total:     "1000",
		AvgCredit: "2000",
		AvgDebit:  "3000",
	}*/
	mail.ParseTemplate(t, data)
	return mail
}

func (m *Mail) ParseTemplate(file string, data interface{}) error {
	cwd, _ := os.Getwd()
	path := cwd + "/templates/" + file

	t, err := template.ParseFiles(path)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, data); err != nil {
		return err
	}
	m.Content = buffer.String()
	return nil
}

func (m *Mail) SendTo(to string) {
	body := "To: " + to + "\r\nSubject: " + m.Subject + "\r\n" + m.Mime + "\r\n" + m.Content
	client := NewClient()
	err := smtp.SendMail(client.Host+":587", client.Auth, client.User, []string{to}, []byte(body))
	if err != nil {
		panic(err)
	}
}
