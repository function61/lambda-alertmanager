package main

// when row is inserted into DynamoDB (= the alert has been ingested), finally forward
// the alert forward SNS, which can then deliver it by email, SMS etc.

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/function61/gokit/envvar"
	"github.com/function61/lambda-alertmanager/alertmanager/pkg/alertmanagertypes"
	"github.com/function61/lambda-alertmanager/alertmanager/pkg/wtfgo"
	"log"
)

// invoked on any changes to the DynamoDB table
func handleDynamoDbEvent(ctx context.Context, event events.DynamoDBEvent) error {
	if len(event.Records) != 1 { // should not happen, as trigger config: BatchSize=1
		return fmt.Errorf("got %d record(s) from DynamoDB; we only support one", len(event.Records))
	}

	record := event.Records[0]

	if record.EventName != "INSERT" {
		log.Printf("not interested in %s", record.EventName)
		return nil
	}

	alert, err := deserializeAlertFromDynamoDb(dynamoEventImageToDynamoType(record.Change.NewImage))
	if err != nil {
		return err
	}

	return publishAlert(*alert)
}

func publishAlert(alert alertmanagertypes.Alert) error {
	messageText := alert.Subject + "\n\n" + alert.Details

	alertTopic, err := envvar.Get("ALERT_TOPIC")
	if err != nil {
		return err
	}

	messagePerProtocol := struct {
		Default string `json:"default"` // email etc.
		Sms     string `json:"sms"`
	}{
		Default: "Ack: " + ackLink(alert) + "\n\n" + wtfgo.Truncate(messageText, 4*1024),
		Sms:     wtfgo.Substr(messageText, 0, 160-7), // -7 for "ALERT >" prefix in SMS messages
	}

	messagePerProtocolJson, err := json.Marshal(&messagePerProtocol)
	if err != nil {
		return err
	}

	_, err = snsSvc.Publish(&sns.PublishInput{
		TopicArn:         aws.String(alertTopic),
		Subject:          aws.String(alert.Subject),
		Message:          aws.String(string(messagePerProtocolJson)),
		MessageStructure: aws.String("json"),
	})
	return err
}
