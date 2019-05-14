package main

// This design is not pretty.. https://stackoverflow.com/a/52572943

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
)

// we introduce just enough fields to determine what type of trigger this is, so we can
// deserialize JSON with proper type
type eventTypeProbe struct {
	HttpMethod string `json:"httpMethod"`
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
func (e *eventTypeProbe) Identify() (interface{}, error) {
	if e.HttpMethod != "" {
		return &events.APIGatewayProxyRequest{}, nil
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

type multiLambdaEventTypeDispatcher struct{}

func (h multiLambdaEventTypeDispatcher) Invoke(ctx context.Context, reqRaw []byte) ([]byte, error) {
	probe := &eventTypeProbe{}
	polymorphicEvent, err := probe.IdentifyAndUnmarshal(reqRaw)
	if err != nil {
		return nil, err
	}

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

	switch event := polymorphicEvent.(type) {
	case *events.SNSEvent:
		return nil, handleSnsIngest(ctx, *event)
	case *events.DynamoDBEvent:
		return nil, handleDynamoDbEvent(ctx, *event)
	case *events.APIGatewayProxyRequest:
		resp, err := handleRestCall(ctx, *event)

		return jsonOutHandler(resp, resp == nil, err)
	default:
		return nil, errors.New("cannot identify type of request")
	}
}
