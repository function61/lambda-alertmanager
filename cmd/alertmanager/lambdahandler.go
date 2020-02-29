package main

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/function61/lambda-alertmanager/pkg/lambdautils"
)

func lambdaHandler() {
	restApi := newRestApi(context.Background())

	handler := func(ctx context.Context, polymorphicEvent interface{}) ([]byte, error) {
		switch event := polymorphicEvent.(type) {
		case *events.CloudWatchEvent:
			return nil, handleCloudwatchScheduledEvent(ctx, event.Time)
		case *events.SNSEvent:
			return nil, handleSnsIngest(ctx, *event)
		case *events.APIGatewayProxyRequest:
			return lambdautils.ServeApiGatewayProxyRequestUsingHttpHandler(
				ctx,
				event,
				restApi)
		default:
			return nil, errors.New("cannot identify type of request")
		}
	}

	lambda.StartHandler(lambdautils.NewMultiEventTypeHandler(handler))
}
