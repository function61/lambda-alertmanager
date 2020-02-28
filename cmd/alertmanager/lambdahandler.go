package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func lambdaHandler() {
	// respIsNil because:
	// https://stackoverflow.com/questions/13476349/check-for-nil-and-nil-interface-in-go
	jsonOutHandler := func(resp interface{}, respIsNil bool, err error) ([]byte, error) {
		if respIsNil {
			return nil, err
		}

		asJson, errMarshal := json.Marshal(resp)
		if errMarshal != nil {
			return nil, errMarshal
		}

		return asJson, err
	}

	lambda.StartHandler(multiLambdaEventTypeDispatcher{func(ctx context.Context, polymorphicEvent interface{}) ([]byte, error) {
		switch event := polymorphicEvent.(type) {
		case *events.CloudWatchEvent:
			return nil, handleCloudwatchScheduledEvent(ctx, event.Time)
		case *events.SNSEvent:
			return nil, handleSnsIngest(ctx, *event)
		case *events.APIGatewayProxyRequest:
			resp, err := handleRestCall(ctx, *event)

			return jsonOutHandler(resp, resp == nil, err)
		default:
			return nil, errors.New("cannot identify type of request")
		}
	}})
}
