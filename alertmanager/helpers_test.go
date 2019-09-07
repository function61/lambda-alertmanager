package main

import (
	"github.com/function61/gokit/assert"
	"github.com/function61/lambda-alertmanager/alertmanager/pkg/alertmanagertypes"
	"testing"
	"time"
)

func TestMarshalToDynamoAndBack(t *testing.T) {
	asDynamo, err := marshalToDynamoDb(&alertmanagertypes.Alert{
		Key:       "314",
		Timestamp: time.Date(2019, 9, 6, 13, 00, 00, 0, time.UTC),
		Subject:   "example.com offline",
		Details:   "Connection timeout in 2000 ms",
	})
	if err != nil {
		panic(err)
	}

	al := alertmanagertypes.Alert{}
	if err := unmarshalFromDynamoDb(asDynamo, &al); err != nil {
		panic(err)
	}

	assert.EqualString(t, al.Key, "314")
	assert.EqualString(t, al.Timestamp.Format(time.RFC3339Nano), "2019-09-06T13:00:00Z")
	assert.EqualString(t, al.Subject, "example.com offline")
	assert.EqualString(t, al.Details, "Connection timeout in 2000 ms")
}
