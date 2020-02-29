package lambdautils

// This design is not pretty.. https://stackoverflow.com/a/52572943

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type multiEventTypeHandlerFn func(ctx context.Context, polymorphicEvent interface{}) ([]byte, error)

type multiEventTypeHandler struct {
	fn multiEventTypeHandlerFn
}

func NewMultiEventTypeHandler(fn multiEventTypeHandlerFn) lambda.Handler {
	return &multiEventTypeHandler{fn}
}

func (m *multiEventTypeHandler) Invoke(ctx context.Context, reqRaw []byte) ([]byte, error) {
	probe := &eventTypeProbe{}
	polymorphicEvent, err := probe.IdentifyAndUnmarshal(reqRaw)
	if err != nil {
		return nil, err
	}

	return m.fn(ctx, polymorphicEvent)
}

// we introduce just enough fields to determine what type of trigger this is, so we can
// deserialize JSON with proper type
type eventTypeProbe struct {
	HttpMethod string `json:"httpMethod"`  // APIGatewayProxyRequest
	DetailType string `json:"detail-type"` // CloudWatchEvent
	Records    []struct {
		// curiously, SNS and DynamoDB event EventSource and EventVersion fields differ in case
		EventSourceDynamoDb string `json:"eventSource"` // always "aws:dynamodb"
		EventSourceSns      string `json:"EventSource"` // always "aws:sns"
	} `json:"records"`
}

// triggers we need to handle:
// - API Gateway events
// - SNS trigger
// - DynamoDB trigger
// - CloudWatch scheduled event
func (e *eventTypeProbe) Identify() (interface{}, error) {
	if e.HttpMethod != "" {
		return &events.APIGatewayProxyRequest{}, nil
	}

	if e.DetailType == "Scheduled Event" {
		return &events.CloudWatchEvent{}, nil
	}

	if e.Records[0].EventSourceDynamoDb == "aws:dynamodb" {
		return &events.DynamoDBEvent{}, nil
	}

	if e.Records[0].EventSourceSns == "aws:sns" {
		return &events.SNSEvent{}, nil
	}

	return nil, errors.New("cannot identify type of request")
}

func (e *eventTypeProbe) IdentifyAndUnmarshal(reqRaw []byte) (interface{}, error) {
	if err := json.Unmarshal(reqRaw, e); err != nil {
		return nil, err
	}

	typeOfRequest, err := e.Identify()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(reqRaw, typeOfRequest); err != nil {
		return nil, fmt.Errorf("request unmarshal: %v", err)
	}

	return typeOfRequest, nil
}
