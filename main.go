package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type RequestPayload struct {
	AccountID    string `json:"account_id"`
	ConnectionID string `json:"connection_id"`
}

type ResponsePayload struct {
	KeyId     string `json:"key_id"`
	KeySecret string `json:"key_secret"`
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var payload RequestPayload
	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return responseWithError(errors.New("Unauthorized"), 401), nil
	}

	if payload.AccountID == "" || payload.ConnectionID == "" {
		return responseWithError(errors.New("Unauthorized"), 401), nil
	}

	keyId := os.Getenv("KEY_ID")
	KeySecret := os.Getenv("KEY_SECRET")

	if keyId == "" || KeySecret == "" {
		return responseWithError(errors.New("missing required environment variables"), 500), nil
	}

	response := ResponsePayload{
		KeyId:     keyId,
		KeySecret: KeySecret,
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		return responseWithError(err, 500), nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}, nil
}

func responseWithError(err error, statusCode int) events.APIGatewayProxyResponse {
	errorResponse := map[string]string{
		"error": err.Error(),
	}
	body, _ := json.Marshal(errorResponse)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(body),
	}
}

func main() {
	lambda.Start(handler)
}
