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
func handleSnsIngest(ctx context.Context, req events.SNSEvent) error {
	_, err := ingestAlert(alertmanagertypes.Alert{
		Subject:   req.Records[0].SNS.Subject,
		Details:   req.Records[0].SNS.Message,
		Timestamp: req.Records[0].SNS.Timestamp,
	})

	return err
}

func ingestAlert(item alertmanagertypes.Alert) (bool, error) {
	triesRemaining := 5

	for {
		created, err := tryIngestAlertOnce(item)
		if err == nil {
			return created, nil
		}

		triesRemaining--

		if triesRemaining == 0 {
			return false, fmt.Errorf("save tries exceeded; last error: %v", err)
		}
	}
}

func tryIngestAlertOnce(alert alertmanagertypes.Alert) (bool, error) {
	maxFiringAlerts, err := getMaxFiringAlerts()
	if err != nil {
		return false, err
	}

	result, err := dynamodbSvc.Scan(&dynamodb.ScanInput{
		TableName: alertsDynamoDbTableName,
		Limit:     aws.Int64(1000), // whichever comes first, 1 MB or 1 000 records
	})
	if err != nil {
		return false, err
	}

	if len(result.Items) >= maxFiringAlerts {
		log.Println("Max alerts already firing. Discarding the submitted alert.")
		return false, nil // do not save more
	}

	largestNumber := 0

	for _, rawAlertInDb := range result.Items {
		alertInDb, err := deserializeAlertFromDynamoDb(rawAlertInDb)
		if err != nil {
			return false, err
		}

		if alertInDb.Subject == alert.Subject {
			log.Println("This alert is already firing. Discarding the submitted alert.")
			return false, nil
		}

		num, err := strconv.Atoi(alertInDb.Key)
		if err != nil {
			return false, err
		}

		largestNumber = wtfgo.MaxInt(largestNumber, num)
	}

	// if you want to test ConditionalCheckFailedException, don't increment this
	alert.Key = strconv.Itoa(largestNumber + 1)

	_, err = dynamodbSvc.PutItem(&dynamodb.PutItemInput{
		// http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.SpecifyingConditions.html
		ConditionExpression: aws.String("attribute_not_exists(alert_key)"),

		TableName: alertsDynamoDbTableName,

		Item: serializeAlertToDynamoDb(alert),
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
