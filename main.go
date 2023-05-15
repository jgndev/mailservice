package main

import (
	"bytes"
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"html/template"
	"jgnovak.com/mailservice/models"
	"log"
	"os"
)

var kApiKey string = ""

func init() {
	kApiKey = os.Getenv("SENDGRID_API_KEY")
}

func HandleRequest(ctx context.Context, request models.MailRequest) (string, error) {
	from := mail.NewEmail("Example User", "test@example.com")
	subject := request.Subject
	to := mail.NewEmail("Example User", request.To)

	tmpl := template.Must(template.ParseFiles("mailTemplate.html"))
	data := models.MailData{
		From:    request.From,
		Subject: request.Subject,
		Body:    request.Body,
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		return "Failed to execute template", err
	}

	htmlContent := buf.String()
	message := mail.NewSingleEmail(from, subject, to, request.Body, htmlContent)

	client := sendgrid.NewSendClient(kApiKey)
	_, err := client.Send(message)

	if err != nil {
		log.Println(err)
		return "Failed to send email", err
	} else {
		log.Println("Email sent successfully")
		return "Email sent successfully", nil
	}
}

func main() {
	lambda.Start(HandleRequest)
}
