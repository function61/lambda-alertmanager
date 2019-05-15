package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/function61/lambda-alertmanager/alertmanager/pkg/alertmanagertypes"
	"github.com/function61/lambda-alertmanager/alertmanager/pkg/apigatewayutils"
	"os"
	"sort"
)

func handleRestCall(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	synopsis := req.HTTPMethod + " " + req.Path

	switch synopsis {
	case "GET /alerts":
		return handleGetAlerts(ctx, req)
	case "GET /alerts/acknowledge":
		// this endpoint should really be a POST (since it mutates state), but we've to be
		// pragmatic here because we want acks to be ack-able from emails
		key := req.QueryStringParameters["key"]
		if key == "" {
			return apigatewayutils.BadRequest("key not specified"), nil
		}

		return handleAcknowledgeAlert(ctx, key)
	case "POST /alerts/ingest":
		item := alertmanagertypes.Alert{}

		if err := json.Unmarshal([]byte(req.Body), &item); err != nil {
			return nil, err
		}

		created, err := ingestAlert(item)
		if err != nil {
			return apigatewayutils.InternalServerError(err.Error()), nil
		}

		if created {
			return apigatewayutils.Created(), nil
		} else {
			return apigatewayutils.NoContent(), nil

		}
	case "POST /prometheus-alertmanager/api/v1/alerts":
		return apigatewayutils.InternalServerError("not implemented yet"), nil
	default:
		return apigatewayutils.BadRequest("unknown endpoint"), nil
	}
}

func handleGetAlerts(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	alerts, err := getAlerts()
	if err != nil {
		return apigatewayutils.InternalServerError(err.Error()), nil
	}

	return apigatewayutils.RespondJson(alerts)
}

func handleAcknowledgeAlert(ctx context.Context, alertKey string) (*events.APIGatewayProxyResponse, error) {
	if _, err := dynamodbSvc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName:           alertsDynamoDbTableName,
		ConditionExpression: aws.String("attribute_exists(alert_key)"), // to get error if item-to-delete not found
		Key: dynamoDbRecord{
			"alert_key": mkDynamoString(alertKey),
		},
	}); err != nil {
		if errAws, ok := err.(awserr.Error); ok && errAws.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			// API Gateway doesn't process APIGatewayProxyResponse if Lambda is reported as failure
			return apigatewayutils.NotFound(fmt.Sprintf("alert %s does not exist", alertKey)), nil
		} else {
			return apigatewayutils.InternalServerError(err.Error()), nil
		}
	}

	return apigatewayutils.NoContent(), nil
}

func getAlerts() ([]alertmanagertypes.Alert, error) {
	result, err := dynamodbSvc.Scan(&dynamodb.ScanInput{
		TableName: alertsDynamoDbTableName,
		Limit:     aws.Int64(1000), // whichever comes first, 1 MB or 1 000 records
	})
	if err != nil {
		return nil, err
	}

	alerts := []alertmanagertypes.Alert{}

	for _, alertDb := range result.Items {
		alert, err := deserializeAlertFromDynamoDb(alertDb)
		if err != nil {
			return nil, err
		}

		alerts = append(alerts, *alert)
	}

	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].Key < alerts[j].Key
	})

	return alerts, nil
}

func ackLink(alert alertmanagertypes.Alert) string {
	return os.Getenv("API_ENDPOINT") + "/alerts/acknowledge?key=" + alert.Key
}
