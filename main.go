package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	coldStart     = true
	bootStartTime = time.Now()
	globalConfigs = make(map[string]*Config)
)

// Configuration Structs
type Config struct {
	Database  DatabaseConfig  `json:"database"`
	Features  FeaturesConfig  `json:"features"`
	Logging   LoggingConfig   `json:"logging"`
	Services  []ServiceConfig `json:"services"`
	LargeData string          `json:"large_data"`
}

type DatabaseConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	MaxConnections int    `json:"max_connections"`
	Timeout        int    `json:"timeout"`
}

type FeaturesConfig struct {
	EnableFeatureX bool     `json:"enable_feature_x"`
	EnableFeatureY bool     `json:"enable_feature_y"`
	BetaUsers      []string `json:"beta_users"`
}

type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
	Output string `json:"output"`
}

type ServiceConfig struct {
	Name    string `json:"name"`
	Url     string `json:"url"`
	Retries int    `json:"retries"`
}

func init() {
	fmt.Printf("Cold Start initialized at: %v\n", bootStartTime.Format(time.RFC3339Nano))
}

func loadConfig() {
	start := time.Now()
	fmt.Println("Loading configurations...")

	files := []string{"config.json", "config_1.json", "config_2.json", "config_3.json", "config_4.json"}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading config file %s: %v\n", file, err)
			continue
		}

		var config Config
		err = json.Unmarshal(data, &config)
		if err != nil {
			fmt.Printf("Error parsing config file %s: %v\n", file, err)
			continue
		}
		globalConfigs[file] = &config
		fmt.Printf("Loaded %s in %v (Size: %d bytes)\n", file, time.Since(start), len(data))
	}

	fmt.Printf("All configurations loaded in %v\n", time.Since(start))
}

type MyEvent struct {
	Name string `json:"name"`
}

type Response struct {
	Message string `json:"message"`
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	reqStart := time.Now()
	defer func() {
		fmt.Printf("Request completed in %v\n", time.Since(reqStart))
	}()

	if coldStart {
		fmt.Printf("Processing first request after cold start. Total boot duration: %v\n", time.Since(bootStartTime))
		coldStart = false
	}

	fmt.Printf("Processing request %s %s\n", request.HTTPMethod, request.Path)
	fmt.Printf("Request: %+v\n", request)
	// Routing based on Path and Method
	switch request.Path {
	case "/config":
		if request.HTTPMethod == "GET" {
			return handleGetConfig(request)
		}
	case "/echo":
		if request.HTTPMethod == "POST" {
			return handlePostEcho(request)
		}
	case "/hello":
		if request.HTTPMethod == "GET" {
			return handleGetHello(request)
		}
	}
	fmt.Printf("Not Found: %+v\n", request)
	return events.APIGatewayProxyResponse{
		Body:       `{"error": "Not Found"}`,
		StatusCode: http.StatusNotFound,
	}, nil
}

func handleGetHello(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	msg := fmt.Sprintf("Hello from GET! Path: %s", request.Path)
	return events.APIGatewayProxyResponse{Body: msg, StatusCode: http.StatusOK}, nil
}

func handleGetConfig(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if len(globalConfigs) == 0 {
		return events.APIGatewayProxyResponse{Body: `{"error": "No configurations loaded"}`, StatusCode: http.StatusInternalServerError}, nil
	}

	body, err := json.Marshal(globalConfigs)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: `{"error": "Failed to marshal config"}`, StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func handlePostEcho(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func main() {
	loadConfig()
	fmt.Println("Starting Lambda Handler...")
	lambda.Start(HandleRequest)
}
