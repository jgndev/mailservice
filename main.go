package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mailgun/mailgun-go/v4"
	"html/template"
	"jgnovak.com/mailservice/models"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	kApiKey := os.Getenv("MAILGUN_API_KEY")
	log.Printf("API Key: %s", kApiKey)

	mg := mailgun.NewMailgun("jgnovak.com", kApiKey)

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
	message := mg.NewMessage(mailRequest.From, mailRequest.Subject, body.String(), mailRequest.To)
	message.SetHtml(body.String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	resp, id, err := mg.Send(ctx, message)
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

	log.Printf("ID: %s, Resp: %s\n", id, resp)

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
