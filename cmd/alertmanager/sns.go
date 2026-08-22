package main

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/gokit/strings/stringutils"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
)

func publishAlert(alert amstate.Alert) error {
	awsSession, err := session.NewSession()
	if err != nil {
		return err
	}

	// FIXME: harcoded region id
	snsSvc := sns.New(awsSession, aws.NewConfig().WithRegion("us-east-1"))

	messageText := alert.Subject + "\n\n" + alert.Details

	alertTopic, err := osutil.GetenvRequired("ALERT_TOPIC")
	if err != nil {
		return err
	}

	ackLinkMaybe := ""
	if alert.ID != "" {
		ackLinkMaybe = "Ack: " + ackLink(alert) + "\n\n"
	}

	messagePerProtocol := struct {
		Default string `json:"default"` // email etc.
		Sms     string `json:"sms"`
	}{
		Default: ackLinkMaybe + stringutils.Truncate(messageText, 4*1024),
		Sms:     stringutils.Truncate(messageText, 160-7), // -7 for "ALERT >" prefix in SMS messages
	}

	messagePerProtocolJSON, err := json.Marshal(&messagePerProtocol)
	if err != nil {
		return err
	}

	_, err = snsSvc.Publish(&sns.PublishInput{
		TopicArn:         new(alertTopic),
		Subject:          new(alert.Subject),
		Message:          new(string(messagePerProtocolJSON)),
		MessageStructure: new("json"),
	})
	return err
}
