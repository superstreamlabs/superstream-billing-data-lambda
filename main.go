package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
)

type RequestPayload struct {
	Action                      string            `json:"action"`
	GranularityLevel            types.Granularity `json:"granularity_level"`
	From                        string            `json:"from"`
	To                          string            `json:"to"`
	CostStructure               string            `json:"cost_structure"`
	SuperstreamCostExplorerTag  string            `json:"superstream_cost_explorer_tag"`
	SuperstreamCostExplorerTags map[string]string `json:"superstream_cost_explorer_tags"`
	SuperstreamClusterId        int               `json:"superstream_cluster_id"`
	ClusterArn                  string            `json:"cluster_arn"`
}

type ResponsePayload struct {
	Result []byte `json:"result"`
	Err    string `json:"err"`
}

func handler(ctx context.Context, event json.RawMessage) (events.APIGatewayProxyResponse, error) {
	awsCfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return responseWithError(fmt.Errorf("can not create aws client: %v", err), 500), nil
	}

	var payload RequestPayload
	if err := json.Unmarshal(event, &payload); err != nil {
		return responseWithError(fmt.Errorf("can not parse the request: %v", err), 500), nil
	}

	var result any
	var errRes string
	switch payload.Action {
	case "getCostData":
		costExplorerClient := costexplorer.NewFromConfig(awsCfg)
		keyService := "SERVICE"
		keyUsageType := "USAGE_TYPE"
		input := &costexplorer.GetCostAndUsageInput{
			TimePeriod: &types.DateInterval{
				Start: aws.String(payload.From),
				End:   aws.String(payload.To),
			},
			Granularity: payload.GranularityLevel,
			Metrics:     []string{payload.CostStructure},
			GroupBy: []types.GroupDefinition{
				{
					Key:  &keyService,
					Type: types.GroupDefinitionTypeDimension,
				},
				{
					Key:  &keyUsageType,
					Type: types.GroupDefinitionTypeDimension,
				},
			},
			Filter: &types.Expression{
				Tags: &types.TagValues{
					Key:          aws.String(payload.SuperstreamCostExplorerTag),
					Values:       []string{fmt.Sprintf("%v", payload.SuperstreamClusterId)},
					MatchOptions: []types.MatchOption{"EQUALS"},
				},
			},
		}
		results, errHourlyCosts := costExplorerClient.GetCostAndUsage(context.Background(), input)
		if results != nil {
			result = *results
		}
		if errHourlyCosts != nil {
			errRes = errHourlyCosts.Error()
		}

	case "listTags":
		kafkaClient := kafka.NewFromConfig(awsCfg)
		input := &kafka.ListTagsForResourceInput{
			ResourceArn: aws.String(payload.ClusterArn),
		}

		results, err := kafkaClient.ListTagsForResource(context.Background(), input)
		if results != nil {
			result = *results
		}
		if err != nil {
			errRes = err.Error()
		}

	case "tagResource":
		kafkaClient := kafka.NewFromConfig(awsCfg)
		input := &kafka.TagResourceInput{
			ResourceArn: aws.String(payload.ClusterArn),
			Tags:        payload.SuperstreamCostExplorerTags,
		}

		_, err1 := kafkaClient.TagResource(context.Background(), input)
		if err1 != nil {
			errRes = err1.Error()
		}

	case "listCostAllocationTags":
		costExplorerClient := costexplorer.NewFromConfig(awsCfg)
		input := &costexplorer.ListCostAllocationTagsInput{
			TagKeys: []string{payload.SuperstreamCostExplorerTag},
		}
		results, err := costExplorerClient.ListCostAllocationTags(context.Background(), input)
		if results != nil {
			result = *results
		}
		if err != nil {
			errRes = err.Error()
		}

	case "updateCostAllocationTagsStatus":
		costExplorerClient := costexplorer.NewFromConfig(awsCfg)
		input := &costexplorer.UpdateCostAllocationTagsStatusInput{
			CostAllocationTagsStatus: []types.CostAllocationTagStatusEntry{
				{
					TagKey: &payload.SuperstreamCostExplorerTag,
					Status: "Active",
				},
			},
		}

		_, err1 := costExplorerClient.UpdateCostAllocationTagsStatus(context.Background(), input)
		if err1 != nil {
			errRes = err1.Error()
		}

	case "getPricingData":
		keyService := "SERVICE"
		keyUsageType := "USAGE_TYPE"
		amortizedCost := "AmortizedCost"
		usageQuantity := "UsageQuantity"
		costExplorerClient := costexplorer.NewFromConfig(awsCfg)
		input := &costexplorer.GetCostAndUsageInput{
			TimePeriod: &types.DateInterval{
				Start: aws.String(payload.From),
				End:   aws.String(payload.To),
			},
			Granularity: types.GranularityMonthly,
			Metrics:     []string{amortizedCost, usageQuantity},
			GroupBy: []types.GroupDefinition{
				{
					Key:  &keyService,
					Type: types.GroupDefinitionTypeDimension,
				},
				{
					Key:  &keyUsageType,
					Type: types.GroupDefinitionTypeDimension,
				},
			},
			Filter: &types.Expression{
				Tags: &types.TagValues{
					Key:          aws.String(payload.SuperstreamCostExplorerTag),
					Values:       []string{fmt.Sprintf("%v", payload.SuperstreamClusterId)},
					MatchOptions: []types.MatchOption{"EQUALS"},
				},
			},
		}

		results, err1 := costExplorerClient.GetCostAndUsage(context.Background(), input)
		if results != nil {
			result = *results
		}
		if err1 != nil {
			errRes = err1.Error()
		}

	default:
		return responseWithError(fmt.Errorf("unsupported action %v", payload.Action), 500), nil
	}

	var res []byte
	if result != nil {
		res, err = json.Marshal(result)
		if err != nil {
			return responseWithError(fmt.Errorf("can not marshal the result: %v", err), 500), nil
		}
	}

	response := ResponsePayload{
		Result: res,
		Err:    errRes,
	}
	responseBody, err := json.Marshal(response)
	if err != nil {
		return responseWithError(fmt.Errorf("can not create response: %v", err), 500), nil
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
