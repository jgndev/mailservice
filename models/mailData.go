package models

import "github.com/sendgrid/sendgrid-go/helpers/mail"

type MailData struct {
	To      mail.Email
	From    mail.Email
	Subject string
	Body    string
	Html    string
}
