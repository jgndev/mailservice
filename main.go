package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mailersend/mailersend-go"
	"html/template"
	"jgnovak.com/mailservice/models"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func HandleRequest(_ context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ms := mailersend.NewMailersend(os.Getenv("MAILERSEND_API_KEY"))
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var mailRequest models.MailRequest
	err := json.Unmarshal([]byte(request.Body), &mailRequest)
	if err != nil {
		log.Printf("Could not unmarshal request body to mail request: %v", err.Error())
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Failed to parse request body: %v", err),
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			IsBase64Encoded:   false,
			MultiValueHeaders: nil,
		}, nil
	}

	tmpl, err := template.ParseFiles("mailTemplate.html")
	if err != nil {
		log.Print(err.Error())
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Failed to create html template: %v", err.Error()),
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			IsBase64Encoded:   false,
			MultiValueHeaders: nil,
		}, nil
	}

	body := new(strings.Builder)
	err = tmpl.Execute(body, mailRequest)
	if err != nil {
		log.Print(err.Error())
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Failed to parse html template: %v", err.Error()),
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			IsBase64Encoded:   false,
			MultiValueHeaders: nil,
		}, nil
	}

	log.Printf("To: %s\nFrom: %s\nSubject: %s\nBody: %s\n", mailRequest.To, mailRequest.From, mailRequest.Subject, mailRequest.Body)
	subject := mailRequest.Subject
	text := mailRequest.Body
	html := body.String()
	from := mailersend.From{
		Name:  "Jeremy Novak",
		Email: mailRequest.From,
	}
	recipient := []mailersend.Recipient{
		{
			Name:  "",
			Email: mailRequest.To,
		},
	}

	message := ms.Email.NewMessage()

	message.SetFrom(from)
	message.SetRecipients(recipient)
	message.SetSubject(subject)
	message.SetHTML(html)
	message.SetText(text)
	message.SetInReplyTo("client-id")

	res, err := ms.Email.Send(ctx, message)
	log.Print(res.Header.Get("X-Message-Id"))

	if err != nil {
		apiResponse := events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf("%v", err.Error()),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
		log.Printf("Error doing the thing: %v\n", err.Error())
		return apiResponse, err
	}

	jsonSuccessResponse, err := json.Marshal("ok") // assuming "ok" is the response you want to send
	if err != nil {
		log.Printf("Error marshalling success response: %v", err.Error())
	}
	return events.APIGatewayProxyResponse{
		Body:       string(jsonSuccessResponse),
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
