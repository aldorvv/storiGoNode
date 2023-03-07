// Mailing management module
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

// Create a new SMTP client
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

// Create a new Mail object with default MIME
func NewMail(subject, t string, data interface{}) *Mail {
	mail := &Mail{
		Mime:    "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n",
		Subject: subject,
	}
	mail.ParseTemplate(t, data)
	return mail
}

// Parse html template to text to be sent
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

// Send the email to specified user
func (m *Mail) SendTo(to string) {
	body := "To: " + to + "\r\nSubject: " + m.Subject + "\r\n" + m.Mime + "\r\n" + m.Content
	client := NewClient()
	err := smtp.SendMail(client.Host+":587", client.Auth, client.User, []string{to}, []byte(body))
	if err != nil {
		panic(err)
	}
}
