package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/function61/lambda-alertmanager/alertmanager/pkg/alertmanagertypes"
	"time"
)

type dynamoDbRecord map[string]*dynamodb.AttributeValue

func dynamoEventImageToDynamoType(fucked map[string]events.DynamoDBAttributeValue) dynamoDbRecord {
	ret := dynamoDbRecord{}

	for key, valWrapper := range fucked {
		switch valWrapper.DataType() {
		case events.DataTypeString:
			ret[key] = mkDynamoString(valWrapper.String())
		default:
			panic("unsupported datatype")
		}
	}

	return ret
}

func mkDynamoString(value string) *dynamodb.AttributeValue {
	return &dynamodb.AttributeValue{
		S: aws.String(value),
	}
}

func deserializeAlertFromDynamoDb(record dynamoDbRecord) (*alertmanagertypes.Alert, error) {
	ts, err := time.Parse(time.RFC3339Nano, *record["timestamp"].S)
	if err != nil {
		return nil, err
	}

	return &alertmanagertypes.Alert{
		Key:       *record["alert_key"].S,
		Subject:   *record["subject"].S,
		Details:   *record["details"].S,
		Timestamp: ts,
	}, nil
}

func serializeAlertToDynamoDb(alert alertmanagertypes.Alert) dynamoDbRecord {
	return dynamoDbRecord{
		"alert_key": mkDynamoString(alert.Key),
		"timestamp": mkDynamoString(alert.Timestamp.Format(time.RFC3339Nano)),
		"subject":   mkDynamoString(alert.Subject),
		"details":   mkDynamoString(alert.Details),
	}
}
