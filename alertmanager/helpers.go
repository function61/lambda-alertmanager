package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// this shorthand type does not exist in AWS' library, even though this is rather common
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

func unmarshalFromDynamoDb(record dynamoDbRecord, to interface{}) error {
	return dynamodbattribute.UnmarshalMap(record, to)
}

func marshalToDynamoDb(obj interface{}) (dynamoDbRecord, error) {
	return dynamodbattribute.MarshalMap(obj)
}
