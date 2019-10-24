package main

// Ingesting is the act of taking alert from some system, doing deduplication and alert
// limiting (only N amount of active alerts are allowed). We either accept or drop the alert.

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/function61/lambda-alertmanager/alertmanager/pkg/alertmanagertypes"
	"github.com/function61/lambda-alertmanager/alertmanager/pkg/wtfgo"
	"log"
	"os"
	"strconv"
)

// invoked for "AlertManager-ingest" SNS topic
func handleSnsIngest(ctx context.Context, event events.SNSEvent) error {
	for _, msg := range event.Records {
		if _, err := ingestAlert(alertmanagertypes.Alert{
			Subject:   msg.SNS.Subject,
			Details:   msg.SNS.Message,
			Timestamp: msg.SNS.Timestamp,
		}); err != nil {
			return err
		}
	}

	return nil
}

func ingestAlert(candidateAlert alertmanagertypes.Alert) (bool, error) {
	triesRemaining := 3

	for {
		created, err := tryIngestAlertOnce(candidateAlert)
		if err == nil {
			return created, nil
		}

		triesRemaining--

		if triesRemaining == 0 {
			return false, fmt.Errorf("save tries exceeded; last error: %v", err)
		}
	}
}

func tryIngestAlertOnce(candidateAlert alertmanagertypes.Alert) (bool, error) {
	maxFiringAlerts, err := getMaxFiringAlerts()
	if err != nil {
		return false, err
	}

	firingAlerts, err := getFiringAlerts()
	if err != nil {
		return false, err
	}

	if len(firingAlerts) >= maxFiringAlerts {
		log.Println("Max alerts already firing. Discarding the submitted alert.")
		return false, nil // do not save more
	}

	largestNumber := 0

	for _, firingAlert := range firingAlerts {
		if firingAlert.Subject == candidateAlert.Subject {
			log.Println("This alert is already firing. Discarding the submitted alert.")
			return false, nil
		}

		num, err := strconv.Atoi(firingAlert.Key)
		if err != nil {
			return false, err
		}

		largestNumber = wtfgo.MaxInt(largestNumber, num)
	}

	// if you want to test ConditionalCheckFailedException, don't increment this
	candidateAlert.Key = strconv.Itoa(largestNumber + 1)

	dynamoItem, err := marshalToDynamoDb(&candidateAlert)
	if err != nil {
		return false, err
	}

	_, err = dynamodbSvc.PutItem(&dynamodb.PutItemInput{
		// http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.SpecifyingConditions.html
		ConditionExpression: aws.String("attribute_not_exists(alert_key)"),

		TableName: alertsDynamoDbTableName,

		Item: dynamoItem,
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func getMaxFiringAlerts() (int, error) {
	fromEnvStr := os.Getenv("MAX_FIRING_ALERTS")
	if fromEnvStr == "" {
		return 5, nil // default
	}

	return strconv.Atoi(fromEnvStr)
}
