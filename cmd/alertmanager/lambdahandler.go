package main

import (
	"context"
	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/function61/gokit/app/aws/lambdautils"
)

func lambdaHandler() {
	restAPI := newRestAPI(context.Background())

	handler := func(ctx context.Context, polymorphicEvent any) ([]byte, error) {
		switch event := polymorphicEvent.(type) {
		case *events.CloudWatchEvent:
			return nil, handleCloudwatchScheduledEvent(ctx, event.Time)
		case *events.SNSEvent:
			return nil, handleSnsIngest(ctx, *event)
		case *events.APIGatewayProxyRequest:
			return lambdautils.ServeApiGatewayProxyRequestUsingHttpHandler(
				ctx,
				event,
				restAPI)
		default:
			return nil, errors.New("cannot identify type of request")
		}
	}

	lambda.StartHandler(lambdautils.NewMultiEventTypeHandler(handler))
}
