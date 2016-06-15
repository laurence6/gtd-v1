package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"mime"
	"net/smtp"
	"strings"

	"github.com/laurence6/gtd.go/core"
)

func init() {
	mailSMTP, ok := conf["mail_smtp"].(string)
	if !ok {
		log.Fatalln("Cannot get 'mail_smtp'")
	}
	hp := strings.Split(mailSMTP, ":")
	mailSMTPHost, mailSMTPPort := hp[0], hp[1]

	mailUsername, ok := conf["mail_username"].(string)
	if !ok {
		log.Fatalln("Cannot get 'mail_username'")
	}

	mailPasswordB64, ok := conf["mail_password"].(string)
	if !ok {
		log.Fatalln("Cannot get 'mail_password'")
	}
	mailPasswordByte, err := base64.StdEncoding.DecodeString(mailPasswordB64)
	if err != nil {
		log.Fatalln(err.Error())
	}
	mailPassword := string(mailPasswordByte)

	mailTo, ok := conf["mail_to"].(string)
	if !ok {
		log.Fatalln("Cannot get 'mail_to'")
	}
	mailToList := strings.Split(mailTo, ", ")

	mailNotifier := &MailNotifier{mailSMTPHost, mailSMTPPort, mailUsername, mailPassword, mailToList}
	notifiers = append(notifiers, mailNotifier)
}

type MailNotifier struct {
	smtpHost string
	smtpPort string
	username string
	password string
	to       []string
}

func (m *MailNotifier) Notify(task *gtd.Task) {
	from := "GTD <" + m.username + ">"
	to := m.to
	subject := "Task: " + task.Subject
	body := fmt.Sprintf("%s\n"+
		"Due: %s\n"+
		"Priority: %d\n"+
		"Note: %s",
		task.Subject,
		task.Due.Date()+" "+task.Due.Time(),
		task.Priority,
		task.Note)
	m.sendMail(from, to, subject, body)
}

func (m *MailNotifier) sendMail(from string, to []string, subject, body string) {
	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s",
		from,
		strings.Join(to, ", "),
		mime.BEncoding.Encode("utf-8", subject),
		body)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         m.smtpHost,
	}
	conn, err := tls.Dial("tcp", m.smtpHost+":"+m.smtpPort, tlsConfig)
	if err != nil {
		log.Println(err.Error())
		return
	}

	client, err := smtp.NewClient(conn, m.smtpHost)
	if err != nil {
		log.Println(err.Error())
		return
	}

	auth := smtp.PlainAuth("", m.username, m.password, m.smtpHost)
	if err = client.Auth(auth); err != nil {
		log.Println(err.Error())
		return
	}

	if err = client.Mail(m.username); err != nil {
		log.Println(err.Error())
		return
	}
	for _, r := range to {
		if err = client.Rcpt(r); err != nil {
			log.Println(err.Error())
			return
		}
	}

	w, err := client.Data()
	if err != nil {
		log.Println(err.Error())
		return
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		log.Println(err.Error())
		return
	}
	err = w.Close()
	if err != nil {
		log.Println(err.Error())
		return
	}

	client.Quit()
}
