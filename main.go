package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type MyEvent struct {
	Name string `json:"name"`
}

type Response struct {
	Message string `json:"message"`
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("Processing request %s %s\n", request.HTTPMethod, request.Path)

	switch request.HTTPMethod {
	case "GET":
		return handleGet(request)
	case "POST":
		return handlePost(request)
	default:
		return events.APIGatewayProxyResponse{Body: "Method Not Allowed", StatusCode: http.StatusMethodNotAllowed}, nil
	}
}

func handleGet(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	msg := fmt.Sprintf("Hello from GET! Path: %s", request.Path)
	return events.APIGatewayProxyResponse{Body: msg, StatusCode: http.StatusOK}, nil
}

func handlePost(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var event MyEvent
	err := json.Unmarshal([]byte(request.Body), &event)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Invalid Request Body", StatusCode: http.StatusBadRequest}, nil
	}

	res := Response{
		Message: fmt.Sprintf("Hello %s from POST!", event.Name),
	}

	body, err := json.Marshal(res)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Internal Server Error", StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
