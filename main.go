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
	AccountID    int `json:"account_id"`
	ConnectionID int `json:"connection_id"`
}

type ResponsePayload struct {
	KeyId     string `json:"key_id"`
	KeySecret string `json:"key_secret"`
}

func handler(ctx context.Context, event json.RawMessage) (events.APIGatewayProxyResponse, error) {
	var payload RequestPayload
	if err := json.Unmarshal(event, &payload); err != nil {
		return responseWithError(errors.New("Unautorized"), 401), nil
	}

	if payload.AccountID <= 0 || payload.ConnectionID <= 0 {
		return responseWithError(errors.New("Unautorized"), 401), nil
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
