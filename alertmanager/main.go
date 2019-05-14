package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sns"
)

var (
	awsSession              = session.New()
	dynamodbSvc             = dynamodb.New(awsSession)
	snsSvc                  = sns.New(awsSession)
	alertsDynamoDbTableName = aws.String("alertmanager_alerts")
)

func main() {
	lambda.StartHandler(multiLambdaEventTypeDispatcher{})
}
